package models

import (
	"file-sync/pkg/globalenums"
	"os"
)

type SyncedFile struct {
	os.FileInfo
	Hash     string
	Checksum string
	Status   globalenums.FileStatus
}
