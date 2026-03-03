package tui

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

// ImportScreenModel provides an interface for importing environment snapshots.
type ImportScreenModel struct {
	textInput    textinput.Model
	items        []goenv.EnvVar
	phase        importPhase
	snapshot     persist.Snapshot
	diff         persist.SnapshotDiff
	selectedVars map[string]bool // Variables selected for import
	diffKeys     []string        // Sorted keys for display
	scrollOffset int
	cursorPos    int
	err          error
}

type importPhase int

const (
	importPhasePath importPhase = iota
	importPhasePreview
)

// NewImportScreenModel creates a new import screen.
func NewImportScreenModel() ImportScreenModel {
	ti := textinput.New()
	ti.Placeholder = "/path/to/snapshot.json"
	ti.CharLimit = 256
	ti.Prompt = "> "
	ti.SetStyles(textInputStyles)

	return ImportScreenModel{
		textInput:    ti,
		phase:        importPhasePath,
		selectedVars: make(map[string]bool),
	}
}

// Init initializes the import screen.
func (m *ImportScreenModel) Init() tea.Cmd {
	m.textInput.Focus()
	m.textInput.SetValue("")
	m.phase = importPhasePath
	m.snapshot = persist.Snapshot{}
	m.diff = persist.SnapshotDiff{}
	m.selectedVars = make(map[string]bool)
	m.diffKeys = nil
	m.scrollOffset = 0
	m.cursorPos = 0
	m.err = nil
	return textinput.Blink
}

// SetItems sets the current items for comparison.
func (m *ImportScreenModel) SetItems(items []goenv.EnvVar) {
	m.items = items
}

// Update handles messages for the import screen.
func (m *ImportScreenModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if m.phase == importPhasePath {
				path := strings.TrimSpace(m.textInput.Value())
				if path == "" {
					return nil
				}
				m.err = nil
				return m.loadSnapshot(path)
			}
			// Preview phase - apply selected variables
			return m.applySelected()
		case "esc":
			if m.phase == importPhasePreview {
				// Go back to path input
				m.phase = importPhasePath
				return nil
			}
			return sendMsg(changeScreenMsg(AppScreenList))
		case "q":
			if m.phase == importPhasePreview {
				return sendMsg(changeScreenMsg(AppScreenList))
			}
		case "space":
			// Toggle selection in preview phase
			if m.phase == importPhasePreview && len(m.diffKeys) > 0 {
				key := m.diffKeys[m.cursorPos]
				m.selectedVars[key] = !m.selectedVars[key]
				return nil
			}
		case "a":
			// Select all in preview phase
			if m.phase == importPhasePreview {
				for _, key := range m.diffKeys {
					m.selectedVars[key] = true
				}
				return nil
			}
		case "n":
			// Deselect all in preview phase
			if m.phase == importPhasePreview {
				for _, key := range m.diffKeys {
					m.selectedVars[key] = false
				}
				return nil
			}
		case "j", "down":
			if m.phase == importPhasePreview && m.cursorPos < len(m.diffKeys)-1 {
				m.cursorPos++
				maxShow := 10
				if m.cursorPos >= m.scrollOffset+maxShow {
					m.scrollOffset = m.cursorPos - maxShow + 1
				}
				return nil
			}
		case "k", "up":
			if m.phase == importPhasePreview && m.cursorPos > 0 {
				m.cursorPos--
				if m.cursorPos < m.scrollOffset {
					m.scrollOffset = m.cursorPos
				}
				return nil
			}
		}
	case loadSnapshotSuccessMsg:
		m.snapshot = msg.Snapshot
		m.diff = persist.CompareWithSnapshot(m.items, msg.Snapshot)
		m.buildDiffKeys()
		m.selectAllModified()
		m.phase = importPhasePreview
		return nil
	case loadSnapshotErrorMsg:
		m.err = msg.Err
		return nil
	}

	if m.phase == importPhasePath {
		newTextInputModel, cmd := m.textInput.Update(msg)
		m.textInput = newTextInputModel
		return cmd
	}

	return nil
}

// buildDiffKeys creates a sorted list of keys that have differences.
func (m *ImportScreenModel) buildDiffKeys() {
	m.diffKeys = nil

	// Add modified keys first
	for key := range m.diff.Modified {
		m.diffKeys = append(m.diffKeys, key)
	}

	// Add keys that would be added (in snapshot but not current)
	for key := range m.diff.Added {
		m.diffKeys = append(m.diffKeys, key)
	}

	slices.Sort(m.diffKeys)
}

// selectAllModified selects all modified variables by default.
func (m *ImportScreenModel) selectAllModified() {
	m.selectedVars = make(map[string]bool)
	for key := range m.diff.Modified {
		m.selectedVars[key] = true
	}
}

