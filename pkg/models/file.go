package models

import (
	"file-sync/pkg/enums"
	"os"
	"time"
)

// FileInfo contains information about a file.
type FileInfo struct {
	os.FileInfo
	DebounceTime time.Time
	LastUpdated  time.Time
	Checksum     string
	Status       enums.FileStatus
}

// DirInfo contains information about a directory.
type DirInfo struct {
	os.FileInfo
	DebounceTime time.Time
	LastUpdated  time.Time
}
