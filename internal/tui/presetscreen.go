package tui

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

// PresetScreenModel provides an interface for managing presets.
type PresetScreenModel struct {
	presets       []persist.Preset
	items         []goenv.EnvVar
	selectedIdx   int
	phase         presetPhase
	textInput     textinput.Model
	presetName    string
	err           error
	warning       string // Non-fatal warning (e.g. corrupted files skipped)
	previewDiff   []VarDiff
	previewPreset persist.Preset
	previewScroll int
}

type presetPhase int

const (
	presetPhaseList presetPhase = iota
	presetPhaseCreate
	presetPhaseConfirmDelete
	presetPhasePreview
)

// NewPresetScreenModel creates a new preset screen.
func NewPresetScreenModel() PresetScreenModel {
	ti := textinput.New()
	ti.Placeholder = "preset-name"
	ti.CharLimit = 64
	ti.Prompt = "> "
	ti.SetStyles(textInputStyles)

	return PresetScreenModel{
		textInput: ti,
		phase:     presetPhaseList,
	}
}

// Init initializes the preset screen.
func (m *PresetScreenModel) Init() tea.Cmd {
	m.phase = presetPhaseList
	m.selectedIdx = 0
	m.err = nil
	m.warning = ""
	return m.loadPresets()
}

// SetItems sets the current items for creating new presets.
func (m *PresetScreenModel) SetItems(items []goenv.EnvVar) {
	m.items = items
}

// loadPresets loads all presets from disk and appends built-in presets.
func (m *PresetScreenModel) loadPresets() tea.Cmd {
	return func() tea.Msg {
		presets, skipped, err := persist.ListPresets()
		if err != nil {
			return presetLoadErrorMsg{Err: err}
		}
		presets = append(presets, persist.BuiltinPresets()...)
		return presetLoadSuccessMsg{Presets: presets, Skipped: skipped}
	}
}

// Update handles messages for the preset screen.
func (m *PresetScreenModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch m.phase {
		case presetPhaseList:
			return m.handleListKeys(msg)
		case presetPhaseCreate:
			switch msg.String() {
			case "esc":
				m.phase = presetPhaseList
				return nil
			case "enter":
				name := strings.TrimSpace(m.textInput.Value())
				if name == "" {
					return nil
				}
				m.presetName = name
				return m.createPreset(name)
			}
			// Other keys fall through to textInput.Update below
		case presetPhaseConfirmDelete:
			return m.handleDeleteKeys(msg)
		case presetPhasePreview:
			return m.handlePreviewKeys(msg)
		}
	case presetLoadSuccessMsg:
		m.presets = msg.Presets
		if m.selectedIdx >= len(m.presets) {
			m.selectedIdx = max(0, len(m.presets)-1)
		}
		if msg.Skipped > 0 {
			m.warning = fmt.Sprintf("Skipped %d corrupted preset file(s)", msg.Skipped)
		}
		return nil
	case presetLoadErrorMsg:
		m.err = msg.Err
		return nil
	}

	if m.phase == presetPhaseCreate {
		newTextInputModel, cmd := m.textInput.Update(msg)
		m.textInput = newTextInputModel
		return cmd
	}

	return nil
}

func (m *PresetScreenModel) handleListKeys(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q":
		return sendMsg(changeScreenMsg(AppScreenList))
	case "j", "down":
		if m.selectedIdx < len(m.presets)-1 {
			m.selectedIdx++
		}
		return nil
	case "k", "up":
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return nil
	case "enter":
		if len(m.presets) > 0 && m.selectedIdx < len(m.presets) {
			preset := m.presets[m.selectedIdx]
			m.previewPreset = preset
			m.previewDiff = m.computePresetDiff(preset)
			m.previewScroll = 0
			m.phase = presetPhasePreview
		}
		return nil
	case "n":
		m.phase = presetPhaseCreate
		m.textInput.SetValue("")
		m.textInput.Focus()
		return textinput.Blink
	case "d":
		if len(m.presets) > 0 && m.selectedIdx < len(m.presets) {
			if m.presets[m.selectedIdx].BuiltIn {
				return nil // Cannot delete built-in presets
			}
			m.phase = presetPhaseConfirmDelete
		}
		return nil
	}
	return nil
}

