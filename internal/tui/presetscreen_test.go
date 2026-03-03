package tui

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

// --- PresetScreenModel Tests ---

func TestPresetScreen_NewModel(t *testing.T) {
	m := NewPresetScreenModel()

	if m.phase != presetPhaseList {
		t.Errorf("NewModel: phase = %v, want presetPhaseList", m.phase)
	}
	if m.selectedIdx != 0 {
		t.Errorf("NewModel: selectedIdx = %d, want 0", m.selectedIdx)
	}
}

func TestPresetScreen_Init(t *testing.T) {
	m := NewPresetScreenModel()
	cmd := m.Init()

	if m.phase != presetPhaseList {
		t.Errorf("Init: phase = %v, want presetPhaseList", m.phase)
	}
	if m.err != nil {
		t.Errorf("Init: err = %v, want nil", m.err)
	}
	if cmd == nil {
		t.Error("Init should return load command")
	}
}

func TestPresetScreen_ListNavigation_Down(t *testing.T) {
	m := NewPresetScreenModel()
	m.presets = []persist.Preset{
		{Snapshot: persist.Snapshot{Name: "preset1"}},
		{Snapshot: persist.Snapshot{Name: "preset2"}},
		{Snapshot: persist.Snapshot{Name: "preset3"}},
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

func TestPresetScreen_ListNavigation_Up(t *testing.T) {
	m := NewPresetScreenModel()
	m.presets = []persist.Preset{
		{Snapshot: persist.Snapshot{Name: "preset1"}},
		{Snapshot: persist.Snapshot{Name: "preset2"}},
	}
	m.selectedIdx = 1

	msg := tea.KeyPressMsg{Code: 'k', Text: "k"}
	m.Update(msg)

	if m.selectedIdx != 0 {
		t.Errorf("k: selectedIdx = %d, want 0", m.selectedIdx)
	}
}

func TestPresetScreen_ListNavigation_BoundsCheck(t *testing.T) {
	m := NewPresetScreenModel()
	m.presets = []persist.Preset{
		{Snapshot: persist.Snapshot{Name: "preset1"}},
		{Snapshot: persist.Snapshot{Name: "preset2"}},
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

func TestPresetScreen_EnterShowsPreview(t *testing.T) {
	m := NewPresetScreenModel()
	m.presets = []persist.Preset{
		{Snapshot: persist.Snapshot{Name: "preset1", Variables: map[string]string{"GOPATH": "/go"}}},
	}
	m.items = []goenv.EnvVar{{Key: "GOPATH", Value: "/old"}}
	m.selectedIdx = 0

	// First Enter goes to preview phase
	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	m.Update(msg)

	if m.phase != presetPhasePreview {
		t.Fatalf("first Enter: phase = %v, want presetPhasePreview", m.phase)
	}
	if m.previewPreset.Name != "preset1" {
		t.Errorf("previewPreset.Name = %q, want 'preset1'", m.previewPreset.Name)
	}
	if len(m.previewDiff) != 1 {
		t.Fatalf("previewDiff length = %d, want 1", len(m.previewDiff))
	}

	// Second Enter applies
	cmd := m.Update(msg)
	if cmd == nil {
		t.Fatal("second Enter should return apply command")
	}
	resultMsg := cmd()
	applyMsg, ok := resultMsg.(presetApplyMsg)
	if !ok {
		t.Fatalf("expected presetApplyMsg, got %T", resultMsg)
	}
	if applyMsg.Preset.Name != "preset1" {
		t.Errorf("applyMsg.Preset.Name = %q, want 'preset1'", applyMsg.Preset.Name)
	}
}

func TestPresetScreen_EnterNoPresets_NoAction(t *testing.T) {
	m := NewPresetScreenModel()
	m.presets = []persist.Preset{}
	m.selectedIdx = 0

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	cmd := m.Update(msg)

	if cmd != nil {
		t.Error("Enter with no presets should return nil")
	}
}

func TestPresetScreen_NEntersCreatePhase(t *testing.T) {
	m := NewPresetScreenModel()
	m.phase = presetPhaseList

	msg := tea.KeyPressMsg{Code: 'n', Text: "n"}
	m.Update(msg)

	if m.phase != presetPhaseCreate {
		t.Errorf("n: phase = %v, want presetPhaseCreate", m.phase)
	}
}

func TestPresetScreen_DEntersDeleteConfirm(t *testing.T) {
	m := NewPresetScreenModel()
	m.presets = []persist.Preset{
		{Snapshot: persist.Snapshot{Name: "preset1"}},
	}
	m.phase = presetPhaseList

	msg := tea.KeyPressMsg{Code: 'd', Text: "d"}
	m.Update(msg)

	if m.phase != presetPhaseConfirmDelete {
		t.Errorf("d: phase = %v, want presetPhaseConfirmDelete", m.phase)
	}
}

func TestPresetScreen_DNoPresets_NoAction(t *testing.T) {
	m := NewPresetScreenModel()
	m.presets = []persist.Preset{}
	m.phase = presetPhaseList

	msg := tea.KeyPressMsg{Code: 'd', Text: "d"}
	m.Update(msg)

	if m.phase != presetPhaseList {
		t.Errorf("d with no presets: phase = %v, should stay at presetPhaseList", m.phase)
	}
}

func TestPresetScreen_CreatePhase_EnterCreates(t *testing.T) {
	m := NewPresetScreenModel()
	m.phase = presetPhaseCreate
	m.textInput.SetValue("new-preset")
	m.SetItems([]goenv.EnvVar{{Key: "GOPATH", Value: "/go"}})

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("Enter in create phase should return command")
	}
	if m.presetName != "new-preset" {
		t.Errorf("presetName = %q, want 'new-preset'", m.presetName)
	}
}

func TestPresetScreen_CreatePhase_EmptyName_NoAction(t *testing.T) {
	m := NewPresetScreenModel()
	m.phase = presetPhaseCreate
	m.textInput.SetValue("")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	cmd := m.Update(msg)

	if cmd != nil {
		t.Error("Enter with empty name should return nil")
	}
}

func TestPresetScreen_CreatePhase_EscReturnsToList(t *testing.T) {
	m := NewPresetScreenModel()
	m.phase = presetPhaseCreate

	msg := tea.KeyPressMsg{Code: tea.KeyEscape}
	m.Update(msg)

	if m.phase != presetPhaseList {
		t.Errorf("Esc: phase = %v, want presetPhaseList", m.phase)
	}
}

func TestPresetScreen_DeleteConfirm_YConfirms(t *testing.T) {
	m := NewPresetScreenModel()
	m.presets = []persist.Preset{
		{Snapshot: persist.Snapshot{Name: "preset1"}},
	}
	m.selectedIdx = 0
	m.phase = presetPhaseConfirmDelete

	msg := tea.KeyPressMsg{Code: 'y', Text: "y"}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("y should return delete command")
	}
}

func TestPresetScreen_DeleteConfirm_NCancels(t *testing.T) {
	m := NewPresetScreenModel()
	m.phase = presetPhaseConfirmDelete

	msg := tea.KeyPressMsg{Code: 'n', Text: "n"}
	m.Update(msg)

	if m.phase != presetPhaseList {
		t.Errorf("n: phase = %v, want presetPhaseList", m.phase)
	}
}

func TestPresetScreen_DeleteConfirm_EscCancels(t *testing.T) {
	m := NewPresetScreenModel()
	m.phase = presetPhaseConfirmDelete

	msg := tea.KeyPressMsg{Code: tea.KeyEscape}
	m.Update(msg)

	if m.phase != presetPhaseList {
		t.Errorf("Esc: phase = %v, want presetPhaseList", m.phase)
	}
}

func TestPresetScreen_ListPhase_EscReturnsToMainList(t *testing.T) {
	m := NewPresetScreenModel()
	m.phase = presetPhaseList

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

func TestPresetScreen_PresetLoadSuccess(t *testing.T) {
	m := NewPresetScreenModel()
	m.selectedIdx = 5 // Higher than preset count

	presets := []persist.Preset{
		{Snapshot: persist.Snapshot{Name: "preset1", CreatedAt: time.Now()}},
		{Snapshot: persist.Snapshot{Name: "preset2", CreatedAt: time.Now()}},
	}

	msg := presetLoadSuccessMsg{Presets: presets}
	m.Update(msg)

	if len(m.presets) != 2 {
		t.Errorf("presets length = %d, want 2", len(m.presets))
	}
	// selectedIdx should be clamped to valid range
	if m.selectedIdx > 1 {
		t.Errorf("selectedIdx = %d, should be clamped to <= 1", m.selectedIdx)
	}
}

func TestPresetScreen_SetItems(t *testing.T) {
	m := NewPresetScreenModel()
	items := []goenv.EnvVar{
		{Key: "GOPATH", Value: "/go"},
	}

	m.SetItems(items)

	if len(m.items) != 1 {
		t.Errorf("items length = %d, want 1", len(m.items))
	}
}
