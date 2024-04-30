package fileparser

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var (
	testChecksum = "checksum123456789012345678901234"
	testBlob     = "encrypted_blob_data_here"
	testContent  = fmt.Sprintf("%s\n%s", testChecksum, testBlob)
)

func TestParse(t *testing.T) {
	checksum, blob, err := Parse([]byte(testContent))

	assert.NoError(t, err)
	assert.Equal(t, testChecksum, checksum)
	assert.Equal(t, []byte(testBlob), blob)
}

func TestParse_InvalidFormat(t *testing.T) {
	content := []byte("invalidcontent")

	_, _, err := Parse(content)

	assert.Error(t, err)
	assert.Equal(t, "invalid file format: missing newline separator", err.Error())
}

func TestParseFromReader(t *testing.T) {
	reader := strings.NewReader(testContent)

	checksum, blob, err := ParseFromReader(reader)

	assert.NoError(t, err)
	assert.Equal(t, testChecksum, checksum)
	assert.Equal(t, []byte(testBlob), blob)
}

func TestExtractChecksum(t *testing.T) {
	checksum, err := ExtractChecksum([]byte(testContent))

	assert.NoError(t, err)
	assert.Equal(t, testChecksum, checksum)
}

func TestExtractBlob(t *testing.T) {
	blob, err := ExtractBlob([]byte(testContent))

	assert.NoError(t, err)
	assert.Equal(t, []byte(testBlob), blob)
}

func TestExtractChecksumFromReader(t *testing.T) {
	reader := strings.NewReader(testContent)

	checksum, err := ExtractChecksumFromReader(reader)

	assert.NoError(t, err)
	assert.Equal(t, testChecksum, checksum)
}

func TestExtractBlobFromReader(t *testing.T) {
	reader := strings.NewReader(testContent)

	blob, err := ExtractBlobFromReader(reader)

	assert.NoError(t, err)
	assert.Equal(t, []byte(testBlob), blob)
}
