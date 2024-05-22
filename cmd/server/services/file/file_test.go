package file_test

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os"
	"server/pkg/_mocks"
	"server/services/file"
	"testing"
)

var (
	testBaseDir  = "testdata"
	testUserDir  = "user"
	testDir      = testBaseDir + "/" + testUserDir
	testHash     = "hash123"
	testChecksum = "checksum123456789012345678901234"
	testContent  = []byte("file_content")
)

func TestMain(m *testing.M) {
	code := m.Run()

	// Clean up
	err := os.RemoveAll(testDir)
	if err != nil {
		panic(err)
	}

	os.Exit(code)
}

func TestNewFileService(t *testing.T) {
	fileService, err := file.New(testDir)
	assert.NoError(t, err)
	assert.NotNil(t, fileService)
}

func TestNewFileServiceFactory(t *testing.T) {
	fileCache := &_mocks.MockCache{}
	metaCache := &_mocks.MockCache{}

	factory := file.NewFactory(testBaseDir, fileCache, metaCache)
	assert.NotNil(t, factory)

	fileService, err := factory.New(testUserDir)
	assert.NoError(t, err)
	assert.NotNil(t, fileService)
}

func TestCreateFile(t *testing.T) {
	fileService, err := file.New(testDir)
	assert.NoError(t, err)

	err = fileService.CreateFile(testHash, testChecksum, testContent)
	assert.NoError(t, err)
	assert.Equal(t, testChecksum, fileService.GetFileMap()[testHash].GetChecksum())
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

	var fileBuf *bytes.Buffer
	fileBuf, err = fileService.GetFile(testHash)
	assert.NoError(t, err)
	assert.NotNil(t, fileBuf)
}

func TestGetFile_NotFound(t *testing.T) {
	fileService, err := file.New(testDir)
	assert.NoError(t, err)

	_, err = fileService.GetFile("nonexistent_hash")
	assert.Error(t, err)
	assert.Equal(t, "file not found", err.Error())
}

func TestCreateFile_Error(t *testing.T) {
	fileService, err := file.New(testDir)
	assert.NoError(t, err)

	err = fileService.CreateFile("", "", nil)
	assert.Error(t, err)
}

func TestDeleteFile(t *testing.T) {
	fileService, err := file.New(testDir)
	assert.NoError(t, err)

	err = fileService.DeleteFile(testHash)
	assert.NoError(t, err)
	assert.Nil(t, fileService.GetFileMap()[testHash])
}

func TestDeleteFile_Error(t *testing.T) {
	fileService, err := file.New(testDir)
	assert.NoError(t, err)

	err = fileService.DeleteFile("")
	assert.Error(t, err)
}
