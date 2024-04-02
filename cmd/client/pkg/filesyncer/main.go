package filesyncer

import (
	"file-sync/pkg/models"
	log "github.com/sirupsen/logrus"
)

// FileSyncer is an interface for watching files and directories.
type FileSyncer interface {
	SyncFile(filePath string, fileInfo *models.FileInfo) error
	GetSyncedFileMap() map[string]*models.FileInfo
	Close() error
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
func (w *concreteFileSyncer) SyncFile(filePath string, _ *models.FileInfo) error {
	log.Debug("Syncing file:", filePath)
	return nil
}

// GetSyncedFileMap returns the file map from the server.
func (w *concreteFileSyncer) GetSyncedFileMap() map[string]*models.FileInfo {
	return w.syncedFileMap
}

// Close closes the file syncer.
func (w *concreteFileSyncer) Close() error {
	return nil
}
