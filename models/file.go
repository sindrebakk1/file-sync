package models

import (
	"file-sync/pkg/globalenums"
	"os"
	"time"
)

// File contains information about a file.
type File struct {
	os.FileInfo
	DebounceTime time.Time
	Checksum     string
	Status       globalenums.FileStatus
}

// DirInfo contains information about a directory.
type DirInfo struct {
	os.FileInfo
	DebounceTime time.Time
}
