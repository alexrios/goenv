package tui

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

// DiffStatus represents the status of a variable in a comparison.
type DiffStatus int

const (
	DiffUnchanged DiffStatus = iota
	DiffModified
	DiffAdded
	DiffRemoved
)

// VarDiff represents a single variable difference.
type VarDiff struct {
	Key      string
	Current  string
	Snapshot string
	Status   DiffStatus
}

// CompareScreenModel provides an interface for comparing environments.
type CompareScreenModel struct {
	items        []goenv.EnvVar
	sources      []CompareSource
	selectedIdx  int
	phase        comparePhase
	diff         []VarDiff
	scrollOffset int
	cursorPos    int
	firstSource  *CompareSource // First source for snapshot-to-snapshot comparison
	compareLabel string         // Label describing the comparison
	warning      string         // Non-fatal warning (e.g. corrupted files skipped)
}

type comparePhase int

const (
	comparePhaseSelect       comparePhase = iota
	comparePhaseSelectSecond              // Selecting second source for snapshot-to-snapshot comparison
	comparePhaseView
)

// CompareSource represents a source to compare against.
type CompareSource struct {
	Name     string
	Type     string // "snapshot" or "preset"
	Snapshot persist.Snapshot
}

// NewCompareScreenModel creates a new compare screen.
func NewCompareScreenModel() CompareScreenModel {
	return CompareScreenModel{
		phase: comparePhaseSelect,
	}
}

// Init initializes the compare screen.
func (m *CompareScreenModel) Init() tea.Cmd {
	m.phase = comparePhaseSelect
	m.selectedIdx = 0
	m.diff = nil
	m.scrollOffset = 0
	m.cursorPos = 0
	m.warning = ""
	return m.loadSources()
}

// SetItems sets the current items for comparison.
func (m *CompareScreenModel) SetItems(items []goenv.EnvVar) {
	m.items = items
}

// loadSources loads all available snapshots and presets.
func (m *CompareScreenModel) loadSources() tea.Cmd {
	return func() tea.Msg {
		var sources []CompareSource

		var totalSkipped int

		// Load snapshots
		snapshots, skippedSnap, _ := persist.ListSnapshots()
		totalSkipped += skippedSnap
		for _, s := range snapshots {
			sources = append(sources, CompareSource{
				Name:     s.Name,
				Type:     "snapshot",
				Snapshot: s,
			})
		}

		// Load presets
		presets, skippedPre, _ := persist.ListPresets()
		totalSkipped += skippedPre
		for _, p := range presets {
			sources = append(sources, CompareSource{
				Name:     p.Name,
				Type:     "preset",
				Snapshot: p.Snapshot,
			})
		}

		return compareSourcesLoadedMsg{Sources: sources, Skipped: totalSkipped}
	}
}

// Update handles messages for the compare screen.
func (m *CompareScreenModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch m.phase {
		case comparePhaseSelect:
			return m.handleSelectKeys(msg)
		case comparePhaseSelectSecond:
			return m.handleSelectSecondKeys(msg)
		case comparePhaseView:
			return m.handleViewKeys(msg)
		}
	case compareSourcesLoadedMsg:
		m.sources = msg.Sources
		if msg.Skipped > 0 {
			m.warning = fmt.Sprintf("Skipped %d corrupted file(s)", msg.Skipped)
		}
		return nil
	}
	return nil
}

func (m *CompareScreenModel) handleSelectKeys(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q":
		return sendMsg(changeScreenMsg(AppScreenList))
	case "j", "down":
		if m.selectedIdx < len(m.sources)-1 {
			m.selectedIdx++
		}
		return nil
	case "k", "up":
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return nil
	case "enter":
		if len(m.sources) > 0 && m.selectedIdx < len(m.sources) {
			source := m.sources[m.selectedIdx]
			m.computeDiff(source.Snapshot)
			m.compareLabel = fmt.Sprintf("Current env vs %s", source.Name)
			m.phase = comparePhaseView
			m.cursorPos = 0
			m.scrollOffset = 0
		}
		return nil
	case "v":
		// Enter snapshot-to-snapshot comparison mode
		if len(m.sources) > 0 && m.selectedIdx < len(m.sources) {
			m.firstSource = new(m.sources[m.selectedIdx])
			m.phase = comparePhaseSelectSecond
			m.selectedIdx = 0
		}
		return nil
	}
	return nil
}

func (m *CompareScreenModel) handleSelectSecondKeys(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q":
		m.phase = comparePhaseSelect
		m.firstSource = nil
		return nil
	case "j", "down":
		if m.selectedIdx < len(m.sources)-1 {
			m.selectedIdx++
		}
		return nil
	case "k", "up":
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return nil
	case "enter":
		if len(m.sources) > 0 && m.selectedIdx < len(m.sources) && m.firstSource != nil {
			second := m.sources[m.selectedIdx]
			m.computeDiffBetweenSnapshots(m.firstSource.Snapshot, second.Snapshot)
			m.compareLabel = fmt.Sprintf("%s vs %s", m.firstSource.Name, second.Name)
			m.phase = comparePhaseView
			m.cursorPos = 0
			m.scrollOffset = 0
		}
		return nil
	}
	return nil
}

