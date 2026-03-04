package tui

import (
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

// ListScreenModel displays all Go environment variables in a scrollable list.
// It uses the bubbles/list component for filtering and navigation.
type ListScreenModel struct {
	list list.Model
}

// NewListScreenModel creates a new list screen with the given items.
func NewListScreenModel(items []list.Item) ListScreenModel {
	return NewListScreenModelWithVersion(items, "")
}

// NewListScreenModelWithVersion creates a new list screen with the given items and Go version.
func NewListScreenModelWithVersion(items []list.Item, goVersion string) ListScreenModel {
	return NewListScreenModelWithParsedVersion(items, goVersion, goenv.GoVersion{})
}

// NewListScreenModelWithSortMode creates a new list screen with version warnings, sort mode, and optional watch state.
func NewListScreenModelWithSortMode(items []list.Item, goVersion string, ver goenv.GoVersion, sortMode persist.SortMode, favorites map[string]bool, recentChanges ...map[string]time.Time) ListScreenModel {
	var delegate envVarDelegate
	if len(recentChanges) > 0 && recentChanges[0] != nil {
		delegate = NewEnvVarDelegateWithWatch(sortMode, favorites, recentChanges[0])
	} else {
		delegate = NewEnvVarDelegateWithSortMode(sortMode, favorites)
	}
	itemList := list.New(items, delegate, 0, 0)
	title := appTitle
	if goVersion != "" {
		title = appTitle + " (" + goVersion + ")"
	}

	if !ver.IsZero() {
		if !ver.AtLeast(goenv.MinVersionEnvWrite) {
			title += " [read-only: go env -w requires Go 1.13+]"
		} else if !ver.AtLeast(goenv.MinVersionChangedFlag) {
			title += " [no change detection]"
		}
	}

	itemList.Title = title
	itemList.SetStatusBarItemName("variable", "variables")
	itemList.KeyMap.Quit.Unbind()
	itemList.SetShowHelp(false)
	itemList.SetFilteringEnabled(true)

	return ListScreenModel{
		list: itemList,
	}
}

// NewListScreenModelWithParsedVersion creates a new list screen with version warnings.
func NewListScreenModelWithParsedVersion(items []list.Item, goVersion string, ver goenv.GoVersion) ListScreenModel {
	itemList := list.New(items, NewEnvVarDelegate(), 0, 0)
	title := appTitle
	if goVersion != "" {
		title = appTitle + " (" + goVersion + ")"
	}

	// Add version-specific warnings to the title
	if !ver.IsZero() {
		if !ver.AtLeast(goenv.MinVersionEnvWrite) {
			title += " [read-only: go env -w requires Go 1.13+]"
		} else if !ver.AtLeast(goenv.MinVersionChangedFlag) {
			title += " [no change detection]"
		}
	}

	itemList.Title = title
	itemList.SetStatusBarItemName("variable", "variables")
	itemList.KeyMap.Quit.Unbind()
	itemList.SetShowHelp(false)
	itemList.SetFilteringEnabled(true)

	return ListScreenModel{
		list: itemList,
	}
}

// Init initializes the list screen. Returns nil as no initial command is needed.
func (m ListScreenModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the list screen, including key presses for
// editing, sorting, reloading, copying, and undo/redo operations.
func (m *ListScreenModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Don't intercept shortcut keys when the filter input is active
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "enter":
			if selectedItem, ok := m.list.SelectedItem().(envItem); ok {
				if selectedItem.ReadOnly {
					return sendMsg(showDetailMsg(selectedItem.EnvVar))
				}
				return sendMsg(startEditMsg(selectedItem.EnvVar))
			}
		case "s":
			return sendMsg(toggleSortMsg{})
		case "r":
			return sendMsg(reloadEnvMsg{})
		case "y":
			if selectedItem, ok := m.list.SelectedItem().(envItem); ok {
				return sendMsg(copyValueMsg{Key: selectedItem.Key, Value: selectedItem.Value})
			}
		case "Y":
			if selectedItem, ok := m.list.SelectedItem().(envItem); ok {
				return sendMsg(copyKeyValueMsg{Key: selectedItem.Key, Value: selectedItem.Value})
			}
		case "?":
			return sendMsg(showHelpMsg{})
		case "ctrl+z":
			return sendMsg(undoMsg{})
		case "ctrl+y":
			return sendMsg(redoMsg{})
		case "f":
			if selectedItem, ok := m.list.SelectedItem().(envItem); ok {
				return sendMsg(toggleFavoriteMsg{Key: selectedItem.Key})
			}
		case "e":
			// Export snapshot
			return sendMsg(showExportMsg{})
		case "i":
			// Import snapshot
			return sendMsg(showImportMsg{})
		case "p":
			// Show presets
			return sendMsg(showPresetsMsg{})
		case "c":
			// Compare with snapshot/preset
			return sendMsg(showCompareMsg{})
		case "S":
			// Show stats (Shift+S)
			return sendMsg(showStatsMsg{})
		case "u":
			// Reset current variable to default (not available for read-only vars)
			if selectedItem, ok := m.list.SelectedItem().(envItem); ok {
				if !selectedItem.ReadOnly && selectedItem.Changed {
					return sendMsg(resetEnvMsg{Key: selectedItem.Key})
				}
			}
		case "U":
			// Show batch reset screen
			return sendMsg(showResetScreenMsg{})
		case "x":
			// Show shell export screen
			return sendMsg(showShellExportMsg{})
		case "w":
			// Toggle watch mode
			return sendMsg(toggleWatchMsg{})
		case "t":
			// Cycle theme
			return sendMsg(toggleThemeMsg{})
		case "C":
			// Cycle category filter
			return sendMsg(cycleCategoryMsg{})
		case ":":
			// Open command palette
			return sendMsg(showCommandPaletteMsg{})
		case "ctrl+d":
			// Half-page down
			visibleCount := m.list.Height() / 2
			newIdx := min(m.list.Index()+visibleCount, len(m.list.Items())-1)
			m.list.Select(newIdx)
			return nil
		case "ctrl+u":
			// Half-page up
			visibleCount := m.list.Height() / 2
			newIdx := max(m.list.Index()-visibleCount, 0)
			m.list.Select(newIdx)
			return nil
		}
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	return cmd
}

// View renders the list screen.
func (m ListScreenModel) View() string {
	return m.list.View()
}
