package tui

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

// TestMain runs before all tests to ensure clean state.
func TestMain(m *testing.M) {
	// Clear any persisted history before running tests
	_ = persist.ClearHistory()
	os.Exit(m.Run())
}

// --- envItem Tests ---

func TestEnvItem_Title(t *testing.T) {
	e := envItem{goenv.EnvVar{Key: "GOPATH", Value: "/home/user/go"}}
	if got := e.Title(); got != "GOPATH" {
		t.Errorf("Title() = %q, want %q", got, "GOPATH")
	}
}

func TestEnvItem_Description(t *testing.T) {
	e := envItem{goenv.EnvVar{Key: "GOPATH", Value: "/home/user/go"}}
	if got := e.Description(); got != "/home/user/go" {
		t.Errorf("Description() = %q, want %q", got, "/home/user/go")
	}
}

func TestEnvItem_FilterValue(t *testing.T) {
	e := envItem{goenv.EnvVar{Key: "GOPATH", Value: "/home/user/go"}}
	got := e.FilterValue()
	// FilterValue should include key, value, and description
	if !strings.Contains(got, "GOPATH") {
		t.Errorf("FilterValue() should contain key, got %q", got)
	}
	if !strings.Contains(got, "/home/user/go") {
		t.Errorf("FilterValue() should contain value, got %q", got)
	}
	desc := goenv.GetEnvVarDescription("GOPATH")
	if desc != "" && !strings.Contains(got, desc) {
		t.Errorf("FilterValue() should contain description %q, got %q", desc, got)
	}
}

func TestEnvItem_FilterValue_NoDescription(t *testing.T) {
	e := envItem{goenv.EnvVar{Key: "UNKNOWN_VAR", Value: "some_value"}}
	got := e.FilterValue()
	want := "UNKNOWN_VAR some_value"
	if got != want {
		t.Errorf("FilterValue() = %q, want %q", got, want)
	}
}

// --- Sort Tests ---

func TestSortItems_Alpha(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "GOROOT", Value: "/usr/local/go"},
		{Key: "GOPATH", Value: "/home/user/go"},
		{Key: "GOBIN", Value: "/home/user/go/bin"},
	}

	m := mainModel{items: items, sortMode: persist.SortAlpha, favorites: make(map[string]bool)}
	m.sortItems()

	expected := []string{"GOBIN", "GOPATH", "GOROOT"}
	for i, want := range expected {
		got := m.items[i].Key
		if got != want {
			t.Errorf("items[%d].Key = %q, want %q", i, got, want)
		}
	}
}

func TestSortItems_ModifiedFirst(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "GOROOT", Value: "/usr/local/go", Changed: false},
		{Key: "GOPATH", Value: "/home/user/go", Changed: true},
		{Key: "GOBIN", Value: "/home/user/go/bin", Changed: false},
		{Key: "GOPROXY", Value: "direct", Changed: true},
	}

	m := mainModel{items: items, sortMode: persist.SortModifiedFirst, favorites: make(map[string]bool)}
	m.sortItems()

	// Modified items should come first (GOPATH, GOPROXY), then unmodified (GOBIN, GOROOT)
	// Within each group, alphabetical order
	expected := []string{"GOPATH", "GOPROXY", "GOBIN", "GOROOT"}
	for i, want := range expected {
		got := m.items[i].Key
		if got != want {
			t.Errorf("items[%d].Key = %q, want %q", i, got, want)
		}
	}
}

// --- ListScreenModel Update Tests ---

func TestListScreen_EnterKey(t *testing.T) {
	items := []list.Item{envItem{goenv.EnvVar{Key: "GOPATH", Value: "/go"}}}
	m := NewListScreenModel(items)

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(startEditMsg); !ok {
		t.Errorf("expected startEditMsg, got %T", resultMsg)
	}
}

func TestListScreen_SortToggle(t *testing.T) {
	items := []list.Item{envItem{goenv.EnvVar{Key: "GOPATH", Value: "/go"}}}
	m := NewListScreenModel(items)

	msg := tea.KeyPressMsg{Code: 's', Text: "s"}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(toggleSortMsg); !ok {
		t.Errorf("expected toggleSortMsg, got %T", resultMsg)
	}
}

