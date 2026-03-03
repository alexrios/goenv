package tui

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

// observableModel wraps a tea.Model to capture intermediate state after each
// Update call. This enables waitFor to poll model state while the program runs,
// replacing fragile time.Sleep-based synchronization with condition-based polling
// (inspired by teatest's WaitFor pattern).
type observableModel struct {
	inner     tea.Model
	mu        sync.RWMutex
	snapshot  mainModel
	ready     chan struct{}
	readyOnce sync.Once
}

func newObservableModel(m tea.Model) *observableModel {
	return &observableModel{
		inner: m,
		ready: make(chan struct{}),
	}
}

func (m *observableModel) Init() tea.Cmd {
	return m.inner.Init()
}

func (m *observableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newModel, cmd := m.inner.Update(msg)
	m.inner = newModel
	if mm, ok := newModel.(mainModel); ok {
		m.mu.Lock()
		m.snapshot = mm
		m.mu.Unlock()
	}
	m.readyOnce.Do(func() { close(m.ready) })
	return m, cmd
}

func (m *observableModel) View() tea.View {
	return m.inner.View()
}

func (m *observableModel) getSnapshot() mainModel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.snapshot
}

// testProgram wraps a tea.Program for testing with input/output capture.
type testProgram struct {
	t       *testing.T
	program *tea.Program
	output  *bytes.Buffer
	obs     *observableModel
	model   tea.Model
	mu      sync.Mutex
	done    chan struct{}
}

// newTestProgram creates a test program with the given model and terminal size.
func newTestProgram(t *testing.T, m tea.Model, width, height int) *testProgram {
	output := &bytes.Buffer{}
	obs := newObservableModel(m)
	tp := &testProgram{
		t:      t,
		output: output,
		obs:    obs,
		done:   make(chan struct{}),
	}

	tp.program = tea.NewProgram(obs,
		tea.WithInput(strings.NewReader("")),
		tea.WithOutput(output),
		tea.WithWindowSize(width, height),
		tea.WithoutSignalHandler(),
		tea.WithoutRenderer(),
	)

	return tp
}

// start runs the program in a goroutine and waits for initialization.
func (tp *testProgram) start() {
	go func() {
		finalModel, err := tp.program.Run()
		if err != nil {
			tp.t.Logf("program finished with error: %v", err)
		}
		tp.mu.Lock()
		tp.model = finalModel
		tp.mu.Unlock()
		close(tp.done)
	}()
	// Wait for the program to process its first Update (e.g. WindowSizeMsg)
	select {
	case <-tp.obs.ready:
	case <-time.After(2 * time.Second):
		tp.t.Fatal("timeout waiting for program to initialize")
	}
}

// send sends a message to the program.
func (tp *testProgram) send(msg tea.Msg) {
	tp.program.Send(msg)
}

// sendKey sends a key press message.
func (tp *testProgram) sendKey(key string) {
	var msg tea.KeyPressMsg
	switch key {
	case "enter":
		msg = tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		msg = tea.KeyPressMsg{Code: tea.KeyEscape}
	case "up":
		msg = tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		msg = tea.KeyPressMsg{Code: tea.KeyDown}
	case "q":
		msg = tea.KeyPressMsg{Code: 'q', Text: "q"}
	case "?":
		msg = tea.KeyPressMsg{Code: '?', Text: "?"}
	case "s":
		msg = tea.KeyPressMsg{Code: 's', Text: "s"}
	case "r":
		msg = tea.KeyPressMsg{Code: 'r', Text: "r"}
	case "ctrl+z":
		msg = tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl}
	default:
		// For single character keys
		if len(key) == 1 {
			msg = tea.KeyPressMsg{Code: rune(key[0]), Text: key}
		}
	}
	tp.send(msg)
}

