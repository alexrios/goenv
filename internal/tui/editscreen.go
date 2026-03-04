package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
)

// EditScreenModel provides a text input for editing a selected environment variable.
// It displays the original value and tracks length changes during editing.
// When readOnly is true, the text input is disabled and the screen shows
// variable details without allowing modification.
type EditScreenModel struct {
	textInput       textinput.Model
	editingKey      string
	originalValue   string
	suggestions     []string
	suggestionIdx   int
	showSuggestions bool
	validationError string // Current validation error message, if any
	goVersion       goenv.GoVersion
	readOnly        bool
}

// NewEditScreenModel creates a new edit screen with default text input configuration.
func NewEditScreenModel(ver goenv.GoVersion) EditScreenModel {
	ti := textinput.New()
	ti.Placeholder = inputPlaceholder
	ti.CharLimit = 0
	ti.Prompt = inputPrompt
	ti.SetStyles(textInputStyles)

	return EditScreenModel{
		textInput:       ti,
		suggestionIdx:   -1,
		showSuggestions: true,
		goVersion:       ver,
	}
}

// updateSuggestions refreshes the suggestion list based on current input.
func (m *EditScreenModel) updateSuggestions() {
	currentValue := m.textInput.Value()
	m.suggestions = goenv.FilterSuggestionsForVersion(m.editingKey, currentValue, m.goVersion)

	// Fall back to path suggestions if no standard suggestions exist
	if len(m.suggestions) == 0 {
		pathSuggestions := goenv.GetPathSuggestions(m.editingKey)
		if len(pathSuggestions) > 0 {
			if currentValue == "" {
				m.suggestions = pathSuggestions
			} else {
				for _, p := range pathSuggestions {
					if strings.HasPrefix(p, currentValue) {
						m.suggestions = append(m.suggestions, p)
					}
				}
			}
		}
	}

	// Reset selection if suggestions changed
	if m.suggestionIdx >= len(m.suggestions) {
		m.suggestionIdx = -1
	}
}

// Init initializes the edit screen by focusing the text input and starting cursor blink.
// In read-only mode, the text input is not focused and no cursor blink is started.
func (m *EditScreenModel) Init() tea.Cmd {
	m.suggestionIdx = -1
	m.showSuggestions = true
	m.validationError = ""
	if m.readOnly {
		m.textInput.Blur()
		return nil
	}
	m.textInput.Focus()
	m.updateSuggestions()
	return textinput.Blink
}

// Update handles messages for the edit screen, including Enter to save and Esc to cancel.
// In read-only mode, only Esc (back) and copy keys (y/Y) are handled.
func (m *EditScreenModel) Update(msg tea.Msg) tea.Cmd {
	if m.readOnly {
		if msg, ok := msg.(tea.KeyPressMsg); ok {
			switch msg.String() {
			case "esc":
				return sendMsg(changeScreenMsg(AppScreenList))
			case "y":
				return sendMsg(copyValueMsg{Key: m.editingKey, Value: m.originalValue})
			case "Y":
				return sendMsg(copyKeyValueMsg{Key: m.editingKey, Value: m.originalValue})
			}
		}
		return nil
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			// If a suggestion is selected, apply it first
			if m.suggestionIdx >= 0 && m.suggestionIdx < len(m.suggestions) {
				m.textInput.SetValue(m.suggestions[m.suggestionIdx])
				m.suggestionIdx = -1
				m.validationError = ""
				m.updateSuggestions()
				return nil
			}
			// Otherwise validate and save the current value
			newValue := m.textInput.Value()
			if err := goenv.ValidateEnvValueForVersion(m.editingKey, newValue, m.goVersion); err != nil {
				m.validationError = goenv.FormatValidationError(err)
				return nil
			}
			m.validationError = ""
			editedEnv := goenv.EnvVar{Key: m.editingKey, Value: newValue, Changed: true}
			return sendMsg(saveEnvRequestMsg{Env: editedEnv, OriginalValue: m.originalValue})
		case "esc":
			return sendMsg(changeScreenMsg(AppScreenList))
		case "tab":
			// Navigate to next suggestion
			if len(m.suggestions) > 0 && m.showSuggestions {
				m.suggestionIdx++
				if m.suggestionIdx >= len(m.suggestions) {
					m.suggestionIdx = 0
				}
				return nil
			}
		case "shift+tab":
			// Navigate to previous suggestion
			if len(m.suggestions) > 0 && m.showSuggestions {
				m.suggestionIdx--
				if m.suggestionIdx < 0 {
					m.suggestionIdx = len(m.suggestions) - 1
				}
				return nil
			}
		case "ctrl+n":
			// Alternative: next suggestion
			if len(m.suggestions) > 0 && m.showSuggestions {
				m.suggestionIdx++
				if m.suggestionIdx >= len(m.suggestions) {
					m.suggestionIdx = 0
				}
				return nil
			}
		case "ctrl+p":
			// Alternative: previous suggestion
			if len(m.suggestions) > 0 && m.showSuggestions {
				m.suggestionIdx--
				if m.suggestionIdx < 0 {
					m.suggestionIdx = len(m.suggestions) - 1
				}
				return nil
			}
		case "ctrl+r":
			m.textInput.SetValue(m.originalValue)
			m.validationError = ""
			m.suggestionIdx = -1
			m.updateSuggestions()
			return nil
		}
	}

	prevValue := m.textInput.Value()
	newTextInputModel, cmd := m.textInput.Update(msg)
	m.textInput = newTextInputModel

	// Update suggestions and clear validation error if the value changed
	if m.textInput.Value() != prevValue {
		m.validationError = ""
		m.updateSuggestions()
	}

	return cmd
}

