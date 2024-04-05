package main

import (
	"client/pkg/filesyncer"
	"client/pkg/filewatcher"
	"file-sync/pkg/globalmodels"
	"file-sync/pkg/utils"
	"flag"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"path"
	"time"
)

var (
	watchDir         string
	debounceDuration = 200 * time.Millisecond
)

func main() {
	syncer, err := filesyncer.New()
	if err != nil {
		log.Fatal(err)
	}
	defer syncer.Close()

	watcher, err := filewatcher.NewFileWatcher(syncer)
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.WatchDirectory(watchDir)
	err = watcher.ListenForEvents(map[fsnotify.Op]filewatcher.EventHandler{
		fsnotify.Write:  handleWrite,
		fsnotify.Create: handleCreate,
		fsnotify.Remove: handleRemove,
		fsnotify.Rename: handleRename,
	})

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	log.Info("Shutting down...")
}

func handleWrite(filePath string, fileInfo *globalmodels.FileInfo, _ filesyncer.FileSyncer) {
	log.Debug("WRITE:", filePath, "status:", fileInfo.Status)
}

func handleCreate(filePath string, fileInfo *globalmodels.FileInfo, _ filesyncer.FileSyncer) {
	log.Debug("CREATE:", filePath, "status:", fileInfo.Status)
}

func handleRemove(filePath string, fileInfo *globalmodels.FileInfo, _ filesyncer.FileSyncer) {
	log.Debug("REMOVE:", filePath, "status:", fileInfo.Status)
}

func handleRename(filePath string, fileInfo *globalmodels.FileInfo, _ filesyncer.FileSyncer) {
	log.Debug("RENAME:", filePath, "status:", fileInfo.Status)
}

func init() {
	// Parse command line flags
	flag.StringVar(&watchDir, "dir", "~/", "directory to monitor")
	flag.Parse()
	normalizedPath, err := utils.NormalizePath(watchDir)
	if err != nil {
		log.Fatal(err)
	}
	watchDir = path.Clean(normalizedPath)

	// Configure logging
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
}