// quit sends a quit command and waits for the program to finish.
func (tp *testProgram) quit(timeout time.Duration) tea.Model {
	tp.program.Quit()
	select {
	case <-tp.done:
	case <-time.After(timeout):
		tp.t.Log("timeout waiting for program to quit")
	}
	tp.mu.Lock()
	defer tp.mu.Unlock()
	// Unwrap observableModel to return the inner model
	if obs, ok := tp.model.(*observableModel); ok {
		return obs.inner
	}
	return tp.model
}

// getFinalModel returns the final model after program completion.
func (tp *testProgram) getFinalModel() mainModel {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	model := tp.model
	// Unwrap observableModel
	if obs, ok := model.(*observableModel); ok {
		model = obs.inner
	}
	if model == nil {
		return mainModel{}
	}
	if m, ok := model.(mainModel); ok {
		return m
	}
	tp.t.Fatalf("unexpected model type: %T", model)
	return mainModel{}
}

// waitFor polls the model state until condition returns true or timeout expires.
// This replaces time.Sleep-based synchronization with explicit condition checks,
// making tests faster (no over-sleeping) and less flaky (no under-sleeping).
func (tp *testProgram) waitFor(condition func(mainModel) bool, timeout time.Duration) mainModel {
	tp.t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		snap := tp.obs.getSnapshot()
		if condition(snap) {
			return snap
		}
		select {
		case <-tp.done:
			// Program exited -- check condition one last time
			snap = tp.obs.getSnapshot()
			if condition(snap) {
				return snap
			}
			tp.t.Fatalf("waitFor: program exited before condition was met")
			return mainModel{}
		case <-time.After(5 * time.Millisecond):
		}
	}
	tp.t.Fatalf("waitFor: condition not met within %v", timeout)
	return mainModel{}
}

// waitForOutput polls the model's View output until it contains the expected
// string or timeout expires. Useful for verifying screen content transitions.
func (tp *testProgram) waitForOutput(contains string, timeout time.Duration) {
	tp.t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		snap := tp.obs.getSnapshot()
		view := snap.View()
		content := fmt.Sprint(view.Content)
		if strings.Contains(content, contains) {
			return
		}
		select {
		case <-tp.done:
			snap = tp.obs.getSnapshot()
			view = snap.View()
			content = fmt.Sprint(view.Content)
			if strings.Contains(content, contains) {
				return
			}
			tp.t.Fatalf("waitForOutput: program exited before %q was found in view", contains)
			return
		case <-time.After(5 * time.Millisecond):
		}
	}
	tp.t.Fatalf("waitForOutput: %q not found in view within %v", contains, timeout)
}

// --- Integration Tests ---

// TestIntegration_ListNavigation verifies arrow keys navigate the list.
func TestIntegration_ListNavigation(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "GOPATH", Value: "/home/user/go"},
		{Key: "GOROOT", Value: "/usr/local/go"},
		{Key: "GOBIN", Value: "/home/user/go/bin"},
	}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.calculateScreenSizes()

	tp := newTestProgram(t, m, 80, 24)
	tp.start()

	// Navigate down
	tp.sendKey("down")
	tp.sendKey("down")

	// Quit and verify
	finalModel := tp.quit(time.Second)
	if finalModel == nil {
		t.Fatal("expected final model, got nil")
	}

	fm, ok := finalModel.(mainModel)
	if !ok {
		t.Fatalf("unexpected model type: %T", finalModel)
	}

	if fm.currentScreen != AppScreenList {
		t.Errorf("expected AppScreenList, got %v", fm.currentScreen)
	}
}