// loadSnapshot creates a command to load a snapshot file.
func (m *ImportScreenModel) loadSnapshot(path string) tea.Cmd {
	return func() tea.Msg {
		snapshot, err := persist.ImportSnapshot(path)
		if err != nil {
			return loadSnapshotErrorMsg{Err: err}
		}
		return loadSnapshotSuccessMsg{Snapshot: snapshot}
	}
}

// applySelected applies the selected variables.
func (m *ImportScreenModel) applySelected() tea.Cmd {
	var varsToApply []goenv.EnvVar
	for key, selected := range m.selectedVars {
		if !selected {
			continue
		}
		if val, ok := m.snapshot.Variables[key]; ok {
			varsToApply = append(varsToApply, goenv.EnvVar{Key: key, Value: val, Changed: true})
		}
	}

	if len(varsToApply) == 0 {
		return sendMsg(changeScreenMsg(AppScreenList))
	}

	return sendMsg(importApplyMsg{Variables: varsToApply})
}

// View renders the import screen.
func (m ImportScreenModel) View() string {
	var parts []string

	parts = append(parts, helpTitleStyle.Render("Import Environment Snapshot"))
	if m.phase == importPhasePath {
		parts = append(parts, suggestionHintStyle.Render("Step 1/2: Select file"))
	} else {
		parts = append(parts, suggestionHintStyle.Render("Step 2/2: Review changes"))
	}
	parts = append(parts, "")

	if m.phase == importPhasePath {
		parts = append(parts, "Enter path to snapshot file:")
		parts = append(parts, m.textInput.View())
		if m.err != nil {
			parts = append(parts, "")
			parts = append(parts, errorMessageStyle.Render(m.err.Error()))
		}
		parts = append(parts, "")
		parts = append(parts, suggestionHintStyle.Render("Tip: Use Tab to autocomplete file paths"))
		parts = append(parts, "")
		parts = append(parts, helpFooterStyle.Render("(Enter to load, Esc to cancel)"))
	} else {
		// Preview phase
		parts = append(parts, m.renderPreview())
		parts = append(parts, "")
		parts = append(parts, helpFooterStyle.Render("(Space: toggle, a: all, n: none, Enter: apply, Esc: back)"))
	}

	return strings.Join(parts, "\n")
}

// renderPreview renders the diff preview.
func (m ImportScreenModel) renderPreview() string {
	var lines []string

	if m.snapshot.Name != "" {
		lines = append(lines, fmt.Sprintf("Snapshot: %s", m.snapshot.Name))
	}
	if m.snapshot.Description != "" {
		lines = append(lines, fmt.Sprintf("Description: %s", m.snapshot.Description))
	}
	if m.snapshot.GoVersion != "" {
		lines = append(lines, fmt.Sprintf("Go Version: %s", m.snapshot.GoVersion))
	}
	lines = append(lines, "")

	// Summary
	modCount := len(m.diff.Modified)
	addCount := len(m.diff.Added)
	selectedCount := 0
	for _, selected := range m.selectedVars {
		if selected {
			selectedCount++
		}
	}

	lines = append(lines, fmt.Sprintf("Changes: %d modified, %d new | Selected: %d",
		modCount, addCount, selectedCount))
	lines = append(lines, "")

	if len(m.diffKeys) == 0 {
		lines = append(lines, suggestionHintStyle.Render("No changes to apply"))
		return strings.Join(lines, "\n")
	}

	// Show diff items
	maxShow := 10
	start := m.scrollOffset
	end := start + maxShow
	if end > len(m.diffKeys) {
		end = len(m.diffKeys)
	}

	for i := start; i < end; i++ {
		key := m.diffKeys[i]
		selected := m.selectedVars[key]

		checkbox := "[ ]"
		if selected {
			checkbox = "[x]"
		}

		cursor := "  "
		if i == m.cursorPos {
			cursor = "> "
		}

		var detail string
		if mod, ok := m.diff.Modified[key]; ok {
			detail = fmt.Sprintf("%s -> %s",
				truncateForDiff(mod.Current, 20),
				truncateForDiff(mod.Snapshot, 20))
		} else if val, ok := m.diff.Added[key]; ok {
			detail = fmt.Sprintf("(new) %s", truncateForDiff(val, 30))
		}

		line := fmt.Sprintf("%s%s %s: %s", cursor, checkbox, key, detail)
		if i == m.cursorPos {
			lines = append(lines, suggestionSelectedStyle.Render(line))
		} else {
			lines = append(lines, suggestionStyle.Render(line))
		}
	}

	if len(m.diffKeys) > maxShow {
		lines = append(lines, suggestionHintStyle.Render(
			fmt.Sprintf("  ... showing %d-%d of %d", start+1, end, len(m.diffKeys))))
	}

	return strings.Join(lines, "\n")
}
