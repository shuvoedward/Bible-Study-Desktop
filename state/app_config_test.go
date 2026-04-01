package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

type MockFileSystem struct {
	Files map[string][]byte
}

func (fs MockFileSystem) ReadFile(path string) ([]byte, error) {
	data, exists := fs.Files[path]
	if !exists {
		return nil, os.ErrNotExist
	}
	return data, nil
}
func (fs MockFileSystem) WriteFile(path string, data []byte, perm fs.FileMode) error {
	fs.Files[path] = data
	return nil
}
func (fs MockFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	return nil
}
func (fs MockFileSystem) Stat(path string) (fs.FileInfo, error) {
	_, ok := fs.Files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return nil, nil
}

func TestWriteAppConfig(t *testing.T) {
	tests := []struct {
		cfg AppConfig
	}{
		{cfg: AppConfig{VaultPath: "/User/cornelius/Document/"}},
	}

	fs := NewAppFileSystem()

	for _, tt := range tests {
		if err := WriteAppConfig(tt.cfg, fs); err != nil {
			t.Fatalf("write app config failed: %v", err)
		}

		basePath, err := os.UserConfigDir()
		if err != nil {
			t.Fatal("getting config path failed")
		}

		filePath := filepath.Join(basePath, "bible-app", "config.json")
		data, err := fs.ReadFile(filePath)
		if err != nil {
			t.Fatalf("read file failed")
		}
		var cfg AppConfig
		err = json.Unmarshal(data, &cfg)
		if err != nil {
			t.Fatal("failed to unmarshal json")
		}
		fmt.Println(cfg.VaultPath)
		if cfg.VaultPath != tt.cfg.VaultPath {
			t.Error("Vault path are wrong")
		}

		dirPath := filepath.Join(basePath, "bible-app")
		err = os.RemoveAll(dirPath)
		if err != nil {
			fmt.Println("Error deleting directory tree: ", err)
			return

		}
	}
}

func TestWriteAppConfig_MockFS(t *testing.T) {
	mockFS := &MockFileSystem{Files: map[string][]byte{}}
	cfg := AppConfig{VaultPath: "/User/cornelius/vault/"}

	err := WriteAppConfig(cfg, mockFS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	basePath, _ := os.UserConfigDir()
	expectedPath := filepath.Join(basePath, "bible-app", "config.json")

	// 1. was the file written to the correct path?
	data, ok := mockFS.Files[expectedPath]
	if !ok {
		t.Fatal("config.json was not written to the expected path")
	}

	// 2. is the content correct?
	var result AppConfig
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("written data is not valid JSON: %v", err)
	}

	if result.VaultPath != cfg.VaultPath {
		t.Errorf("got VaultPath %q, want %q", result.VaultPath, cfg.VaultPath)
	}

}

func TestLoadAppConfig(t *testing.T) {
	fs := NewAppFileSystem()

	_, err := LoadAppConfig(fs)
	if err != nil && !errors.Is(err, ErrConfigNotFound) {
		t.Errorf("failed error handling: %v", err)
	}

	cfg := AppConfig{VaultPath: "/User/cornelius/Documents/"}
	err = WriteAppConfig(cfg, fs)
	defer deleteTestDir()
	if err != nil {
		t.Fatalf("failed to write app configuration: %v", err)
	}

	appCfg, err := LoadAppConfig(fs)
	if err != nil {
		t.Fatalf("failed to read app config: %v", err)
	}

	if appCfg.VaultPath != cfg.VaultPath {
		t.Errorf("got vault path: %s, expected: %s", appCfg.VaultPath, cfg.VaultPath)
	}

}

func deleteTestDir() {
	basePath, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Error getting basepath")
	}
	dirPath := filepath.Join(basePath, "bible-app")
	err = os.RemoveAll(dirPath)
	if err != nil {
		fmt.Println("Error deleting directory tree: ", err)
	}
}
