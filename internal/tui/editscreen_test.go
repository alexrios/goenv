package tui

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
)

// stripANSI removes ANSI escape codes from a string.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string { return ansiRe.ReplaceAllString(s, "") }

func TestNewEditScreenModel_Defaults(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{Major: 1, Minor: 21, Patch: 0})
	if m.editingKey != "" {
		t.Errorf("editingKey = %q, want empty", m.editingKey)
	}
	if m.suggestionIdx != -1 {
		t.Errorf("suggestionIdx = %d, want -1", m.suggestionIdx)
	}
	if !m.showSuggestions {
		t.Error("showSuggestions should be true by default")
	}
}

func TestEditScreen_UpdateSuggestions_GOOS(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{Major: 1, Minor: 21, Patch: 0})
	m.editingKey = "GOOS"
	m.textInput.SetValue("lin")
	m.updateSuggestions()

	if len(m.suggestions) == 0 {
		t.Fatal("expected suggestions for GOOS with prefix 'lin'")
	}
	found := false
	for _, s := range m.suggestions {
		if s == "linux" {
			found = true
		}
		if !strings.HasPrefix(s, "lin") {
			t.Errorf("suggestion %q should start with 'lin'", s)
		}
	}
	if !found {
		t.Error("expected 'linux' in suggestions")
	}
}

func TestEditScreen_UpdateSuggestions_NoMatch(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{Major: 1, Minor: 21, Patch: 0})
	m.editingKey = "GOOS"
	m.textInput.SetValue("zzz")
	m.updateSuggestions()

	if len(m.suggestions) != 0 {
		t.Errorf("expected no suggestions for 'zzz', got %d", len(m.suggestions))
	}
}

func TestEditScreen_TabCyclesSuggestions(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{Major: 1, Minor: 21, Patch: 0})
	m.editingKey = "CGO_ENABLED"
	m.Init()

	// Should have suggestions (0 and 1)
	if len(m.suggestions) == 0 {
		t.Fatal("expected suggestions for CGO_ENABLED")
	}

	// Press tab -- should select first suggestion
	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.suggestionIdx != 0 {
		t.Errorf("after tab: suggestionIdx = %d, want 0", m.suggestionIdx)
	}

	// Press tab again -- should advance
	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.suggestionIdx != 1 {
		t.Errorf("after 2nd tab: suggestionIdx = %d, want 1", m.suggestionIdx)
	}
}

func TestEditScreen_ShiftTabCyclesBackward(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{Major: 1, Minor: 21, Patch: 0})
	m.editingKey = "CGO_ENABLED"
	m.Init()

	if len(m.suggestions) == 0 {
		t.Fatal("expected suggestions for CGO_ENABLED")
	}

	// Shift+Tab from -1 should wrap to last
	m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if m.suggestionIdx != len(m.suggestions)-1 {
		t.Errorf("shift+tab from start: suggestionIdx = %d, want %d", m.suggestionIdx, len(m.suggestions)-1)
	}
}

func TestEditScreen_CtrlR_ResetsValue(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{Major: 1, Minor: 21, Patch: 0})
	m.editingKey = "GOOS"
	m.originalValue = "linux"
	m.textInput.SetValue("darwin")
	m.Init()

	// Ctrl+R should reset to original
	m.Update(tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl})

	if m.textInput.Value() != "linux" {
		t.Errorf("after Ctrl+R: value = %q, want 'linux'", m.textInput.Value())
	}
}

func TestEditScreen_EnterWithSuggestion_AppliesSuggestion(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{Major: 1, Minor: 21, Patch: 0})
	m.editingKey = "CGO_ENABLED"
	m.Init()

	if len(m.suggestions) == 0 {
		t.Fatal("expected suggestions")
	}

	// Select first suggestion via tab
	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	firstSuggestion := m.suggestions[0]

	// Press enter -- should apply suggestion, not save
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.textInput.Value() != firstSuggestion {
		t.Errorf("after enter with suggestion: value = %q, want %q", m.textInput.Value(), firstSuggestion)
	}
	// Should reset suggestion index
	if m.suggestionIdx != -1 {
		t.Errorf("after applying suggestion: suggestionIdx = %d, want -1", m.suggestionIdx)
	}
}

func TestEditScreen_RenderDiff_NoChange(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{})
	result := m.renderDiff("same", "same")
	if !strings.Contains(result, "Current:") {
		t.Error("expected 'Current:' when values are the same")
	}
}

func TestEditScreen_RenderDiff_WithChange(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{})
	result := stripANSI(m.renderDiff("old", "new"))
	if !strings.Contains(result, "Old:") || !strings.Contains(result, "New:") {
		t.Errorf("diff should contain Old:/New: labels, got %q", result)
	}
}

func TestEditScreen_RenderSuggestions_Empty(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{})
	m.suggestions = nil
	result := m.renderSuggestions()
	if result != "" {
		t.Errorf("expected empty string for no suggestions, got %q", result)
	}
}

