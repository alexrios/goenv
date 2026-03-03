package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

// --- ImportScreenModel Tests ---

func TestImportScreen_Init(t *testing.T) {
	m := NewImportScreenModel()
	m.Init()

	if m.phase != importPhasePath {
		t.Errorf("Init: phase = %v, want importPhasePath", m.phase)
	}
	if m.textInput.Value() != "" {
		t.Errorf("Init: textInput value = %q, want empty", m.textInput.Value())
	}
	if len(m.selectedVars) != 0 {
		t.Errorf("Init: selectedVars length = %d, want 0", len(m.selectedVars))
	}
	if m.cursorPos != 0 {
		t.Errorf("Init: cursorPos = %d, want 0", m.cursorPos)
	}
}

func TestImportScreen_EnterOnEmptyPath_NoAction(t *testing.T) {
	m := NewImportScreenModel()
	m.Init()
	m.textInput.SetValue("") // empty path

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	cmd := m.Update(msg)

	if cmd != nil {
		t.Error("Enter on empty path should return nil command")
	}
}

func TestImportScreen_EnterOnValidPath_LoadsSnapshot(t *testing.T) {
	m := NewImportScreenModel()
	m.Init()
	m.textInput.SetValue("/path/to/snapshot.json")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("Enter on valid path should return load command")
	}
	// The command will try to load the file (which may fail in tests)
	// We just verify the command is returned
}

func TestImportScreen_EscFromPath_ReturnsToList(t *testing.T) {
	m := NewImportScreenModel()
	m.Init()

	msg := tea.KeyPressMsg{Code: tea.KeyEscape}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("Esc from path should return command")
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

func TestImportScreen_EscFromPreview_GoesBackToPath(t *testing.T) {
	m := NewImportScreenModel()
	m.Init()
	m.phase = importPhasePreview // Simulate being in preview phase

	msg := tea.KeyPressMsg{Code: tea.KeyEscape}
	m.Update(msg)

	if m.phase != importPhasePath {
		t.Errorf("Esc from preview: phase = %v, want importPhasePath", m.phase)
	}
}

func TestImportScreen_SpaceTogglesSelection(t *testing.T) {
	m := NewImportScreenModel()
	m.Init()
	m.phase = importPhasePreview
	m.diffKeys = []string{"GOPATH", "GOROOT", "GOBIN"}
	m.selectedVars = map[string]bool{"GOPATH": true, "GOROOT": false, "GOBIN": false}
	m.cursorPos = 0

	msg := tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	m.Update(msg)

	// GOPATH should now be false (toggled)
	if m.selectedVars["GOPATH"] {
		t.Error("GOPATH should be deselected after toggle")
	}

	// Toggle again
	m.Update(msg)
	if !m.selectedVars["GOPATH"] {
		t.Error("GOPATH should be selected after second toggle")
	}
}

func TestImportScreen_SelectAll(t *testing.T) {
	m := NewImportScreenModel()
	m.Init()
	m.phase = importPhasePreview
	m.diffKeys = []string{"GOPATH", "GOROOT", "GOBIN"}
	m.selectedVars = map[string]bool{"GOPATH": false, "GOROOT": false, "GOBIN": false}

	msg := tea.KeyPressMsg{Code: 'a', Text: "a"}
	m.Update(msg)

	for _, key := range m.diffKeys {
		if !m.selectedVars[key] {
			t.Errorf("'a' should select all: %s not selected", key)
		}
	}
}

func TestImportScreen_DeselectAll(t *testing.T) {
	m := NewImportScreenModel()
	m.Init()
	m.phase = importPhasePreview
	m.diffKeys = []string{"GOPATH", "GOROOT", "GOBIN"}
	m.selectedVars = map[string]bool{"GOPATH": true, "GOROOT": true, "GOBIN": true}

	msg := tea.KeyPressMsg{Code: 'n', Text: "n"}
	m.Update(msg)

	for _, key := range m.diffKeys {
		if m.selectedVars[key] {
			t.Errorf("'n' should deselect all: %s still selected", key)
		}
	}
}

func TestImportScreen_Navigation_Down(t *testing.T) {
	m := NewImportScreenModel()
	m.Init()
	m.phase = importPhasePreview
	m.diffKeys = []string{"GOPATH", "GOROOT", "GOBIN"}
	m.cursorPos = 0

	msg := tea.KeyPressMsg{Code: 'j', Text: "j"}
	m.Update(msg)

	if m.cursorPos != 1 {
		t.Errorf("j: cursorPos = %d, want 1", m.cursorPos)
	}

	// Also test 'down' key
	msg = tea.KeyPressMsg{Code: tea.KeyDown}
	m.Update(msg)

	if m.cursorPos != 2 {
		t.Errorf("down: cursorPos = %d, want 2", m.cursorPos)
	}
}

func TestImportScreen_Navigation_Up(t *testing.T) {
	m := NewImportScreenModel()
	m.Init()
	m.phase = importPhasePreview
	m.diffKeys = []string{"GOPATH", "GOROOT", "GOBIN"}
	m.cursorPos = 2

	msg := tea.KeyPressMsg{Code: 'k', Text: "k"}
	m.Update(msg)

	if m.cursorPos != 1 {
		t.Errorf("k: cursorPos = %d, want 1", m.cursorPos)
	}

	// Also test 'up' key
	msg = tea.KeyPressMsg{Code: tea.KeyUp}
	m.Update(msg)

	if m.cursorPos != 0 {
		t.Errorf("up: cursorPos = %d, want 0", m.cursorPos)
	}
}

func TestImportScreen_Navigation_BoundsCheck(t *testing.T) {
	m := NewImportScreenModel()
	m.Init()
	m.phase = importPhasePreview
	m.diffKeys = []string{"GOPATH", "GOROOT"}
	m.cursorPos = 0

	// Try to go up from start
	msg := tea.KeyPressMsg{Code: 'k', Text: "k"}
	m.Update(msg)
	if m.cursorPos != 0 {
		t.Errorf("k at start: cursorPos = %d, should stay at 0", m.cursorPos)
	}

	// Go to end
	m.cursorPos = 1

	// Try to go down from end
	msg = tea.KeyPressMsg{Code: 'j', Text: "j"}
	m.Update(msg)
	if m.cursorPos != 1 {
		t.Errorf("j at end: cursorPos = %d, should stay at 1", m.cursorPos)
	}
}

func TestImportScreen_LoadSnapshotSuccess(t *testing.T) {
	m := NewImportScreenModel()
	m.Init()
	m.SetItems([]goenv.EnvVar{
		{Key: "GOPATH", Value: "/current"},
	})

	snapshot := persist.Snapshot{
		Name: "test-snapshot",
		Variables: map[string]string{
			"GOPATH": "/new",
			"GOROOT": "/usr/local/go",
		},
	}

	msg := loadSnapshotSuccessMsg{Snapshot: snapshot}
	m.Update(msg)

	if m.phase != importPhasePreview {
		t.Errorf("after load success: phase = %v, want importPhasePreview", m.phase)
	}
	if m.snapshot.Name != "test-snapshot" {
		t.Errorf("snapshot.Name = %q, want 'test-snapshot'", m.snapshot.Name)
	}
	// GOPATH should be modified, GOROOT should be added
	if len(m.diffKeys) < 1 {
		t.Error("diffKeys should contain at least one entry")
	}
}

func TestImportScreen_ApplySelected_NoSelection(t *testing.T) {
	m := NewImportScreenModel()
	m.Init()
	m.phase = importPhasePreview
	m.diffKeys = []string{"GOPATH"}
	m.selectedVars = map[string]bool{"GOPATH": false}
	m.snapshot = persist.Snapshot{Variables: map[string]string{"GOPATH": "/new"}}

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("Enter with no selection should return command")
	}

	resultMsg := cmd()
	changeMsg, ok := resultMsg.(changeScreenMsg)
	if !ok {
		t.Fatalf("expected changeScreenMsg (return to list), got %T", resultMsg)
	}
	if AppScreen(changeMsg) != AppScreenList {
		t.Errorf("with no selection, should return to list")
	}
}