func TestListScreen_ReloadKey(t *testing.T) {
	items := []list.Item{envItem{goenv.EnvVar{Key: "GOPATH", Value: "/go"}}}
	m := NewListScreenModel(items)

	msg := tea.KeyPressMsg{Code: 'r', Text: "r"}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(reloadEnvMsg); !ok {
		t.Errorf("expected reloadEnvMsg, got %T", resultMsg)
	}
}

func TestListScreen_CopyValueKey(t *testing.T) {
	items := []list.Item{envItem{goenv.EnvVar{Key: "GOPATH", Value: "/go"}}}
	m := NewListScreenModel(items)

	msg := tea.KeyPressMsg{Code: 'y', Text: "y"}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	resultMsg := cmd()
	copyMsg, ok := resultMsg.(copyValueMsg)
	if !ok {
		t.Errorf("expected copyValueMsg, got %T", resultMsg)
	}
	if copyMsg.Key != "GOPATH" || copyMsg.Value != "/go" {
		t.Errorf("copyValueMsg = {%q, %q}, want {GOPATH, /go}", copyMsg.Key, copyMsg.Value)
	}
}

func TestListScreen_ShowHelpKey(t *testing.T) {
	items := []list.Item{envItem{goenv.EnvVar{Key: "GOPATH", Value: "/go"}}}
	m := NewListScreenModel(items)

	msg := tea.KeyPressMsg{Code: '?', Text: "?"}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(showHelpMsg); !ok {
		t.Errorf("expected showHelpMsg, got %T", resultMsg)
	}
}

func TestListScreen_UndoKey(t *testing.T) {
	items := []list.Item{envItem{goenv.EnvVar{Key: "GOPATH", Value: "/go"}}}
	m := NewListScreenModel(items)

	msg := tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(undoMsg); !ok {
		t.Errorf("expected undoMsg, got %T", resultMsg)
	}
}

func TestListScreen_RedoKey(t *testing.T) {
	items := []list.Item{envItem{goenv.EnvVar{Key: "GOPATH", Value: "/go"}}}
	m := NewListScreenModel(items)

	msg := tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(redoMsg); !ok {
		t.Errorf("expected redoMsg, got %T", resultMsg)
	}
}

// --- EditScreenModel Update Tests ---

func TestEditScreen_EnterSaves(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{})
	m.editingKey = "GOPATH"
	m.originalValue = "/old/path"
	m.textInput.SetValue("/new/path")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	resultMsg := cmd()
	saveReq, ok := resultMsg.(saveEnvRequestMsg)
	if !ok {
		t.Errorf("expected saveEnvRequestMsg, got %T", resultMsg)
	}
	if saveReq.Env.Key != "GOPATH" || saveReq.Env.Value != "/new/path" {
		t.Errorf("saveEnvRequestMsg.Env = {%q, %q}, want {GOPATH, /new/path}", saveReq.Env.Key, saveReq.Env.Value)
	}
	if saveReq.OriginalValue != "/old/path" {
		t.Errorf("saveEnvRequestMsg.OriginalValue = %q, want /old/path", saveReq.OriginalValue)
	}
}

func TestEditScreen_EscCancels(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{})
	m.editingKey = "GOPATH"

	msg := tea.KeyPressMsg{Code: tea.KeyEscape}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	resultMsg := cmd()
	changeMsg, ok := resultMsg.(changeScreenMsg)
	if !ok {
		t.Errorf("expected changeScreenMsg, got %T", resultMsg)
	}
	if AppScreen(changeMsg) != AppScreenList {
		t.Errorf("changeScreenMsg = %v, want AppScreenList", changeMsg)
	}
}

func TestEditScreen_CtrlR_ResetsToOriginal(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{})
	m.editingKey = "GOPATH"
	m.originalValue = "/orig"
	m.textInput.SetValue("/changed")
	m.validationError = "some error"

	msg := tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl}
	cmd := m.Update(msg)

	if cmd != nil {
		t.Error("Ctrl+R should return nil command (stay in edit screen)")
	}
	if m.textInput.Value() != "/orig" {
		t.Errorf("textInput value = %q, want '/orig'", m.textInput.Value())
	}
	if m.validationError != "" {
		t.Errorf("validationError = %q, want empty", m.validationError)
	}
	if m.suggestionIdx != -1 {
		t.Errorf("suggestionIdx = %d, want -1", m.suggestionIdx)
	}
}

// --- Undo/Redo History Tests ---