func (m *CompareScreenModel) handleViewKeys(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q":
		m.phase = comparePhaseSelect
		return nil
	case "j", "down":
		if m.cursorPos < len(m.diff)-1 {
			m.cursorPos++
			// Adjust scroll if needed
			maxVisible := 15
			if m.cursorPos >= m.scrollOffset+maxVisible {
				m.scrollOffset = m.cursorPos - maxVisible + 1
			}
		}
		return nil
	case "k", "up":
		if m.cursorPos > 0 {
			m.cursorPos--
			if m.cursorPos < m.scrollOffset {
				m.scrollOffset = m.cursorPos
			}
		}
		return nil
	}
	return nil
}

// computeDiff calculates the difference between current env and a snapshot.
func (m *CompareScreenModel) computeDiff(snapshot persist.Snapshot) {
	m.diff = nil

	// Build current env map
	current := make(map[string]string)
	for _, ev := range m.items {
		current[ev.Key] = ev.Value
	}

	// Check all keys
	allKeys := make(map[string]bool)
	for k := range current {
		allKeys[k] = true
	}
	for k := range snapshot.Variables {
		allKeys[k] = true
	}

	keys := slices.Sorted(maps.Keys(allKeys))

	for _, key := range keys {
		currentVal, inCurrent := current[key]
		snapshotVal, inSnapshot := snapshot.Variables[key]

		var status DiffStatus
		if inCurrent && inSnapshot {
			if currentVal == snapshotVal {
				status = DiffUnchanged
			} else {
				status = DiffModified
			}
		} else if inCurrent {
			status = DiffRemoved
		} else {
			status = DiffAdded
		}

		m.diff = append(m.diff, VarDiff{
			Key:      key,
			Current:  currentVal,
			Snapshot: snapshotVal,
			Status:   status,
		})
	}

	// Sort: modified first, then added, then removed, then unchanged
	slices.SortFunc(m.diff, func(a, b VarDiff) int {
		if a.Status != b.Status {
			// Order: Modified, Added, Removed, Unchanged
			order := map[DiffStatus]int{
				DiffModified:  0,
				DiffAdded:     1,
				DiffRemoved:   2,
				DiffUnchanged: 3,
			}
			return cmp.Compare(order[a.Status], order[b.Status])
		}
		return cmp.Compare(a.Key, b.Key)
	})
}

// computeDiffBetweenSnapshots calculates the difference between two snapshots.
func (m *CompareScreenModel) computeDiffBetweenSnapshots(a, b persist.Snapshot) {
	m.diff = nil

	allKeys := make(map[string]bool)
	for k := range a.Variables {
		allKeys[k] = true
	}
	for k := range b.Variables {
		allKeys[k] = true
	}

	keys := slices.Sorted(maps.Keys(allKeys))

	for _, key := range keys {
		aVal, inA := a.Variables[key]
		bVal, inB := b.Variables[key]

		var status DiffStatus
		if inA && inB {
			if aVal == bVal {
				status = DiffUnchanged
			} else {
				status = DiffModified
			}
		} else if inA {
			status = DiffRemoved // In first but not second
		} else {
			status = DiffAdded // In second but not first
		}

		m.diff = append(m.diff, VarDiff{
			Key:      key,
			Current:  aVal,
			Snapshot: bVal,
			Status:   status,
		})
	}

	// Sort: modified first, then added, then removed, then unchanged
	slices.SortFunc(m.diff, func(a, b VarDiff) int {
		if a.Status != b.Status {
			order := map[DiffStatus]int{
				DiffModified:  0,
				DiffAdded:     1,
				DiffRemoved:   2,
				DiffUnchanged: 3,
			}
			return cmp.Compare(order[a.Status], order[b.Status])
		}
		return cmp.Compare(a.Key, b.Key)
	})
}