func TestImportScreen_QFromPreview_ReturnsToList(t *testing.T) {
	m := NewImportScreenModel()
	m.Init()
	m.phase = importPhasePreview

	msg := tea.KeyPressMsg{Code: 'q', Text: "q"}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("q from preview should return command")
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

func TestImportScreen_QFromPath_DoesNotExit(t *testing.T) {
	m := NewImportScreenModel()
	m.Init()
	m.phase = importPhasePath

	msg := tea.KeyPressMsg{Code: 'q', Text: "q"}
	cmd := m.Update(msg)

	// q in path phase should NOT exit -- it falls through to textInput
	if cmd != nil {
		resultMsg := cmd()
		if _, ok := resultMsg.(changeScreenMsg); ok {
			t.Error("q in path phase should NOT return changeScreenMsg")
		}
	}

	// Phase should still be path
	if m.phase != importPhasePath {
		t.Errorf("phase = %v, should stay at importPhasePath", m.phase)
	}
}

func TestImportScreen_View_ShowsStepIndicator(t *testing.T) {
	m := NewImportScreenModel()
	m.Init()

	// Path phase should show Step 1/2
	view := m.View()
	if !strings.Contains(view, "Step 1/2") {
		t.Error("Import path phase should show 'Step 1/2'")
	}

	// Preview phase should show Step 2/2
	m.phase = importPhasePreview
	view = m.View()
	if !strings.Contains(view, "Step 2/2") {
		t.Error("Import preview phase should show 'Step 2/2'")
	}
}

func TestImportScreen_SetItems(t *testing.T) {
	m := NewImportScreenModel()
	items := []goenv.EnvVar{
		{Key: "GOPATH", Value: "/go"},
	}

	m.SetItems(items)

	if len(m.items) != 1 {
		t.Errorf("items length = %d, want 1", len(m.items))
	}
}