func TestUndoRedo_RecordsEdit(t *testing.T) {
	_ = persist.ClearHistory() // Ensure clean state
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/old"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	// Process envSetSuccessMsg with a new value (simulating successful save)
	msg := envSetSuccessMsg{
		Env:           goenv.EnvVar{Key: "GOPATH", Value: "/new", Changed: true},
		OriginalValue: "/old",
		IsUndoRedo:    false,
	}
	newModel, _ := m.Update(msg)
	m = newModel.(mainModel)

	if len(m.editHistory) != 1 {
		t.Fatalf("editHistory length = %d, want 1", len(m.editHistory))
	}

	record := m.editHistory[0]
	if record.Key != "GOPATH" {
		t.Errorf("record.Key = %q, want GOPATH", record.Key)
	}
	if record.OldValue != "/old" {
		t.Errorf("record.OldValue = %q, want /old", record.OldValue)
	}
	if record.NewValue != "/new" {
		t.Errorf("record.NewValue = %q, want /new", record.NewValue)
	}
	if m.historyIdx != 1 {
		t.Errorf("historyIdx = %d, want 1", m.historyIdx)
	}
}

func TestUndoRedo_NoRecordWhenValueUnchanged(t *testing.T) {
	_ = persist.ClearHistory() // Ensure clean state
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/same"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	// Process envSetSuccessMsg with the same value
	msg := envSetSuccessMsg{
		Env:           goenv.EnvVar{Key: "GOPATH", Value: "/same", Changed: true},
		OriginalValue: "/same",
		IsUndoRedo:    false,
	}
	newModel, _ := m.Update(msg)
	m = newModel.(mainModel)

	if len(m.editHistory) != 0 {
		t.Errorf("editHistory length = %d, want 0 (no change)", len(m.editHistory))
	}
}

func TestUndoRedo_UndoRestores(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/new"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	// Setup history manually
	m.editHistory = []persist.EditRecord{
		{Key: "GOPATH", OldValue: "/old", NewValue: "/new"},
	}
	m.historyIdx = 1

	// Process undoMsg
	msg := undoMsg{}
	newModel, _ := m.Update(msg)
	m = newModel.(mainModel)

	if m.historyIdx != 0 {
		t.Errorf("historyIdx = %d, want 0", m.historyIdx)
	}

	// Check that the item was restored
	for _, item := range m.items {
		if item.Key == "GOPATH" {
			if item.Value != "/old" {
				t.Errorf("GOPATH value = %q, want /old", item.Value)
			}
			return
		}
	}
	t.Error("GOPATH not found in items")
}

func TestUndoRedo_RedoReapplies(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/old"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	// Setup history manually (already undone)
	m.editHistory = []persist.EditRecord{
		{Key: "GOPATH", OldValue: "/old", NewValue: "/new"},
	}
	m.historyIdx = 0

	// Process redoMsg
	msg := redoMsg{}
	newModel, _ := m.Update(msg)
	m = newModel.(mainModel)

	if m.historyIdx != 1 {
		t.Errorf("historyIdx = %d, want 1", m.historyIdx)
	}

	// Check that the item was reapplied
	for _, item := range m.items {
		if item.Key == "GOPATH" {
			if item.Value != "/new" {
				t.Errorf("GOPATH value = %q, want /new", item.Value)
			}
			return
		}
	}
	t.Error("GOPATH not found in items")
}

func TestUndoRedo_NothingToUndo(t *testing.T) {
	_ = persist.ClearHistory() // Ensure clean state
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	// historyIdx is 0, no history
	msg := undoMsg{}
	newModel, _ := m.Update(msg)
	m = newModel.(mainModel)

	if m.successMessage != "Nothing to undo" {
		t.Errorf("successMessage = %q, want 'Nothing to undo'", m.successMessage)
	}
}

func TestUndoRedo_NothingToRedo(t *testing.T) {
	_ = persist.ClearHistory() // Ensure clean state
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	// historyIdx equals history length, nothing to redo
	msg := redoMsg{}
	newModel, _ := m.Update(msg)
	m = newModel.(mainModel)

	if m.successMessage != "Nothing to redo" {
		t.Errorf("successMessage = %q, want 'Nothing to redo'", m.successMessage)
	}
}

