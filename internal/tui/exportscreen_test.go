package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
)

// --- ExportScreenModel Tests ---

func TestExportScreen_Init(t *testing.T) {
	m := NewExportScreenModel()
	m.Init()

	if m.phase != exportPhaseName {
		t.Errorf("Init: phase = %v, want exportPhaseName", m.phase)
	}
	if m.snapshotName != "" {
		t.Errorf("Init: snapshotName = %q, want empty", m.snapshotName)
	}
	if m.textInput.Value() != "" {
		t.Errorf("Init: textInput value = %q, want empty", m.textInput.Value())
	}
}

func TestExportScreen_EnterOnEmptyName_NoAction(t *testing.T) {
	m := NewExportScreenModel()
	m.Init()
	m.textInput.SetValue("") // empty name

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	cmd := m.Update(msg)

	if cmd != nil {
		t.Error("Enter on empty name should return nil command")
	}
	if m.phase != exportPhaseName {
		t.Errorf("phase = %v, should stay at exportPhaseName", m.phase)
	}
}

func TestExportScreen_EnterOnValidName_MovesToDescription(t *testing.T) {
	m := NewExportScreenModel()
	m.Init()
	m.textInput.SetValue("my-snapshot")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	cmd := m.Update(msg)

	if cmd != nil {
		t.Error("should return nil command (not export yet)")
	}
	if m.phase != exportPhaseDescription {
		t.Errorf("phase = %v, want exportPhaseDescription", m.phase)
	}
	if m.snapshotName != "my-snapshot" {
		t.Errorf("snapshotName = %q, want 'my-snapshot'", m.snapshotName)
	}
	if m.textInput.Value() != "" {
		t.Errorf("textInput should be cleared for description, got %q", m.textInput.Value())
	}
}

func TestExportScreen_EnterOnDescription_TriggersExport(t *testing.T) {
	m := NewExportScreenModel()
	m.Init()
	m.textInput.SetValue("my-snapshot")

	// Move to description phase
	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	m.Update(msg)

	// Set description and trigger export
	m.textInput.SetValue("A test snapshot")
	m.SetItems([]goenv.EnvVar{{Key: "GOPATH", Value: "/go"}})

	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("Enter on description should return export command")
	}

	resultMsg := cmd()
	exportMsg, ok := resultMsg.(exportSnapshotMsg)
	if !ok {
		t.Fatalf("expected exportSnapshotMsg, got %T", resultMsg)
	}
	if exportMsg.Name != "my-snapshot" {
		t.Errorf("exportMsg.Name = %q, want 'my-snapshot'", exportMsg.Name)
	}
	if exportMsg.Description != "A test snapshot" {
		t.Errorf("exportMsg.Description = %q, want 'A test snapshot'", exportMsg.Description)
	}
	if len(exportMsg.Items) != 1 {
		t.Errorf("exportMsg.Items length = %d, want 1", len(exportMsg.Items))
	}
}

func TestExportScreen_EscReturnsToList(t *testing.T) {
	m := NewExportScreenModel()
	m.Init()

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

func TestExportScreen_SetItems(t *testing.T) {
	m := NewExportScreenModel()
	items := []goenv.EnvVar{
		{Key: "GOPATH", Value: "/go"},
		{Key: "GOROOT", Value: "/usr/local/go"},
	}

	m.SetItems(items)

	if len(m.items) != 2 {
		t.Errorf("items length = %d, want 2", len(m.items))
	}
}

func TestExportScreen_View_ShowsStepIndicator(t *testing.T) {
	m := NewExportScreenModel()
	m.Init()

	// Name phase should show Step 1/2
	view := m.View()
	if !strings.Contains(view, "Step 1/2") {
		t.Error("Export name phase should show 'Step 1/2'")
	}

	// Move to description phase
	m.textInput.SetValue("test")
	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	m.Update(msg)

	// Description phase should show Step 2/2
	view = m.View()
	if !strings.Contains(view, "Step 2/2") {
		t.Error("Export description phase should show 'Step 2/2'")
	}
}

func TestExportScreen_View_NamePhase(t *testing.T) {
	m := NewExportScreenModel()
	m.Init()
	m.SetItems([]goenv.EnvVar{{Key: "GOPATH", Value: "/go"}})

	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
	// View should mention "name"
	if !strings.Contains(view, "name") && !strings.Contains(view, "Name") {
		t.Error("View in name phase should mention 'name'")
	}
}

func TestExportScreen_View_DescriptionPhase(t *testing.T) {
	m := NewExportScreenModel()
	m.Init()
	m.textInput.SetValue("test")

	// Move to description phase
	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	m.Update(msg)

	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
	// View should show snapshot name
	if !strings.Contains(view, "test") {
		t.Error("View in description phase should show snapshot name")
	}
}