// TestIntegration_EnterOpensEdit verifies Enter transitions to edit screen.
func TestIntegration_EnterOpensEdit(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "GOPATH", Value: "/home/user/go"},
	}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.calculateScreenSizes()

	tp := newTestProgram(t, m, 80, 24)
	tp.start()

	// Press Enter to edit -- waitFor replaces time.Sleep
	tp.sendKey("enter")
	tp.waitFor(func(m mainModel) bool {
		return m.editScreen.editingKey == "GOPATH"
	}, time.Second)

	// Quit
	finalModel := tp.quit(time.Second)
	if finalModel == nil {
		t.Fatal("expected final model, got nil")
	}

	fm, ok := finalModel.(mainModel)
	if !ok {
		t.Fatalf("unexpected model type: %T", finalModel)
	}

	if fm.editScreen.editingKey != "GOPATH" {
		t.Errorf("expected editingKey to be GOPATH, got %q", fm.editScreen.editingKey)
	}
}

// TestIntegration_EscCancelsEdit verifies Esc returns from edit to list.
func TestIntegration_EscCancelsEdit(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "GOPATH", Value: "/home/user/go"},
	}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.calculateScreenSizes()

	tp := newTestProgram(t, m, 80, 24)
	tp.start()

	// Press Enter to edit
	tp.sendKey("enter")
	tp.waitFor(func(m mainModel) bool {
		return m.currentScreen == AppScreenEdit
	}, time.Second)

	// Press Esc to cancel
	tp.sendKey("esc")
	tp.waitFor(func(m mainModel) bool {
		return m.currentScreen == AppScreenList
	}, time.Second)

	// Quit
	finalModel := tp.quit(time.Second)
	if finalModel == nil {
		t.Fatal("expected final model, got nil")
	}

	fm, ok := finalModel.(mainModel)
	if !ok {
		t.Fatalf("unexpected model type: %T", finalModel)
	}

	// Should be back on list screen
	if fm.currentScreen != AppScreenList {
		t.Errorf("expected AppScreenList after Esc, got %v", fm.currentScreen)
	}
}

// TestIntegration_HelpScreen verifies ? opens and closes help.
func TestIntegration_HelpScreen(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.calculateScreenSizes()

	tp := newTestProgram(t, m, 80, 24)
	tp.start()

	// Press ? for help -- now we CAN verify intermediate state
	tp.sendKey("?")
	tp.waitFor(func(m mainModel) bool {
		return m.currentScreen == AppScreenHelp
	}, time.Second)

	// Close help with Esc
	tp.sendKey("esc")
	tp.waitFor(func(m mainModel) bool {
		return m.currentScreen == AppScreenList
	}, time.Second)

	// Quit
	finalModel := tp.quit(time.Second)
	if finalModel == nil {
		t.Fatal("expected final model, got nil")
	}

	fm, ok := finalModel.(mainModel)
	if !ok {
		t.Fatalf("unexpected model type: %T", finalModel)
	}

	// Should be back on list screen
	if fm.currentScreen != AppScreenList {
		t.Errorf("expected AppScreenList after closing help, got %v", fm.currentScreen)
	}
}

// TestIntegration_SortToggle verifies s key toggles sort mode.
func TestIntegration_SortToggle(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "GOROOT", Value: "/usr/local/go"},
		{Key: "GOPATH", Value: "/home/user/go"},
	}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.calculateScreenSizes()

	tp := newTestProgram(t, m, 80, 24)
	tp.start()

	// Initial sort mode should be alpha
	initialModel := m
	if initialModel.sortMode != persist.SortAlpha {
		t.Errorf("expected initial sortMode SortAlpha, got %v", initialModel.sortMode)
	}

	// Press s to toggle sort
	tp.sendKey("s")
	tp.waitFor(func(m mainModel) bool {
		return m.sortMode == persist.SortModifiedFirst
	}, time.Second)

	// Quit
	finalModel := tp.quit(time.Second)
	if finalModel == nil {
		t.Fatal("expected final model, got nil")
	}

	fm, ok := finalModel.(mainModel)
	if !ok {
		t.Fatalf("unexpected model type: %T", finalModel)
	}

	// Sort mode should be toggled
	if fm.sortMode != persist.SortModifiedFirst {
		t.Errorf("expected sortMode SortModifiedFirst after toggle, got %v", fm.sortMode)
	}
}

