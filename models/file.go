package models

import (
	"encoding/binary"
	"time"
)

const (
	FileHashSize  = 8
	ChecksumSize  = 32
	TimestampSize = 2
	FileInfoSize  = FileHashSize + ChecksumSize + TimestampSize
)

type FileHash [FileHashSize]byte

type FileChecksum [ChecksumSize]byte

type FileTimestamp [TimestampSize]byte

type FileInfoBytes [FileInfoSize]byte

func (f *FileInfoBytes) GetHash() string {
	return string(f[:FileHashSize])
}

func (f *FileInfoBytes) GetChecksum() string {
	return string(f[FileHashSize : FileHashSize+ChecksumSize])
}

func (f *FileInfoBytes) GetTimestamp() time.Time {
	return time.Unix(int64(binary.BigEndian.Uint16(f[FileHashSize+ChecksumSize:FileHashSize+ChecksumSize+TimestampSize])), 0)
}

func NewFileInfoBytes(hash string, checksum string, timestamp time.Time) *FileInfoBytes {
	var fileInfoBytes FileInfoBytes
	copy(fileInfoBytes[:FileHashSize], hash)
	copy(fileInfoBytes[FileHashSize:FileHashSize+ChecksumSize], checksum)
	binary.BigEndian.PutUint16(fileInfoBytes[FileHashSize+ChecksumSize:FileInfoSize], uint16(timestamp.Unix()))
	return &fileInfoBytes
}
