package filesyncer

import (
	"file-sync/pkg/models"
	log "github.com/sirupsen/logrus"
)

// FileSyncer is an interface for watching files and directories.
type FileSyncer interface {
	SyncFile(filePath string, fileInfo *models.FileInfo) error
	SyncedFileMap() map[string]*models.FileInfo
}

// EventHandler is a function that handles file events.
type EventHandler func(string, *models.FileInfo)

// concreteFileSyncer implements the FileWatcher interface.
type concreteFileSyncer struct {
	syncedFileMap map[string]*models.FileInfo
}

// New creates a new instance of FileSyncer.
func New() (FileSyncer, error) {
	return &concreteFileSyncer{
		make(map[string]*models.FileInfo),
	}, nil
}

// SyncFile queries the server and syncs the file.
func (w *concreteFileSyncer) SyncFile(filePath string, fileInfo *models.FileInfo) error {
	log.Debug("Syncing file:", filePath)
	return nil
}

// SyncedFileMap returns the file map from the server.
func (w *concreteFileSyncer) SyncedFileMap() map[string]*models.FileInfo {
	return w.syncedFileMap
}
