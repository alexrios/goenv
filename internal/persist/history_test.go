package persist

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexrios/goenv/internal/testutil"
)

func TestSaveAndLoadHistory(t *testing.T) {
	// Use a temp dir to avoid polluting the real config
	testutil.TempConfig(t)

	history := SessionHistory{
		Records: []EditRecord{
			{Key: "GOPATH", OldValue: "/old", NewValue: "/new"},
			{Key: "GOOS", OldValue: "darwin", NewValue: "linux"},
		},
		CurrentIdx: 1,
	}

	if err := SaveHistory(history, 50); err != nil {
		t.Fatalf("SaveHistory: %v", err)
	}

	loaded, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}

	if len(loaded.Records) != 2 {
		t.Errorf("loaded %d records, want 2", len(loaded.Records))
	}
	if loaded.CurrentIdx != 1 {
		t.Errorf("currentIdx = %d, want 1", loaded.CurrentIdx)
	}
	if loaded.Records[0].Key != "GOPATH" {
		t.Errorf("first record key = %q, want GOPATH", loaded.Records[0].Key)
	}
}

func TestLoadHistory_MissingFile(t *testing.T) {
	testutil.TempConfig(t)

	history, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory on missing file: %v", err)
	}
	if len(history.Records) != 0 {
		t.Errorf("expected empty records, got %d", len(history.Records))
	}
}

func TestLoadHistory_CorruptedFile(t *testing.T) {
	tmpDir := testutil.TempConfig(t)

	dir := filepath.Join(tmpDir, ConfigDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, historyFileName), []byte("{invalid json"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadHistory()
	if err == nil {
		t.Error("expected error for corrupted history file, got nil")
	}
}

func TestSaveHistory_TrimsToMaxEntries(t *testing.T) {
	testutil.TempConfig(t)

	// Create history with 10 records
	var records []EditRecord
	for i := range 10 {
		records = append(records, EditRecord{
			Key:      "GOPATH",
			OldValue: "old",
			NewValue: string(rune('a' + i)),
		})
	}
	history := SessionHistory{
		Records:    records,
		CurrentIdx: 9,
	}

	// Save with max 5
	if err := SaveHistory(history, 5); err != nil {
		t.Fatalf("SaveHistory: %v", err)
	}

	loaded, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}

	if len(loaded.Records) != 5 {
		t.Errorf("trimmed to %d records, want 5", len(loaded.Records))
	}
	// Should keep the most recent (last 5)
	if loaded.Records[0].NewValue != "f" {
		t.Errorf("first kept record = %q, want 'f'", loaded.Records[0].NewValue)
	}
}

func TestSaveHistory_DefaultMaxOnZero(t *testing.T) {
	testutil.TempConfig(t)

	history := SessionHistory{
		Records:    make([]EditRecord, 60),
		CurrentIdx: 59,
	}

	// maxEntries=0 should use default (50)
	if err := SaveHistory(history, 0); err != nil {
		t.Fatalf("SaveHistory: %v", err)
	}

	loaded, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}

	if len(loaded.Records) != defaultMaxHistoryEntries {
		t.Errorf("got %d records, want %d (default)", len(loaded.Records), defaultMaxHistoryEntries)
	}
}

func TestClearHistory(t *testing.T) {
	testutil.TempConfig(t)

	history := SessionHistory{
		Records:    []EditRecord{{Key: "GOPATH", OldValue: "a", NewValue: "b"}},
		CurrentIdx: 0,
	}
	if err := SaveHistory(history, 50); err != nil {
		t.Fatalf("SaveHistory: %v", err)
	}

	if err := ClearHistory(); err != nil {
		t.Fatalf("ClearHistory: %v", err)
	}

	// Should load empty after clear
	loaded, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory after clear: %v", err)
	}
	if len(loaded.Records) != 0 {
		t.Errorf("expected 0 records after clear, got %d", len(loaded.Records))
	}
}

func TestLoadHistory_ValidatesIndices(t *testing.T) {
	testutil.TempConfig(t)

	// Save with invalid currentIdx (too high)
	history := SessionHistory{
		Records:    []EditRecord{{Key: "GOPATH", OldValue: "a", NewValue: "b"}},
		CurrentIdx: 100,
	}
	if err := SaveHistory(history, 50); err != nil {
		t.Fatalf("SaveHistory: %v", err)
	}

	loaded, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}

	// Should clamp to exactly len(records)
	if loaded.CurrentIdx != len(loaded.Records) {
		t.Errorf("currentIdx = %d, want %d (clamped to len)", loaded.CurrentIdx, len(loaded.Records))
	}
}
