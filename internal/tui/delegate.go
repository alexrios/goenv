package tui

import (
	"fmt"
	"io"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

// envVarDelegate is a custom list delegate that shows the documentation
// description as a third line only for the selected item.
type envVarDelegate struct {
	list.DefaultDelegate
	sortMode      persist.SortMode
	favorites     map[string]bool
	recentChanges map[string]time.Time
}

// NewEnvVarDelegate creates a new delegate for rendering environment variables.
func NewEnvVarDelegate() envVarDelegate {
	d := list.NewDefaultDelegate()
	return envVarDelegate{DefaultDelegate: d}
}

// NewEnvVarDelegateWithSortMode creates a delegate that renders category headers in category sort mode.
func NewEnvVarDelegateWithSortMode(sortMode persist.SortMode, favorites map[string]bool) envVarDelegate {
	d := list.NewDefaultDelegate()
	return envVarDelegate{DefaultDelegate: d, sortMode: sortMode, favorites: favorites}
}

// NewEnvVarDelegateWithWatch creates a delegate that includes watch mode change highlighting.
func NewEnvVarDelegateWithWatch(sortMode persist.SortMode, favorites map[string]bool, recentChanges map[string]time.Time) envVarDelegate {
	d := list.NewDefaultDelegate()
	return envVarDelegate{DefaultDelegate: d, sortMode: sortMode, favorites: favorites, recentChanges: recentChanges}
}

// Height returns the height of each list item.
// Selected items with descriptions need 3 lines, others need 2.
func (d envVarDelegate) Height() int {
	return 3 // Reserve space for description line
}

// Spacing returns the spacing between items.
func (d envVarDelegate) Spacing() int {
	return 0
}

// Render renders a list item with optional inline description for selected items.
func (d envVarDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ei, ok := item.(envItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	width := m.Width()
	if width <= 0 {
		width = 80 // Default width
	}

	// Styles for rendering
	var titleStyle, descStyle lipgloss.Style

	if isSelected {
		titleStyle = d.Styles.SelectedTitle.Width(width)
		descStyle = d.Styles.SelectedDesc.Width(width)
	} else {
		titleStyle = d.Styles.NormalTitle.Width(width)
		descStyle = d.Styles.NormalDesc.Width(width)
	}

	// Render title (variable name) with optional category tag, read-only tag, and change indicator
	title := ei.Title()
	if ei.ReadOnly {
		title = readOnlyLabelStyle.Render("[RO]") + " " + title
	}
	if tag := d.categoryTag(m, index, ei); tag != "" {
		title = categoryLabelStyle.Render(tag) + " " + title
	}
	if d.isRecentlyChanged(ei.Key) {
		title = watchChangedStyle.Render("\u25C6 ") + title
	}
	fmt.Fprint(w, titleStyle.Render(title))

	// Render value (description in list.Item terms)
	value := ei.Description()
	if value == "" {
		value = "(empty)"
	}
	// Truncate if too long
	value = truncateString(value, width-4)
	fmt.Fprint(w, "\n")
	fmt.Fprint(w, descStyle.Render(value))

	// Add doc description only for selected item
	fmt.Fprint(w, "\n")
	if isSelected {
		if desc := goenv.GetEnvVarDescription(ei.Key); desc != "" {
			// Truncate description if too long
			desc = truncateString(desc, width-4)
			fmt.Fprint(w, inlineDescStyle.Width(width).Render(desc))
		} else {
			// Empty line to maintain consistent height
			fmt.Fprint(w, "")
		}
	} else {
		// Empty line for unselected items to maintain consistent height
		fmt.Fprint(w, "")
	}
}

// isRecentlyChanged returns true if the variable changed within the last 10 seconds.
func (d envVarDelegate) isRecentlyChanged(key string) bool {
	if d.recentChanges == nil {
		return false
	}
	changedAt, ok := d.recentChanges[key]
	if !ok {
		return false
	}
	return time.Since(changedAt) < 10*time.Second
}

// categoryTag returns a category label like "[Architecture]" if this item is the first
// in its category group (only in SortCategory mode). Returns empty string otherwise.
func (d envVarDelegate) categoryTag(m list.Model, index int, ei envItem) string {
	if d.sortMode != persist.SortCategory {
		return ""
	}
	// Skip category tags when filter is active (filtered subset changes item ordering)
	if m.FilterState() != list.Unfiltered {
		return ""
	}
	// Skip category tags for favorited items (they sort to the top regardless)
	if d.favorites[ei.Key] {
		return ""
	}

	currentCat := goenv.GetEnvVarCategory(ei.Key)

	// Check previous item's category
	if index > 0 {
		if prevItem, ok := m.Items()[index-1].(envItem); ok {
			// If previous item is a favorite, this is the first non-favorite -- show tag
			if d.favorites[prevItem.Key] {
				return "[" + string(currentCat) + "]"
			}
			prevCat := goenv.GetEnvVarCategory(prevItem.Key)
			if prevCat == currentCat {
				return "" // Same category, no tag
			}
		}
	}

	return "[" + string(currentCat) + "]"
}

// truncateString truncates a string to maxWidth display columns, adding "..." if needed.
// Uses rune-based iteration with per-rune width measurement to handle multibyte UTF-8 correctly.
func truncateString(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		runes := []rune(s)
		var w, i int
		for i = 0; i < len(runes) && w < maxWidth; i++ {
			w += lipgloss.Width(string(runes[i]))
		}
		return string(runes[:i])
	}
	runes := []rune(s)
	var w, i int
	for i = 0; i < len(runes) && w < maxWidth-3; i++ {
		w += lipgloss.Width(string(runes[i]))
	}
	return string(runes[:i]) + "..."
}

// truncateForDiff truncates a string for diff display, adding ellipsis if needed.
func truncateForDiff(s string, maxWidth int) string {
	return truncateString(s, maxWidth)
}
