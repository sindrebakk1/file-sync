package services

import (
	"bytes"
	"file-sync/pkg/utils"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"server/models"
	"server/pkg/fileparser"
	"sync"
)

type FileService interface {
	// GetFileInfo returns the file with the given ID.
	GetFileInfo(hash string) (fileInfo *models.SyncedFile, found bool)
	// GetFile returns a file reader for the file with the given ID.
	GetFile(hash string) (file *bytes.Buffer, err error)
	// CreateFile adds a new file to the file service.
	CreateFile(hash string, checksum string, stream []byte) (err error)
	// DeleteFile deletes the file with the given ID.
	DeleteFile(hash string) (err error)
	// GetFileMap returns the file map.
	GetFileMap() map[string]*models.SyncedFile
}

type FileServiceFactory interface {
	NewFileService(dir string) (FileService, error)
}

type concreteFileServiceFactory struct {
	baseDir string
}

func NewFileServiceFactory(baseDir string) FileServiceFactory {
	return &concreteFileServiceFactory{
		baseDir,
	}
}

func (f *concreteFileServiceFactory) NewFileService(dir string) (FileService, error) {
	return newFileService(filepath.Join(f.baseDir, dir))
}

type concreteFileService struct {
	baseDir       string
	syncedFileMap map[string]*models.SyncedFile
	mutexes       *sync.Map
}

func newFileService(baseDir string) (FileService, error) {
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

func (s *concreteFileService) GetFileInfo(hash string) (fileInfo *models.SyncedFile, found bool) {
	fileInfo, found = s.syncedFileMap[hash]
	return fileInfo, found
}

func (s *concreteFileService) GetFile(hash string) (fileBuffer *bytes.Buffer, err error) {
	syncedFile, found := s.syncedFileMap[hash]
	if !found {
		return nil, fmt.Errorf("file not found")
	}

	// Lock the file
	mutex, _ := s.mutexes.Load(syncedFile.Hash)
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

	s.syncedFileMap[hash] = &models.SyncedFile{
		Hash:     hash,
		Checksum: checksum,
		FileInfo: fileInfo,
	}
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

func (s *concreteFileService) GetFileMap() map[string]*models.SyncedFile {
	return s.syncedFileMap
}

func initFileMap(baseDir string) (fileMap map[string]*models.SyncedFile, mutexes *sync.Map, err error) {
	var normalizedBaseDir string
	normalizedBaseDir, err = utils.NormalizePath(baseDir)
	if err != nil {
		return nil, nil, err
	}
	fileMap = make(map[string]*models.SyncedFile)
	mutexes = &sync.Map{}
	err = filepath.Walk(normalizedBaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		mutex, _ := mutexes.LoadOrStore(info.Name(), &sync.Mutex{})
		mutex.(*sync.Mutex).Lock()
		defer mutex.(*sync.Mutex).Unlock()

		var (
			file     *os.File
			checksum []byte
		)
		file, err = os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		checksum, err = fileparser.ExtractChecksumFromReader(file)
		if err != nil {
			return err
		}

		var fileInfo os.FileInfo
		fileInfo, err = file.Stat()
		if err != nil {
			return err
		}

		fileMap[info.Name()] = &models.SyncedFile{
			Hash:     info.Name(),
			Checksum: string(checksum),
			FileInfo: fileInfo,
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return fileMap, mutexes, nil
}
