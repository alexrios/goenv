package persist

import (
	"testing"

	"github.com/alexrios/goenv/internal/goenv"
)

// --- CompareWithSnapshot Tests ---

func TestCompareWithSnapshot_IdenticalEnvironments(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "GOPATH", Value: "/go"},
		{Key: "GOROOT", Value: "/usr/local/go"},
	}
	snapshot := Snapshot{
		Variables: map[string]string{
			"GOPATH": "/go",
			"GOROOT": "/usr/local/go",
		},
	}

	diff := CompareWithSnapshot(items, snapshot)

	if len(diff.Added) != 0 {
		t.Errorf("Added count = %d, want 0", len(diff.Added))
	}
	if len(diff.Removed) != 0 {
		t.Errorf("Removed count = %d, want 0", len(diff.Removed))
	}
	if len(diff.Modified) != 0 {
		t.Errorf("Modified count = %d, want 0", len(diff.Modified))
	}
	if len(diff.Unchanged) != 2 {
		t.Errorf("Unchanged count = %d, want 2", len(diff.Unchanged))
	}
}

func TestCompareWithSnapshot_ModifiedVariable(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "GOPATH", Value: "/new/path"},
	}
	snapshot := Snapshot{
		Variables: map[string]string{
			"GOPATH": "/old/path",
		},
	}

	diff := CompareWithSnapshot(items, snapshot)

	if len(diff.Modified) != 1 {
		t.Fatalf("Modified count = %d, want 1", len(diff.Modified))
	}
	mod, ok := diff.Modified["GOPATH"]
	if !ok {
		t.Fatal("GOPATH not in Modified map")
	}
	if mod.Current != "/new/path" {
		t.Errorf("Modified.Current = %q, want /new/path", mod.Current)
	}
	if mod.Snapshot != "/old/path" {
		t.Errorf("Modified.Snapshot = %q, want /old/path", mod.Snapshot)
	}
	if len(diff.Unchanged) != 0 {
		t.Errorf("Unchanged count = %d, want 0", len(diff.Unchanged))
	}
}

func TestCompareWithSnapshot_AddedVariable(t *testing.T) {
	// Variable in snapshot but not in current
	items := []goenv.EnvVar{}
	snapshot := Snapshot{
		Variables: map[string]string{
			"GOPATH": "/go",
		},
	}

	diff := CompareWithSnapshot(items, snapshot)

	if len(diff.Added) != 1 {
		t.Fatalf("Added count = %d, want 1", len(diff.Added))
	}
	val, ok := diff.Added["GOPATH"]
	if !ok {
		t.Fatal("GOPATH not in Added map")
	}
	if val != "/go" {
		t.Errorf("Added[GOPATH] = %q, want /go", val)
	}
}

func TestCompareWithSnapshot_RemovedVariable(t *testing.T) {
	// Variable in current but not in snapshot
	items := []goenv.EnvVar{
		{Key: "GOPATH", Value: "/go"},
	}
	snapshot := Snapshot{
		Variables: map[string]string{},
	}

	diff := CompareWithSnapshot(items, snapshot)

	if len(diff.Removed) != 1 {
		t.Fatalf("Removed count = %d, want 1", len(diff.Removed))
	}
	val, ok := diff.Removed["GOPATH"]
	if !ok {
		t.Fatal("GOPATH not in Removed map")
	}
	if val != "/go" {
		t.Errorf("Removed[GOPATH] = %q, want /go", val)
	}
}

func TestCompareWithSnapshot_MixedChanges(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "GOPATH", Value: "/go"},       // unchanged
		{Key: "GOROOT", Value: "/new/root"}, // modified
		{Key: "GOBIN", Value: "/bin"},        // removed (not in snapshot)
	}
	snapshot := Snapshot{
		Variables: map[string]string{
			"GOPATH":  "/go",           // unchanged
			"GOROOT":  "/old/root",     // modified
			"GOPROXY": "https://proxy", // added (not in current)
		},
	}

	diff := CompareWithSnapshot(items, snapshot)

	if len(diff.Unchanged) != 1 {
		t.Errorf("Unchanged count = %d, want 1", len(diff.Unchanged))
	}
	if diff.Unchanged[0] != "GOPATH" {
		t.Errorf("Unchanged[0] = %q, want GOPATH", diff.Unchanged[0])
	}

	if len(diff.Modified) != 1 {
		t.Errorf("Modified count = %d, want 1", len(diff.Modified))
	}
	if _, ok := diff.Modified["GOROOT"]; !ok {
		t.Error("GOROOT should be in Modified")
	}

	if len(diff.Added) != 1 {
		t.Errorf("Added count = %d, want 1", len(diff.Added))
	}
	if _, ok := diff.Added["GOPROXY"]; !ok {
		t.Error("GOPROXY should be in Added")
	}

	if len(diff.Removed) != 1 {
		t.Errorf("Removed count = %d, want 1", len(diff.Removed))
	}
	if _, ok := diff.Removed["GOBIN"]; !ok {
		t.Error("GOBIN should be in Removed")
	}
}

func TestCompareWithSnapshot_EmptySnapshot(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "GOPATH", Value: "/go"},
		{Key: "GOROOT", Value: "/usr/local/go"},
	}
	snapshot := Snapshot{
		Variables: map[string]string{},
	}

	diff := CompareWithSnapshot(items, snapshot)

	// All current vars should be "removed" (exist in current but not in snapshot)
	if len(diff.Removed) != 2 {
		t.Errorf("Removed count = %d, want 2", len(diff.Removed))
	}
	if len(diff.Added) != 0 {
		t.Errorf("Added count = %d, want 0", len(diff.Added))
	}
	if len(diff.Modified) != 0 {
		t.Errorf("Modified count = %d, want 0", len(diff.Modified))
	}
	if len(diff.Unchanged) != 0 {
		t.Errorf("Unchanged count = %d, want 0", len(diff.Unchanged))
	}
}

func TestCompareWithSnapshot_EmptyCurrent(t *testing.T) {
	items := []goenv.EnvVar{}
	snapshot := Snapshot{
		Variables: map[string]string{
			"GOPATH": "/go",
			"GOROOT": "/usr/local/go",
		},
	}

	diff := CompareWithSnapshot(items, snapshot)

	// All snapshot vars should be "added" (exist in snapshot but not in current)
	if len(diff.Added) != 2 {
		t.Errorf("Added count = %d, want 2", len(diff.Added))
	}
	if len(diff.Removed) != 0 {
		t.Errorf("Removed count = %d, want 0", len(diff.Removed))
	}
	if len(diff.Modified) != 0 {
		t.Errorf("Modified count = %d, want 0", len(diff.Modified))
	}
	if len(diff.Unchanged) != 0 {
		t.Errorf("Unchanged count = %d, want 0", len(diff.Unchanged))
	}
}

func TestCompareWithSnapshot_BothEmpty(t *testing.T) {
	items := []goenv.EnvVar{}
	snapshot := Snapshot{
		Variables: map[string]string{},
	}

	diff := CompareWithSnapshot(items, snapshot)

	if len(diff.Added) != 0 {
		t.Errorf("Added count = %d, want 0", len(diff.Added))
	}
	if len(diff.Removed) != 0 {
		t.Errorf("Removed count = %d, want 0", len(diff.Removed))
	}
	if len(diff.Modified) != 0 {
		t.Errorf("Modified count = %d, want 0", len(diff.Modified))
	}
	if len(diff.Unchanged) != 0 {
		t.Errorf("Unchanged count = %d, want 0", len(diff.Unchanged))
	}
}
