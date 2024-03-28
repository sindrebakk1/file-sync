package main

import (
	"file-sync/pkg/utils"
	"flag"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"
)

type SyncedFileInfo struct {
	os.FileInfo
	checksum     string
	debounceTime time.Time
	//status       filestatus.FileStatus
}

var (
	watchDir         *string
	debounceDuration = 200 * time.Millisecond
	fileMutexes      sync.Map
	fileMap          = make(map[string]SyncedFileInfo)
)

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func(watcher *fsnotify.Watcher) {
		err := watcher.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(watcher)

	err = filepath.Walk(*watchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				log.Error("Error adding", path, "to watcher", err)
			}
			checksum, err := utils.CalculateSHA256Checksum(path)
			if err != nil {
				log.Error("Error calculating checksum while adding", path, "to watcher", err)
			}

			fileMutexes.Store(path, &sync.Mutex{})

			fileInfo, exists := fileMap[path]
			if !exists {
				fileMap[path] = SyncedFileInfo{
					info,
					checksum,
					time.Now(),
					//filestatus.Stale,
				}
				// sync file
			} else if fileInfo.checksum != checksum {
				fileInfo.checksum = checksum
				fileInfo.debounceTime = time.Now()
				// sync file
			} else {
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					debounceEvent(event.Name, handleCreate)
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					debounceEvent(event.Name, handleWrite)
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					debounceEvent(event.Name, handleRemove)
				}
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					debounceEvent(event.Name, handleRemove)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error("Error", err)
			}
		}
	}()

	log.Debug("Monitoring directory:", *watchDir)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	log.Info("Shutting down...")
}

func debounceEvent(filePath string, handler func(string)) {
	mutex, exists := fileMutexes.Load(filePath)
	if !exists {
		log.Debug("Error loading mutex for file:", filePath, ". Creating new mutex and retrying")
		fileMutexes.Store(filePath, &sync.Mutex{})
		debounceEvent(filePath, handler)
	}
	mutex.(*sync.Mutex).Lock()
	defer mutex.(*sync.Mutex).Unlock()

	fileInfo, exists := fileMap[filePath]
	if !exists || time.Since(fileInfo.debounceTime) >= debounceDuration {
		// Either the file is not in the debounce map or enough time has passed since the last event
		fileInfo.debounceTime = time.Now()
		fileMap[filePath] = fileInfo
		go handler(filePath)
	}
}

func handleWrite(filePath string) {
	// This function is executed when the debounced event is triggered
	log.Debug("Wrote to file:", filePath)
	checksum, err := utils.CalculateSHA256Checksum(filePath)
	if err != nil {
		log.Error("Error calculating checksum of file:", filePath, err)
		return
	}
	fileInfo, exists := fileMap[filePath]
	if !exists {
		log.Warning("File not found in map:", filePath)
		newFileInfo, err := os.Stat(filePath)
		if err != nil {
			log.Error("Error getting file info for new file:", filePath)
			return
		}
		fileMap[filePath] = SyncedFileInfo{
			newFileInfo,
			checksum,
			time.Now(),
			//filestatus.Stale,
		}
	}
	if exists && checksum == fileInfo.checksum {
		log.Debug("No change")
		return
	}

	log.Debug("File modification detected")
}

func handleCreate(filePath string) {
	log.Debug("Created file:", filePath)
	checksum, err := utils.CalculateSHA256Checksum(filePath)
	if err != nil {
		log.Error("Error calculating checksum of file:", filePath, err)
		return
	}
	_, exists := fileMap[filePath]
	if !exists {
		log.Debug("New file:", filePath)
		newFileInfo, err := os.Stat(filePath)
		if err != nil {
			log.Error("Error getting file info for new file:", filePath)
			return
		}
		fileMap[filePath] = SyncedFileInfo{
			newFileInfo,
			checksum,
			time.Now(),
			//filestatus.Stale,
		}
		return
	}
}

func handleRemove(filepath string) {
	log.Debug("Removed file:", filepath)
	delete(fileMap, filepath)
}

func init() {
	watchDir = flag.String("dir", ".", "directory to monitor")
	flag.Parse()
	log.SetLevel(log.DebugLevel)
}
