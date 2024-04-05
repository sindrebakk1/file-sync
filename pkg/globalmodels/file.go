package globalmodels

import (
	"file-sync/pkg/globalenums"
	"os"
	"time"
)

// FileInfo contains information about a file.
type FileInfo struct {
	os.FileInfo  `json:"os_._file_info"`
	DebounceTime time.Time              `json:"debounce_time"`
	LastUpdated  time.Time              `json:"last_updated"`
	Checksum     string                 `json:"checksum"`
	Status       globalenums.FileStatus `json:"status"`
}

// DirInfo contains information about a directory.
type DirInfo struct {
	os.FileInfo  `json:"os_._file_info"`
	DebounceTime time.Time `json:"debounce_time"`
	LastUpdated  time.Time `json:"last_updated"`
}
