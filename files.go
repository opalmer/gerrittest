package gerrittest

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// File contains information about a file to be place on disk inside
// of a repository.
type File struct {
	// Mode is the file mode to be written to disk.
	Mode os.FileMode `json:"mode"`

	// Contents is the data to be written to disk.
	Contents []byte `json:"contents"`
}

// NewFile reads the provided path and produces a *File struct.
func NewFile(root string, relative string, mode os.FileMode) (*File, error) {
	file, err := os.Open(filepath.Join(root, relative))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	contents := &bytes.Buffer{}
	if _, err := io.Copy(contents, file); err != nil {
		return nil, err
	}
	return &File{Mode: mode, Contents: contents.Bytes()}, nil
}

// FileRepository represents a file structure on disk as a set of
// structures.
type FileRepository struct {
	Root  string           `json:"root"`
	Files map[string]*File `json:"files"`
}

// Write takes the file repository and writes it to disk.
func (f *FileRepository) Write(root string, perm os.FileMode) error {
	if err := os.MkdirAll(root, perm); err != nil {
		return err
	}
	for relative, file := range f.Files {
		if err := ioutil.WriteFile(filepath.Join(root, relative), file.Contents, file.Mode); err != nil {
			return err
		}
	}
	return nil
}

// NewFileRepository transforms a file structure on disk into a
// *FileRepository. This can then be used to write out a json structure
// to disk for later testing.
func NewFileRepository(root string) (*FileRepository, error) {
	files := map[string]*File{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		file, err := NewFile(root, relative, info.Mode())
		if err != nil {
			return err
		}
		files[relative] = file
		return nil
	})
	return &FileRepository{Root: root, Files: files}, err
}
