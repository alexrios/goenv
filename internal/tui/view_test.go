package tui

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

// TestEditScreenView_ShowsLengthDiff verifies that the edit screen displays
// the length difference when the value has been modified.
func TestEditScreenView_ShowsLengthDiff(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{})
	m.editingKey = "GOPATH"
	m.originalValue = "/short"
	m.textInput.SetValue("/much/longer/path")

	view := m.View()

	// Should show positive diff when value is longer
	if !strings.Contains(view, "(+") {
		t.Error("View should show positive length difference when new value is longer")
	}

	// Test negative diff
	m.originalValue = "/a/very/long/original/path"
	m.textInput.SetValue("/short")
	view = m.View()

	// Should show negative diff when value is shorter
	if !strings.Contains(view, "(-") {
		t.Error("View should show negative length difference when new value is shorter")
	}
}

// TestEditScreenView_ShowsOriginalValue verifies that the edit screen
// displays the original value for reference using visual diff.
func TestEditScreenView_ShowsOriginalValue(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{})
	m.editingKey = "GOPROXY"
	m.originalValue = "https://proxy.golang.org,direct"
	m.textInput.SetValue("off")

	view := m.View()

	// When value is changed, should show "New:" label (the Old: label may have ANSI codes)
	// and the new value should be visible
	if !strings.Contains(view, "New:") {
		t.Error("View should display 'New:' label when value differs")
	}
	if !strings.Contains(view, "off") {
		t.Error("View should display the new value")
	}

	// Should show the editing key
	if !strings.Contains(view, "Editing: GOPROXY") {
		t.Error("View should display the key being edited")
	}

	// Test unchanged value shows "Current:"
	m.textInput.SetValue(m.originalValue)
	view = m.View()
	if !strings.Contains(view, "Current:") {
		t.Error("View should display 'Current:' when value is unchanged")
	}
}

// TestEditScreenView_ShowsKeyboardHints verifies that keyboard shortcuts are displayed.
func TestEditScreenView_ShowsKeyboardHints(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{})
	m.editingKey = "GOPATH"
	m.originalValue = "/go"

	view := m.View()

	if !strings.Contains(view, "Enter") {
		t.Error("View should show Enter key hint")
	}
	if !strings.Contains(view, "Esc") {
		t.Error("View should show Esc key hint")
	}
	if !strings.Contains(view, "Ctrl+R") {
		t.Error("View should show Ctrl+R key hint")
	}
}

// TestMainModelView_ShowsStatusBar verifies that the main model view
// includes status bar information when on the list screen.
func TestMainModelView_ShowsStatusBar(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "GOPATH", Value: "/go"},
		{Key: "GOROOT", Value: "/usr/local/go"},
	}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.calculateScreenSizes()

	view := m.View()
	content := fmt.Sprint(view.Content)

	// Status bar should show item count
	if !strings.Contains(content, "variables") {
		t.Error("Status bar should indicate number of variables")
	}
}

// TestMainModelView_ShowsErrorMessage verifies that error messages
// are displayed in the view.
func TestMainModelView_ShowsErrorMessage(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.saveError = "Error setting GOPATH: permission denied"
	m.calculateScreenSizes()

	view := m.View()
	content := fmt.Sprint(view.Content)

	if !strings.Contains(content, "Error setting GOPATH") {
		t.Error("View should display error message when saveError is set")
	}
}

// TestMainModelView_ShowsSuccessMessage verifies that success messages
// are displayed in the status bar with the success icon.
func TestMainModelView_ShowsSuccessMessage(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.successMessage = "Saved 'GOPATH' successfully!"
	m.calculateScreenSizes()

	view := m.View()
	content := fmt.Sprint(view.Content)

	// Look for the success icon or the success message content
	if !strings.Contains(content, IconSuccess) && !strings.Contains(content, "Saved 'GOPATH'") {
		t.Error("View should display success icon or message when successMessage is set")
	}
}

// TestHelpScreenView_ContainsAllShortcuts verifies that all documented
// keyboard shortcuts are present in the help screen.
func TestHelpScreenView_ContainsAllShortcuts(t *testing.T) {
	m := NewHelpScreenModel()
	view := m.View()

	// Navigation shortcuts
	navigationShortcuts := []string{
		"j/k",
		"/",
		"Enter",
		"q",
	}

	for _, shortcut := range navigationShortcuts {
		if !strings.Contains(view, shortcut) {
			t.Errorf("Help screen should document %q navigation shortcut", shortcut)
		}
	}

	// Action shortcuts
	actionShortcuts := []string{
		"y",
		"Y",
		"f",
		"s",
		"r",
		"Ctrl+Z",
		"Ctrl+Y",
	}

	for _, shortcut := range actionShortcuts {
		if !strings.Contains(view, shortcut) {
			t.Errorf("Help screen should document %q action shortcut", shortcut)
		}
	}

	// Edit mode shortcuts
	editShortcuts := []string{
		"Enter",
		"Esc",
		"Ctrl+R",
		"Tab",
		"Shift+Tab",
	}

	for _, shortcut := range editShortcuts {
		if !strings.Contains(view, shortcut) {
			t.Errorf("Help screen should document %q edit mode shortcut", shortcut)
		}
	}
}