// View renders the edit screen with the variable name, visual diff,
// text input, suggestions, length information, and keyboard hints.
// In read-only mode, it shows the value as plain text with description and docs.
func (m EditScreenModel) View() string {
	if m.readOnly {
		return m.readOnlyView()
	}
	currentValue := m.textInput.Value()
	currentLen := len(currentValue)
	originalLen := len(m.originalValue)

	// Build length info with diff indicator
	lengthInfo := fmt.Sprintf("Length: %d", currentLen)
	if currentLen != originalLen {
		diff := currentLen - originalLen
		if diff > 0 {
			lengthInfo = fmt.Sprintf("Length: %d (+%d)", currentLen, diff)
		} else {
			lengthInfo = fmt.Sprintf("Length: %d (%d)", currentLen, diff)
		}
	}

	// Build visual diff section
	diffSection := m.renderDiff(m.originalValue, currentValue)

	// Build suggestions section
	suggestionSection := m.renderSuggestions()

	// Build help text with suggestion hints if applicable
	helpText := editHelpText
	if len(m.suggestions) > 0 && m.showSuggestions {
		helpText = "(Tab/Shift+Tab: cycle suggestions, Enter: accept/save, Esc: cancel)"
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("Editing: %s", m.editingKey))
	parts = append(parts, diffSection)
	parts = append(parts, m.textInput.View())
	if m.validationError != "" {
		parts = append(parts, editValidationErrorStyle.Render(m.validationError))
	}
	if suggestionSection != "" {
		parts = append(parts, suggestionSection)
	}

	// CSV value breakdown
	if csvSection := m.renderCSVBreakdown(); csvSection != "" {
		parts = append(parts, csvSection)
	}

	// Extended documentation
	if doc := goenv.GetExtendedDoc(m.editingKey); doc != "" {
		parts = append(parts, editDocStyle.Render(doc))
	}

	parts = append(parts, editLengthInfoStyle.Render(lengthInfo))
	parts = append(parts, helpText)

	return strings.Join(parts, "\n\n")
}

// renderSuggestions renders the suggestion list.
func (m EditScreenModel) renderSuggestions() string {
	if !m.showSuggestions || len(m.suggestions) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, suggestionHintStyle.Render("Suggestions:"))

	// Show up to 8 suggestions
	maxShow := 8
	if len(m.suggestions) < maxShow {
		maxShow = len(m.suggestions)
	}

	for i := range maxShow {
		suggestion := m.suggestions[i]
		if i == m.suggestionIdx {
			lines = append(lines, suggestionSelectedStyle.Render("\u25B6 "+suggestion))
		} else {
			lines = append(lines, suggestionStyle.Render("  "+suggestion))
		}
	}

	if len(m.suggestions) > maxShow {
		lines = append(lines, suggestionHintStyle.Render(fmt.Sprintf("  ... and %d more", len(m.suggestions)-maxShow)))
	}

	return strings.Join(lines, "\n")
}

// renderCSVBreakdown renders a parsed list of comma-separated items for CSV variables.
func (m EditScreenModel) renderCSVBreakdown() string {
	if !goenv.IsCSVVariable(m.editingKey) {
		return ""
	}

	currentValue := m.textInput.Value()
	if currentValue == "" {
		return ""
	}

	items := strings.Split(currentValue, ",")
	if len(items) <= 1 {
		return "" // Only show breakdown for multiple items
	}

	var lines []string
	lines = append(lines, suggestionHintStyle.Render("Current items:"))
	num := 0
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			num++
			lines = append(lines, editLengthInfoStyle.Render(fmt.Sprintf("  %d. %s", num, item)))
		}
	}
	return strings.Join(lines, "\n")
}

// readOnlyView renders the read-only detail view for a variable.
func (m EditScreenModel) readOnlyView() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("%s %s (read-only)", IconReadOnly, m.editingKey))

	parts = append(parts, editOriginalValueStyle.Render(fmt.Sprintf("Value: %s", m.originalValue)))

	if desc := goenv.GetEnvVarDescription(m.editingKey); desc != "" {
		parts = append(parts, editDocStyle.Render(desc))
	}

	if doc := goenv.GetExtendedDoc(m.editingKey); doc != "" {
		parts = append(parts, editDocStyle.Render(doc))
	}

	parts = append(parts, editLengthInfoStyle.Render(fmt.Sprintf("Length: %d", len(m.originalValue))))

	parts = append(parts, readOnlyHelpText)

	return strings.Join(parts, "\n\n")
}

// renderDiff creates a visual comparison between original and new values.
func (m EditScreenModel) renderDiff(original, current string) string {
	if original == current {
		// No change - just show the current value
		return editOriginalValueStyle.Render(fmt.Sprintf("Current: %s", original))
	}

	// Values are different - show side-by-side comparison
	oldLabel := editDiffOldStyle.Render("Old: ")
	oldValue := editDiffOldStyle.Render(truncateForDiff(original, 40))

	newLabel := editDiffNewStyle.Render("New: ")
	newValue := editDiffChangedStyle.Render(truncateForDiff(current, 40))

	return fmt.Sprintf("%s%s\n%s%s",
		oldLabel, oldValue,
		newLabel, newValue,
	)
}
