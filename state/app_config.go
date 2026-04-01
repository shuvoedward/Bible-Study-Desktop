package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
)

var (
	// ErrConfigNotFound is returned instead of os.ErrNotExist so callers
	// don't depend on OS-level errors directly.
	ErrConfigNotFound = errors.New("config not found")
	ErrJSONUnmarshal  = errors.New("json unmarshall failed")
)

type Path string

type AppConfig struct {
	Vaults      []Path `json:"vault_path"`
	ActiveVault Path   `json:"active_vault"`
}

func (cfg *AppConfig) AddVault(path Path) {
	cfg.Vaults = append(cfg.Vaults, path)
}

func (cfg *AppConfig) DeleteVault(path Path) {
	cfg.Vaults = slices.DeleteFunc(cfg.Vaults, func(vp Path) bool {
		return vp == path
	})
}

func (cfg *AppConfig) SetActiveVault(path Path) {
	cfg.ActiveVault = path
}

// LoadAppConfig reads the app config from the OS config directory.
// Returns ErrConfigNotFound (likely first launch) or ErrJSONUnmarshal
func LoadAppConfig(fs FileSystem) (*AppConfig, error) {
	basePath, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	// 	macOS:   /Users/cornelius/Library/Application Support/bible-app/config.json
	// Windows: C:\Users\cornelius\AppData\Roaming\bible-app\config.json
	// Linux:   /home/cornelius/.config/bible-app/config.json

	filePath := filepath.Join(basePath, "bible-app", "config.json")

	// = "/Users/cornelius/Library/Application Support/bible-app/config.json"

	file, err := fs.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrConfigNotFound
		}

		return nil, err
	}

	var appConfig AppConfig
	if err = json.Unmarshal(file, &appConfig); err != nil {
		return nil, ErrJSONUnmarshal
	}

	return &appConfig, nil
}

// WriteAppConfig writes AppConfig to the OS config directory.
// Errors are to be buble it up to the user.
func WriteAppConfig(cfg AppConfig, fs FileSystem) error {
	basePath, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	dirPath := filepath.Join(basePath, "bible-app")

	// 0755 - octal, 7 - owner (you): can read, write, execute (enter the folder)
	// 55 - group and others: can read + execute (enter the folder)
	err = fs.MkdirAll(dirPath, 0755)
	if err != nil {
		return err
	}

	data, err := json.Marshal(&cfg)
	if err != nil {
		return err
	}

	filePath := filepath.Join(dirPath, "config.json")

	// 0644  → owner: read+write, everyone else: read only
	err = fs.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
