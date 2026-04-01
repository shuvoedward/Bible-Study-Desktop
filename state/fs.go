package state

import (
	"io/fs"
	"os"
)

// FileSystem is an interface for OS-level file operations.
// Services depend on this, never on `os` directly.
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
	Stat(path string) (fs.FileInfo, error)
}

type AppFileSystem struct{}

func NewAppFileSystem() FileSystem {
	return AppFileSystem{}
}

func (AppFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (AppFileSystem) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (AppFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (AppFileSystem) Stat(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}
