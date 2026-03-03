package persist

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alexrios/goenv/internal/goenv"
)

// Snapshot represents a saved state of Go environment variables.
type Snapshot struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	GoVersion   string            `json:"goVersion,omitempty"`
	Variables   map[string]string `json:"variables"`
}

// NewSnapshot creates a new snapshot from the given environment variables.
func NewSnapshot(name string, items []goenv.EnvVar, goVersion string) Snapshot {
	variables := make(map[string]string)
	for _, ev := range items {
		variables[ev.Key] = ev.Value
	}

	return Snapshot{
		Name:      name,
		CreatedAt: time.Now(),
		GoVersion: goVersion,
		Variables: variables,
	}
}

// ExportSnapshot saves a snapshot to a JSON file.
func ExportSnapshot(snapshot Snapshot, filePath string) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write snapshot file: %w", err)
	}

	return nil
}

// ImportSnapshot loads a snapshot from a JSON file.
func ImportSnapshot(filePath string) (Snapshot, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Snapshot{}, fmt.Errorf("failed to read snapshot file: %w", err)
	}

	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("failed to parse snapshot: %w", err)
	}

	return snapshot, nil
}

// SnapshotDiff represents the difference between current env and a snapshot.
type SnapshotDiff struct {
	Added    map[string]string // In snapshot but not current
	Removed  map[string]string // In current but not snapshot
	Modified map[string]struct {
		Current  string
		Snapshot string
	}
	Unchanged []string
}

// CompareWithSnapshot compares current environment with a snapshot.
func CompareWithSnapshot(items []goenv.EnvVar, snapshot Snapshot) SnapshotDiff {
	diff := SnapshotDiff{
		Added:   make(map[string]string),
		Removed: make(map[string]string),
		Modified: make(map[string]struct {
			Current  string
			Snapshot string
		}),
	}

	// Build current env map
	current := make(map[string]string)
	for _, ev := range items {
		current[ev.Key] = ev.Value
	}

	// Check snapshot variables
	for key, snapshotVal := range snapshot.Variables {
		if currentVal, exists := current[key]; exists {
			if currentVal != snapshotVal {
				diff.Modified[key] = struct {
					Current  string
					Snapshot string
				}{Current: currentVal, Snapshot: snapshotVal}
			} else {
				diff.Unchanged = append(diff.Unchanged, key)
			}
		} else {
			diff.Added[key] = snapshotVal
		}
	}

	// Check for removed variables (in current but not in snapshot)
	for key, currentVal := range current {
		if _, exists := snapshot.Variables[key]; !exists {
			diff.Removed[key] = currentVal
		}
	}

	return diff
}

// DefaultSnapshotDir returns the default directory for snapshots.
func DefaultSnapshotDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ConfigDirName, "snapshots"), nil
}

// ListSnapshots returns all snapshots in the default directory.
// Also returns the count of files that could not be parsed.
func ListSnapshots() (snapshots []Snapshot, skipped int, err error) {
	dir, err := DefaultSnapshotDir()
	if err != nil {
		return nil, 0, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		snapshot, err := ImportSnapshot(path)
		if err == nil {
			snapshots = append(snapshots, snapshot)
		} else {
			skipped++
		}
	}

	return snapshots, skipped, nil
}

// SanitizeFilename removes or replaces invalid filename characters.
func SanitizeFilename(name string) string {
	// Replace spaces and invalid characters
	name = replaceChars(name, " /\\:*?\"<>|", '_')
	if name == "" {
		name = "snapshot"
	}
	return name
}

func replaceChars(s string, chars string, replacement byte) string {
	result := []byte(s)
	for i := range result {
		for j := 0; j < len(chars); j++ {
			if result[i] == chars[j] {
				result[i] = replacement
				break
			}
		}
	}
	return string(result)
}

// UniqueFilePath returns a file path that doesn't collide with existing files.
// If dir/base.ext exists, it tries dir/base_2.ext, dir/base_3.ext, etc.
func UniqueFilePath(dir, base, ext string) string {
	isAvailable := func(path string) bool {
		_, err := os.Stat(path)
		return err != nil // ErrNotExist or any other error (e.g. permission)
	}

	path := filepath.Join(dir, base+ext)
	if isAvailable(path) {
		return path
	}
	for i := 2; i < 1000; i++ {
		path = filepath.Join(dir, fmt.Sprintf("%s_%d%s", base, i, ext))
		if isAvailable(path) {
			return path
		}
	}
	// Fallback: timestamp-based suffix to avoid data loss
	return filepath.Join(dir, fmt.Sprintf("%s_%d%s", base, time.Now().UnixNano(), ext))
}