func TestUndoRedo_BlockedWhileSaving(t *testing.T) {
	_ = persist.ClearHistory()
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	// Add a history entry so undo/redo would normally work
	m.editHistory = []persist.EditRecord{{Key: "GOPATH", OldValue: "/old", NewValue: "/go"}}
	m.historyIdx = 1

	// Simulate saving in progress
	m.isSaving = true

	// Undo should be blocked
	newModel, _ := m.Update(undoMsg{})
	m = newModel.(mainModel)
	if m.successMessage != "Please wait, saving in progress..." {
		t.Errorf("undo while saving: successMessage = %q, want saving message", m.successMessage)
	}
	if m.historyIdx != 1 {
		t.Errorf("historyIdx should not change during save, got %d", m.historyIdx)
	}

	// Redo should also be blocked
	m.historyIdx = 0 // reset to allow redo
	newModel, _ = m.Update(redoMsg{})
	m = newModel.(mainModel)
	if m.successMessage != "Please wait, saving in progress..." {
		t.Errorf("redo while saving: successMessage = %q, want saving message", m.successMessage)
	}
	if m.historyIdx != 0 {
		t.Errorf("historyIdx should not change during save, got %d", m.historyIdx)
	}
}

// --- Screen Transition Tests ---

func TestMainModel_ToggleSort(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "GOROOT", Value: "/usr/local/go"},
		{Key: "GOPATH", Value: "/home/user/go"},
	}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	if m.sortMode != persist.SortAlpha {
		t.Errorf("initial sortMode = %v, want SortAlpha", m.sortMode)
	}

	// Toggle sort: Alpha -> ModifiedFirst
	msg := toggleSortMsg{}
	newModel, _ := m.Update(msg)
	m = newModel.(mainModel)

	if m.sortMode != persist.SortModifiedFirst {
		t.Errorf("sortMode after first toggle = %v, want SortModifiedFirst", m.sortMode)
	}

	// Toggle again: ModifiedFirst -> Category
	newModel, _ = m.Update(msg)
	m = newModel.(mainModel)

	if m.sortMode != persist.SortCategory {
		t.Errorf("sortMode after second toggle = %v, want SortCategory", m.sortMode)
	}

	// Toggle again: Category -> Alpha (full cycle)
	newModel, _ = m.Update(msg)
	m = newModel.(mainModel)

	if m.sortMode != persist.SortAlpha {
		t.Errorf("sortMode after third toggle = %v, want SortAlpha", m.sortMode)
	}
}

func TestMainModel_StartEdit(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	msg := startEditMsg{Key: "GOPATH", Value: "/go"}
	newModel, _ := m.Update(msg)
	m = newModel.(mainModel)

	if m.currentScreen != AppScreenEdit {
		t.Errorf("currentScreen = %v, want AppScreenEdit", m.currentScreen)
	}
	if m.editScreen.editingKey != "GOPATH" {
		t.Errorf("editingKey = %q, want GOPATH", m.editScreen.editingKey)
	}
	if m.editScreen.originalValue != "/go" {
		t.Errorf("originalValue = %q, want /go", m.editScreen.originalValue)
	}
}

func TestMainModel_ChangeScreen(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.currentScreen = AppScreenEdit

	msg := changeScreenMsg(AppScreenList)
	newModel, _ := m.Update(msg)
	m = newModel.(mainModel)

	if m.currentScreen != AppScreenList {
		t.Errorf("currentScreen = %v, want AppScreenList", m.currentScreen)
	}
}

func TestMainModel_ShowHelp(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	msg := showHelpMsg{}
	newModel, _ := m.Update(msg)
	m = newModel.(mainModel)

	if m.currentScreen != AppScreenHelp {
		t.Errorf("currentScreen = %v, want AppScreenHelp", m.currentScreen)
	}
}

// =============================================================================
// REGRESSION TESTS - Verify fixes for critical issues
// =============================================================================

// TestRegression_EditScreenReturnsSaveRequest verifies that EditScreen returns
// saveEnvRequestMsg (not envSetSuccessMsg), so the actual save is triggered.
// This was a critical bug where edits appeared successful but weren't persisted.
func TestRegression_EditScreenReturnsSaveRequest(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{})
	m.editingKey = "GOPATH"
	m.originalValue = "/old"
	m.textInput.SetValue("/new")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	resultMsg := cmd()

	// CRITICAL: Must return saveEnvRequestMsg, NOT envSetSuccessMsg
	_, isSaveRequest := resultMsg.(saveEnvRequestMsg)
	_, isSuccess := resultMsg.(envSetSuccessMsg)

	if !isSaveRequest {
		t.Errorf("EditScreen must return saveEnvRequestMsg to trigger actual save, got %T", resultMsg)
	}
	if isSuccess {
		t.Error("EditScreen must NOT return envSetSuccessMsg directly - this bypasses the actual save!")
	}
}

