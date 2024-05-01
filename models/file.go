package models

import (
	"file-sync/enums"
	"os"
	"time"
)

// File contains information about a file.
type File struct {
	os.FileInfo
	DebounceTime time.Time
	Checksum     string
	Status       enums.FileStatus
}

// DirInfo contains information about a directory.
type DirInfo struct {
	os.FileInfo
	DebounceTime time.Time
}
