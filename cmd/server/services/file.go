package services

import (
	"file-sync/pkg/enums"
	"file-sync/pkg/models"
)

type FileService interface {
	// GetStatus returns the file with the given ID.
	GetStatus(hash string, fileInfo *models.FileInfo) (status enums.FileStatus, err error)
	// UpdateStatus updates the status of the file with the given ID.
	UpdateStatus(hash string, status enums.FileStatus) (err error)
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
	GetFileMap() map[string]*models.FileInfo
}

type concreteFileService struct {
	syncedFileMap map[string]*models.FileInfo
}

func NewFileService() (FileService, error) {
	return &concreteFileService{
		make(map[string]*models.FileInfo),
	}, nil
}

func (s *concreteFileService) GetStatus(hash string, fileInfo *models.FileInfo) (status enums.FileStatus, err error) {
	return s.syncedFileMap[hash].Status, nil
}

func (s *concreteFileService) UpdateStatus(hash string, status enums.FileStatus) (err error) {
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

func (s *concreteFileService) GetFileMap() map[string]*models.FileInfo {
	return s.syncedFileMap
}
