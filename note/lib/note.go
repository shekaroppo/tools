package lib

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type NoteDirNotSetError bool
type EditorNotSetError bool
type NoFilesError bool
type MultipleFilesError struct {
	files []string
}

func (e NoteDirNotSetError) Error() string {
	return "'NOTES_DIR' environment variable not defined."
}

func (e EditorNotSetError) Error() string {
	return "'EDITOR' environment variable not defined."
}

func (e NoFilesError) Error() string {
	return "No files matched given name."
}

func (e *MultipleFilesError) Error() string {
	errorMsg := "Multiple files match the name: \n"
	errorMsg += strings.Join(BaseNames(e.files), "\n") + "\n"
	errorMsg += "Please choose a more exact one."
	return errorMsg
}

type FileList []string

func (f FileList) Len() int {
	return len(f)
}

func (f FileList) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f FileList) Less(i, j int) bool {
	atimi := Atime(f[i])
	atimj := Atime(f[j])
	return atimi < atimj
}

func BaseNames(files []string) []string {
	var basenames []string
	for _, file := range files {
		basenames = append(basenames, path.Base(file))
	}
	return basenames
}

func List(name string) ([]string, error) {
	dir := os.Getenv("NOTES_DIR")
	if dir == "" {
		return nil, NoteDirNotSetError(true)
	}
	var files []string
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Print(err)
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if name == "" || strings.Contains(path, name) {
			files = append(files, path)
		}
		return nil
	}
	filepath.Walk(dir, walkFn)
	sort.Sort(FileList(files))
	return files, nil
}

func Grep(pattern string) ([]string, error) {
	files, err := List("")
	if err != nil {
		return nil, err
	}
	patternBytes := []byte(pattern)
	var matchingFiles []string
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			log.Print(err)
			continue
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if bytes.Contains(scanner.Bytes(), patternBytes) {
				matchingFiles = append(matchingFiles, file)
				break
			}
		}
	}
	return matchingFiles, nil
}

func Clean() error {
	files, err := List("")
	if err != nil {
		return err
	}
	tempFileRegex, _ := regexp.Compile(".*(swp|swo|swn)")
	for _, name := range files {
		if tempFileRegex.MatchString(path.Base(name)) {
			os.Remove(name)
		}
	}
	return nil
}

func Edit(name string, create bool) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return EditorNotSetError(true)
	}
	matchingFiles, err := List(name)
	if err != nil {
		return err
	}
	var file string
	if len(matchingFiles) == 0 {
		if !create {
			return NoFilesError(true)
		}
		file = path.Join(os.Getenv("NOTES_DIR"), name)
		ioutil.WriteFile(file, []byte(""), 0644)
	} else if len(matchingFiles) > 1 {
		return &MultipleFilesError{matchingFiles}
	} else {
		file = matchingFiles[0]
	}
	cmd := exec.Command(editor, file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	return err
}
