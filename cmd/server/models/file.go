package models

import (
	"file-sync/pkg/globalenums"
	"time"
)

type SyncedFile struct {
	Hash        string
	Checksum    string
	Status      globalenums.FileStatus
	LastUpdated time.Time
}
