// Package testutil provides configuration isolation for tests.
package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// ConfigEnv returns the variable used by os.UserConfigDir on this platform.
func ConfigEnv() string {
	switch runtime.GOOS {
	case "darwin", "ios":
		return "HOME"
	case "windows":
		return "AppData"
	case "plan9":
		return "home"
	default:
		return "XDG_CONFIG_HOME"
	}
}

// ConfigDir checks that the resolved configuration directory stays inside root.
func ConfigDir(root string) (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, dir)
	if err != nil || !filepath.IsLocal(rel) {
		return "", fmt.Errorf("configuration directory %q is outside temporary root %q", dir, root)
	}
	return dir, nil
}

// TempConfig isolates configuration until the test completes and returns its directory.
func TempConfig(t testing.TB) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv(ConfigEnv(), root)
	dir, err := ConfigDir(root)
	if err != nil {
		t.Fatal(err)
	}
	return dir
}
