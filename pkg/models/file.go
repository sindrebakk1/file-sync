package models

import (
	"file-sync/pkg/enums"
	"os"
	"time"
)

// FileInfo contains information about a file.
type FileInfo struct {
	os.FileInfo  `json:"os_._file_info,omitempty"`
	DebounceTime time.Time        `json:"debounce_time"`
	LastUpdated  time.Time        `json:"last_updated"`
	Checksum     string           `json:"checksum,omitempty"`
	Status       enums.FileStatus `json:"status,omitempty"`
}

// DirInfo contains information about a directory.
type DirInfo struct {
	os.FileInfo  `json:"os_._file_info,omitempty"`
	DebounceTime time.Time `json:"debounce_time"`
	LastUpdated  time.Time `json:"last_updated"`
}
