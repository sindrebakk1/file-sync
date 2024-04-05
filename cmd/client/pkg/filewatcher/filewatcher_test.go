package filewatcher

import (
	"client/pkg/_mocks"
	"file-sync/pkg/utils"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"
)

var (
	baseDir       = "./test"
	backupsDir    = baseDir + "/backups"
	testDir       = baseDir + "/test_dir"
	watcher       FileWatcher
	testFiles     = make(map[string]*os.File)
	testFilePaths = []string{
		path.Join(testDir, "test_file1.txt"),
		path.Join(testDir, "test_file2.txt"),
		path.Join(testDir, "test_file3.txt"),
		path.Join(testDir, "test_sub_dir", "test_file4.txt"),
		path.Join(testDir, "test_sub_dir", "test_file5.txt"),
	}
)

func TestMain(m *testing.M) {
	code := m.Run()

	// clean up in case of a panic in TestSetup or TestTeardown
	if watcher != nil {
		watcher.Close()
		watcher = nil
	}
	tearDownTestDir()
	os.Exit(code)
}

func TestSetup(t *testing.T) {
	setupTestDir()
	createTestFiles()
	fileInfoMap := makeFileInfoMap()
	syncer := _mocks.NewMockFileSyncer(fileInfoMap)
	fileWatcher, err := NewFileWatcher(syncer)
	if err != nil {
		t.Fatal(err)
	}
	watcher = fileWatcher
}

func TestTeardown(t *testing.T) {
	if watcher != nil {
		watcher.Close()
		watcher = nil
	}
	tearDownTestDir()
}

func setupTestDir() {
	err := os.MkdirAll(testDir, os.ModeDir)
	if err != nil {
		panic(err)
	}
}

func tearDownTestDir() {
	for _, testFile := range testFiles {
		err := testFile.Close()
		if err != nil {
			panic(err)
		}
	}
	err := os.RemoveAll(testDir)
	if err != nil {
		panic(err)
	}
	testFiles = make(map[string]*os.File)
}

func createTestFiles() {
	for _, testFilePath := range testFilePaths {
		filePath, file := createTestFile(testFilePath)
		testFiles[filePath] = file
	}
}

func createTestFile(filePath string) (testFilePath string, file *os.File) {
	var err error
	testFilePath, err = utils.NormalizePath(filePath)
	if err != nil {
		panic(err)
	}

	var testFile *os.File
	testFile, err = os.OpenFile(testFilePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		// Check if the error is due to the directory not existing
		if os.IsNotExist(err) {
			// Create the directory and try opening the file again
			if err = os.MkdirAll(filepath.Dir(testFilePath), 0755); err != nil {
				panic(err)
			}
			testFile, err = os.OpenFile(testFilePath, os.O_CREATE|os.O_RDWR, 0666)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	var backupFilePath string
	fileName := path.Base(filePath)
	backupFilePath, err = utils.NormalizePath(path.Join(backupsDir, fileName+".bak"))
	if err != nil {
		panic(err)
	}

	log.Debug()
	var backupFile *os.File
	backupFile, err = os.OpenFile(backupFilePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer backupFile.Close()

	// Copy the contents of the backup testFile to the new testFile
	_, err = io.Copy(testFile, backupFile)
	if err != nil {
		panic(err)
	}

	return testFilePath, testFile
}

func makeFileInfoMap() (fileInfoMap map[string]*globalmodels.FileInfo) {
	fileInfoMap = make(map[string]*globalmodels.FileInfo)
	for filePath, testFile := range testFiles {
		info, err := testFile.Stat()
		if err != nil {
			panic(err)
		}
		checksum, err := utils.CalculateSHA256Checksum(filePath)
		fileInfo := &globalmodels.FileInfo{
			FileInfo: info,
			Status:   globalenums.Synced,
			Checksum: checksum,
		}
		fileInfoMap[filePath] = fileInfo
	}
	return fileInfoMap
}
