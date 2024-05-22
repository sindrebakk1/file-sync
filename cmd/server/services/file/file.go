package file

import (
	"bytes"
	"filesync/models"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"server/pkg/cache"
	"server/pkg/fileparser"
	"sync"
)

type Service interface {
	// GetFileInfo returns the file with the given TransactionID.
	GetFileInfo(hash string) (fileInfo *models.FileInfoBytes, found bool)
	// GetFile returns a file reader for the file with the given TransactionID.
	GetFile(hash string) (file *bytes.Buffer, err error)
	// CreateFile adds a new file to the file service.
	CreateFile(hash string, checksum string, stream []byte) (err error)
	// DeleteFile deletes the file with the given TransactionID.
	DeleteFile(hash string) (err error)
	// GetFileMap returns the file map.
	GetFileMap() map[string]*models.FileInfoBytes
}

type Factory interface {
	New(dir string) (Service, error)
}

type concreteFactory struct {
	baseDir   string
	fileCache cache.Cache
	metaCache cache.Cache
}

func NewFactory(baseDir string, fileCache cache.Cache, metaCache cache.Cache) Factory {
	return &concreteFactory{
		baseDir,
		fileCache,
		metaCache,
	}
}

func (f *concreteFactory) New(userDir string) (Service, error) {
	return New(filepath.Join(f.baseDir, userDir))
}

type concreteService struct {
	dir           string
	syncedFileMap map[string]*models.FileInfoBytes
	mutexes       *sync.Map
}

func New(dir string) (Service, error) {
	fileMap, mutexes, err := initFileMap(dir)
	if err != nil {
		return nil, err
	}
	return &concreteService{
		dir,
		fileMap,
		mutexes,
	}, nil
}

func (s *concreteService) GetFileInfo(hash string) (fileInfo *models.FileInfoBytes, found bool) {
	fileInfo, found = s.syncedFileMap[hash]
	return fileInfo, found
}

func (s *concreteService) GetFile(hash string) (fileBuffer *bytes.Buffer, err error) {
	syncedFile, found := s.syncedFileMap[hash]
	if !found {
		return nil, fmt.Errorf("file not found")
	}

	// Lock the file
	mutex, _ := s.mutexes.LoadOrStore(syncedFile.GetHash(), &sync.Mutex{})
	mutex.(*sync.Mutex).Lock()
	defer mutex.(*sync.Mutex).Unlock()

	// Open the file
	var file *os.File
	file, err = os.Open(filepath.Join(s.dir, hash))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the file into a buffer
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, file)
	if err != nil {
		return nil, err
	}

	// Discard the checksum
	_, err = buffer.ReadString('\n')
	if err != nil {
		return nil, err
	}

	return &buffer, nil
}

func (s *concreteService) CreateFile(hash string, checksum string, stream []byte) (err error) {
	mutex, _ := s.mutexes.LoadOrStore(hash, &sync.Mutex{})
	mutex.(*sync.Mutex).Lock()
	defer mutex.(*sync.Mutex).Unlock()

	var file *os.File
	file, err = os.Create(filepath.Join(s.dir, hash))
	if err != nil {
		return err
	}
	defer file.Close()

	checksumBytes := []byte(fmt.Sprintf("%s\n", checksum))
	_, err = file.Write(checksumBytes)

	_, err = file.Write(stream)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte("\n"))
	if err != nil {
		return err
	}

	var fileInfo os.FileInfo
	fileInfo, err = file.Stat()
	if err != nil {
		return err
	}

	s.syncedFileMap[hash] = models.NewFileInfoBytes(hash, checksum, fileInfo.ModTime())
	return nil
}

func (s *concreteService) DeleteFile(hash string) (err error) {
	mutex, _ := s.mutexes.LoadOrStore(hash, &sync.Mutex{})
	mutex.(*sync.Mutex).Lock()
	defer func(mutex *sync.Mutex) {
		mutex.Unlock()
		s.mutexes.Delete(hash)
	}(mutex.(*sync.Mutex))

	var info os.FileInfo
	info, err = os.Stat(filepath.Join(s.dir, hash))
	if os.IsNotExist(err) {
		return fmt.Errorf("file not found")
	}

	if info.IsDir() {
		return fmt.Errorf("file is a directory")
	}

	err = os.Remove(filepath.Join(s.dir, hash))
	if err != nil {
		return err
	}

	delete(s.syncedFileMap, hash)
	return nil
}

func (s *concreteService) GetFileMap() map[string]*models.FileInfoBytes {
	return s.syncedFileMap
}

func initFileMap(baseDir string) (fileMap map[string]*models.FileInfoBytes, mutexes *sync.Map, err error) {
	var normalizedBaseDir string
	normalizedBaseDir, err = filepath.Abs(baseDir)
	if err != nil {
		return nil, nil, err
	}
	normalizedBaseDir = filepath.Clean(normalizedBaseDir)

	_, err = os.Stat(normalizedBaseDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(normalizedBaseDir, os.ModePerm)
		if err != nil {
			return nil, nil, err
		}
	}

	fileMap = make(map[string]*models.FileInfoBytes)
	mutexes = &sync.Map{}

	err = filepath.Walk(normalizedBaseDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip directories
			return nil
		}

		mutex, _ := mutexes.LoadOrStore(info.Name(), &sync.Mutex{})
		mutex.(*sync.Mutex).Lock()
		defer mutex.(*sync.Mutex).Unlock()

		var file *os.File
		file, err = os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		var checksum string
		checksum, err = fileparser.ExtractChecksumFromReader(file)
		if err != nil {
			return err
		}

		var fileInfo os.FileInfo
		fileInfo, err = file.Stat()
		if err != nil {
			return err
		}

		fileMap[info.Name()] = models.NewFileInfoBytes(info.Name(), checksum, fileInfo.ModTime())
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return fileMap, mutexes, nil
}