// TestRegression_HistoryRecordedOnlyAfterSaveSuccess verifies that history is
// recorded only when save succeeds, not when the edit is initiated.
func TestRegression_HistoryRecordedOnlyAfterSaveSuccess(t *testing.T) {
	_ = persist.ClearHistory() // Ensure clean state
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/old"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	// saveEnvRequestMsg should NOT record history (save hasn't completed yet)
	saveReq := saveEnvRequestMsg{
		Env:           goenv.EnvVar{Key: "GOPATH", Value: "/new", Changed: true},
		OriginalValue: "/old",
	}
	newModel, _ := m.Update(saveReq)
	m = newModel.(mainModel)

	if len(m.editHistory) != 0 {
		t.Error("saveEnvRequestMsg should NOT record history - save hasn't completed")
	}

	// envSetSuccessMsg SHOULD record history (save completed)
	successMsg := envSetSuccessMsg{
		Env:           goenv.EnvVar{Key: "GOPATH", Value: "/new", Changed: true},
		OriginalValue: "/old",
		IsUndoRedo:    false,
	}
	newModel, _ = m.Update(successMsg)
	m = newModel.(mainModel)

	if len(m.editHistory) != 1 {
		t.Errorf("envSetSuccessMsg should record history, got %d entries", len(m.editHistory))
	}
}

// TestRegression_UndoRedoDoesNotRecordHistory verifies that undo/redo operations
// don't create new history entries (they navigate existing history).
func TestRegression_UndoRedoDoesNotRecordHistory(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/new"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	// Setup initial history
	m.editHistory = []persist.EditRecord{
		{Key: "GOPATH", OldValue: "/old", NewValue: "/new"},
	}
	m.historyIdx = 1

	// Simulate undo success message with IsUndoRedo=true
	successMsg := envSetSuccessMsg{
		Env:           goenv.EnvVar{Key: "GOPATH", Value: "/old", Changed: true},
		OriginalValue: "/new",
		IsUndoRedo:    true, // CRITICAL: This must prevent history recording
	}
	newModel, _ := m.Update(successMsg)
	m = newModel.(mainModel)

	// History should still be 1 entry, not 2
	if len(m.editHistory) != 1 {
		t.Errorf("undo/redo must NOT record new history, got %d entries (want 1)", len(m.editHistory))
	}
}

// TestRegression_UpdateItemByKeyReturnsBoolean verifies that updateItemByKey
// returns false when key is not found (previously returned nothing/silent fail).
func TestRegression_UpdateItemByKeyReturnsBoolean(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	// Updating existing key should return true
	found := m.updateItemByKey("GOPATH", goenv.EnvVar{Key: "GOPATH", Value: "/new"})
	if !found {
		t.Error("updateItemByKey should return true for existing key")
	}

	// Updating non-existent key should return false
	found = m.updateItemByKey("NONEXISTENT", goenv.EnvVar{Key: "NONEXISTENT", Value: "value"})
	if found {
		t.Error("updateItemByKey should return false for non-existent key")
	}
}

// TestRegression_SaveEnvRequestIncludesOriginalValue verifies that saveEnvRequestMsg
// includes the original value, which is needed for proper history recording.
func TestRegression_SaveEnvRequestIncludesOriginalValue(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{})
	m.editingKey = "GOPATH"
	m.originalValue = "/original/path"
	m.textInput.SetValue("/new/path")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	cmd := m.Update(msg)
	resultMsg := cmd()

	saveReq, ok := resultMsg.(saveEnvRequestMsg)
	if !ok {
		t.Fatalf("expected saveEnvRequestMsg, got %T", resultMsg)
	}

	if saveReq.OriginalValue != "/original/path" {
		t.Errorf("OriginalValue = %q, want /original/path", saveReq.OriginalValue)
	}
}

