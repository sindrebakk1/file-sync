package main

import (
	"client/pkg/filewatcher"
	"file-sync/pkg/models"
	"flag"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"time"
)

var (
	watchDir         string
	debounceDuration = 200 * time.Millisecond
)

func main() {
	fileMap := make(map[string]*models.FileInfo)
	watcher, err := filewatcher.NewFileWatcher(fileMap)
	if err != nil {
		log.Fatal(err)
	}
	defer func(watcher filewatcher.FileWatcher) {
		err = watcher.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(watcher)

	err = watcher.WatchDirectory(watchDir)
	err = watcher.ListenForEvents(map[fsnotify.Op]filewatcher.EventHandler{
		fsnotify.Write:  handleWrite,
		fsnotify.Create: handleCreate,
		fsnotify.Remove: handleRemove,
	})

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	log.Info("Shutting down...")
}

func handleWrite(filePath string, fileInfo *models.FileInfo) {
	log.Debug("Wrote to file:", filePath, "status:", fileInfo.Status)
}

func handleCreate(filePath string, fileInfo *models.FileInfo) {
	log.Debug("Wrote to file:", filePath, "status:", fileInfo.Status)
}

func handleRemove(filePath string, fileInfo *models.FileInfo) {
	log.Debug("Wrote to file:", filePath, "status:", fileInfo.Status)
}

func init() {
	// Parse command line flags
	flag.StringVar(&watchDir, "dir", ".", "directory to monitor")
	flag.Parse()
	absPath, err := filepath.Abs(watchDir)
	if err != nil {
		log.Fatal(err)
	}
	watchDir = path.Clean(absPath)

	// Configure logging
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
	log.SetFormatter(&log.JSONFormatter{})
}
