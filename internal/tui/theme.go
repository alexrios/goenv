package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Theme defines the color scheme for the application.
type Theme struct {
	Name    string
	Primary color.Color // borders, prompts
	Accent  color.Color // cursor, titles
	Dim     color.Color // secondary text
	Success color.Color
	Warning color.Color
	Error   color.Color

	// Status bar colors
	StatusBg      color.Color
	StatusFg      color.Color
	StatusKeyBg   color.Color
	StatusKeyFg   color.Color
	HelpKey       color.Color
	HelpDesc      color.Color
	FavoriteStar  color.Color
	CategoryLabel color.Color
}

// Available theme names.
const (
	ThemeDefault = "default"
	ThemeNord    = "nord"
	ThemeDracula = "dracula"
)

// DefaultTheme returns the default color theme.
func DefaultTheme() Theme {
	return Theme{
		Name:          ThemeDefault,
		Primary:       lipgloss.Color("62"),  // Purple/blue
		Accent:        lipgloss.Color("205"), // Pink
		Dim:           lipgloss.Color("240"), // Gray
		Success:       lipgloss.Color("42"),  // Green
		Warning:       lipgloss.Color("214"), // Orange
		Error:         lipgloss.Color("196"), // Red
		StatusBg:      lipgloss.Color("#353533"),
		StatusFg:      lipgloss.Color("#C1C6B2"),
		StatusKeyBg:   lipgloss.Color("#8bc427"),
		StatusKeyFg:   lipgloss.Color("#FFFDF5"),
		HelpKey:       lipgloss.Color("86"),
		HelpDesc:      lipgloss.Color("252"),
		FavoriteStar:  lipgloss.Color("220"), // Gold
		CategoryLabel: lipgloss.Color("39"),  // Cyan
	}
}

// NordTheme returns a Nord-inspired color theme.
func NordTheme() Theme {
	return Theme{
		Name:          ThemeNord,
		Primary:       lipgloss.Color("#5E81AC"), // Nord blue
		Accent:        lipgloss.Color("#88C0D0"), // Nord frost
		Dim:           lipgloss.Color("#4C566A"), // Nord polar night
		Success:       lipgloss.Color("#A3BE8C"), // Nord green
		Warning:       lipgloss.Color("#EBCB8B"), // Nord yellow
		Error:         lipgloss.Color("#BF616A"), // Nord red
		StatusBg:      lipgloss.Color("#3B4252"),
		StatusFg:      lipgloss.Color("#ECEFF4"),
		StatusKeyBg:   lipgloss.Color("#5E81AC"),
		StatusKeyFg:   lipgloss.Color("#ECEFF4"),
		HelpKey:       lipgloss.Color("#88C0D0"),
		HelpDesc:      lipgloss.Color("#E5E9F0"),
		FavoriteStar:  lipgloss.Color("#EBCB8B"),
		CategoryLabel: lipgloss.Color("#81A1C1"),
	}
}

// DraculaTheme returns a Dracula-inspired color theme.
func DraculaTheme() Theme {
	return Theme{
		Name:          ThemeDracula,
		Primary:       lipgloss.Color("#BD93F9"), // Dracula purple
		Accent:        lipgloss.Color("#FF79C6"), // Dracula pink
		Dim:           lipgloss.Color("#6272A4"), // Dracula comment
		Success:       lipgloss.Color("#50FA7B"), // Dracula green
		Warning:       lipgloss.Color("#FFB86C"), // Dracula orange
		Error:         lipgloss.Color("#FF5555"), // Dracula red
		StatusBg:      lipgloss.Color("#282A36"),
		StatusFg:      lipgloss.Color("#F8F8F2"),
		StatusKeyBg:   lipgloss.Color("#BD93F9"),
		StatusKeyFg:   lipgloss.Color("#282A36"),
		HelpKey:       lipgloss.Color("#8BE9FD"),
		HelpDesc:      lipgloss.Color("#F8F8F2"),
		FavoriteStar:  lipgloss.Color("#F1FA8C"),
		CategoryLabel: lipgloss.Color("#8BE9FD"),
	}
}

// GetTheme returns a theme by name. Falls back to default if not found.
func GetTheme(name string) Theme {
	switch name {
	case ThemeNord:
		return NordTheme()
	case ThemeDracula:
		return DraculaTheme()
	default:
		return DefaultTheme()
	}
}

// AvailableThemes returns a list of available theme names.
func AvailableThemes() []string {
	return []string{ThemeDefault, ThemeNord, ThemeDracula}
}

// NextTheme returns the next theme name in the cycle.
func NextTheme(current string) string {
	switch current {
	case ThemeDefault:
		return ThemeNord
	case ThemeNord:
		return ThemeDracula
	case ThemeDracula:
		return ThemeDefault
	default:
		return ThemeDefault
	}
}