// TestIntegration_QuitFromList verifies q quits from list screen.
func TestIntegration_QuitFromList(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.calculateScreenSizes()

	tp := newTestProgram(t, m, 80, 24)
	tp.start()

	// Press q to quit
	tp.sendKey("q")

	// Wait for program to finish
	select {
	case <-tp.done:
		// Good, program quit
	case <-time.After(time.Second):
		t.Fatal("program didn't quit on 'q' key")
	}
}

// TestIntegration_QuitBlockedInEdit verifies q doesn't quit from edit screen.
func TestIntegration_QuitBlockedInEdit(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.calculateScreenSizes()

	tp := newTestProgram(t, m, 80, 24)
	tp.start()

	// Enter edit mode
	tp.sendKey("enter")
	tp.waitFor(func(m mainModel) bool {
		return m.currentScreen == AppScreenEdit
	}, time.Second)

	// Press q (should NOT quit because we're in edit mode)
	tp.sendKey("q")
	// Testing a negative requires a small sleep -- there's no state change to poll for
	time.Sleep(50 * time.Millisecond)

	// Program should still be running
	select {
	case <-tp.done:
		t.Fatal("program quit unexpectedly - 'q' should be blocked in edit mode")
	default:
		// Good, program is still running
	}

	// Force quit with program.Quit()
	tp.quit(time.Second)
}

// TestIntegration_FullEditCycle tests a complete edit cycle without external commands.
// This tests the internal message flow, not the actual go env -w command.
func TestIntegration_FullEditCycle(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "GOPATH", Value: "/old/path"},
	}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.calculateScreenSizes()

	// Test the message flow by directly calling Update
	// This is a hybrid approach - testing the model transitions

	// 1. Start on list screen
	if m.currentScreen != AppScreenList {
		t.Fatalf("expected AppScreenList, got %v", m.currentScreen)
	}

	// 2. Trigger startEditMsg
	newModel, _ := m.Update(startEditMsg{Key: "GOPATH", Value: "/old/path"})
	m = newModel.(mainModel)

	if m.currentScreen != AppScreenEdit {
		t.Fatalf("expected AppScreenEdit after startEditMsg, got %v", m.currentScreen)
	}
	if m.editScreen.editingKey != "GOPATH" {
		t.Fatalf("expected editingKey GOPATH, got %q", m.editScreen.editingKey)
	}

	// 3. Simulate successful save
	newModel, _ = m.Update(envSetSuccessMsg{
		Env:           goenv.EnvVar{Key: "GOPATH", Value: "/new/path", Changed: true},
		OriginalValue: "/old/path",
		IsUndoRedo:    false,
	})
	m = newModel.(mainModel)

	// 4. Verify back on list screen with updated value
	if m.currentScreen != AppScreenList {
		t.Errorf("expected AppScreenList after save, got %v", m.currentScreen)
	}

	// 5. Verify item was updated
	found := false
	for _, item := range m.items {
		if item.Key == "GOPATH" {
			if item.Value != "/new/path" {
				t.Errorf("GOPATH value = %q, want /new/path", item.Value)
			}
			found = true
			break
		}
	}
	if !found {
		t.Error("GOPATH not found in items after edit")
	}

	// 6. Verify history was recorded
	if len(m.editHistory) != 1 {
		t.Errorf("editHistory length = %d, want 1", len(m.editHistory))
	}
}