func (m *PresetScreenModel) handleDeleteKeys(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "y", "Y":
		if m.selectedIdx < len(m.presets) {
			preset := m.presets[m.selectedIdx]
			return m.deletePreset(preset.Name)
		}
		return nil
	case "n", "N", "esc":
		m.phase = presetPhaseList
		return nil
	}
	return nil
}

func (m *PresetScreenModel) handlePreviewKeys(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		return sendMsg(presetApplyMsg{Preset: m.previewPreset})
	case "esc":
		m.phase = presetPhaseList
		return nil
	case "q":
		return sendMsg(changeScreenMsg(AppScreenList))
	case "j", "down":
		if m.previewScroll < len(m.previewDiff)-1 {
			m.previewScroll++
		}
		return nil
	case "k", "up":
		if m.previewScroll > 0 {
			m.previewScroll--
		}
		return nil
	}
	return nil
}

// computePresetDiff compares current items against a preset's variables.
func (m *PresetScreenModel) computePresetDiff(preset persist.Preset) []VarDiff {
	currentVars := make(map[string]string)
	for _, ev := range m.items {
		currentVars[ev.Key] = ev.Value
	}

	var diffs []VarDiff
	for key, presetVal := range preset.Variables {
		current, exists := currentVars[key]
		if !exists {
			diffs = append(diffs, VarDiff{Key: key, Current: "", Snapshot: presetVal, Status: DiffAdded})
		} else if current != presetVal {
			diffs = append(diffs, VarDiff{Key: key, Current: current, Snapshot: presetVal, Status: DiffModified})
		}
	}
	slices.SortFunc(diffs, func(a, b VarDiff) int {
		// Modified first, then added
		if a.Status != b.Status {
			if a.Status == DiffModified {
				return -1
			}
			return 1
		}
		return cmp.Compare(a.Key, b.Key)
	})
	return diffs
}

func (m *PresetScreenModel) createPreset(name string) tea.Cmd {
	items := m.items
	return func() tea.Msg {
		preset := persist.Preset{
			Snapshot: persist.NewSnapshot(name, items, ""),
		}

		dir, err := persist.PresetsDir()
		if err != nil {
			return presetCreateErrorMsg{Err: err}
		}

		base := persist.SanitizeFilename(name)
		filePath := persist.UniqueFilePath(dir, base, ".json")

		if err := persist.ExportSnapshot(preset.Snapshot, filePath); err != nil {
			return presetCreateErrorMsg{Err: err}
		}

		return presetCreateSuccessMsg{Name: name}
	}
}

func (m *PresetScreenModel) deletePreset(name string) tea.Cmd {
	return func() tea.Msg {
		dir, err := persist.PresetsDir()
		if err != nil {
			return presetDeleteErrorMsg{Err: err}
		}

		filename := persist.SanitizeFilename(name) + ".json"
		filePath := filepath.Join(dir, filename)

		if err := os.Remove(filePath); err != nil {
			return presetDeleteErrorMsg{Err: err}
		}

		return presetDeleteSuccessMsg{Name: name}
	}
}

