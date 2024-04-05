package services

import (
	"file-sync/pkg/globalenums"
	"file-sync/pkg/globalmodels"
	"file-sync/pkg/utils"
	"os"
	"path/filepath"
	"server/models"
	"server/pkg/fileparser"
	"sync"
)

type FileService interface {
	// GetStatus returns the file with the given ID.
	GetStatus(hash string, fileInfo *globalmodels.FileInfo) (status globalenums.FileStatus, err error)
	// UpdateStatus updates the status of the file with the given ID.
	UpdateStatus(hash string, status globalenums.FileStatus) (err error)
	// GetChecksum returns the checksum of the file with the given ID.
	GetChecksum(hash string) (checksum string, err error)
	// UpdateChecksum updates the checksum of the file with the given ID.
	UpdateChecksum(hash string, checksum string) (err error)
	// GetLastUpdated returns the last updated time of the file with the given ID.
	GetLastUpdated(hash string) (lastUpdated string, err error)
	// UpdateLastUpdated updates the last updated time of the file with the given ID.
	UpdateLastUpdated(hash string, lastUpdated string) (err error)
	// GetUploadSession returns the upload session ID of the file with the given ID.
	GetUploadSession(hash string) (sessionId string, err error)
	// UploadChunk uploads a chunk of the file with the given session ID.
	UploadChunk(hash string, sessionId string, chunk []byte) (err error)
	// CommitChunks commits the chunks of the file with the given session ID and updates the file.
	CommitChunks(hash string, sessionId string) (err error)
	// GetFileStream returns a stream of the file with the given ID.
	GetFileStream(hash string) (stream []byte, err error)
	// GetFileMap returns the file map.
	GetFileMap() map[string]*models.SyncedFile
}

type concreteFileService struct {
	baseDir       string
	syncedFileMap map[string]*models.SyncedFile
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

func (s *concreteFileService) GetStatus(hash string, fileInfo *globalmodels.FileInfo) (status globalenums.FileStatus, err error) {
	return s.syncedFileMap[hash].Status, nil
}

func (s *concreteFileService) UpdateStatus(hash string, status globalenums.FileStatus) (err error) {
	s.syncedFileMap[hash].Status = status
	return nil
}

func (s *concreteFileService) GetChecksum(hash string) (checksum string, err error) {
	return s.syncedFileMap[hash].Checksum, nil
}

func (s *concreteFileService) UpdateChecksum(hash string, checksum string) (err error) {
	s.syncedFileMap[hash].Checksum = checksum
	return nil
}

func (s *concreteFileService) GetLastUpdated(hash string) (lastUpdated string, err error) {
	return s.syncedFileMap[hash].LastUpdated.String(), nil
}

func (s *concreteFileService) UpdateLastUpdated(hash string, lastUpdated string) (err error) {
	return nil
}

func (s *concreteFileService) GetUploadSession(hash string) (sessionId string, err error) {
	return "", nil
}

func (s *concreteFileService) UploadChunk(hash string, sessionId string, chunk []byte) (err error) {
	return nil
}

func (s *concreteFileService) CommitChunks(hash string, sessionId string) (err error) {
	return nil
}

func (s *concreteFileService) GetFileStream(hash string) (stream []byte, err error) {
	return nil, nil
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

		mutexes.Store(info.Name(), &sync.Mutex{})
		fileMap[info.Name()] = &models.SyncedFile{
			Hash:        info.Name(),
			Checksum:    string(checksum),
			Status:      globalenums.Unknown,
			LastUpdated: info.ModTime(),
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return fileMap, mutexes, nil
}
