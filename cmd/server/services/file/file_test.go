package file_test

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os"
	"server/services/file"
	"testing"
)

var (
	testBaseDir  = "testdata"
	testUserDir  = "user"
	testDir      = testBaseDir + "/" + testUserDir
	testHash     = "hash123"
	testChecksum = "checksum123456789012345678901234"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestNewFileServiceFactory(t *testing.T) {
	fileCache := &mocks.MockCache{}
	metaCache := &mocks.MockCache{}

	factory := file.NewFactory(testBaseDir, fileCache, metaCache)
	assert.NotNil(t, factory)

	fileService, err := factory.New(testUserDir)
	assert.NoError(t, err)
	assert.NotNil(t, fileService)
}

func TestNewFileService(t *testing.T) {
	fileService, err := file.New(testDir)
	assert.NoError(t, err)
	assert.NotNil(t, fileService)
}

func TestGetFileInfo(t *testing.T) {
	fileService, err := file.New(testDir)
	assert.NoError(t, err)

	fileInfo, found := fileService.GetFileInfo(testHash)
	assert.True(t, found)
	assert.Equal(t, testHash, fileInfo.GetHash())
	assert.Equal(t, testChecksum, fileInfo.GetChecksum())
}

func TestGetFileInfo_NotFound(t *testing.T) {
	fileService, err := file.New(testDir)
	assert.NoError(t, err)

	fileInfo, found := fileService.GetFileInfo("nonexistent_hash")
	assert.False(t, found)
	assert.Nil(t, fileInfo)
}

func TestGetFile(t *testing.T) {
	fileService, err := file.New(testDir)
	assert.NoError(t, err)

	var file *bytes.Buffer
	file, err = fileService.GetFile(testHash)
	assert.NoError(t, err)
	assert.NotNil(t, file)

}

func TestGetFile_NotFound(t *testing.T) {
	fileService, err := file.New(testDir)
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

	fileService, err := file.New(testDir)
	assert.NoError(t, err)

	err = fileService.CreateFile(mockHash, mockChecksum, mockContent)
	assert.NoError(t, err)
	assert.Equal(t, mockChecksum, fileService.GetFileMap()[mockHash].GetChecksum())
}

func TestCreateFile_Error(t *testing.T) {
	fileService, err := file.New(testDir)
	assert.NoError(t, err)

	err = fileService.CreateFile("", "", nil)
	assert.Error(t, err)
}

func TestDeleteFile(t *testing.T) {
	mockHash := "mockHash"

	fileService, err := file.New(testDir)
	assert.NoError(t, err)

	err = fileService.DeleteFile(mockHash)
	assert.NoError(t, err)
	assert.Nil(t, fileService.GetFileMap()[mockHash])
}

func TestDeleteFile_Error(t *testing.T) {
	fileService, err := file.New(testDir)
	assert.NoError(t, err)

	err = fileService.DeleteFile("")
	assert.Error(t, err)
}
