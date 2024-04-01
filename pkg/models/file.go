package models

import (
	filestatus "file-sync/pkg/enums"
	"os"
	"time"
)

// FileInfo contains information about a file.
type FileInfo struct {
	os.FileInfo
	DebounceTime time.Time
	LastUpdated  time.Time
	Checksum     string
	Status       filestatus.FileStatus
}

// DirInfo contains information about a directory.
type DirInfo struct {
	os.FileInfo
	DebounceTime time.Time
	LastUpdated  time.Time
}