// TestHelpScreenView_ContainsSections verifies that help screen has
// organized sections for different shortcut categories.
func TestHelpScreenView_ContainsSections(t *testing.T) {
	m := NewHelpScreenModel()
	view := m.View()

	sections := []string{
		"Navigation",
		"Actions",
		"Edit Mode",
	}

	for _, section := range sections {
		if !strings.Contains(view, section) {
			t.Errorf("Help screen should have %q section", section)
		}
	}
}

// TestHelpScreenView_ContainsTitle verifies that the help screen has a title.
func TestHelpScreenView_ContainsTitle(t *testing.T) {
	m := NewHelpScreenModel()
	view := m.View()

	if !strings.Contains(view, "GO ENV") {
		t.Error("Help screen should contain 'GO ENV' in the title")
	}
	if !strings.Contains(view, "Keyboard Shortcuts") {
		t.Error("Help screen should mention 'Keyboard Shortcuts' in the title")
	}
}

// TestEditScreenView_ShowsExtendedDoc verifies that the edit screen shows
// extended documentation for known variables.
func TestEditScreenView_ShowsExtendedDoc(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{})
	m.editingKey = "GOOS"
	m.originalValue = "linux"
	m.textInput.SetValue("linux")

	view := m.View()
	doc := goenv.GetExtendedDoc("GOOS")
	if doc != "" && !strings.Contains(view, "Examples") {
		t.Error("View should display extended documentation for GOOS")
	}
}

// TestEditScreenView_ShowsCSVBreakdown verifies that CSV variables show item breakdown.
func TestEditScreenView_ShowsCSVBreakdown(t *testing.T) {
	m := NewEditScreenModel(goenv.GoVersion{})
	m.editingKey = "GODEBUG"
	m.originalValue = "gctrace=1,schedtrace=1000"
	m.textInput.SetValue("gctrace=1,schedtrace=1000")

	view := m.View()
	if !strings.Contains(view, "Current items:") {
		t.Error("View should show CSV breakdown for GODEBUG")
	}
	if !strings.Contains(view, "gctrace=1") {
		t.Error("View should show individual CSV items")
	}
}

// TestMainModelView_ShowsScreenNameOnEditScreen verifies that status bar
// shows the screen name when on the edit screen.
func TestMainModelView_ShowsScreenNameOnEditScreen(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.currentScreen = AppScreenEdit
	m.editScreen.editingKey = "GOPATH"
	m.editScreen.originalValue = "/go"
	m.calculateScreenSizes()

	view := m.View()
	content := fmt.Sprint(view.Content)

	// Should not show the "Press Enter to edit" help text on edit screen
	if strings.Contains(content, "Press Enter to edit") {
		t.Error("Edit screen should not show list screen help text")
	}

	// Should show screen name in status bar
	if !strings.Contains(content, "Edit") {
		t.Error("Edit screen should show 'Edit' in status bar")
	}

	// Should show "(Esc: back)" help text
	if !strings.Contains(content, "Esc: back") {
		t.Error("Edit screen should show 'Esc: back' help text")
	}
}

// TestMainModelView_StatusBarShowsErrorOnSecondaryScreen verifies that
// error messages still appear on non-list screens.
func TestMainModelView_StatusBarShowsErrorOnSecondaryScreen(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.currentScreen = AppScreenHelp
	m.saveError = "Something went wrong"
	m.calculateScreenSizes()

	view := m.View()
	content := fmt.Sprint(view.Content)

	if !strings.Contains(content, "Something went wrong") {
		t.Error("Error messages should be visible on secondary screens")
	}
}

// TestMainModelView_StatusBarShowsSuccessOnSecondaryScreen verifies that
// success messages appear on non-list screens.
func TestMainModelView_StatusBarShowsSuccessOnSecondaryScreen(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.width = 80
	m.height = 24
	m.currentScreen = AppScreenExport
	m.successMessage = "Export completed"
	m.calculateScreenSizes()

	view := m.View()
	content := fmt.Sprint(view.Content)

	if !strings.Contains(content, "Export completed") {
		t.Error("Success messages should be visible on secondary screens")
	}
}

// Ensure list.Item import is used
var _ list.Item = envItem{}