// View renders the compare screen.
func (m CompareScreenModel) View() string {
	var parts []string

	parts = append(parts, helpTitleStyle.Render("Environment Comparison"))
	switch m.phase {
	case comparePhaseSelect:
		parts = append(parts, suggestionHintStyle.Render("Select source to compare"))
	case comparePhaseSelectSecond:
		parts = append(parts, suggestionHintStyle.Render("Select second source"))
	case comparePhaseView:
		parts = append(parts, suggestionHintStyle.Render("Viewing differences"))
	}
	parts = append(parts, "")

	switch m.phase {
	case comparePhaseSelect:
		if len(m.sources) == 0 {
			parts = append(parts, suggestionHintStyle.Render("No snapshots or presets available to compare."))
			parts = append(parts, "")
			parts = append(parts, "Export a snapshot (e) or create a preset (p) first.")
		} else {
			parts = append(parts, "Select a snapshot or preset to compare:")
			parts = append(parts, "")
			for i, source := range m.sources {
				cursor := "  "
				if i == m.selectedIdx {
					cursor = "> "
				}
				typeLabel := fmt.Sprintf("[%s]", source.Type)
				line := fmt.Sprintf("%s%s %s", cursor, typeLabel, source.Name)
				if i == m.selectedIdx {
					parts = append(parts, suggestionSelectedStyle.Render(line))
				} else {
					parts = append(parts, suggestionStyle.Render(line))
				}
			}
		}
		parts = append(parts, "")
		parts = append(parts, helpFooterStyle.Render("(Enter: compare vs current, v: compare two snapshots, Esc: back)"))

	case comparePhaseSelectSecond:
		if m.firstSource != nil {
			parts = append(parts, fmt.Sprintf("First: %s [%s]", m.firstSource.Name, m.firstSource.Type))
			parts = append(parts, "Select second source to compare against:")
			parts = append(parts, "")
			for i, source := range m.sources {
				cursor := "  "
				if i == m.selectedIdx {
					cursor = "> "
				}
				typeLabel := fmt.Sprintf("[%s]", source.Type)
				line := fmt.Sprintf("%s%s %s", cursor, typeLabel, source.Name)
				if i == m.selectedIdx {
					parts = append(parts, suggestionSelectedStyle.Render(line))
				} else {
					parts = append(parts, suggestionStyle.Render(line))
				}
			}
		}
		parts = append(parts, "")
		parts = append(parts, helpFooterStyle.Render("(Enter: compare, Esc: back)"))

	case comparePhaseView:
		if m.compareLabel != "" {
			parts = append(parts, suggestionHintStyle.Render(m.compareLabel))
			parts = append(parts, "")
		}
		parts = append(parts, m.renderDiffView())
		parts = append(parts, "")
		parts = append(parts, helpFooterStyle.Render("(j/k: navigate, Esc: back to selection)"))
	}

	if m.warning != "" {
		parts = append(parts, "")
		parts = append(parts, errorMessageStyle.Render(m.warning))
	}

	return strings.Join(parts, "\n")
}

// renderDiffView renders the diff view.
func (m CompareScreenModel) renderDiffView() string {
	var lines []string

	// Summary
	var modCount, addCount, remCount, unchCount int
	for _, d := range m.diff {
		switch d.Status {
		case DiffModified:
			modCount++
		case DiffAdded:
			addCount++
		case DiffRemoved:
			remCount++
		case DiffUnchanged:
			unchCount++
		}
	}

	lines = append(lines, fmt.Sprintf("Modified: %d | Added: %d | Removed: %d | Unchanged: %d",
		modCount, addCount, remCount, unchCount))
	lines = append(lines, "")

	// Diff items
	maxVisible := 15
	start := m.scrollOffset
	end := start + maxVisible
	if end > len(m.diff) {
		end = len(m.diff)
	}

	for i := start; i < end; i++ {
		d := m.diff[i]
		cursor := "  "
		if i == m.cursorPos {
			cursor = "> "
		}

		var statusIcon string
		var style lipgloss.Style
		switch d.Status {
		case DiffModified:
			statusIcon = "~"
			style = lipgloss.NewStyle().Foreground(colorWarning)
		case DiffAdded:
			statusIcon = "+"
			style = lipgloss.NewStyle().Foreground(colorSuccess)
		case DiffRemoved:
			statusIcon = "-"
			style = lipgloss.NewStyle().Foreground(colorError)
		case DiffUnchanged:
			statusIcon = " "
			style = lipgloss.NewStyle().Foreground(colorDim)
		}

		var detail string
		switch d.Status {
		case DiffModified:
			detail = fmt.Sprintf("%s: %s -> %s", d.Key,
				truncateForDiff(d.Current, 15),
				truncateForDiff(d.Snapshot, 15))
		case DiffAdded:
			detail = fmt.Sprintf("%s: (new) %s", d.Key, truncateForDiff(d.Snapshot, 25))
		case DiffRemoved:
			detail = fmt.Sprintf("%s: (removed) %s", d.Key, truncateForDiff(d.Current, 20))
		case DiffUnchanged:
			detail = fmt.Sprintf("%s: %s", d.Key, truncateForDiff(d.Current, 30))
		}

		line := fmt.Sprintf("%s%s %s", cursor, statusIcon, detail)
		lines = append(lines, style.Render(line))
	}

	if len(m.diff) > maxVisible {
		lines = append(lines, suggestionHintStyle.Render(
			fmt.Sprintf("  Showing %d-%d of %d", start+1, end, len(m.diff))))
	}

	return strings.Join(lines, "\n")
}
