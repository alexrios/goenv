package persist

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const historyFileName = "history.json"
const defaultMaxHistoryEntries = 50

// EditRecord stores a single edit operation for undo/redo functionality.
type EditRecord struct {
	Key      string `json:"key"`
	OldValue string `json:"oldValue"`
	NewValue string `json:"newValue"`
}

// SessionHistory stores edit history that persists across sessions.
type SessionHistory struct {
	Records    []EditRecord `json:"records"`
	CurrentIdx int          `json:"currentIdx"`
}

// historyPath returns the full path to the history file.
func historyPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ConfigDirName, historyFileName), nil
}

// LoadHistory loads history from disk.
// Returns empty history if file doesn't exist.
func LoadHistory() (SessionHistory, error) {
	path, err := historyPath()
	if err != nil {
		return SessionHistory{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return SessionHistory{}, nil
		}
		return SessionHistory{}, err
	}

	var history SessionHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return SessionHistory{}, err
	}

	// Validate indices
	if history.CurrentIdx < 0 {
		history.CurrentIdx = 0
	}
	if history.CurrentIdx > len(history.Records) {
		history.CurrentIdx = len(history.Records)
	}

	return history, nil
}

// SaveHistory persists history to disk.
// maxEntries controls the maximum number of history records to keep.
// If maxEntries <= 0, the default of 50 is used.
func SaveHistory(history SessionHistory, maxEntries int) error {
	if maxEntries <= 0 {
		maxEntries = defaultMaxHistoryEntries
	}

	path, err := historyPath()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Trim history to max entries
	if len(history.Records) > maxEntries {
		// Keep the most recent entries
		start := len(history.Records) - maxEntries
		history.Records = history.Records[start:]
		// Adjust currentIdx
		history.CurrentIdx -= start
		if history.CurrentIdx < 0 {
			history.CurrentIdx = 0
		}
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return AtomicWriteFile(path, data, 0o644)
}

// ClearHistory removes the history file.
func ClearHistory() error {
	path, err := historyPath()
	if err != nil {
		return err
	}
	return os.Remove(path)
}
