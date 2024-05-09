package services

import (
	"bytes"
	"file-sync/models"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"server/pkg/cache"
	"server/pkg/fileparser"
	"sync"
)

type FileService interface {
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

type FileServiceFactory interface {
	NewFileService(dir string) (FileService, error)
}

type concreteFileServiceFactory struct {
	baseDir   string
	fileCache cache.Cache
	metaCache cache.Cache
}

func NewFileServiceFactory(baseDir string, fileCache cache.Cache, metaCache cache.Cache) FileServiceFactory {
	return &concreteFileServiceFactory{
		baseDir,
		fileCache,
		metaCache,
	}
}

func (f *concreteFileServiceFactory) NewFileService(dir string) (FileService, error) {
	return NewFileService(filepath.Join(f.baseDir, dir))
}

type concreteFileService struct {
	baseDir       string
	syncedFileMap map[string]*models.FileInfoBytes
	mutexes       *sync.Map
}

func NewFileService(baseDir string) (FileService, error) {
	fileMap, mutexes, err := initFileMap(baseDir)
	if err != nil {
		return nil, err
	}
	return &concreteFileService{
		baseDir,
		fileMap,
		mutexes,
	}, nil
}

func (s *concreteFileService) GetFileInfo(hash string) (fileInfo *models.FileInfoBytes, found bool) {
	fileInfo, found = s.syncedFileMap[hash]
	return fileInfo, found
}

func (s *concreteFileService) GetFile(hash string) (fileBuffer *bytes.Buffer, err error) {
	syncedFile, found := s.syncedFileMap[hash]
	if !found {
		return nil, fmt.Errorf("file not found")
	}

	// Lock the file
	mutex, _ := s.mutexes.Load(syncedFile.GetHash())
	mutex.(*sync.Mutex).Lock()
	defer mutex.(*sync.Mutex).Unlock()

	// Open the file
	var file *os.File
	file, err = os.Open(filepath.Join(s.baseDir, hash))
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

func (s *concreteFileService) CreateFile(hash string, checksum string, stream []byte) (err error) {
	mutex, _ := s.mutexes.LoadOrStore(hash, &sync.Mutex{})
	mutex.(*sync.Mutex).Lock()
	defer mutex.(*sync.Mutex).Unlock()

	var file *os.File
	file, err = os.Create(filepath.Join(s.baseDir, hash))
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

	var fileInfo os.FileInfo
	fileInfo, err = file.Stat()
	if err != nil {
		return err
	}

	s.syncedFileMap[hash] = models.NewFileInfoBytes(hash, checksum, fileInfo.ModTime())
	return nil
}

func (s *concreteFileService) DeleteFile(hash string) (err error) {
	mutex, _ := s.mutexes.Load(hash)
	mutex.(*sync.Mutex).Lock()
	defer func(mutex *sync.Mutex) {
		mutex.Unlock()
		s.mutexes.Delete(hash)
	}(mutex.(*sync.Mutex))

	err = os.Remove(filepath.Join(s.baseDir, hash))
	if err != nil {
		return err
	}

	delete(s.syncedFileMap, hash)
	return nil
}

func (s *concreteFileService) GetFileMap() map[string]*models.FileInfoBytes {
	return s.syncedFileMap
}

func initFileMap(baseDir string) (fileMap map[string]*models.FileInfoBytes, mutexes *sync.Map, err error) {
	var normalizedBaseDir string
	normalizedBaseDir, err = filepath.Abs(baseDir)
	if err != nil {
		return nil, nil, err
	}
	normalizedBaseDir = filepath.Clean(normalizedBaseDir)
	if err != nil {
		return nil, nil, err
	}

	fileMap = make(map[string]*models.FileInfoBytes)
	mutexes = &sync.Map{}

	err = filepath.Walk(normalizedBaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
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

		fileMap[info.Name()] = models.NewFileInfoBytes(info.Name(), string(checksum), fileInfo.ModTime())
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return fileMap, mutexes, nil
}
