package filewatcher

import (
	"file-sync/pkg/enums"
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

// EventHandler is a function that handles testFile1 events.
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
					Status:       enums.New,
				}
				w.SetFileInfo(absPath, fileInfo)
			}
			if fileInfo.Checksum != checksum {
				fileInfo.Checksum = checksum
				fileInfo.DebounceTime = time.Now()
				fileInfo.Status = enums.Dirty
				w.SetFileInfo(absPath, fileInfo)
			}
			// sync testFile1

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

// ListenForEvents listens for testFile1 changes.
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

// Close closes the testFile1 watcher.
func (w *concreteFileWatcher) Close() error {
	err := w.watcher.Close()
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

// debounceEvent debounces events for a given testFile1 path.
// If enough time has passed since the last event, it triggers the provided handler.
//   - filePath: The path of the testFile1 to debounce events for.
//   - handler: The event handler function to call.
func (w *concreteFileWatcher) debounceEvent(filePath string, handler EventHandler) {
	normalizedPath, err := utils.NormalizePath(filePath)
	if err != nil {
		log.Error("Error getting absolute filePath for ", normalizedPath, err)
		return
	}

	// Lock the mutex for the testFile1 to prevent concurrent access
	mutex, ok := w.mutexes.Load(normalizedPath)
	if !ok {
		log.Error("Error loading mutex for testFile1:", normalizedPath, err)
		return
	}
	mutex.(*sync.Mutex).Lock()
	defer mutex.(*sync.Mutex).Unlock()

	fileInfo, ok := w.GetFileInfo(normalizedPath)
	if !ok {
		log.Error("Error getting testFile1 info for new testFile1: ", normalizedPath)
		return
	}
	if time.Since(fileInfo.DebounceTime) >= w.debounceDuration {
		// Either the testFile1 is not in the debounce map or enough time has passed since the last event
		fileInfo.DebounceTime = time.Now()
		w.SetFileInfo(normalizedPath, fileInfo)
		go w.handleEvent(normalizedPath, fileInfo, handler)
		return
	}
}

// handleEvent update the testFile1 status and Checksum before calling the handler.
//   - filePath: The path of the testFile1 to handle.
//   - fileInfo: The testFile1 info to update.
//   - handler: The event handler function to call.
func (w *concreteFileWatcher) handleEvent(filePath string, fileInfo *models.FileInfo, handler EventHandler) {
	checksum, err := utils.CalculateSHA256Checksum(filePath)
	if err != nil {
		log.Error("Error calculating Checksum of testFile1: ", filePath, err)
		return
	}
	if fileInfo.Checksum != checksum {
		log.Debug("File modification detected: ", filePath, " status: ", fileInfo.Status, " Checksum: ", fileInfo.Checksum, " new Checksum: ", checksum)
		fileInfo.Status = enums.Dirty
		fileInfo.Checksum = checksum
		w.SetFileInfo(filePath, fileInfo)
		handler(filePath, fileInfo)
	} else {
		log.Debug("No testFile1 modification detected: ", filePath, " status: ", fileInfo.Status, " Checksum: ", fileInfo.Checksum, " new Checksum: ", checksum)
	}
}
