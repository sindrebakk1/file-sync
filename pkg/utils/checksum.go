package utils

import (
	"crypto/sha1"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

func CalculateSHA256Checksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Debug("Error from os.Open", path)
		return "", err
	}
	defer file.Close()

	hasher := sha1.New()

	_, err = io.Copy(hasher, file)
	if err != nil {
		log.Debug("Error from io.Copy", file.Name())
		return "", err
	}
	checksum := hex.EncodeToString(hasher.Sum(nil))

	return checksum, nil
}
