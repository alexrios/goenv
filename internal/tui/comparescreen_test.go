package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

// --- CompareScreenModel Tests ---

func TestCompareScreen_NewModel(t *testing.T) {
	m := NewCompareScreenModel()

	if m.phase != comparePhaseSelect {
		t.Errorf("NewModel: phase = %v, want comparePhaseSelect", m.phase)
	}
}

func TestCompareScreen_Init(t *testing.T) {
	m := NewCompareScreenModel()
	cmd := m.Init()

	if m.phase != comparePhaseSelect {
		t.Errorf("Init: phase = %v, want comparePhaseSelect", m.phase)
	}
	if m.selectedIdx != 0 {
		t.Errorf("Init: selectedIdx = %d, want 0", m.selectedIdx)
	}
	if cmd == nil {
		t.Error("Init should return load sources command")
	}
}

func TestCompareScreen_SelectNavigation_Down(t *testing.T) {
	m := NewCompareScreenModel()
	m.sources = []CompareSource{
		{Name: "snapshot1", Type: "snapshot"},
		{Name: "snapshot2", Type: "snapshot"},
		{Name: "preset1", Type: "preset"},
	}
	m.selectedIdx = 0

	msg := tea.KeyPressMsg{Code: 'j', Text: "j"}
	m.Update(msg)

	if m.selectedIdx != 1 {
		t.Errorf("j: selectedIdx = %d, want 1", m.selectedIdx)
	}

	// Also test 'down' key
	msg = tea.KeyPressMsg{Code: tea.KeyDown}
	m.Update(msg)

	if m.selectedIdx != 2 {
		t.Errorf("down: selectedIdx = %d, want 2", m.selectedIdx)
	}
}

func TestCompareScreen_SelectNavigation_Up(t *testing.T) {
	m := NewCompareScreenModel()
	m.sources = []CompareSource{
		{Name: "snapshot1", Type: "snapshot"},
		{Name: "snapshot2", Type: "snapshot"},
	}
	m.selectedIdx = 1

	msg := tea.KeyPressMsg{Code: 'k', Text: "k"}
	m.Update(msg)

	if m.selectedIdx != 0 {
		t.Errorf("k: selectedIdx = %d, want 0", m.selectedIdx)
	}
}

func TestCompareScreen_SelectNavigation_BoundsCheck(t *testing.T) {
	m := NewCompareScreenModel()
	m.sources = []CompareSource{
		{Name: "snapshot1", Type: "snapshot"},
		{Name: "snapshot2", Type: "snapshot"},
	}

	// At start, try to go up
	m.selectedIdx = 0
	msg := tea.KeyPressMsg{Code: 'k', Text: "k"}
	m.Update(msg)
	if m.selectedIdx != 0 {
		t.Errorf("k at start: selectedIdx = %d, should stay at 0", m.selectedIdx)
	}

	// At end, try to go down
	m.selectedIdx = 1
	msg = tea.KeyPressMsg{Code: 'j', Text: "j"}
	m.Update(msg)
	if m.selectedIdx != 1 {
		t.Errorf("j at end: selectedIdx = %d, should stay at 1", m.selectedIdx)
	}
}

func TestCompareScreen_EnterLoadsAndComputesDiff(t *testing.T) {
	m := NewCompareScreenModel()
	m.SetItems([]goenv.EnvVar{
		{Key: "GOPATH", Value: "/current"},
		{Key: "GOROOT", Value: "/usr/local/go"},
	})
	m.sources = []CompareSource{
		{
			Name: "snapshot1",
			Type: "snapshot",
			Snapshot: persist.Snapshot{
				Variables: map[string]string{
					"GOPATH": "/old",          // modified
					"GOROOT": "/usr/local/go", // unchanged
					"GOBIN":  "/bin",          // added
				},
			},
		},
	}
	m.selectedIdx = 0

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	m.Update(msg)

	if m.phase != comparePhaseView {
		t.Errorf("Enter: phase = %v, want comparePhaseView", m.phase)
	}
	if len(m.diff) == 0 {
		t.Error("diff should be populated after Enter")
	}
}