// TestRegression_MainModelHandlesSaveRequest verifies that mainModel properly
// handles saveEnvRequestMsg by setting isSaving and returning the save command.
func TestRegression_MainModelHandlesSaveRequest(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/old"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	saveReq := saveEnvRequestMsg{
		Env:           goenv.EnvVar{Key: "GOPATH", Value: "/new", Changed: true},
		OriginalValue: "/old",
	}
	newModel, cmd := m.Update(saveReq)
	m = newModel.(mainModel)

	// Should be in saving state
	if !m.isSaving {
		t.Error("mainModel should set isSaving=true when handling saveEnvRequestMsg")
	}

	// Should return a command (the actual save operation)
	if cmd == nil {
		t.Error("mainModel should return a command for the save operation")
	}
}

// TestRegression_ListSelectionPreserved verifies that refreshListScreen() preserves
// the current selection index when refreshing the list.
func TestRegression_ListSelectionPreserved(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "AAA", Value: "1"},
		{Key: "BBB", Value: "2"},
		{Key: "CCC", Value: "3"},
	}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	// Select 2nd item (index 1)
	m.listScreen.list.Select(1)

	// Refresh list
	m.refreshListScreen()

	// Verify selection preserved
	if m.listScreen.list.Index() != 1 {
		t.Errorf("expected index 1, got %d", m.listScreen.list.Index())
	}
}

// TestRegression_ListSelectionClamped verifies that selection is clamped
// when items are removed and the current index would be out of bounds.
func TestRegression_ListSelectionClamped(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "AAA", Value: "1"},
		{Key: "BBB", Value: "2"},
		{Key: "CCC", Value: "3"},
	}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	// Select last item (index 2)
	m.listScreen.list.Select(2)

	// Reduce items to 2
	m.items = m.items[:2]
	m.refreshListScreen()

	// Verify selection clamped to last valid index (1)
	if m.listScreen.list.Index() != 1 {
		t.Errorf("expected index 1 (clamped), got %d", m.listScreen.list.Index())
	}
}

// =============================================================================
// ERROR PATH TESTS - Verify error handling behavior
// =============================================================================

// TestMainModel_EnvSetError_ShowsError verifies that envSetErrorMsg sets saveError
// and clears isSaving state.
func TestMainModel_EnvSetError_ShowsError(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.isSaving = true
	m.currentScreen = AppScreenEdit

	errMsg := envSetErrorMsg{Key: "GOPATH", Err: fmt.Errorf("permission denied")}
	newModel, cmd := m.Update(errMsg)
	m = newModel.(mainModel)

	if m.isSaving {
		t.Error("isSaving should be false after error")
	}
	if m.saveError == "" {
		t.Error("saveError should be set")
	}
	if !strings.Contains(m.saveError, "GOPATH") {
		t.Errorf("saveError should mention the key, got: %q", m.saveError)
	}
	if !strings.Contains(m.saveError, "permission denied") {
		t.Errorf("saveError should contain error message, got: %q", m.saveError)
	}
	if cmd == nil {
		t.Error("should return command to clear message after timeout")
	}
}

// TestMainModel_ReloadError_ShowsError verifies that reloadEnvErrorMsg sets
// saveError and clears isReloading state.
func TestMainModel_ReloadError_ShowsError(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.isReloading = true

	errMsg := reloadEnvErrorMsg{Err: fmt.Errorf("go env command failed")}
	newModel, cmd := m.Update(errMsg)
	m = newModel.(mainModel)

	if m.isReloading {
		t.Error("isReloading should be false after error")
	}
	if m.saveError == "" {
		t.Error("saveError should be set")
	}
	if !strings.Contains(m.saveError, "Reload failed") {
		t.Errorf("saveError should indicate reload failure, got: %q", m.saveError)
	}
	if cmd == nil {
		t.Error("should return command to clear message after timeout")
	}
}

// TestMainModel_ClipboardError_ShowsError verifies that clipboardResultMsg
// with Success=false sets saveError.
func TestMainModel_ClipboardError_ShowsError(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	clipMsg := clipboardResultMsg{Success: false, Message: "clipboard not available"}
	newModel, cmd := m.Update(clipMsg)
	m = newModel.(mainModel)

	if m.saveError != "clipboard not available" {
		t.Errorf("saveError = %q, want 'clipboard not available'", m.saveError)
	}
	if m.successMessage != "" {
		t.Errorf("successMessage should be empty on error, got: %q", m.successMessage)
	}
	if cmd == nil {
		t.Error("should return command to clear message after timeout")
	}
}

