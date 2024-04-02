package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCalculateSHA256Checksum(t *testing.T) {
	var testFilePath = "../../test/data/checksum/test_file.txt"
	assert.FileExistsf(t, testFilePath, "test file exists")

	var checksum, err = CalculateSHA256Checksum(testFilePath)
	assert.Equal(t, err, nil, "err should be nil")
	assert.Equal(t, checksum, "4fc34956ab2b0b2955399a773220a79686fde748", "checksum should be the expected value")
}