func TestEditScreen_RenderSuggestions_HiddenWhenDisabled(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{})
	m.suggestions = []string{"linux", "darwin"}
	m.showSuggestions = false
	result := m.renderSuggestions()
	if result != "" {
		t.Errorf("expected empty string when suggestions disabled, got %q", result)
	}
}

func TestEditScreen_RenderCSVBreakdown_NonCSVVar(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{})
	m.editingKey = "GOPATH"
	m.textInput.SetValue("a,b,c")
	result := m.renderCSVBreakdown()
	if result != "" {
		t.Errorf("expected empty for non-CSV var, got %q", result)
	}
}

func TestEditScreen_RenderCSVBreakdown_CSVVar(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{})
	m.editingKey = "GOEXPERIMENT"
	m.textInput.SetValue("rangefunc,arenas")
	result := m.renderCSVBreakdown()
	if result == "" {
		t.Fatal("expected CSV breakdown for GOEXPERIMENT")
	}
	if !strings.Contains(result, "rangefunc") || !strings.Contains(result, "arenas") {
		t.Error("CSV breakdown should contain both items")
	}
}

func TestEditScreen_RenderCSVBreakdown_SingleItem(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{})
	m.editingKey = "GOEXPERIMENT"
	m.textInput.SetValue("rangefunc")
	result := m.renderCSVBreakdown()
	if result != "" {
		t.Error("expected empty for single CSV item (no breakdown needed)")
	}
}

func TestEditScreen_View_ContainsEditingKey(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{Major: 1, Minor: 21, Patch: 0})
	m.editingKey = "GOPATH"
	m.originalValue = "/go"
	m.textInput.SetValue("/go")
	m.Init()

	view := m.View()
	if !strings.Contains(view, "GOPATH") {
		t.Error("view should contain the editing key name")
	}
}

// --- Read-Only Mode Tests ---

func TestEditScreen_ReadOnly_InitDoesNotFocus(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{Major: 1, Minor: 21, Patch: 0})
	m.editingKey = "GOVERSION"
	m.originalValue = "go1.21.0"
	m.readOnly = true
	cmd := m.Init()

	if cmd != nil {
		t.Error("read-only Init should return nil (no blink)")
	}
}

func TestEditScreen_ReadOnly_EscReturnsToList(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{Major: 1, Minor: 21, Patch: 0})
	m.editingKey = "GOVERSION"
	m.originalValue = "go1.21.0"
	m.readOnly = true
	m.Init()

	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected command from Esc in read-only mode")
	}
	msg := cmd()
	if csm, ok := msg.(changeScreenMsg); !ok || AppScreen(csm) != AppScreenList {
		t.Errorf("expected changeScreenMsg(AppScreenList), got %T", msg)
	}
}

func TestEditScreen_ReadOnly_IgnoresEnterAndTyping(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{Major: 1, Minor: 21, Patch: 0})
	m.editingKey = "GOVERSION"
	m.originalValue = "go1.21.0"
	m.readOnly = true
	m.Init()

	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil {
		t.Error("read-only mode should ignore Enter key")
	}

	cmd = m.Update(tea.KeyPressMsg{Code: 'a'})
	if cmd != nil {
		t.Error("read-only mode should ignore character input")
	}
}

func TestEditScreen_ReadOnly_CopyValueKey(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{Major: 1, Minor: 21, Patch: 0})
	m.editingKey = "GOVERSION"
	m.originalValue = "go1.21.0"
	m.readOnly = true
	m.Init()

	cmd := m.Update(tea.KeyPressMsg{Code: 'y'})
	if cmd == nil {
		t.Fatal("expected command from y in read-only mode")
	}
	msg := cmd()
	if cvm, ok := msg.(copyValueMsg); !ok || cvm.Key != "GOVERSION" {
		t.Errorf("expected copyValueMsg for GOVERSION, got %T", msg)
	}
}

func TestEditScreen_ReadOnly_CopyKeyValueKey(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{Major: 1, Minor: 21, Patch: 0})
	m.editingKey = "GOVERSION"
	m.originalValue = "go1.21.0"
	m.readOnly = true
	m.Init()

	cmd := m.Update(tea.KeyPressMsg{Code: 'Y'})
	if cmd == nil {
		t.Fatal("expected command from Y in read-only mode")
	}
	msg := cmd()
	if ckvm, ok := msg.(copyKeyValueMsg); !ok || ckvm.Key != "GOVERSION" {
		t.Errorf("expected copyKeyValueMsg for GOVERSION, got %T", msg)
	}
}

func TestEditScreen_ReadOnly_ViewShowsReadOnlyLabel(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{Major: 1, Minor: 21, Patch: 0})
	m.editingKey = "GOVERSION"
	m.originalValue = "go1.21.0"
	m.readOnly = true
	m.Init()

	view := m.View()
	if !strings.Contains(view, "read-only") {
		t.Error("read-only view should contain 'read-only' label")
	}
	if !strings.Contains(view, "GOVERSION") {
		t.Error("read-only view should contain the variable name")
	}
	if !strings.Contains(view, "go1.21.0") {
		t.Error("read-only view should contain the value")
	}
}
