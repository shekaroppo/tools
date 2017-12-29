package lib

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setAtime(files []string) {
	for idx, file := range files {
		atime := time.Date(2006, time.February, 1, 3, 4, idx, 0, time.UTC)
		os.Chtimes(file, atime, atime)
	}
}

func createTestFiles() (string, []string) {
	dir, err := ioutil.TempDir("", "note")
	if err != nil {
		log.Fatal(err)
	}
	var files []string
	for i := 0; i < 10; i++ {
		file, _ := ioutil.TempFile(dir, "file")
		files = append(files, file.Name())
	}
	setAtime(files)
	return dir, files
}

func TestBaseNames(t *testing.T) {
	dir, files := createTestFiles()
	defer os.RemoveAll(dir)
	var expectedNames []string
	for _, file := range files {
		expectedNames = append(expectedNames, path.Base(file))
	}
	assert.Equal(t, expectedNames, BaseNames(files))
}

func TestList(t *testing.T) {
	// Return error if NOTES_DIR is not set.
	os.Unsetenv("NOTES_DIR")
	returnedFileNames, err := List("")
	assert.Nil(t, returnedFileNames)
	assert.Equal(t, NoteDirNotSetError(true), err)

	// Return a list of notes when it is available
	dir, files := createTestFiles()
	defer os.RemoveAll(dir)
	os.Setenv("NOTES_DIR", dir)
	returnedFileNames, _ = List("")
	assert.Equal(t, files, returnedFileNames)

	// Match only given names.
	var newFiles []string
	for idx, fileName := range files {
		if idx%3 == 0 {
			oldFileName := fileName
			fileName = fileName + "foo" + strconv.Itoa(idx)
			os.Rename(oldFileName, fileName)
			newFiles = append(newFiles, fileName)
		}
	}
	setAtime(newFiles)
	returnedFileNames, _ = List("foo")
	assert.Equal(t, newFiles, returnedFileNames)
}

func TestClean(t *testing.T) {
	// Return error if NOTES_DIR is not set.
	os.Unsetenv("NOTES_DIR")
	err := Clean()
	assert.Equal(t, NoteDirNotSetError(true), err)

	// Return a list of notes when it is available
	dir, files := createTestFiles()
	defer os.RemoveAll(dir)
	os.Setenv("NOTES_DIR", dir)
	var removedFiles []string
	for idx, fileName := range files {
		if idx%3 == 0 {
			var extension string
			if idx%9 == 0 {
				extension = "swp"
			} else if idx%6 == 0 {
				extension = "swo"
			} else {
				extension = "swn"
			}
			oldFileName := fileName
			fileName = fileName + extension
			os.Rename(oldFileName, fileName)
			removedFiles = append(removedFiles, fileName)
		}
	}
	err = Clean()
	assert.Nil(t, err)
	for _, removedFile := range removedFiles {
		_, err = os.Stat(removedFile)
		errorMsg := removedFile + " exists"
		assert.True(t, os.IsNotExist(err), errorMsg)
	}
}

func TestGrep(t *testing.T) {
	// Return error if NOTES_DIR is not set.
	os.Unsetenv("NOTES_DIR")
	err := Clean()
	assert.Equal(t, NoteDirNotSetError(true), err)

	dir, files := createTestFiles()
	defer os.RemoveAll(dir)
	os.Setenv("NOTES_DIR", dir)
	fooBytes := []byte("bar\nfoo is bar\nfoo is bar\nbar")
	var expectedFileNames []string
	for idx, fileName := range files {
		if idx%3 == 0 {
			fmt.Println(fileName)
			ioutil.WriteFile(fileName, fooBytes, 0644)
			expectedFileNames = append(expectedFileNames, fileName)
		}
	}
	setAtime(expectedFileNames)
	returnedFileNames, _ := Grep("foo is bar")
	assert.Equal(t, expectedFileNames, returnedFileNames)
}

func TestEditor(t *testing.T) {
	os.Unsetenv("EDITOR")
	err := Edit("foo", false)
	assert.Equal(t, EditorNotSetError(true), err)

	os.Setenv("EDITOR", "touch")

	// Return error if NOTES_DIR is not set.
	os.Unsetenv("NOTES_DIR")
	err = Edit("foo", false)
	assert.Equal(t, NoteDirNotSetError(true), err)

	dir, files := createTestFiles()
	defer os.RemoveAll(dir)

	// Return when file doesn't exist and create is false.
	os.Setenv("NOTES_DIR", dir)
	err = Edit("foo", false)
	assert.Equal(t, NoFilesError(true), err)

	// Create and edit the file when it doesn't exist and create=true.
	os.Setenv("EDITOR", "cat")
	err = Edit("bar", true)
	assert.Nil(t, err)
	_, err = os.Stat(path.Join(dir, "bar"))
	assert.Nil(t, err)

	var chosenFile string
	for idx, fileName := range files[:3] {
		oldFileName := fileName
		fileName = fileName + "foo" + strconv.Itoa(idx)
		if idx == 1 {
			chosenFile = fileName
		}
		os.Rename(oldFileName, fileName)
	}
	err = Edit("foo", false)
	_, ok := err.(*MultipleFilesError)
	assert.True(t, ok)

	// Test editing of file. Since EDITOR is set to touch, it will
	// change the access time of the file.
	os.Setenv("EDITOR", "touch")
	prevAtime := Atime(chosenFile)
	err = Edit("foo1", false)
	assert.Nil(t, err)
	newAtime := Atime(chosenFile)
	assert.NotEqual(t, prevAtime, newAtime)

	err = Edit("bar", false)

	// Returns error when editor doesn't exist.
	os.Setenv("EDITOR", "unknowneditor")
	err = Edit("foo1", false)
	assert.NotNil(t, err)
}
