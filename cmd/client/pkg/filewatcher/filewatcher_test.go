package filewatcher

import (
	"file-sync/pkg/models"
	"file-sync/pkg/utils"
	"fmt"
	"io"
	"os"
	"path"
	"testing"
)

var (
	baseDir       = "../test/data/filewatcher"
	testDir       = baseDir + "/test_dir"
	watcher       FileWatcher
	fileMap       = make(map[string]*models.FileInfo)
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
	// setup
	setupTestDir()

	exitCode := m.Run()

	// teardown
	if watcher != nil {
		err := watcher.Close()
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	}
	teardownTestDir()

	os.Exit(exitCode)
}

func TestSetup(t *testing.T) {
	if watcher != nil {
		fileWatcher := watcher
		err := fileWatcher.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	fileWatcher, err := NewFileWatcher(fileMap)
	if err != nil {
		t.Fatal(err)
	}
	watcher = fileWatcher
}

func TestTeardown(t *testing.T) {
	if watcher != nil {
		fileWatcher := watcher
		err := fileWatcher.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func setupTestDir() {
	err := os.MkdirAll(testDir, os.ModeDir)
	if err != nil {
		panic(err)
	}
}

func teardownTestDir() {
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
	for _, fileName := range testFilePaths {
		filePath, file := createTestFile(fileName)
		testFiles[filePath] = file
	}
}

func createTestFile(fileName string) (filePath string, file *os.File) {
	testFilePath := path.Join(testDir, fileName)
	var err error
	testFilePath, err = utils.NormalizePath(testFilePath)
	if err != nil {
		panic(err)
	}

	var testFile *os.File
	testFile, err = os.Create(testFilePath)
	if err != nil {
		panic(err)
	}

	var backupFilePath string
	var backupFile *os.File
	backupFilePath, err = utils.NormalizePath(path.Join(baseDir, fileName+".bak"))
	if err != nil {
		panic(err)
	}
	backupFile, err = os.OpenFile(backupFilePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	// Copy the contents of the backup testFile to the new testFile
	_, err = io.Copy(testFile, backupFile)
	if err != nil {
		panic(err)
	}

	err = backupFile.Close()
	if err != nil {
		panic(err)
	}

	return testFilePath, testFile
}
