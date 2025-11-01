package dir

import (
	"os"
	"path/filepath"
)

// Dir represents a directory in the filesystem
type Dir struct {
	path string
}

// New creates a new Dir instance with the given path
func New(path string) *Dir {
	return &Dir{path: path}
}

// Create creates the directory if it does not exist
// also creates parent directories as needed
func (d *Dir) Create(perm os.FileMode) error {
	return os.MkdirAll(d.path, perm)
}

// Exists checks if the directory exists
func (d *Dir) Exists() bool {
	v, err := os.Stat(d.path)
	return err == nil && v.IsDir()
}

// List returns a list of directory entries
func (d *Dir) List() ([]os.DirEntry, error) {
	return os.ReadDir(d.path)
}

// Name returns the name of the directory
func (d *Dir) Name() string {
	return filepath.Base(d.path)
}

// Path returns the path of the directory
func (d *Dir) Path() string {
	return d.path
}

// Remove removes the directory and all its contents
func (d *Dir) Remove() error {
	return os.RemoveAll(d.path)
}