// TestIntegration_WindowResize verifies window resize handling.
func TestIntegration_WindowResize(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	tp := newTestProgram(t, m, 80, 24)
	tp.start()

	// Send window resize
	tp.send(tea.WindowSizeMsg{Width: 120, Height: 40})
	tp.waitFor(func(m mainModel) bool {
		return m.width == 120 && m.height == 40
	}, time.Second)

	// Quit
	finalModel := tp.quit(time.Second)
	if finalModel == nil {
		t.Fatal("expected final model, got nil")
	}

	fm, ok := finalModel.(mainModel)
	if !ok {
		t.Fatalf("unexpected model type: %T", finalModel)
	}

	// Verify dimensions were updated
	if fm.width != 120 {
		t.Errorf("expected width 120, got %d", fm.width)
	}
	if fm.height != 40 {
		t.Errorf("expected height 40, got %d", fm.height)
	}
}

// TestIntegration_UndoRedo tests undo/redo navigation through history.
func TestIntegration_UndoRedo(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "GOPATH", Value: "/new/path"},
	}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.calculateScreenSizes()

	// Setup history
	m.editHistory = []persist.EditRecord{
		{Key: "GOPATH", OldValue: "/old/path", NewValue: "/new/path"},
	}
	m.historyIdx = 1

	// Process undo
	newModel, _ := m.Update(undoMsg{})
	m = newModel.(mainModel)

	// Verify historyIdx was decremented
	if m.historyIdx != 0 {
		t.Errorf("historyIdx after undo = %d, want 0", m.historyIdx)
	}

	// Verify value was restored (the undo triggers a save command,
	// but we verify the in-memory state)
	for _, item := range m.items {
		if item.Key == "GOPATH" {
			if item.Value != "/old/path" {
				t.Errorf("GOPATH after undo = %q, want /old/path", item.Value)
			}
			break
		}
	}

	// Simulate save completion before redo
	m.isSaving = false

	// Process redo
	newModel, _ = m.Update(redoMsg{})
	m = newModel.(mainModel)

	// Verify historyIdx was incremented
	if m.historyIdx != 1 {
		t.Errorf("historyIdx after redo = %d, want 1", m.historyIdx)
	}

	// Verify value was reapplied
	for _, item := range m.items {
		if item.Key == "GOPATH" {
			if item.Value != "/new/path" {
				t.Errorf("GOPATH after redo = %q, want /new/path", item.Value)
			}
			break
		}
	}
}

// --- View Tests for Integration ---

// TestIntegration_ListViewContainsItems verifies the list view shows items.
func TestIntegration_ListViewContainsItems(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "GOPATH", Value: "/home/user/go"},
		{Key: "GOROOT", Value: "/usr/local/go"},
	}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.calculateScreenSizes()

	view := m.View()
	content := fmt.Sprint(view.Content)

	if !strings.Contains(content, "GOPATH") {
		t.Error("view should contain GOPATH")
	}
	if !strings.Contains(content, "GOROOT") {
		t.Error("view should contain GOROOT")
	}
}

// TestIntegration_EditViewContainsEditingInfo verifies edit screen shows editing info.
func TestIntegration_EditViewContainsEditingInfo(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.currentScreen = AppScreenEdit
	m.editScreen.editingKey = "GOPATH"
	m.editScreen.originalValue = "/old/value"
	m.calculateScreenSizes()

	view := m.View()
	content := fmt.Sprint(view.Content)

	if !strings.Contains(content, "Editing:") {
		t.Error("edit view should contain 'Editing:'")
	}
	if !strings.Contains(content, "GOPATH") {
		t.Error("edit view should contain the key being edited")
	}
}

// TestIntegration_HelpViewContainsShortcuts verifies help screen shows shortcuts.
func TestIntegration_HelpViewContainsShortcuts(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.currentScreen = AppScreenHelp
	m.calculateScreenSizes()

	view := m.View()
	content := fmt.Sprint(view.Content)

	if !strings.Contains(content, "Keyboard Shortcuts") {
		t.Error("help view should contain 'Keyboard Shortcuts'")
	}
	if !strings.Contains(content, "Navigation") {
		t.Error("help view should contain 'Navigation' section")
	}
}

// Ensure output buffer is used to prevent unused warning
var _ io.Writer = (*bytes.Buffer)(nil)
