package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func CalculateSHA256Checksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}
	defer file.Close()

	hasher := sha1.New()

	_, err = io.Copy(hasher, file)
	if err != nil {
		fmt.Println("Error", err)
		return "", err
	}
	checksum := hex.EncodeToString(hasher.Sum(nil))

	return checksum, nil
}
