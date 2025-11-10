package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// File represents a file in the filesystem
type File struct {
	path string
}

// New creates a new File instance with the given path
func New(path string) *File {
	return &File{path: path}
}

// Append appends data to the file, creating it if it does not exist
func (f *File) Append(data []byte, perm os.FileMode) (int, error) {
	v, err := os.OpenFile(f.path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, perm)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", f.path, err)
	}
	defer v.Close()

	return v.Write(data)
}

// CopyTo copies the file to the destination path with the given permissions
func (f *File) CopyTo(dest string, perm os.FileMode) error {
	src, err := os.Open(f.path)
	if err != nil {
		return fmt.Errorf("open source %s: %w", f.path, err)
	}
	defer src.Close()

	df, err := os.OpenFile(filepath.Clean(dest), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("open destination %s: %w", dest, err)
	}
	defer df.Close()

	if _, err = io.Copy(df, src); err != nil {
		return fmt.Errorf("copy from %s to %s: %w", f.path, dest, err)
	}
	return nil
}

// Create creates the file with the given permissions
func (f *File) Create(perm os.FileMode) error {
	v, err := os.OpenFile(f.path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err != nil {
		return fmt.Errorf("open %s: %w", f.path, err)
	}
	if err := v.Close(); err != nil {
		return fmt.Errorf("close %s: %w", f.path, err)
	}
	return nil
}

// Exists checks if the file exists
func (f *File) Exists() bool {
	_, err := os.Stat(f.path)
	return err == nil
}

// MoveTo moves the file to the destination path with the given permissions
func (f *File) MoveTo(dest string, perm os.FileMode) error {
	if err := os.Rename(f.path, dest); err == nil {
		return nil
	}
	if err := f.CopyTo(dest, perm); err != nil {
		return fmt.Errorf("copy to %s: %w", dest, err)
	}
	if err := f.Remove(); err != nil {
		return fmt.Errorf("remove source %s: %w", f.path, err)
	}
	return nil
}

// Name returns the name of the file
func (f *File) Name() string {
	return filepath.Base(f.path)
}

// Path returns the path of the file
func (f *File) Path() string {
	return f.path
}

// Read reads the contents of the file
func (f *File) Read() ([]byte, error) {
	return os.ReadFile(f.path)
}

// Remove removes the file
func (f *File) Remove() error {
	return os.Remove(f.path)
}

// Size returns the size of the file in bytes
func (f *File) Size() (int64, error) {
	v, err := os.Stat(f.path)
	if err != nil {
		return 0, fmt.Errorf("stat %s: %w", f.path, err)
	}
	return v.Size(), nil
}

// Write writes data to the file with the given permissions
func (f *File) Write(data []byte, perm os.FileMode) error {
	if err := os.WriteFile(f.path, data, perm); err != nil {
		return fmt.Errorf("write %s: %w", f.path, err)
	}
	return nil
}
