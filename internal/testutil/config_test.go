package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTempConfigPreservesExistingFiles(t *testing.T) {
	// Simulate an existing user configuration without accessing the real one.
	root := t.TempDir()
	t.Setenv(ConfigEnv(), root)
	originalDir, err := ConfigDir(root)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{"config.json", "history.json", filepath.Join("presets", "preset.json")}
	for _, name := range names {
		path := filepath.Join(originalDir, "goenv", name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("original"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("isolated fixture writes", func(t *testing.T) {
		dir := TempConfig(t)
		if dir == originalDir {
			t.Fatal("temporary configuration reused the original directory")
		}
		for _, name := range names {
			path := filepath.Join(dir, "goenv", name)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("fixture"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	})

	if dir, err := os.UserConfigDir(); err != nil || dir != originalDir {
		t.Fatalf("configuration not restored: dir=%q err=%v", dir, err)
	}
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(originalDir, "goenv", name))
		if err != nil || string(data) != "original" {
			t.Errorf("original %s changed: data=%q err=%v", name, data, err)
		}
	}
	if _, err := ConfigDir(t.TempDir()); err == nil {
		t.Fatal("accepted a configuration directory outside the temporary root")
	}
}
