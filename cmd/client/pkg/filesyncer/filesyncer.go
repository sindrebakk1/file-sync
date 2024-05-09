package filesyncer

import (
	"file-sync/models"
	log "github.com/sirupsen/logrus"
)

// FileSyncer is an interface for watching files and directories.
type FileSyncer interface {
	SyncFile(filePath string, fileInfo *models.File) error
	GetSyncedFileMap() map[string]*models.File
	Close() error
}

// EventHandler is a function that handles file events.
type EventHandler func(string, *models.File)

// concreteFileSyncer implements the FileWatcher interface.
type concreteFileSyncer struct {
	syncedFileMap map[string]*models.File
}

// New creates a new instance of FileSyncer.
func New() (FileSyncer, error) {
	return &concreteFileSyncer{
		make(map[string]*models.File),
	}, nil
}

// SyncFile queries the server and syncs the file.
func (w *concreteFileSyncer) SyncFile(filePath string, _ *models.File) error {
	log.Debug("Syncing file:", filePath)
	return nil
}

// GetSyncedFileMap returns the file map from the server.
func (w *concreteFileSyncer) GetSyncedFileMap() map[string]*models.File {
	return w.syncedFileMap
}

// Close closes the file syncer.
func (w *concreteFileSyncer) Close() error {
	return nil
}
