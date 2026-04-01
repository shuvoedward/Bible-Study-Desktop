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
	t.Cleanup(deleteTestDir)
	tests := []struct {
		cfg AppConfig
	}{
		{cfg: AppConfig{Vaults: []Path{"/User/cornelius/Document/1"}}},
		{cfg: AppConfig{Vaults: []Path{"/User/cornelius/Document/2", "/User/cornelius/3"}}},
	}

	fs := NewAppFileSystem()

	for _, tt := range tests {
		if err := WriteAppConfig(tt.cfg, fs); err != nil {
			t.Fatalf("write app config failed: %v", err)
		}

		appCfg, err := LoadAppConfig(fs)
		if err != nil {
			t.Fatalf("failed reading config: %v", err)
		}

		if len(tt.cfg.Vaults) != len(appCfg.Vaults) {
			t.Fatalf("config doesn't have same number of vault paths\n exp: %d, got:%d", len(tt.cfg.Vaults), len(appCfg.Vaults))
		}

		for i, path := range tt.cfg.Vaults {
			if appCfg.Vaults[i] != path {
				t.Error("Vault path are wrong")
			}
		}
	}
}

func TestWriteAppConfig_MockFS(t *testing.T) {
	mockFS := &MockFileSystem{Files: map[string][]byte{}}
	cfg := AppConfig{Vaults: []Path{"/User/cornelius/vault/"}}

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

	if result.Vaults[0] != cfg.Vaults[0] {
		t.Errorf("got VaultPath %q, want %q", result.Vaults[0], cfg.Vaults[0])
	}

}

func TestLoadAppConfig(t *testing.T) {
	fs := NewAppFileSystem()

	_, err := LoadAppConfig(fs)
	if err != nil && !errors.Is(err, ErrConfigNotFound) {
		t.Errorf("failed error handling: %v", err)
	}

	cfg := AppConfig{Vaults: []Path{"/User/cornelius/Documents/"}}
	err = WriteAppConfig(cfg, fs)
	defer deleteTestDir()
	if err != nil {
		t.Fatalf("failed to write app configuration: %v", err)
	}

	appCfg, err := LoadAppConfig(fs)
	if err != nil {
		t.Fatalf("failed to read app config: %v", err)
	}

	if appCfg.Vaults[0] != cfg.Vaults[0] {
		t.Errorf("got vault path: %s, expected: %s", appCfg.Vaults[0], cfg.Vaults[0])
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
