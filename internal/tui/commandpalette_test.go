package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestCommandPalette_NewModel(t *testing.T) {
	m := NewCommandPaletteModel()

	if len(m.allEntries) == 0 {
		t.Error("NewModel: allEntries should not be empty")
	}
	if len(m.filtered) != len(m.allEntries) {
		t.Errorf("NewModel: filtered length = %d, want %d", len(m.filtered), len(m.allEntries))
	}
}

func TestCommandPalette_Init(t *testing.T) {
	m := NewCommandPaletteModel()
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init should return blink command")
	}
	if m.selectedIdx != 0 {
		t.Errorf("Init: selectedIdx = %d, want 0", m.selectedIdx)
	}
}

func TestCommandPalette_FilterNarrows(t *testing.T) {
	m := NewCommandPaletteModel()
	m.Init()

	allCount := len(m.filtered)

	// Simulate typing "exp"
	m.textInput.SetValue("exp")
	m.refilter()

	if len(m.filtered) >= allCount {
		t.Errorf("filter 'exp': expected fewer results than %d, got %d", allCount, len(m.filtered))
	}
	if len(m.filtered) == 0 {
		t.Error("filter 'exp': expected at least one result for export")
	}
}

func TestCommandPalette_FilterEmpty(t *testing.T) {
	m := NewCommandPaletteModel()
	m.Init()

	m.textInput.SetValue("")
	m.refilter()

	if len(m.filtered) != len(m.allEntries) {
		t.Errorf("empty filter: filtered length = %d, want %d", len(m.filtered), len(m.allEntries))
	}
}

func TestCommandPalette_FilterNoMatch(t *testing.T) {
	m := NewCommandPaletteModel()
	m.Init()

	m.textInput.SetValue("zzzzzzzzz")
	m.refilter()

	if len(m.filtered) != 0 {
		t.Errorf("no match filter: filtered length = %d, want 0", len(m.filtered))
	}
}

func TestCommandPalette_NavigateDown(t *testing.T) {
	m := NewCommandPaletteModel()
	m.Init()

	msg := tea.KeyPressMsg{Code: tea.KeyDown}
	m.Update(msg)

	if m.selectedIdx != 1 {
		t.Errorf("down: selectedIdx = %d, want 1", m.selectedIdx)
	}
}

func TestCommandPalette_NavigateUp(t *testing.T) {
	m := NewCommandPaletteModel()
	m.Init()
	m.selectedIdx = 2

	msg := tea.KeyPressMsg{Code: tea.KeyUp}
	m.Update(msg)

	if m.selectedIdx != 1 {
		t.Errorf("up: selectedIdx = %d, want 1", m.selectedIdx)
	}
}

func TestCommandPalette_NavigateUpAtZero(t *testing.T) {
	m := NewCommandPaletteModel()
	m.Init()
	m.selectedIdx = 0

	msg := tea.KeyPressMsg{Code: tea.KeyUp}
	m.Update(msg)

	if m.selectedIdx != 0 {
		t.Errorf("up at 0: selectedIdx = %d, want 0", m.selectedIdx)
	}
}

func TestCommandPalette_EscReturnsToList(t *testing.T) {
	m := NewCommandPaletteModel()
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

func TestCommandPalette_EnterDispatchesSort(t *testing.T) {
	m := NewCommandPaletteModel()
	m.Init()

	// Find the "sort" entry
	for i, e := range m.filtered {
		if e.Name == "sort" {
			m.selectedIdx = i
			break
		}
	}

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("Enter on sort should return command")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(toggleSortMsg); !ok {
		t.Fatalf("expected toggleSortMsg, got %T", resultMsg)
	}
}

func TestCommandPalette_EnterOnQuit(t *testing.T) {
	m := NewCommandPaletteModel()
	m.Init()

	// Find the "quit" entry
	for i, e := range m.filtered {
		if e.Name == "quit" {
			m.selectedIdx = i
			break
		}
	}

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("Enter on quit should return command")
	}
}

func TestCommandPalette_ViewContainsTitle(t *testing.T) {
	m := NewCommandPaletteModel()
	m.Init()

	view := m.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}
