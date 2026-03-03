package persist

import (
	"cmp"
	"errors"
	"os"
	"path/filepath"
	"slices"
)

// Preset represents a named environment configuration.
type Preset struct {
	Snapshot
	BuiltIn bool
}

// PresetsDir returns the directory for storing presets.
func PresetsDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(configDir, ConfigDirName, "presets")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// ListPresets returns all saved presets.
// Also returns the count of files that could not be parsed.
func ListPresets() (presets []Preset, skipped int, err error) {
	dir, err := PresetsDir()
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
			presets = append(presets, Preset{Snapshot: snapshot})
		} else {
			skipped++
		}
	}

	// Sort by name
	slices.SortFunc(presets, func(a, b Preset) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return presets, skipped, nil
}

// BuiltinPresets returns the set of built-in starter presets.
// These are shipped with goenv and cannot be deleted by the user.
func BuiltinPresets() []Preset {
	return []Preset{
		{
			Snapshot: Snapshot{
				Name:        "Static Linux Build",
				Description: "Static binary for Linux amd64 (no CGO)",
				Variables: map[string]string{
					"CGO_ENABLED": "0",
					"GOOS":        "linux",
					"GOARCH":      "amd64",
				},
			},
			BuiltIn: true,
		},
		{
			Snapshot: Snapshot{
				Name:        "WASM Target",
				Description: "WebAssembly build target",
				Variables: map[string]string{
					"GOOS":   "js",
					"GOARCH": "wasm",
				},
			},
			BuiltIn: true,
		},
		{
			Snapshot: Snapshot{
				Name:        "Cross-compile ARM64",
				Description: "Cross-compile for Linux ARM64",
				Variables: map[string]string{
					"GOOS":   "linux",
					"GOARCH": "arm64",
				},
			},
			BuiltIn: true,
		},
		{
			Snapshot: Snapshot{
				Name:        "Debug Verbose",
				Description: "Enable GC and scheduler debug tracing",
				Variables: map[string]string{
					"GODEBUG": "gctrace=1,schedtrace=1000",
				},
			},
			BuiltIn: true,
		},
		{
			Snapshot: Snapshot{
				Name:        "Minimal Proxy",
				Description: "Direct module fetching, skip proxy and checksum",
				Variables: map[string]string{
					"GOPROXY":   "direct",
					"GONOSUMDB": "*",
				},
			},
			BuiltIn: true,
		},
	}
}
