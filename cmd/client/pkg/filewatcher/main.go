package filewatcher

import (
	filestatus "file-sync/pkg/enums"
	"file-sync/pkg/models"
	"file-sync/pkg/utils"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// FileWatcher is an interface for watching files and directories.
type FileWatcher interface {
	WatchDirectory(watchDir string) error
	ListenForEvents(eventMap map[fsnotify.Op]EventHandler) error
	Close() error
	GetFileInfo(filePath string) (*models.FileInfo, bool)
	SetFileInfo(filePath string, fileInfo *models.FileInfo)
	DeleteFileInfo(filePath string)
}

// EventHandler is a function that handles file events.
type EventHandler func(string, *models.FileInfo)

// concreteFileWatcher implements the FileWatcher interface.
type concreteFileWatcher struct {
	watcher          *fsnotify.Watcher
	mutexes          sync.Map
	fileMap          map[string]*models.FileInfo
	dirMap           map[string]*models.DirInfo
	debounceDuration time.Duration
}

// NewFileWatcher creates a new instance of FileWatcher.
func NewFileWatcher(fileMap map[string]*models.FileInfo) (FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &concreteFileWatcher{
		watcher,
		sync.Map{},
		fileMap,
		make(map[string]*models.DirInfo),
		200 * time.Millisecond,
	}, nil
}

// WatchDirectory starts watching the directory for changes.
func (w *concreteFileWatcher) WatchDirectory(watchDir string) error {
	log.Debug("Watching directory:", watchDir)
	err := filepath.Walk(watchDir, func(path string, info os.FileInfo, walkError error) error {
		if walkError != nil {
			return walkError
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			log.Error("Error getting absolute path for ", path, err)
			return err
		}
		absPath = filepath.Clean(absPath)
		if !info.IsDir() {
			checksum, checksumErr := utils.CalculateSHA256Checksum(absPath)
			if checksumErr != nil {
				log.Error("Error calculating Checksum while adding ", absPath, " to watcher ", checksumErr)
				return checksumErr
			}

			fileInfo, exists := w.GetFileInfo(absPath)
			if !exists {
				fileInfo = &models.FileInfo{
					FileInfo:     info,
					DebounceTime: time.Now(),
					LastUpdated:  time.Now(),
					Checksum:     checksum,
					Status:       filestatus.New,
				}
				w.SetFileInfo(absPath, fileInfo)
			}
			if fileInfo.Checksum != checksum {
				fileInfo.Checksum = checksum
				fileInfo.DebounceTime = time.Now()
				fileInfo.Status = filestatus.Dirty
				w.SetFileInfo(absPath, fileInfo)
			}
			// sync file

			err = w.watcher.Add(absPath)
			if err != nil {
				log.Error("Error adding ", absPath, " to watcher ", err)
				return err
			}
			w.mutexes.Store(absPath, &sync.Mutex{})
		} else {
			if _, exists := w.dirMap[absPath]; !exists {
				w.dirMap[absPath] = &models.DirInfo{
					FileInfo:     info,
					DebounceTime: time.Now(),
					LastUpdated:  time.Now(),
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// ListenForEvents listens for file changes.
func (w *concreteFileWatcher) ListenForEvents(eventMap map[fsnotify.Op]EventHandler) error {
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				handler, exists := eventMap[event.Op]
				if exists {
					w.debounceEvent(event.Name, handler)
				} else {
					log.Debug("Unhandled event: ", event)
				}
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Error("Error: ", err)
			}
		}
	}()

	var dirs []string
	for dir := range w.dirMap {
		dirs = append(dirs, dir)
	}
	var eventNames []string
	for event := range eventMap {
		eventNames = append(eventNames, event.String())
	}
	log.Debug("Monitoring directories: ", strings.Join(dirs, ", "), " for events: ", strings.Join(eventNames, ", "))
	return nil
}

// Close closes the file watcher.
func (w *concreteFileWatcher) Close() error {
	// Implement the logic for closing the file watcher.
	err := w.watcher.Close()
	if err != nil {
		return err
	}
	err = w.watcher.Close()
	if err != nil {
		return err
	}
	return nil
}

func (w *concreteFileWatcher) GetFileInfo(filePath string) (*models.FileInfo, bool) {
	fileInfo, ok := w.fileMap[filePath]
	if !ok {
		return nil, false
	}
	return fileInfo, true
}

func (w *concreteFileWatcher) SetFileInfo(filePath string, fileInfo *models.FileInfo) {
	w.fileMap[filePath] = fileInfo
}

func (w *concreteFileWatcher) DeleteFileInfo(filePath string) {
	delete(w.fileMap, filePath)
}

func (w *concreteFileWatcher) debounceEvent(filePath string, handler EventHandler) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		log.Error("Error getting absolute path for ", filePath, err)
		return
	}
	absPath = filepath.Clean(absPath)
	filePath = filepath.Clean(absPath)

	// Lock the mutex for the file to prevent concurrent access
	mutex, ok := w.mutexes.Load(filePath)
	if !ok {
		log.Error("Error loading mutex for file:", filePath, err)
		return
	}
	mutex.(*sync.Mutex).Lock()
	defer mutex.(*sync.Mutex).Unlock()

	fileInfo, ok := w.GetFileInfo(filePath)
	if !ok {
		log.Error("Error getting file info for new file: ", filePath)
		return
	}
	if time.Since(fileInfo.DebounceTime) >= w.debounceDuration {
		// Either the file is not in the debounce map or enough time has passed since the last event
		fileInfo.DebounceTime = time.Now()
		w.SetFileInfo(filePath, fileInfo)
		go w.handleEvent(filePath, fileInfo, handler)
		return
	}
}

// handleEvent update the file status and Checksum before calling the handler.
func (w *concreteFileWatcher) handleEvent(filePath string, fileInfo *models.FileInfo, handler EventHandler) {
	checksum, err := utils.CalculateSHA256Checksum(filePath)
	if err != nil {
		log.Error("Error calculating Checksum of file: ", filePath, err)
		return
	}
	if fileInfo.Checksum != checksum {
		log.Debug("File modification detected: ", filePath, " status: ", fileInfo.Status, " Checksum: ", fileInfo.Checksum, " new Checksum: ", checksum)
		fileInfo.Status = filestatus.Dirty
		fileInfo.Checksum = checksum
		w.SetFileInfo(filePath, fileInfo)
		handler(filePath, fileInfo)
	} else {
		log.Debug("No file modification detected: ", filePath, " status: ", fileInfo.Status, " Checksum: ", fileInfo.Checksum, " new Checksum: ", checksum)
	}
}
