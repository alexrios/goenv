package tui

import (
	"image/color"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

// activeTheme holds the current theme for global style access.
// Set during initialization in NewMainModel.
var activeTheme = DefaultTheme()

// SetActiveTheme updates the global theme and regenerates all styles.
func SetActiveTheme(theme Theme) {
	activeTheme = theme
	regenerateStyles()
}

// Color palette - derived from active theme
var (
	colorDim      color.Color
	colorPrimary  color.Color
	colorAccent   color.Color
	colorHelpKey  color.Color
	colorHelpDesc color.Color
	colorSuccess  color.Color
	colorWarning  color.Color
	colorError    color.Color
)

// --- Styles ---
// These are initialized in init() and regenerated when theme changes.
var docStyle = lipgloss.NewStyle().Margin(1, 2)

var (
	statusBarStyle      lipgloss.Style
	statusKeyStyle      lipgloss.Style
	statusModifiedStyle lipgloss.Style
	statusErrorStyle    lipgloss.Style
	helpTextStyle       lipgloss.Style
	statusSavingStyle   lipgloss.Style
	statusSuccessStyle  lipgloss.Style
	errorMessageStyle   lipgloss.Style
)

// --- TextInput Styles ---
var textInputStyles textinput.Styles

// --- Help Screen Styles ---
var (
	helpTitleStyle   lipgloss.Style
	helpSectionStyle lipgloss.Style
	helpKeyStyle     lipgloss.Style
	helpDescStyle    lipgloss.Style
	helpBoxStyle     lipgloss.Style
	helpFooterStyle  lipgloss.Style
)

// --- Edit Screen Styles ---
var (
	editOriginalValueStyle   lipgloss.Style
	editLengthInfoStyle      lipgloss.Style
	editDiffOldStyle         lipgloss.Style
	editDiffNewStyle         lipgloss.Style
	editDiffChangedStyle     lipgloss.Style
	editValidationErrorStyle lipgloss.Style
	editDocStyle             lipgloss.Style
	suggestionStyle          lipgloss.Style
	suggestionSelectedStyle  lipgloss.Style
	suggestionHintStyle      lipgloss.Style
)

// --- List Screen Styles ---
var (
	favoriteStarStyle  lipgloss.Style
	categoryLabelStyle lipgloss.Style
	readOnlyLabelStyle lipgloss.Style
)

var inlineDescStyle lipgloss.Style
var spinnerStyle lipgloss.Style
var watchChangedStyle lipgloss.Style

// init initializes styles with the default theme.
func init() {
	regenerateStyles()
}

// regenerateStyles rebuilds all styles from the active theme.
func regenerateStyles() {
	// Update color palette from theme
	colorDim = activeTheme.Dim
	colorPrimary = activeTheme.Primary
	colorAccent = activeTheme.Accent
	colorHelpKey = activeTheme.HelpKey
	colorHelpDesc = activeTheme.HelpDesc
	colorSuccess = activeTheme.Success
	colorWarning = activeTheme.Warning
	colorError = activeTheme.Error

	// Status bar styles
	statusBarStyle = lipgloss.NewStyle().
		Foreground(activeTheme.StatusFg).
		Background(activeTheme.StatusBg)

	statusKeyStyle = lipgloss.NewStyle().
		Inherit(statusBarStyle).
		Foreground(activeTheme.StatusKeyFg).
		Background(activeTheme.StatusKeyBg).
		Padding(0, 1).
		MarginRight(1)

	statusModifiedStyle = statusKeyStyle.Background(colorWarning)
	statusErrorStyle = statusKeyStyle.Background(colorError)
	helpTextStyle = lipgloss.NewStyle().Inherit(statusBarStyle)
	statusSavingStyle = statusKeyStyle.Background(colorAccent)
	statusSuccessStyle = statusKeyStyle.Background(colorSuccess)

	errorMessageStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(activeTheme.StatusKeyFg).
		Background(colorError).
		Padding(0, 1).
		MarginTop(1).
		MarginBottom(1)

	// TextInput styles
	textInputStyles = textinput.Styles{
		Focused: textinput.StyleState{
			Prompt: lipgloss.NewStyle().Foreground(colorPrimary),
			Text:   lipgloss.NewStyle().Foreground(colorAccent),
		},
		Blurred: textinput.StyleState{
			Prompt: lipgloss.NewStyle().Foreground(colorPrimary),
			Text:   lipgloss.NewStyle().Foreground(colorDim),
		},
		Cursor: textinput.CursorStyle{
			Color: colorAccent,
		},
	}

	// Help screen styles
	helpTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorAccent).
		MarginBottom(1)

	helpSectionStyle = lipgloss.NewStyle().
		MarginTop(1).
		MarginBottom(1)

	helpKeyStyle = lipgloss.NewStyle().
		Foreground(colorHelpKey).
		Bold(true).
		Width(12)

	helpDescStyle = lipgloss.NewStyle().
		Foreground(colorHelpDesc)

	helpBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2)

	helpFooterStyle = lipgloss.NewStyle().
		Foreground(colorDim)

	// Edit screen styles
	editOriginalValueStyle = lipgloss.NewStyle().
		Foreground(colorDim).
		Italic(true)

	editLengthInfoStyle = lipgloss.NewStyle().
		Foreground(colorDim)

	editDiffOldStyle = lipgloss.NewStyle().
		Foreground(colorDim).
		Strikethrough(true)

	editDiffNewStyle = lipgloss.NewStyle().
		Foreground(colorSuccess).
		Bold(true)

	editDiffChangedStyle = lipgloss.NewStyle().
		Foreground(colorWarning)

	editValidationErrorStyle = lipgloss.NewStyle().
		Foreground(colorError).
		Bold(true)

	editDocStyle = lipgloss.NewStyle().
		Foreground(colorDim).
		Italic(true)

	suggestionStyle = lipgloss.NewStyle().
		Foreground(colorDim).
		PaddingLeft(2)

	suggestionSelectedStyle = lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true).
		PaddingLeft(2)

	suggestionHintStyle = lipgloss.NewStyle().
		Foreground(colorDim).
		Italic(true)

	// List screen styles
	favoriteStarStyle = lipgloss.NewStyle().
		Foreground(activeTheme.FavoriteStar)

	categoryLabelStyle = lipgloss.NewStyle().
		Foreground(activeTheme.CategoryLabel).
		Italic(true)

	readOnlyLabelStyle = lipgloss.NewStyle().
		Foreground(colorDim).
		Italic(true)

	// Inline description style (shown under selected item)
	inlineDescStyle = lipgloss.NewStyle().
		Foreground(colorAccent).
		Italic(true).
		PaddingLeft(2)

	// Spinner style
	spinnerStyle = lipgloss.NewStyle().
		Foreground(colorAccent)

	// Watch mode changed item highlight
	watchChangedStyle = lipgloss.NewStyle().
		Foreground(colorWarning).
		Bold(true)
}

// --- Application Text Constants ---
const (
	appTitle         = "GO ENV"
	inputPrompt      = "> "
	inputPlaceholder = "Enter new value..."
	editHelpText     = "(Enter: save, Esc: cancel, Ctrl+R: reset)"
	readOnlyHelpText = "(y: copy value, Y: copy KEY=VALUE, Esc: back)"
	listHelpText     = "Press Enter to edit, 'q' or Ctrl+C to quit"
	helpScreenTitle  = "GO ENV - Keyboard Shortcuts"
	helpScreenFooter = "Press Esc, q, or ? to return..."
)

// Status icons for visual indicators.
const (
	IconSuccess  = "\u2713"
	IconError    = "\u2717"
	IconModified = "\u25CF"
	IconSaving   = "\u27F3"
	IconReload   = "\u21BB"
	IconFavorite = "\u2605"
	IconWatch    = "\u2299"
	IconReadOnly = "\U0001F512"
)
