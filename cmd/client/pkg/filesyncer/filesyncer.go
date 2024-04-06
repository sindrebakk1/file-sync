package filesyncer

import (
	"file-sync/pkg/globalmodels"
	log "github.com/sirupsen/logrus"
)

// FileSyncer is an interface for watching files and directories.
type FileSyncer interface {
	SyncFile(filePath string, fileInfo *globalmodels.File) error
	GetSyncedFileMap() map[string]*globalmodels.File
	Close() error
}

// EventHandler is a function that handles file events.
type EventHandler func(string, *globalmodels.File)

// concreteFileSyncer implements the FileWatcher interface.
type concreteFileSyncer struct {
	syncedFileMap map[string]*globalmodels.File
}

// New creates a new instance of FileSyncer.
func New() (FileSyncer, error) {
	return &concreteFileSyncer{
		make(map[string]*globalmodels.File),
	}, nil
}

// SyncFile queries the server and syncs the file.
func (w *concreteFileSyncer) SyncFile(filePath string, _ *globalmodels.File) error {
	log.Debug("Syncing file:", filePath)
	return nil
}

// GetSyncedFileMap returns the file map from the server.
func (w *concreteFileSyncer) GetSyncedFileMap() map[string]*globalmodels.File {
	return w.syncedFileMap
}

// Close closes the file syncer.
func (w *concreteFileSyncer) Close() error {
	return nil
}
