package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
)

// ResetScreenModel provides a screen for selecting variables to reset to defaults.
type ResetScreenModel struct {
	items          []goenv.EnvVar
	selected       map[string]bool
	cursor         int
	confirmMode    bool
	singleResetKey string // Non-empty when resetting a single variable
}

// NewResetScreenModel creates a new reset screen model.
func NewResetScreenModel() ResetScreenModel {
	return ResetScreenModel{
		selected: make(map[string]bool),
	}
}

// SetItems sets the available items for batch reset.
func (m *ResetScreenModel) SetItems(items []goenv.EnvVar) {
	m.items = items
	m.selected = make(map[string]bool)
	m.cursor = 0
	m.confirmMode = false
	m.singleResetKey = ""
}

// SetSingleReset sets the screen for confirming a single variable reset.
func (m *ResetScreenModel) SetSingleReset(key string) {
	m.singleResetKey = key
	m.confirmMode = true
}

// Init initializes the reset screen.
func (m *ResetScreenModel) Init() tea.Cmd {
	return nil
}

// getModifiedItems returns only the modified items.
func (m *ResetScreenModel) getModifiedItems() []goenv.EnvVar {
	var modified []goenv.EnvVar
	for _, ev := range m.items {
		if ev.Changed {
			modified = append(modified, ev)
		}
	}
	return modified
}

// Update handles messages for the reset screen.
func (m *ResetScreenModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Single reset confirmation mode
		if m.singleResetKey != "" {
			switch msg.String() {
			case "y", "Y", "enter":
				key := m.singleResetKey
				m.singleResetKey = ""
				m.confirmMode = false
				return sendMsg(confirmResetMsg{Key: key})
			case "n", "N", "esc", "q":
				m.singleResetKey = ""
				m.confirmMode = false
				return sendMsg(changeScreenMsg(AppScreenList))
			}
			return nil
		}

		// Batch reset mode
		modified := m.getModifiedItems()
		switch msg.String() {
		case "esc", "q":
			return sendMsg(changeScreenMsg(AppScreenList))
		case "j", "down":
			if len(modified) > 0 {
				m.cursor = (m.cursor + 1) % len(modified)
			}
		case "k", "up":
			if len(modified) > 0 {
				m.cursor = (m.cursor - 1 + len(modified)) % len(modified)
			}
		case " ":
			// Toggle selection
			if len(modified) > 0 {
				key := modified[m.cursor].Key
				m.selected[key] = !m.selected[key]
			}
		case "a":
			// Select all
			for _, ev := range modified {
				m.selected[ev.Key] = true
			}
		case "A":
			// Deselect all
			m.selected = make(map[string]bool)
		case "enter":
			// Confirm reset of selected items
			if len(m.selected) > 0 {
				m.confirmMode = true
			}
		case "y", "Y":
			// If in confirm mode, proceed with reset
			if m.confirmMode {
				var keys []string
				for key, selected := range m.selected {
					if selected {
						keys = append(keys, key)
					}
				}
				m.confirmMode = false
				m.selected = make(map[string]bool)
				return performBatchUnsetEnvCmd(keys)
			}
		case "n", "N":
			// Cancel confirmation
			if m.confirmMode {
				m.confirmMode = false
			}
		}
	}
	return nil
}

// View renders the reset screen.
func (m ResetScreenModel) View() string {
	var b strings.Builder

	// Single reset confirmation
	if m.singleResetKey != "" {
		b.WriteString(helpTitleStyle.Render("Reset to Default") + "\n\n")
		b.WriteString(fmt.Sprintf("Reset '%s' to its default value?\n\n", m.singleResetKey))
		b.WriteString("This will remove any user-set value for this variable.\n\n")
		b.WriteString(helpKeyStyle.Render("y") + helpDescStyle.Render("Yes, reset to default") + "\n")
		b.WriteString(helpKeyStyle.Render("n") + helpDescStyle.Render("No, cancel") + "\n")
		return helpBoxStyle.Render(b.String())
	}

	b.WriteString(helpTitleStyle.Render("Batch Reset to Defaults") + "\n\n")

	modified := m.getModifiedItems()
	if len(modified) == 0 {
		b.WriteString("No modified variables to reset.\n\n")
		b.WriteString(helpFooterStyle.Render("Press Esc to return..."))
		return helpBoxStyle.Render(b.String())
	}

	// Show confirmation prompt
	if m.confirmMode {
		count := 0
		for _, v := range m.selected {
			if v {
				count++
			}
		}
		b.WriteString(fmt.Sprintf("Reset %d variable(s) to default?\n\n", count))
		b.WriteString(helpKeyStyle.Render("y") + helpDescStyle.Render("Yes, reset all selected") + "\n")
		b.WriteString(helpKeyStyle.Render("n") + helpDescStyle.Render("No, go back") + "\n")
		return helpBoxStyle.Render(b.String())
	}

	b.WriteString("Select variables to reset:\n\n")

	// Show modified variables
	for i, ev := range modified {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		checkbox := "[ ]"
		if m.selected[ev.Key] {
			checkbox = "[x]"
		}

		line := fmt.Sprintf("%s%s %s", cursor, checkbox, ev.Key)
		if i == m.cursor {
			b.WriteString(suggestionSelectedStyle.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}

	// Show help
	b.WriteString("\n")
	b.WriteString(helpSectionStyle.Render("Keys") + "\n")
	b.WriteString(helpKeyStyle.Render("Space") + helpDescStyle.Render("Toggle selection") + "\n")
	b.WriteString(helpKeyStyle.Render("a/A") + helpDescStyle.Render("Select/deselect all") + "\n")
	b.WriteString(helpKeyStyle.Render("Enter") + helpDescStyle.Render("Confirm reset") + "\n")
	b.WriteString(helpKeyStyle.Render("Esc") + helpDescStyle.Render("Cancel") + "\n")

	return helpBoxStyle.Render(b.String())
}