// TestMainModel_ClipboardSuccess_ShowsSuccess verifies that clipboardResultMsg
// with Success=true sets successMessage.
func TestMainModel_ClipboardSuccess_ShowsSuccess(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	clipMsg := clipboardResultMsg{Success: true, Message: "Copied GOPATH value"}
	newModel, _ := m.Update(clipMsg)
	m = newModel.(mainModel)

	if m.successMessage != "Copied GOPATH value" {
		t.Errorf("successMessage = %q, want 'Copied GOPATH value'", m.successMessage)
	}
	if m.saveError != "" {
		t.Errorf("saveError should be empty on success, got: %q", m.saveError)
	}
}

// TestMainModel_ClearMessage_ClearsError verifies that clearMessageMsg
// clears both saveError and successMessage.
func TestMainModel_ClearMessage_ClearsError(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.saveError = "some error"
	m.successMessage = "some success"

	newModel, _ := m.Update(clearMessageMsg{})
	m = newModel.(mainModel)

	if m.saveError != "" {
		t.Errorf("saveError should be cleared, got: %q", m.saveError)
	}
	if m.successMessage != "" {
		t.Errorf("successMessage should be cleared, got: %q", m.successMessage)
	}
}

// TestMainModel_ErrorDisplay_InView verifies that View() includes error
// when saveError is set.
func TestMainModel_ErrorDisplay_InView(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.saveError = "Test error message"
	m.calculateScreenSizes()

	view := m.View()
	content := fmt.Sprint(view.Content)
	if !strings.Contains(content, "Test error message") {
		t.Error("View should display saveError when set")
	}
}

// TestMainModel_SuccessDisplay_InView verifies that View() includes success
// message when successMessage is set.
func TestMainModel_SuccessDisplay_InView(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.successMessage = "Operation successful"
	m.calculateScreenSizes()

	view := m.View()
	content := fmt.Sprint(view.Content)
	if !strings.Contains(content, "Operation successful") {
		t.Error("View should display successMessage when set")
	}
}

// TestRegression_KeyPressMsg_DelegatedToListScreen verifies that KeyPressMsg
// is properly delegated to the ListScreen when on the list screen.
// This was a regression where keys were not being forwarded to sub-screens.
func TestRegression_KeyPressMsg_DelegatedToListScreen(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	// Press Enter which should trigger startEditMsg
	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	newModel, cmd := m.Update(msg)
	m = newModel.(mainModel)

	if cmd == nil {
		t.Fatal("KeyPressMsg should return a command from ListScreen, got nil")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(startEditMsg); !ok {
		t.Errorf("Expected startEditMsg from Enter key, got %T", resultMsg)
	}
}

// TestRegression_KeyPressMsg_DelegatedToEditScreen verifies that KeyPressMsg
// is properly delegated to the EditScreen when on the edit screen.
func TestRegression_KeyPressMsg_DelegatedToEditScreen(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.currentScreen = AppScreenEdit
	m.editScreen.editingKey = "GOPATH"
	m.editScreen.originalValue = "/go"
	m.editScreen.textInput.SetValue("/new/path")

	// Press Enter which should trigger saveEnvRequestMsg
	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	newModel, cmd := m.Update(msg)
	m = newModel.(mainModel)

	if cmd == nil {
		t.Fatal("KeyPressMsg should return a command from EditScreen, got nil")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(saveEnvRequestMsg); !ok {
		t.Errorf("Expected saveEnvRequestMsg from Enter key on edit screen, got %T", resultMsg)
	}
}

// =============================================================================
// NEW FEATURE TESTS - Vim Navigation, Config, Descriptions
// =============================================================================

// TestListScreen_CtrlD_HalfPageDown verifies Ctrl+D moves down by half page.
func TestListScreen_CtrlD_HalfPageDown(t *testing.T) {
	items := make([]list.Item, 20)
	for i := range items {
		items[i] = envItem{goenv.EnvVar{Key: fmt.Sprintf("VAR%02d", i), Value: "value"}}
	}
	m := NewListScreenModel(items)
	m.list.SetHeight(10) // So half page is 5

	// Start at index 0
	m.list.Select(0)

	msg := tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl}
	cmd := m.Update(msg)

	// Should return nil (no follow-up command needed)
	if cmd != nil {
		t.Errorf("Ctrl+D should return nil command, got %v", cmd)
	}

	// Should have moved down by half page
	newIdx := m.list.Index()
	if newIdx < 1 {
		t.Errorf("Ctrl+D should move selection down, got index %d", newIdx)
	}
}

// TestListScreen_CtrlU_HalfPageUp verifies Ctrl+U moves up by half page.
func TestListScreen_CtrlU_HalfPageUp(t *testing.T) {
	items := make([]list.Item, 20)
	for i := range items {
		items[i] = envItem{goenv.EnvVar{Key: fmt.Sprintf("VAR%02d", i), Value: "value"}}
	}
	m := NewListScreenModel(items)
	m.list.SetHeight(10) // So half page is 5

	// Start at index 10
	m.list.Select(10)

	msg := tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl}
	cmd := m.Update(msg)

	// Should return nil (no follow-up command needed)
	if cmd != nil {
		t.Errorf("Ctrl+U should return nil command, got %v", cmd)
	}

	// Should have moved up by half page
	newIdx := m.list.Index()
	if newIdx > 9 {
		t.Errorf("Ctrl+U should move selection up, got index %d", newIdx)
	}
}

