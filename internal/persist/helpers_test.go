package persist

import (
	"os"
	"path/filepath"
	"testing"
)

// --- SanitizeFilename Tests ---

func TestSanitizeFilename_Spaces(t *testing.T) {
	result := SanitizeFilename("my snapshot")
	if result != "my_snapshot" {
		t.Errorf("SanitizeFilename('my snapshot') = %q, want 'my_snapshot'", result)
	}
}

func TestSanitizeFilename_Slashes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"path/to/file", "path_to_file"},
		{"path\\to\\file", "path_to_file"},
	}

	for _, tt := range tests {
		result := SanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSanitizeFilename_SpecialChars(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"file:name", "file_name"},
		{"file*name", "file_name"},
		{"file?name", "file_name"},
		{"file\"name", "file_name"},
		{"file<name", "file_name"},
		{"file>name", "file_name"},
		{"file|name", "file_name"},
	}

	for _, tt := range tests {
		result := SanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSanitizeFilename_MultipleSpecialChars(t *testing.T) {
	result := SanitizeFilename("my:file/name*here")
	expected := "my_file_name_here"
	if result != expected {
		t.Errorf("SanitizeFilename('my:file/name*here') = %q, want %q", result, expected)
	}
}

func TestSanitizeFilename_EmptyString(t *testing.T) {
	result := SanitizeFilename("")
	if result != "snapshot" {
		t.Errorf("SanitizeFilename('') = %q, want 'snapshot'", result)
	}
}

func TestSanitizeFilename_AlreadyClean(t *testing.T) {
	result := SanitizeFilename("clean-filename_123")
	if result != "clean-filename_123" {
		t.Errorf("SanitizeFilename('clean-filename_123') = %q, want 'clean-filename_123'", result)
	}
}

// --- UniqueFilePath Tests ---

func TestUniqueFilePath_NoCollision(t *testing.T) {
	dir := t.TempDir()
	path := UniqueFilePath(dir, "test", ".json")
	if filepath.Base(path) != "test.json" {
		t.Errorf("expected test.json, got %s", filepath.Base(path))
	}
}

func TestUniqueFilePath_WithCollision(t *testing.T) {
	dir := t.TempDir()
	// Create the first file
	if err := os.WriteFile(filepath.Join(dir, "test.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := UniqueFilePath(dir, "test", ".json")
	if filepath.Base(path) != "test_2.json" {
		t.Errorf("expected test_2.json, got %s", filepath.Base(path))
	}
}

func TestUniqueFilePath_MultipleCollisions(t *testing.T) {
	dir := t.TempDir()
	// Create test.json, test_2.json, test_3.json
	for _, name := range []string{"test.json", "test_2.json", "test_3.json"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	path := UniqueFilePath(dir, "test", ".json")
	if filepath.Base(path) != "test_4.json" {
		t.Errorf("expected test_4.json, got %s", filepath.Base(path))
	}
}
