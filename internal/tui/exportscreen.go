package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
)

// ExportScreenModel provides an interface for exporting environment snapshots.
type ExportScreenModel struct {
	textInput    textinput.Model
	items        []goenv.EnvVar
	snapshotName string
	phase        exportPhase
}

type exportPhase int

const (
	exportPhaseName exportPhase = iota
	exportPhaseDescription
)

// NewExportScreenModel creates a new export screen.
func NewExportScreenModel() ExportScreenModel {
	ti := textinput.New()
	ti.Placeholder = "snapshot-name"
	ti.CharLimit = 64
	ti.Prompt = "> "
	ti.SetStyles(textInputStyles)

	return ExportScreenModel{
		textInput: ti,
		phase:     exportPhaseName,
	}
}

// Init initializes the export screen.
func (m *ExportScreenModel) Init() tea.Cmd {
	m.textInput.Focus()
	m.textInput.SetValue("")
	m.snapshotName = ""
	m.phase = exportPhaseName
	m.textInput.Placeholder = "snapshot-name"
	return textinput.Blink
}

// SetItems sets the items to export.
func (m *ExportScreenModel) SetItems(items []goenv.EnvVar) {
	m.items = items
}

// Update handles messages for the export screen.
func (m *ExportScreenModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if m.phase == exportPhaseName {
				name := strings.TrimSpace(m.textInput.Value())
				if name == "" {
					return nil
				}
				// Store name and move to description phase
				m.snapshotName = name
				m.phase = exportPhaseDescription
				m.textInput.SetValue("")
				m.textInput.Placeholder = "Description (optional)"
				return nil
			}
			// Description phase - trigger export
			description := strings.TrimSpace(m.textInput.Value())
			return sendMsg(exportSnapshotMsg{
				Name:        m.snapshotName,
				Description: description,
				Items:       m.items,
			})
		case "esc":
			return sendMsg(changeScreenMsg(AppScreenList))
		}
	}

	newTextInputModel, cmd := m.textInput.Update(msg)
	m.textInput = newTextInputModel
	return cmd
}

// View renders the export screen.
func (m ExportScreenModel) View() string {
	var parts []string

	parts = append(parts, helpTitleStyle.Render("Export Environment Snapshot"))
	if m.phase == exportPhaseName {
		parts = append(parts, suggestionHintStyle.Render("Step 1/2: Name"))
	} else {
		parts = append(parts, suggestionHintStyle.Render("Step 2/2: Description"))
	}
	parts = append(parts, "")

	if m.phase == exportPhaseName {
		parts = append(parts, "Enter a name for this snapshot:")
		parts = append(parts, m.textInput.View())
		parts = append(parts, "")
		parts = append(parts, suggestionHintStyle.Render(fmt.Sprintf("Will export %d variables", len(m.items))))
	} else {
		parts = append(parts, fmt.Sprintf("Snapshot: %s", m.snapshotName))
		parts = append(parts, "")
		parts = append(parts, "Enter a description (optional):")
		parts = append(parts, m.textInput.View())
	}

	parts = append(parts, "")
	parts = append(parts, helpFooterStyle.Render("(Enter to continue, Esc to cancel)"))

	return strings.Join(parts, "\n")
}