// TestConfig_DefaultConfig verifies default configuration values.
func TestConfig_DefaultConfig(t *testing.T) {
	cfg := persist.DefaultConfig()
	if cfg.SortMode != "alpha" {
		t.Errorf("DefaultConfig().SortMode = %q, want 'alpha'", cfg.SortMode)
	}
}

// TestConfig_SortModeConversion verifies SortMode string conversions.
func TestConfig_SortModeConversion(t *testing.T) {
	tests := []struct {
		str  string
		mode persist.SortMode
	}{
		{"alpha", persist.SortAlpha},
		{"modified_first", persist.SortModifiedFirst},
		{"invalid", persist.SortAlpha}, // defaults to alpha
		{"", persist.SortAlpha},        // defaults to alpha
	}

	for _, tt := range tests {
		got := persist.SortModeFromString(tt.str)
		if got != tt.mode {
			t.Errorf("SortModeFromString(%q) = %v, want %v", tt.str, got, tt.mode)
		}
	}

	// Test reverse conversion
	if persist.SortModeToString(persist.SortAlpha) != "alpha" {
		t.Errorf("SortModeToString(SortAlpha) = %q, want 'alpha'", persist.SortModeToString(persist.SortAlpha))
	}
	if persist.SortModeToString(persist.SortModifiedFirst) != "modified_first" {
		t.Errorf("SortModeToString(SortModifiedFirst) = %q, want 'modified_first'", persist.SortModeToString(persist.SortModifiedFirst))
	}
}

// TestGetEnvVarDescription verifies description lookup.
func TestGetEnvVarDescription(t *testing.T) {
	tests := []struct {
		key      string
		wantDesc bool
	}{
		{"GOPATH", true},
		{"GOROOT", true},
		{"GOPROXY", true},
		{"CGO_ENABLED", true},
		{"NONEXISTENT_VAR", false},
	}

	for _, tt := range tests {
		desc := goenv.GetEnvVarDescription(tt.key)
		if tt.wantDesc && desc == "" {
			t.Errorf("GetEnvVarDescription(%q) = empty, want description", tt.key)
		}
		if !tt.wantDesc && desc != "" {
			t.Errorf("GetEnvVarDescription(%q) = %q, want empty", tt.key, desc)
		}
	}
}

// TestMainModel_ConfigApplied verifies config is applied at startup.
func TestMainModel_ConfigApplied(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}

	// Test with modified_first sort mode
	cfg := persist.Config{SortMode: "modified_first"}
	m := NewMainModel(items, nil, cfg)

	if m.sortMode != persist.SortModifiedFirst {
		t.Errorf("NewMainModel with modified_first config: sortMode = %v, want SortModifiedFirst", m.sortMode)
	}

	// Test with alpha sort mode
	cfg = persist.Config{SortMode: "alpha"}
	m = NewMainModel(items, nil, cfg)

	if m.sortMode != persist.SortAlpha {
		t.Errorf("NewMainModel with alpha config: sortMode = %v, want SortAlpha", m.sortMode)
	}
}
