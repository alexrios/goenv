package persist

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const configFileName = "config.json"
const ConfigDirName = "goenv"

// SortMode determines how environment variables are sorted in the list.
type SortMode int

// Sort mode options.
const (
	SortAlpha         SortMode = iota // Alphabetical by key
	SortModifiedFirst                 // Modified variables first, then alphabetical
	SortCategory                      // Grouped by category, then alphabetical
)

// String returns a human-readable name for the sort mode.
func (s SortMode) String() string {
	switch s {
	case SortAlpha:
		return "Alphabetical"
	case SortModifiedFirst:
		return "Modified first"
	case SortCategory:
		return "By category"
	default:
		return "Unknown"
	}
}

// Next returns the next sort mode in the cycle.
func (s SortMode) Next() SortMode {
	switch s {
	case SortAlpha:
		return SortModifiedFirst
	case SortModifiedFirst:
		return SortCategory
	case SortCategory:
		return SortAlpha
	default:
		return SortAlpha
	}
}

// SortModeFromString converts a string to SortMode.
func SortModeFromString(s string) SortMode {
	switch s {
	case "modified_first":
		return SortModifiedFirst
	case "category":
		return SortCategory
	default:
		return SortAlpha
	}
}

// SortModeToString converts a SortMode to string.
func SortModeToString(mode SortMode) string {
	switch mode {
	case SortModifiedFirst:
		return "modified_first"
	case SortCategory:
		return "category"
	default:
		return "alpha"
	}
}

// Config holds user preferences that persist across sessions.
type Config struct {
	SortMode      string   `json:"sortMode"`
	Theme         string   `json:"theme,omitempty"`
	Favorites     []string `json:"favorites,omitzero"`
	WatchInterval int      `json:"watchInterval,omitempty"`
	MaxHistory    int      `json:"maxHistory,omitempty"`
}

// ThemeDefault is the default theme name.
const ThemeDefault = "default"

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		SortMode:  "alpha",
		Theme:     ThemeDefault,
		Favorites: []string{},
	}
}

// configPath returns the full path to the config file.
func configPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ConfigDirName, configFileName), nil
}

// LoadConfig loads configuration from disk.
// Returns default config if file doesn't exist.
func LoadConfig() (Config, error) {
	path, err := configPath()
	if err != nil {
		return DefaultConfig(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}
		return DefaultConfig(), err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err
	}

	return cfg, nil
}

// SaveConfig persists configuration to disk.
func SaveConfig(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return AtomicWriteFile(path, data, 0o644)
}

// AtomicWriteFile writes data to a file atomically by writing to a temp file
// and renaming (atomic on POSIX).
func AtomicWriteFile(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
