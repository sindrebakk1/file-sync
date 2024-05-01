package services

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	testDir      = "test_dir"
	testHash     = "hash123"
	testChecksum = "checksum123456789012345678901234"
)

// MockCache is a mock implementation of the cache.Cache interface for testing purposes.
type MockCache struct{}

func (c *MockCache) Get(key string) (interface{}, bool) {
	return nil, false
}

func (c *MockCache) Set(key string, value interface{}) {
}

func (c *MockCache) Delete(key string) {
}

func TestNewFileServiceFactory(t *testing.T) {
	fileCache := &MockCache{}
	metaCache := &MockCache{}

	factory := NewFileServiceFactory(testDir, fileCache, metaCache)

	assert.NotNil(t, factory)
}

func TestNewFileService(t *testing.T) {
	fileService, err := newFileService(testDir)
	assert.NoError(t, err)
	assert.NotNil(t, fileService)
}

func TestGetFileInfo(t *testing.T) {
	fileService, err := newFileService(testDir)
	assert.NoError(t, err)

	fileInfo, found := fileService.GetFileInfo(testHash)
	assert.True(t, found)
	assert.Equal(t, testHash, fileInfo.GetHash())
	assert.Equal(t, testChecksum, fileInfo.GetChecksum())
}

func TestGetFileInfo_NotFound(t *testing.T) {
	fileService, err := newFileService(testDir)
	assert.NoError(t, err)

	fileInfo, found := fileService.GetFileInfo("nonexistent_hash")
	assert.False(t, found)
	assert.Nil(t, fileInfo)
}

func TestGetFile_NotFound(t *testing.T) {
	fileService, err := newFileService(testDir)
	assert.NoError(t, err)

	_, err = fileService.GetFile("nonexistent_hash")
	assert.Error(t, err)
	assert.Equal(t, "file not found", err.Error())
}

func TestCreateFile(t *testing.T) {
	// Mocking file content and behavior
	mockHash := "mockHash"
	mockChecksum := "checksum123"
	mockContent := []byte("file_content")

	fileService, err := newFileService(testDir)
	assert.NoError(t, err)

	err = fileService.CreateFile(mockHash, mockChecksum, mockContent)
	assert.NoError(t, err)
	assert.Equal(t, mockChecksum, fileService.GetFileMap()[mockHash].GetChecksum())
}

func TestCreateFile_Error(t *testing.T) {
	fileService, err := newFileService(testDir)
	assert.NoError(t, err)

	err = fileService.CreateFile("", "", nil)
	assert.Error(t, err)
}

func TestDeleteFile(t *testing.T) {
	mockHash := "mockHash"

	fileService, err := newFileService(testDir)
	assert.NoError(t, err)

	err = fileService.DeleteFile(mockHash)
	assert.NoError(t, err)
	assert.Nil(t, fileService.GetFileMap()[mockHash])
}

func TestDeleteFile_Error(t *testing.T) {
	fileService, err := newFileService(testDir)
	assert.NoError(t, err)

	err = fileService.DeleteFile("")
	assert.Error(t, err)
}
