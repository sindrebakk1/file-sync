package models

import "os"

type SyncedFile struct {
	os.FileInfo
	Hash     string
	Checksum string
}
