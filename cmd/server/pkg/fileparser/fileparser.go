package fileparser

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
)

// Parse extracts the checksum and the raw encrypted file blob from the file content.
func Parse(content []byte) (checksum []byte, blob []byte, err error) {
	// Split content by newline to separate the header and the blob
	parts := bytes.SplitN(content, []byte("\n"), 2)
	if len(parts) != 2 {
		return nil, nil, errors.New("invalid file format: missing newline separator")
	}

	// Extract checksum from the header (first 32 bytes)
	checksumBytes := parts[0][:32]
	checksumHex := string(checksumBytes)
	checksumBytes, err = hex.DecodeString(checksumHex)
	if err != nil {
		return nil, nil, errors.New("invalid checksum format")
	}

	// Extract blob from the content after the newline separator
	blob = parts[1]

	return checksum, blob, nil
}

// ParseFromReader reads the content from an io.Reader and parses it.
func ParseFromReader(reader io.Reader) (checksum []byte, blob []byte, err error) {
	var content bytes.Buffer
	_, err = io.Copy(&content, reader)
	if err != nil {
		return nil, nil, err
	}
	return Parse(content.Bytes())
}

// ExtractChecksum extracts the checksum from the file content.
func ExtractChecksum(content []byte) ([]byte, error) {
	checksum, _, err := Parse(content)
	return checksum, err
}

// ExtractBlob extracts the raw encrypted file blob from the file content.
func ExtractBlob(content []byte) ([]byte, error) {
	_, blob, err := Parse(content)
	return blob, err
}

// ExtractChecksumFromReader reads the content from an io.Reader and extracts the checksum.
func ExtractChecksumFromReader(reader io.Reader) ([]byte, error) {
	checksum, _, err := ParseFromReader(reader)
	return checksum, err
}

// ExtractBlobFromReader reads the content from an io.Reader and extracts the blob.
func ExtractBlobFromReader(reader io.Reader) ([]byte, error) {
	_, blob, err := ParseFromReader(reader)
	return blob, err
}
