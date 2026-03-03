package tui

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

// StatsScreenModel displays statistics about environment variables.
type StatsScreenModel struct {
	items       []goenv.EnvVar
	favorites   map[string]bool
	editHistory []persist.EditRecord
	stats       EnvStats
}

// EnvStats holds computed statistics about the environment.
type EnvStats struct {
	TotalVars      int
	ModifiedCount  int
	FavoritesCount int
	ByCategory     map[goenv.Category]int
	RecentEdits    []persist.EditRecord
}

// NewStatsScreenModel creates a new stats screen.
func NewStatsScreenModel() StatsScreenModel {
	return StatsScreenModel{}
}

// Init initializes the stats screen.
func (m *StatsScreenModel) Init() tea.Cmd {
	m.computeStats()
	return nil
}

// SetData sets the data needed for stats computation.
func (m *StatsScreenModel) SetData(items []goenv.EnvVar, favorites map[string]bool, history []persist.EditRecord) {
	m.items = items
	m.favorites = favorites
	m.editHistory = history
}

// computeStats calculates statistics from the current data.
func (m *StatsScreenModel) computeStats() {
	m.stats = EnvStats{
		ByCategory: make(map[goenv.Category]int),
	}

	for _, ev := range m.items {
		m.stats.TotalVars++

		if ev.Changed {
			m.stats.ModifiedCount++
		}

		if m.favorites[ev.Key] {
			m.stats.FavoritesCount++
		}

		cat := goenv.GetEnvVarCategory(ev.Key)
		m.stats.ByCategory[cat]++
	}

	// Get recent edits (last 5)
	maxRecent := 5
	if len(m.editHistory) < maxRecent {
		maxRecent = len(m.editHistory)
	}
	if maxRecent > 0 {
		m.stats.RecentEdits = m.editHistory[len(m.editHistory)-maxRecent:]
		// Reverse to show most recent first
		slices.Reverse(m.stats.RecentEdits)
	}
}

// Update handles messages for the stats screen.
func (m *StatsScreenModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "q", "S":
			return sendMsg(changeScreenMsg(AppScreenList))
		}
	}
	return nil
}

// View renders the stats screen.
func (m StatsScreenModel) View() string {
	var parts []string

	parts = append(parts, helpTitleStyle.Render("Environment Statistics"))
	parts = append(parts, "")

	// Overview
	parts = append(parts, helpSectionStyle.Render("Overview"))
	parts = append(parts, fmt.Sprintf("  Total Variables: %d", m.stats.TotalVars))
	parts = append(parts, fmt.Sprintf("  Modified:        %d", m.stats.ModifiedCount))
	parts = append(parts, fmt.Sprintf("  Favorites:       %d", m.stats.FavoritesCount))
	parts = append(parts, "")

	// Category breakdown
	parts = append(parts, helpSectionStyle.Render("By Category"))

	// Sort categories by order
	categories := slices.Collect(maps.Keys(m.stats.ByCategory))
	slices.SortFunc(categories, func(a, b goenv.Category) int {
		return cmp.Compare(goenv.CategoryIndex(a), goenv.CategoryIndex(b))
	})

	for _, cat := range categories {
		count := m.stats.ByCategory[cat]
		parts = append(parts, fmt.Sprintf("  %-15s %d", string(cat)+":", count))
	}
	parts = append(parts, "")

	// Recent edits
	if len(m.stats.RecentEdits) > 0 {
		parts = append(parts, helpSectionStyle.Render("Recent Edits"))
		for _, edit := range m.stats.RecentEdits {
			parts = append(parts, fmt.Sprintf("  %s: %s -> %s",
				edit.Key,
				truncateForDiff(edit.OldValue, 15),
				truncateForDiff(edit.NewValue, 15)))
		}
		parts = append(parts, "")
	}

	parts = append(parts, helpFooterStyle.Render("(Press Esc or S to return)"))

	return strings.Join(parts, "\n")
}
