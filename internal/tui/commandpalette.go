package tui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// CommandEntry represents a single command available in the palette.
type CommandEntry struct {
	Name string  // Display name
	Desc string  // Short description
	Key  string  // Keyboard shortcut (for display)
	Msg  tea.Msg // Message to dispatch when selected
}

// CommandPaletteModel provides a fuzzy-filtered command picker.
type CommandPaletteModel struct {
	textInput   textinput.Model
	allEntries  []CommandEntry
	filtered    []CommandEntry
	selectedIdx int
	maxVisible  int
}

// NewCommandPaletteModel creates a new command palette.
func NewCommandPaletteModel() CommandPaletteModel {
	ti := textinput.New()
	ti.Placeholder = "type a command..."
	ti.CharLimit = 64
	ti.Prompt = ": "
	ti.SetStyles(textInputStyles)

	entries := defaultCommands()

	return CommandPaletteModel{
		textInput:  ti,
		allEntries: entries,
		filtered:   entries,
		maxVisible: 12,
	}
}

// defaultCommands returns the full list of palette commands.
func defaultCommands() []CommandEntry {
	return []CommandEntry{
		{Name: "edit", Desc: "Edit selected variable", Key: "Enter", Msg: nil}, // special: handled inline
		{Name: "sort", Desc: "Toggle sort mode", Key: "s", Msg: toggleSortMsg{}},
		{Name: "reload", Desc: "Reload environment", Key: "r", Msg: reloadEnvMsg{}},
		{Name: "copy", Desc: "Copy value to clipboard", Key: "y", Msg: nil},    // special
		{Name: "copykey", Desc: "Copy KEY=VALUE", Key: "Y", Msg: nil},          // special
		{Name: "help", Desc: "Show keyboard shortcuts", Key: "?", Msg: showHelpMsg{}},
		{Name: "undo", Desc: "Undo last change", Key: "Ctrl+Z", Msg: undoMsg{}},
		{Name: "redo", Desc: "Redo last change", Key: "Ctrl+Y", Msg: redoMsg{}},
		{Name: "favorite", Desc: "Toggle favorite", Key: "f", Msg: nil}, // special
		{Name: "export", Desc: "Export snapshot", Key: "e", Msg: showExportMsg{}},
		{Name: "import", Desc: "Import snapshot", Key: "i", Msg: showImportMsg{}},
		{Name: "presets", Desc: "Manage presets", Key: "p", Msg: showPresetsMsg{}},
		{Name: "compare", Desc: "Compare with snapshot/preset", Key: "c", Msg: showCompareMsg{}},
		{Name: "stats", Desc: "Show statistics", Key: "S", Msg: showStatsMsg{}},
		{Name: "reset", Desc: "Reset variable to default", Key: "u", Msg: nil}, // special
		{Name: "resetall", Desc: "Batch reset to defaults", Key: "U", Msg: showResetScreenMsg{}},
		{Name: "shellexport", Desc: "Export as shell commands", Key: "x", Msg: showShellExportMsg{}},
		{Name: "watch", Desc: "Toggle watch mode", Key: "w", Msg: toggleWatchMsg{}},
		{Name: "theme", Desc: "Cycle color theme", Key: "t", Msg: toggleThemeMsg{}},
		{Name: "category", Desc: "Cycle category filter", Key: "C", Msg: cycleCategoryMsg{}},
		{Name: "quit", Desc: "Quit goenv", Key: "q", Msg: nil}, // special: tea.Quit
	}
}

// Init initializes the command palette.
func (m *CommandPaletteModel) Init() tea.Cmd {
	m.textInput.SetValue("")
	m.textInput.Focus()
	m.filtered = m.allEntries
	m.selectedIdx = 0
	return textinput.Blink
}

// Update handles messages for the command palette.
func (m *CommandPaletteModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return sendMsg(changeScreenMsg(AppScreenList))
		case "enter":
			if len(m.filtered) > 0 && m.selectedIdx < len(m.filtered) {
				return m.dispatchSelected()
			}
			return nil
		case "up", "ctrl+k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
			return nil
		case "down", "ctrl+j":
			if m.selectedIdx < len(m.filtered)-1 {
				m.selectedIdx++
			}
			return nil
		}
	}

	// Update text input and re-filter
	prevValue := m.textInput.Value()
	newTextInput, cmd := m.textInput.Update(msg)
	m.textInput = newTextInput

	if m.textInput.Value() != prevValue {
		m.refilter()
	}

	return cmd
}

// dispatchSelected returns the command for the selected entry.
func (m *CommandPaletteModel) dispatchSelected() tea.Cmd {
	entry := m.filtered[m.selectedIdx]

	// Handle special commands that need list context
	switch entry.Name {
	case "quit":
		return tea.Quit
	case "edit":
		return sendMsg(commandPaletteEditMsg{})
	case "copy":
		return sendMsg(commandPaletteCopyMsg{})
	case "copykey":
		return sendMsg(commandPaletteCopyKeyMsg{})
	case "favorite":
		return sendMsg(commandPaletteFavoriteMsg{})
	case "reset":
		return sendMsg(commandPaletteResetMsg{})
	}

	if entry.Msg != nil {
		return sendMsg(entry.Msg)
	}
	return nil
}

// refilter updates the filtered list based on current input.
func (m *CommandPaletteModel) refilter() {
	query := strings.ToLower(m.textInput.Value())
	if query == "" {
		m.filtered = m.allEntries
		m.selectedIdx = 0
		return
	}

	var matches []CommandEntry
	for _, e := range m.allEntries {
		target := strings.ToLower(e.Name + " " + e.Desc)
		if strings.Contains(target, query) {
			matches = append(matches, e)
		}
	}
	m.filtered = matches
	m.selectedIdx = 0
}

// View renders the command palette.
func (m CommandPaletteModel) View() string {
	var parts []string

	parts = append(parts, helpTitleStyle.Render("Command Palette"))
	parts = append(parts, m.textInput.View())
	parts = append(parts, "")

	if len(m.filtered) == 0 {
		parts = append(parts, suggestionHintStyle.Render("No matching commands"))
	} else {
		endIdx := m.maxVisible
		if endIdx > len(m.filtered) {
			endIdx = len(m.filtered)
		}

		// Ensure selected item is visible
		startIdx := 0
		if m.selectedIdx >= endIdx {
			startIdx = m.selectedIdx - m.maxVisible + 1
			endIdx = m.selectedIdx + 1
		}

		for i := startIdx; i < endIdx; i++ {
			e := m.filtered[i]
			line := "  " + e.Name
			if e.Key != "" {
				line += "  [" + e.Key + "]"
			}
			if e.Desc != "" {
				line += "  " + e.Desc
			}
			if i == m.selectedIdx {
				parts = append(parts, suggestionSelectedStyle.Render("> "+e.Name+
					"  "+suggestionHintStyle.Render(e.Desc)))
			} else {
				parts = append(parts, suggestionStyle.Render(line))
			}
		}

		if len(m.filtered) > m.maxVisible {
			parts = append(parts, "")
			parts = append(parts, suggestionHintStyle.Render("... and more"))
		}
	}

	parts = append(parts, "")
	parts = append(parts, helpFooterStyle.Render("(Enter: run, Esc: cancel, arrows: navigate)"))

	return helpBoxStyle.Render(strings.Join(parts, "\n"))
}
