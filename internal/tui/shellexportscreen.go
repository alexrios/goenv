package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
)

// shellExportPhase represents the current phase of the shell export flow.
type shellExportPhase int

const (
	shellExportPhaseOptions shellExportPhase = iota // Selecting options
	shellExportPhasePreview                         // Previewing output
)

// ShellExportScreenModel provides a screen for generating shell export commands.
type ShellExportScreenModel struct {
	items        []goenv.EnvVar
	shell        goenv.ShellType
	filter       goenv.ExportFilter
	phase        shellExportPhase
	previewLines []string
	scrollOffset int
	copied       bool
	copyError    string
	maxLines     int
}

// NewShellExportScreenModel creates a new shell export screen model.
func NewShellExportScreenModel() ShellExportScreenModel {
	return ShellExportScreenModel{
		shell:    goenv.ShellBash,
		filter:   goenv.ExportModified,
		maxLines: 20,
	}
}

// SetItems sets the available items for export.
func (m *ShellExportScreenModel) SetItems(items []goenv.EnvVar) {
	m.items = items
	m.phase = shellExportPhaseOptions
	m.scrollOffset = 0
	m.copied = false
}

// Init initializes the shell export screen.
func (m *ShellExportScreenModel) Init() tea.Cmd {
	m.phase = shellExportPhaseOptions
	m.copied = false
	return nil
}

// Update handles messages for the shell export screen.
func (m *ShellExportScreenModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch m.phase {
		case shellExportPhaseOptions:
			return m.updateOptions(msg)
		case shellExportPhasePreview:
			return m.updatePreview(msg)
		}
	}
	return nil
}

// updateOptions handles key presses in the options phase.
func (m *ShellExportScreenModel) updateOptions(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q":
		return sendMsg(changeScreenMsg(AppScreenList))
	case "s", "1":
		// Toggle shell type
		m.shell = m.shell.Next()
	case "f", "2":
		// Toggle filter
		m.filter = m.filter.Next()
	case "enter", "p":
		// Generate and show preview
		output := goenv.GenerateShellExport(m.items, m.shell, m.filter)
		m.previewLines = strings.Split(output, "\n")
		m.scrollOffset = 0
		m.phase = shellExportPhasePreview
		m.copied = false
		m.copyError = ""
	}
	return nil
}

// updatePreview handles key presses in the preview phase.
func (m *ShellExportScreenModel) updatePreview(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "b":
		// Back to options
		m.phase = shellExportPhaseOptions
		m.copied = false
	case "q":
		return sendMsg(changeScreenMsg(AppScreenList))
	case "j", "down":
		// Scroll down
		if m.scrollOffset < len(m.previewLines)-m.maxLines {
			m.scrollOffset++
		}
	case "k", "up":
		// Scroll up
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
	case "c", "y":
		// Copy to clipboard using Bubble Tea's native OSC52 support
		output := strings.Join(m.previewLines, "\n")
		m.copied = true
		m.copyError = ""
		return tea.SetClipboard(output)
	case "g":
		// Go to top
		m.scrollOffset = 0
	case "G":
		// Go to bottom
		if len(m.previewLines) > m.maxLines {
			m.scrollOffset = len(m.previewLines) - m.maxLines
		}
	}
	return nil
}

// View renders the shell export screen.
func (m ShellExportScreenModel) View() string {
	var b strings.Builder

	switch m.phase {
	case shellExportPhaseOptions:
		return m.viewOptions(&b)
	case shellExportPhasePreview:
		return m.viewPreview(&b)
	}

	return ""
}

// viewOptions renders the options selection phase.
func (m ShellExportScreenModel) viewOptions(b *strings.Builder) string {
	b.WriteString(helpTitleStyle.Render("Shell Export") + "\n")
	b.WriteString(suggestionHintStyle.Render("Step 1/2: Configure") + "\n\n")
	b.WriteString("Generate shell commands to set Go environment variables.\n\n")

	// Count exportable variables
	count := goenv.CountExportableVars(m.items, m.filter)

	// Shell type selection
	b.WriteString(helpSectionStyle.Render("Options") + "\n")
	b.WriteString(helpKeyStyle.Render("[1/s]") + " Shell: ")
	b.WriteString(suggestionSelectedStyle.Render(m.shell.String()) + "\n")

	// Filter selection
	b.WriteString(helpKeyStyle.Render("[2/f]") + " Filter: ")
	b.WriteString(suggestionSelectedStyle.Render(m.filter.String()) + "\n")

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Variables to export: %d\n", count))

	// Help
	b.WriteString("\n" + helpSectionStyle.Render("Actions") + "\n")
	b.WriteString(helpKeyStyle.Render("Enter/p") + helpDescStyle.Render("Preview export") + "\n")
	b.WriteString(helpKeyStyle.Render("Esc/q") + helpDescStyle.Render("Cancel") + "\n")

	return helpBoxStyle.Render(b.String())
}

// viewPreview renders the preview phase.
func (m ShellExportScreenModel) viewPreview(b *strings.Builder) string {
	b.WriteString(helpTitleStyle.Render("Export Preview") + "\n")
	b.WriteString(suggestionHintStyle.Render("Step 2/2: Preview & copy") + "\n\n")

	if len(m.previewLines) == 0 {
		b.WriteString("No variables to export.\n\n")
		b.WriteString(helpKeyStyle.Render("b/Esc") + helpDescStyle.Render("Back to options") + "\n")
		return helpBoxStyle.Render(b.String())
	}

	// Show visible lines
	endIdx := m.scrollOffset + m.maxLines
	if endIdx > len(m.previewLines) {
		endIdx = len(m.previewLines)
	}

	for _, line := range m.previewLines[m.scrollOffset:endIdx] {
		if strings.HasPrefix(line, "#") {
			b.WriteString(suggestionHintStyle.Render(line) + "\n")
		} else if strings.HasPrefix(line, "export ") || strings.HasPrefix(line, "set -x ") || strings.HasPrefix(line, "$env:") {
			// Highlight the key name
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				b.WriteString(helpKeyStyle.Render(parts[0]+" ") + parts[1] + "\n")
			} else {
				b.WriteString(line + "\n")
			}
		} else {
			b.WriteString(line + "\n")
		}
	}

	// Scroll indicator
	if len(m.previewLines) > m.maxLines {
		b.WriteString("\n")
		b.WriteString(suggestionHintStyle.Render(
			fmt.Sprintf("Lines %d-%d of %d (j/k to scroll, g/G top/bottom)",
				m.scrollOffset+1, endIdx, len(m.previewLines))))
		b.WriteString("\n")
	}

	// Status
	b.WriteString("\n")
	if m.copied {
		b.WriteString(statusSuccessStyle.Render("Copied to clipboard!") + "\n")
	} else if m.copyError != "" {
		b.WriteString(statusErrorStyle.Render(m.copyError) + "\n")
	}

	// Help
	b.WriteString("\n" + helpSectionStyle.Render("Actions") + "\n")
	b.WriteString(helpKeyStyle.Render("c/y") + helpDescStyle.Render("Copy to clipboard") + "\n")
	b.WriteString(helpKeyStyle.Render("b/Esc") + helpDescStyle.Render("Back to options") + "\n")
	b.WriteString(helpKeyStyle.Render("q") + helpDescStyle.Render("Close") + "\n")

	return helpBoxStyle.Render(b.String())
}