func TestCompareScreen_DiffSortedModifiedFirst(t *testing.T) {
	m := NewCompareScreenModel()
	m.SetItems([]goenv.EnvVar{
		{Key: "AAA", Value: "same"},  // unchanged
		{Key: "BBB", Value: "new"},   // modified
		{Key: "CCC", Value: "value"}, // removed (not in snapshot)
	})

	snapshot := persist.Snapshot{
		Variables: map[string]string{
			"AAA": "same",  // unchanged
			"BBB": "old",   // modified
			"DDD": "added", // added
		},
	}

	m.computeDiff(snapshot)

	// Check that modified comes first
	if len(m.diff) < 1 {
		t.Fatal("diff should have entries")
	}

	// First entries should be modified
	foundModifiedFirst := false
	for i, d := range m.diff {
		if d.Status == DiffModified && i == 0 {
			foundModifiedFirst = true
			break
		}
		if d.Status != DiffModified && i == 0 {
			// Check order: Modified should come before Added, Removed, Unchanged
			break
		}
	}
	if !foundModifiedFirst {
		t.Error("modified entries should come first in diff")
	}
}

func TestCompareScreen_ViewPhaseNavigation(t *testing.T) {
	m := NewCompareScreenModel()
	m.phase = comparePhaseView
	m.diff = []VarDiff{
		{Key: "A", Status: DiffModified},
		{Key: "B", Status: DiffAdded},
		{Key: "C", Status: DiffRemoved},
		{Key: "D", Status: DiffUnchanged},
	}
	m.cursorPos = 0

	// Navigate down
	msg := tea.KeyPressMsg{Code: 'j', Text: "j"}
	m.Update(msg)
	if m.cursorPos != 1 {
		t.Errorf("j: cursorPos = %d, want 1", m.cursorPos)
	}

	// Navigate up
	msg = tea.KeyPressMsg{Code: 'k', Text: "k"}
	m.Update(msg)
	if m.cursorPos != 0 {
		t.Errorf("k: cursorPos = %d, want 0", m.cursorPos)
	}
}

func TestCompareScreen_ViewPhase_EscGoesBackToSelect(t *testing.T) {
	m := NewCompareScreenModel()
	m.phase = comparePhaseView

	msg := tea.KeyPressMsg{Code: tea.KeyEscape}
	m.Update(msg)

	if m.phase != comparePhaseSelect {
		t.Errorf("Esc: phase = %v, want comparePhaseSelect", m.phase)
	}
}

func TestCompareScreen_SelectPhase_EscReturnsToList(t *testing.T) {
	m := NewCompareScreenModel()
	m.phase = comparePhaseSelect

	msg := tea.KeyPressMsg{Code: tea.KeyEscape}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("Esc should return command")
	}

	resultMsg := cmd()
	changeMsg, ok := resultMsg.(changeScreenMsg)
	if !ok {
		t.Fatalf("expected changeScreenMsg, got %T", resultMsg)
	}
	if AppScreen(changeMsg) != AppScreenList {
		t.Errorf("changeScreenMsg = %v, want AppScreenList", changeMsg)
	}
}

func TestCompareScreen_ComputeDiff_AllStatuses(t *testing.T) {
	m := NewCompareScreenModel()
	m.SetItems([]goenv.EnvVar{
		{Key: "UNCHANGED", Value: "same"},
		{Key: "MODIFIED", Value: "new_val"},
		{Key: "REMOVED", Value: "current_only"},
	})

	snapshot := persist.Snapshot{
		Variables: map[string]string{
			"UNCHANGED": "same",
			"MODIFIED":  "old_val",
			"ADDED":     "snapshot_only",
		},
	}

	m.computeDiff(snapshot)

	// Count each status
	statusCount := make(map[DiffStatus]int)
	for _, d := range m.diff {
		statusCount[d.Status]++
	}

	if statusCount[DiffUnchanged] != 1 {
		t.Errorf("DiffUnchanged count = %d, want 1", statusCount[DiffUnchanged])
	}
	if statusCount[DiffModified] != 1 {
		t.Errorf("DiffModified count = %d, want 1", statusCount[DiffModified])
	}
	if statusCount[DiffAdded] != 1 {
		t.Errorf("DiffAdded count = %d, want 1", statusCount[DiffAdded])
	}
	if statusCount[DiffRemoved] != 1 {
		t.Errorf("DiffRemoved count = %d, want 1", statusCount[DiffRemoved])
	}
}