// View renders the preset screen.
func (m PresetScreenModel) View() string {
	var parts []string

	parts = append(parts, helpTitleStyle.Render("Presets"))
	switch m.phase {
	case presetPhaseList:
		parts = append(parts, suggestionHintStyle.Render("Step 1/2: Select preset"))
	case presetPhasePreview:
		parts = append(parts, suggestionHintStyle.Render("Step 2/2: Review changes"))
	case presetPhaseCreate:
		parts = append(parts, suggestionHintStyle.Render("Create new preset"))
	case presetPhaseConfirmDelete:
		parts = append(parts, suggestionHintStyle.Render("Confirm deletion"))
	}
	parts = append(parts, "")

	switch m.phase {
	case presetPhaseList:
		if len(m.presets) == 0 {
			parts = append(parts, suggestionHintStyle.Render("No presets saved yet."))
			parts = append(parts, "")
			parts = append(parts, "Press 'n' to create a new preset from current environment.")
		} else {
			for i, preset := range m.presets {
				cursor := "  "
				if i == m.selectedIdx {
					cursor = "> "
				}
				line := fmt.Sprintf("%s%s", cursor, preset.Name)
				if preset.BuiltIn {
					line += " [built-in]"
				}
				if preset.Description != "" {
					line += fmt.Sprintf(" - %s", truncateString(preset.Description, 40))
				}
				if i == m.selectedIdx {
					parts = append(parts, suggestionSelectedStyle.Render(line))
				} else {
					parts = append(parts, suggestionStyle.Render(line))
				}
			}
		}
		parts = append(parts, "")
		parts = append(parts, helpFooterStyle.Render("(Enter: apply, n: new, d: delete, Esc: back)"))

	case presetPhaseCreate:
		parts = append(parts, "Enter a name for the new preset:")
		parts = append(parts, m.textInput.View())
		parts = append(parts, "")
		parts = append(parts, helpFooterStyle.Render("(Enter: create, Esc: cancel)"))

	case presetPhasePreview:
		parts = append(parts, fmt.Sprintf("Preset: %s", suggestionSelectedStyle.Render(m.previewPreset.Name)))
		parts = append(parts, "")
		if len(m.previewDiff) == 0 {
			parts = append(parts, suggestionHintStyle.Render("No changes \u2014 environment already matches this preset."))
		} else {
			modCount, addCount := 0, 0
			for _, d := range m.previewDiff {
				switch d.Status {
				case DiffModified:
					modCount++
				case DiffAdded:
					addCount++
				}
			}
			summary := fmt.Sprintf("%d modified", modCount)
			if addCount > 0 {
				summary += fmt.Sprintf(", %d new", addCount)
			}
			parts = append(parts, suggestionHintStyle.Render(summary))
			parts = append(parts, "")

			maxVisible := 15
			endIdx := m.previewScroll + maxVisible
			if endIdx > len(m.previewDiff) {
				endIdx = len(m.previewDiff)
			}
			for _, d := range m.previewDiff[m.previewScroll:endIdx] {
				switch d.Status {
				case DiffModified:
					parts = append(parts, fmt.Sprintf("  %s %s: %s -> %s",
						editDiffChangedStyle.Render("~"),
						d.Key,
						editDiffOldStyle.Render(truncateString(d.Current, 30)),
						editDiffNewStyle.Render(truncateString(d.Snapshot, 30)),
					))
				case DiffAdded:
					parts = append(parts, fmt.Sprintf("  %s %s: %s",
						editDiffNewStyle.Render("+"),
						d.Key,
						editDiffNewStyle.Render(truncateString(d.Snapshot, 30)),
					))
				}
			}
			if len(m.previewDiff) > maxVisible {
				parts = append(parts, "")
				parts = append(parts, suggestionHintStyle.Render(
					fmt.Sprintf("Showing %d-%d of %d (j/k to scroll)", m.previewScroll+1, endIdx, len(m.previewDiff))))
			}
		}
		parts = append(parts, "")
		parts = append(parts, helpFooterStyle.Render("(Enter: apply, Esc: back, q: cancel)"))

	case presetPhaseConfirmDelete:
		if m.selectedIdx < len(m.presets) {
			parts = append(parts, fmt.Sprintf("Delete preset '%s'?", m.presets[m.selectedIdx].Name))
			parts = append(parts, "")
			parts = append(parts, helpFooterStyle.Render("(y: yes, n: no)"))
		}
	}

	if m.err != nil {
		parts = append(parts, "")
		parts = append(parts, errorMessageStyle.Render(m.err.Error()))
	}
	if m.warning != "" {
		parts = append(parts, "")
		parts = append(parts, errorMessageStyle.Render(m.warning))
	}

	return strings.Join(parts, "\n")
}
