package tui

import (
	"time"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

// AppScreen represents the current active screen in the application.
type AppScreen int

// Application screens.
const (
	AppScreenList           AppScreen = iota // Main list view of environment variables
	AppScreenEdit                            // Edit screen for modifying a variable
	AppScreenHelp                            // Help screen showing keyboard shortcuts
	AppScreenExport                          // Export screen for saving snapshots
	AppScreenImport                          // Import screen for loading snapshots
	AppScreenPresets                         // Presets screen for managing saved configs
	AppScreenCompare                         // Comparison screen for viewing diffs
	AppScreenStats                           // Stats screen for variable statistics
	AppScreenReset                           // Reset screen for batch resetting variables
	AppScreenShellExport                     // Shell export screen for generating export commands
	AppScreenCommandPalette                  // Command palette for quick command access
)

// String returns a human-readable name for the screen.
func (s AppScreen) String() string {
	switch s {
	case AppScreenList:
		return "List"
	case AppScreenEdit:
		return "Edit"
	case AppScreenHelp:
		return "Help"
	case AppScreenExport:
		return "Export"
	case AppScreenImport:
		return "Import"
	case AppScreenPresets:
		return "Presets"
	case AppScreenCompare:
		return "Compare"
	case AppScreenStats:
		return "Stats"
	case AppScreenReset:
		return "Reset"
	case AppScreenShellExport:
		return "Shell Export"
	case AppScreenCommandPalette:
		return "Commands"
	default:
		return ""
	}
}

// Message timeout durations for temporary status messages.
const (
	messageTimeoutSuccess = 4 * time.Second
	messageTimeoutError   = 8 * time.Second
	messageTimeoutInfo    = 3 * time.Second
)

// saveEnvRequestMsg is sent from EditScreen to request saving a variable.
type saveEnvRequestMsg struct {
	Env           goenv.EnvVar
	OriginalValue string
}

// envSetSuccessMsg is sent when SetEnvVar completes successfully.
type envSetSuccessMsg struct {
	Env           goenv.EnvVar
	OriginalValue string
	IsUndoRedo    bool // true if this is from undo/redo (don't record history)
}

// envSetErrorMsg is sent if SetEnvVar returns an error.
type envSetErrorMsg struct {
	Key           string
	OriginalValue string
	IsUndoRedo    bool // true if this is from undo/redo
	Err           error
}

// clearMessageMsg clears temporary messages (error or success) after a delay.
type clearMessageMsg struct{}

// changeScreenMsg is sent from a screen model to request a transition.
type changeScreenMsg AppScreen

// startEditMsg is sent from a screen model to initiate editing a specific variable.
type startEditMsg goenv.EnvVar

// showDetailMsg is sent to view a read-only variable's details without editing.
type showDetailMsg goenv.EnvVar

// toggleSortMsg requests toggling the sort mode.
type toggleSortMsg struct{}

// reloadEnvMsg requests reloading environment variables.
type reloadEnvMsg struct{}

// reloadEnvSuccessMsg is sent when reload completes successfully.
type reloadEnvSuccessMsg struct {
	Items []goenv.EnvVar
}

// reloadEnvErrorMsg is sent if reload returns an error.
type reloadEnvErrorMsg struct {
	Err error
}

// copyValueMsg requests copying a value to clipboard.
type copyValueMsg struct {
	Key   string
	Value string
}

// copyKeyValueMsg requests copying KEY=VALUE to clipboard.
type copyKeyValueMsg struct {
	Key   string
	Value string
}

// clipboardResultMsg is sent when clipboard operation completes.
type clipboardResultMsg struct {
	Success bool
	Message string
}

// showHelpMsg requests showing the help screen.
type showHelpMsg struct{}

// undoMsg requests undoing the last change.
type undoMsg struct{}

// redoMsg requests redoing the last undone change.
type redoMsg struct{}

// toggleFavoriteMsg requests toggling the favorite status of a variable.
type toggleFavoriteMsg struct {
	Key string
}

// showExportMsg requests showing the export screen.
type showExportMsg struct{}

// showImportMsg requests showing the import screen.
type showImportMsg struct{}

// showPresetsMsg requests showing the presets screen.
type showPresetsMsg struct{}

// showCompareMsg requests showing the comparison screen.
type showCompareMsg struct{}

// showStatsMsg requests showing the stats screen.
type showStatsMsg struct{}

// resetEnvMsg requests resetting a variable to its default value.
type resetEnvMsg struct {
	Key string
}

// confirmResetMsg is sent when user confirms a reset action.
type confirmResetMsg struct {
	Key string
}

// envUnsetSuccessMsg is sent when UnsetEnvVar completes successfully.
type envUnsetSuccessMsg struct {
	Key string
}

// envUnsetErrorMsg is sent if UnsetEnvVar returns an error.
type envUnsetErrorMsg struct {
	Key string
	Err error
}

// showResetScreenMsg requests showing the batch reset screen.
type showResetScreenMsg struct{}

// batchUnsetSuccessMsg is sent when batch unset completes.
type batchUnsetSuccessMsg struct {
	Count       int // Number of successful resets
	FailedCount int // Number of failed resets
}

// showShellExportMsg requests showing the shell export screen.
type showShellExportMsg struct{}

// showCommandPaletteMsg requests showing the command palette.
type showCommandPaletteMsg struct{}

// toggleThemeMsg requests cycling to the next theme.
type toggleThemeMsg struct{}

// cycleCategoryMsg requests cycling the category filter.
type cycleCategoryMsg struct{}

// toggleWatchMsg requests toggling watch mode on/off.
type toggleWatchMsg struct{}

// watchTickMsg is sent periodically to trigger auto-reload in watch mode.
type watchTickMsg struct{}

// exportSnapshotMsg is sent when the user wants to export a snapshot.
type exportSnapshotMsg struct {
	Name        string
	Description string
	Items       []goenv.EnvVar
}

// exportSuccessMsg is sent when export completes successfully.
type exportSuccessMsg struct {
	FilePath string
}

// exportErrorMsg is sent when export fails.
type exportErrorMsg struct {
	Err error
}

// loadSnapshotSuccessMsg is sent when snapshot loads successfully.
type loadSnapshotSuccessMsg struct {
	Snapshot persist.Snapshot
}

// loadSnapshotErrorMsg is sent when snapshot loading fails.
type loadSnapshotErrorMsg struct {
	Err error
}

// importApplyMsg is sent when the user wants to apply selected variables.
type importApplyMsg struct {
	Variables []goenv.EnvVar
}

// importSuccessMsg is sent when import completes successfully.
type importSuccessMsg struct {
	Count       int
	FailedCount int
}

// importErrorMsg is sent when import fails.
type importErrorMsg struct {
	Err error
}

// Messages for preset operations
type presetLoadSuccessMsg struct {
	Presets []persist.Preset
	Skipped int
}

type presetLoadErrorMsg struct {
	Err error
}

type presetApplyMsg struct {
	Preset persist.Preset
}

type presetCreateSuccessMsg struct {
	Name string
}

type presetCreateErrorMsg struct {
	Err error
}

type presetDeleteSuccessMsg struct {
	Name string
}

type presetDeleteErrorMsg struct {
	Err error
}

// compareSourcesLoadedMsg is sent when sources are loaded.
type compareSourcesLoadedMsg struct {
	Sources []CompareSource
	Skipped int
}

// Command palette context messages -- dispatched to mainModel which has list context.
type commandPaletteEditMsg struct{}
type commandPaletteCopyMsg struct{}
type commandPaletteCopyKeyMsg struct{}
type commandPaletteFavoriteMsg struct{}
type commandPaletteResetMsg struct{}