func TestCompareScreen_SourcesLoadedMsg(t *testing.T) {
	m := NewCompareScreenModel()

	sources := []CompareSource{
		{Name: "snap1", Type: "snapshot"},
		{Name: "preset1", Type: "preset"},
	}

	msg := compareSourcesLoadedMsg{Sources: sources}
	m.Update(msg)

	if len(m.sources) != 2 {
		t.Errorf("sources length = %d, want 2", len(m.sources))
	}
}

func TestCompareScreen_View_ShowsStepIndicator(t *testing.T) {
	m := NewCompareScreenModel()
	m.phase = comparePhaseSelect

	view := m.View()
	if !strings.Contains(view, "Select source to compare") {
		t.Error("Compare select phase should show 'Select source to compare'")
	}

	m.phase = comparePhaseView
	view = m.View()
	if !strings.Contains(view, "Viewing differences") {
		t.Error("Compare view phase should show 'Viewing differences'")
	}
}

func TestCompareScreen_SetItems(t *testing.T) {
	m := NewCompareScreenModel()
	items := []goenv.EnvVar{
		{Key: "GOPATH", Value: "/go"},
	}

	m.SetItems(items)

	if len(m.items) != 1 {
		t.Errorf("items length = %d, want 1", len(m.items))
	}
}

func TestCompareScreen_SnapshotToSnapshot(t *testing.T) {
	m := NewCompareScreenModel()

	a := persist.Snapshot{
		Name: "first",
		Variables: map[string]string{
			"GOPATH":  "/go1",
			"GOROOT":  "/root",
			"REMOVED": "only_in_a",
		},
	}
	b := persist.Snapshot{
		Name: "second",
		Variables: map[string]string{
			"GOPATH": "/go2",
			"GOROOT": "/root",
			"ADDED":  "only_in_b",
		},
	}

	m.computeDiffBetweenSnapshots(a, b)

	statusCount := make(map[DiffStatus]int)
	for _, d := range m.diff {
		statusCount[d.Status]++
	}

	if statusCount[DiffModified] != 1 {
		t.Errorf("DiffModified count = %d, want 1 (GOPATH)", statusCount[DiffModified])
	}
	if statusCount[DiffUnchanged] != 1 {
		t.Errorf("DiffUnchanged count = %d, want 1 (GOROOT)", statusCount[DiffUnchanged])
	}
	if statusCount[DiffRemoved] != 1 {
		t.Errorf("DiffRemoved count = %d, want 1 (REMOVED)", statusCount[DiffRemoved])
	}
	if statusCount[DiffAdded] != 1 {
		t.Errorf("DiffAdded count = %d, want 1 (ADDED)", statusCount[DiffAdded])
	}
}

func TestCompareScreen_VKeyEntersSnapshotToSnapshotMode(t *testing.T) {
	m := NewCompareScreenModel()
	m.sources = []CompareSource{
		{Name: "snap1", Type: "snapshot", Snapshot: persist.Snapshot{Name: "snap1"}},
		{Name: "snap2", Type: "snapshot", Snapshot: persist.Snapshot{Name: "snap2"}},
	}
	m.phase = comparePhaseSelect
	m.selectedIdx = 0

	msg := tea.KeyPressMsg{Code: 'v', Text: "v"}
	m.Update(msg)

	if m.phase != comparePhaseSelectSecond {
		t.Errorf("phase should be comparePhaseSelectSecond after 'v', got %v", m.phase)
	}
	if m.firstSource == nil {
		t.Error("firstSource should be set after pressing 'v'")
	}
	if m.firstSource.Name != "snap1" {
		t.Errorf("firstSource.Name = %q, want snap1", m.firstSource.Name)
	}
}
