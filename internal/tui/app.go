package tui

import (
	"cmp"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/davecgh/go-spew/spew"

	"github.com/alexrios/goenv/internal/commands"
	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

// mainModel coordinates all screens and manages shared application state.
type mainModel struct {
	currentScreen     AppScreen
	listScreen        ListScreenModel
	editScreen        EditScreenModel
	helpScreen        HelpScreenModel
	exportScreen      ExportScreenModel
	importScreen      ImportScreenModel
	presetScreen      PresetScreenModel
	compareScreen     CompareScreenModel
	statsScreen       StatsScreenModel
	resetScreen       ResetScreenModel
	shellExportScreen ShellExportScreenModel
	commandPalette    CommandPaletteModel

	saveError      string
	isSaving       bool
	isReloading    bool
	successMessage string

	spinner       spinner.Model
	spinnerActive bool

	watchMode     bool
	watchInterval time.Duration
	recentChanges map[string]time.Time

	dump   io.Writer
	width  int
	height int

	items          []goenv.EnvVar
	sortMode       persist.SortMode
	favorites      map[string]bool
	categoryFilter goenv.Category // empty means show all

	editHistory []persist.EditRecord
	historyIdx  int

	cfg           persist.Config
	goVersion     string
	parsedVersion goenv.GoVersion

	hasEditsThisSession bool
	confirmQuit         bool
}

// NewMainModel creates a new main model with the given items, optional debug output, and config.
func NewMainModel(items []goenv.EnvVar, dump *os.File, cfg persist.Config) mainModel {
	// Initialize theme from config
	theme := GetTheme(cfg.Theme)
	SetActiveTheme(theme)

	// Get Go version (ignore error, not critical)
	goVersion, _ := commands.GetGoVersion()

	parsedVersion := goenv.ParseGoVersion(goVersion)

	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(spinnerStyle),
	)

	m := mainModel{
		currentScreen:     AppScreenList,
		editScreen:        NewEditScreenModel(parsedVersion),
		helpScreen:        NewHelpScreenModel(),
		exportScreen:      NewExportScreenModel(),
		importScreen:      NewImportScreenModel(),
		presetScreen:      NewPresetScreenModel(),
		compareScreen:     NewCompareScreenModel(),
		statsScreen:       NewStatsScreenModel(),
		resetScreen:       NewResetScreenModel(),
		shellExportScreen: NewShellExportScreenModel(),
		commandPalette:    NewCommandPaletteModel(),
		items:             items,
		dump:              dump,
		sortMode:          persist.SortModeFromString(cfg.SortMode),
		favorites:         make(map[string]bool),
		cfg:               cfg,
		goVersion:         goVersion,
		parsedVersion:     parsedVersion,
		spinner:           s,
		watchInterval:     watchIntervalFromConfig(cfg.WatchInterval),
		recentChanges:     make(map[string]time.Time),
	}

	// Load favorites from config
	for _, fav := range cfg.Favorites {
		m.favorites[fav] = true
	}

	// Load edit history from disk
	if history, err := persist.LoadHistory(); err == nil {
		m.editHistory = history.Records
		m.historyIdx = history.CurrentIdx
	}

	m.sortItems()
	listItems := envVarsToListItems(m.items)
	m.listScreen = NewListScreenModelWithSortMode(listItems, m.goVersion, m.parsedVersion, m.sortMode, m.favorites, m.recentChanges)
	return m
}

// watchIntervalFromConfig returns a watch interval from config, defaulting to 5s.
// Values <= 0 default to 5 seconds. Clamped to [1s, 24h] to prevent busy loops
// and overflow.
func watchIntervalFromConfig(seconds int) time.Duration {
	const maxWatchSeconds = 86400 // 24 hours
	if seconds <= 0 {
		return 5 * time.Second
	}
	if seconds > maxWatchSeconds {
		return time.Duration(maxWatchSeconds) * time.Second
	}
	return time.Duration(seconds) * time.Second
}

// sortItems sorts the items slice according to the current sort mode.
// Favorites are always sorted first, regardless of sort mode.
func (m *mainModel) sortItems() {
	// Sync Favorite field from favorites map
	for i := range m.items {
		m.items[i].Favorite = m.favorites[m.items[i].Key]
	}

	switch m.sortMode {
	case persist.SortAlpha:
		slices.SortFunc(m.items, func(a, b goenv.EnvVar) int {
			// Favorites first
			iFav, jFav := m.favorites[a.Key], m.favorites[b.Key]
			if iFav != jFav {
				if iFav {
					return -1
				}
				return 1
			}
			return cmp.Compare(a.Key, b.Key)
		})
	case persist.SortModifiedFirst:
		slices.SortFunc(m.items, func(a, b goenv.EnvVar) int {
			// Favorites first
			iFav, jFav := m.favorites[a.Key], m.favorites[b.Key]
			if iFav != jFav {
				if iFav {
					return -1
				}
				return 1
			}
			// Then modified
			if a.Changed != b.Changed {
				if a.Changed {
					return -1
				}
				return 1
			}
			return cmp.Compare(a.Key, b.Key)
		})
	case persist.SortCategory:
		slices.SortFunc(m.items, func(a, b goenv.EnvVar) int {
			// Favorites first
			iFav, jFav := m.favorites[a.Key], m.favorites[b.Key]
			if iFav != jFav {
				if iFav {
					return -1
				}
				return 1
			}
			// Then by category
			iCat := goenv.CategoryIndex(goenv.GetEnvVarCategory(a.Key))
			jCat := goenv.CategoryIndex(goenv.GetEnvVarCategory(b.Key))
			if v := cmp.Compare(iCat, jCat); v != 0 {
				return v
			}
			return cmp.Compare(a.Key, b.Key)
		})
	}
}

// updateItemByKey finds and updates an item by its key.
// Returns true if the item was found and updated, false otherwise.
func (m *mainModel) updateItemByKey(key string, newValue goenv.EnvVar) bool {
	for i, item := range m.items {
		if item.Key == key {
			m.items[i] = newValue
			return true
		}
	}
	m.debugLog(fmt.Sprintf("WARNING: updateItemByKey: key %q not found in items", key))
	return false
}

// refreshListScreen recreates the list screen with current items, preserving selection.
func (m *mainModel) refreshListScreen() tea.Cmd {
	currentIdx := m.listScreen.list.Index()
	filterText := m.listScreen.list.FilterInput.Value()
	filterState := m.listScreen.list.FilterState()

	// Apply category filter if active
	items := m.items
	if m.categoryFilter != "" {
		items = m.filteredItemsByCategory()
	}

	listItems := envVarsToListItems(items)
	m.listScreen = NewListScreenModelWithSortMode(listItems, m.goVersion, m.parsedVersion, m.sortMode, m.favorites, m.recentChanges)
	m.calculateScreenSizes()

	// Restore selection, clamping to valid range
	if currentIdx >= len(listItems) {
		currentIdx = len(listItems) - 1
	}
	if currentIdx >= 0 {
		m.listScreen.list.Select(currentIdx)
	}

	// Restore filter state across refreshes
	if filterText != "" && filterState != list.Unfiltered {
		m.listScreen.list.SetFilterText(filterText)
		m.listScreen.list.SetFilterState(list.FilterApplied)
	}

	return m.listScreen.Init()
}

// filteredItemsByCategory returns items matching the current category filter.
func (m *mainModel) filteredItemsByCategory() []goenv.EnvVar {
	var filtered []goenv.EnvVar
	for _, ev := range m.items {
		if goenv.GetEnvVarCategory(ev.Key) == m.categoryFilter {
			filtered = append(filtered, ev)
		}
	}
	return filtered
}

// clearTempMessages clears temporary error and success messages.
func (m *mainModel) clearTempMessages() {
	m.saveError = ""
	m.successMessage = ""
}

// debugLog writes values to the debug dump if enabled.
func (m *mainModel) debugLog(values ...any) {
	if m.dump != nil {
		spew.Fdump(m.dump, values...)
	}
}

// calculateScreenSizes updates component sizes based on terminal dimensions.
func (m *mainModel) calculateScreenSizes() {
	h, v := docStyle.GetFrameSize()
	footerHeight := lipgloss.Height(statusBarStyle.Render(" "))
	errorAreaHeight := 0
	if m.saveError != "" {
		errorAreaHeight = lipgloss.Height(errorMessageStyle.Render(" "))
	}
	fixedVerticalSpace := v + footerHeight + errorAreaHeight

	availableHeight := clampNonNegative(m.height - fixedVerticalSpace)
	m.listScreen.list.SetSize(m.width-h, availableHeight)

	inputPromptWidth := lipgloss.Width(m.editScreen.textInput.Prompt)
	inputWidth := clampNonNegative(m.width - h - inputPromptWidth)
	m.editScreen.textInput.SetWidth(inputWidth)
}

// Init initializes the main model and returns the initial command.
func (m mainModel) Init() tea.Cmd {
	if m.dump != nil {
		spew.Fdump(m.dump, "NEW DEBUG SESSION")
	}
	return m.listScreen.Init()
}

// Update handles all messages and coordinates screen transitions.
// Uses handler methods from handlers.go for clean separation of concerns.
func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.debugLog(msg)

	var newModel mainModel
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		newModel, cmd = m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		newModel, cmd = m.handleWindowSize(msg)
	case changeScreenMsg:
		newModel, cmd = m.handleChangeScreen(msg)
	case startEditMsg:
		newModel, cmd = m.handleStartEdit(msg)
	case saveEnvRequestMsg:
		newModel, cmd = m.handleSaveEnvRequest(msg)
	case envSetSuccessMsg:
		newModel, cmd = m.handleEnvSetSuccess(msg)
	case envSetErrorMsg:
		newModel, cmd = m.handleEnvSetError(msg)
	case clearMessageMsg:
		newModel, cmd = m.handleClearMessage()
	case toggleSortMsg:
		newModel, cmd = m.handleToggleSort()
	case reloadEnvMsg:
		newModel, cmd = m.handleReloadEnv()
	case reloadEnvSuccessMsg:
		newModel, cmd = m.handleReloadEnvSuccess(msg)
	case reloadEnvErrorMsg:
		newModel, cmd = m.handleReloadEnvError(msg)
	case copyValueMsg:
		newModel, cmd = m.handleCopyValue(msg)
	case copyKeyValueMsg:
		newModel, cmd = m.handleCopyKeyValue(msg)
	case clipboardResultMsg:
		newModel, cmd = m.handleClipboardResult(msg)
	case showHelpMsg:
		newModel, cmd = m.handleShowHelp()
	case undoMsg:
		newModel, cmd = m.handleUndo()
	case redoMsg:
		newModel, cmd = m.handleRedo()
	case toggleFavoriteMsg:
		newModel, cmd = m.handleToggleFavorite(msg)
	case showExportMsg:
		newModel, cmd = m.handleShowExport()
	case showImportMsg:
		newModel, cmd = m.handleShowImport()
	case showPresetsMsg:
		newModel, cmd = m.handleShowPresets()
	case showCompareMsg:
		newModel, cmd = m.handleShowCompare()
	case showStatsMsg:
		newModel, cmd = m.handleShowStats()
	case exportSnapshotMsg:
		newModel, cmd = m.handleExportSnapshot(msg)
	case exportSuccessMsg:
		newModel, cmd = m.handleExportSuccess(msg)
	case exportErrorMsg:
		newModel, cmd = m.handleExportError(msg)
	case importApplyMsg:
		newModel, cmd = m.handleImportApply(msg)
	case importSuccessMsg:
		newModel, cmd = m.handleImportSuccess(msg)
	case importErrorMsg:
		newModel, cmd = m.handleImportError(msg)
	case presetApplyMsg:
		newModel, cmd = m.handlePresetApply(msg)
	case presetCreateSuccessMsg:
		newModel, cmd = m.handlePresetCreateSuccess(msg)
	case presetDeleteSuccessMsg:
		newModel, cmd = m.handlePresetDeleteSuccess(msg)
	case resetEnvMsg:
		newModel, cmd = m.handleResetEnv(msg)
	case confirmResetMsg:
		newModel, cmd = m.handleConfirmReset(msg)
	case envUnsetSuccessMsg:
		newModel, cmd = m.handleEnvUnsetSuccess(msg)
	case envUnsetErrorMsg:
		newModel, cmd = m.handleEnvUnsetError(msg)
	case showResetScreenMsg:
		newModel, cmd = m.handleShowResetScreen()
	case batchUnsetSuccessMsg:
		newModel, cmd = m.handleBatchUnsetSuccess(msg)
	case showShellExportMsg:
		newModel, cmd = m.handleShowShellExport()
	case showCommandPaletteMsg:
		newModel, cmd = m.handleShowCommandPalette()
	case commandPaletteEditMsg:
		newModel, cmd = m.handleCommandPaletteEdit()
	case commandPaletteCopyMsg:
		newModel, cmd = m.handleCommandPaletteCopy()
	case commandPaletteCopyKeyMsg:
		newModel, cmd = m.handleCommandPaletteCopyKey()
	case commandPaletteFavoriteMsg:
		newModel, cmd = m.handleCommandPaletteFavorite()
	case commandPaletteResetMsg:
		newModel, cmd = m.handleCommandPaletteReset()
	case cycleCategoryMsg:
		newModel, cmd = m.handleCycleCategory()
	case toggleThemeMsg:
		newModel, cmd = m.handleToggleTheme()
	case toggleWatchMsg:
		newModel, cmd = m.handleToggleWatch()
	case watchTickMsg:
		newModel, cmd = m.handleWatchTick()
	case spinner.TickMsg:
		if m.spinnerActive {
			var spinnerCmd tea.Cmd
			m.spinner, spinnerCmd = m.spinner.Update(msg)
			return m, spinnerCmd
		}
		return m, nil
	default:
		newModel, cmd = m.handleDefault(msg)
	}

	return newModel, cmd
}

// statusInfo holds text and style for the status bar.
type statusInfo struct {
	text  string
	style lipgloss.Style
}

// getStatusInfo determines the current status text and style.
func (m *mainModel) getStatusInfo() statusInfo {
	currentIdx := m.listScreen.list.Index() + 1
	totalItems := len(m.listScreen.list.Items())
	statusText := fmt.Sprintf("%d/%d variables | %s", currentIdx, totalItems, m.sortMode.String())
	if m.categoryFilter != "" {
		statusText += fmt.Sprintf(" | %s", m.categoryFilter)
	}
	if totalItems == 0 {
		statusText = "No variables found"
	}

	statusStyleToUse := statusKeyStyle

	if m.isReloading {
		if m.spinnerActive {
			statusText = m.spinner.View() + " Reloading..."
		} else {
			statusText = IconReload + " Reloading..."
		}
		statusStyleToUse = statusSavingStyle
	} else if m.isSaving {
		if m.spinnerActive {
			statusText = m.spinner.View() + " Saving..."
		} else {
			statusText = IconSaving + " Saving..."
		}
		statusStyleToUse = statusSavingStyle
	} else if m.successMessage != "" {
		statusText = IconSuccess + " " + m.successMessage
		statusStyleToUse = statusSuccessStyle
	} else {
		// Check if selected item is modified
		if selectedItem := m.listScreen.list.SelectedItem(); selectedItem != nil {
			if ev, ok := selectedItem.(envItem); ok && ev.Changed {
				statusText = fmt.Sprintf("%s %d/%d MODIFIED", IconModified, currentIdx, totalItems)
				statusStyleToUse = statusModifiedStyle
			}
		}
	}

	// Append watch indicator
	if m.watchMode {
		statusText += fmt.Sprintf(" %s Watching (%ds)", IconWatch, int(m.watchInterval.Seconds()))
	}

	return statusInfo{text: statusText, style: statusStyleToUse}
}

// View renders the current screen with status bar and error display.
func (m mainModel) View() tea.View {
	m.debugLog(
		fmt.Sprint("Screen:", m.currentScreen),
		fmt.Sprint("Editing Key:", m.editScreen.editingKey),
		fmt.Sprint("Save Error:", m.saveError),
		fmt.Sprint("Success Msg:", m.successMessage),
		fmt.Sprint("Is Saving:", m.isSaving),
	)

	var screenView string
	var errorArea string
	var footer string

	switch m.currentScreen {
	case AppScreenList:
		screenView = m.listScreen.View()
	case AppScreenEdit:
		screenView = m.editScreen.View()
	case AppScreenHelp:
		screenView = m.helpScreen.View()
	case AppScreenExport:
		screenView = m.exportScreen.View()
	case AppScreenImport:
		screenView = m.importScreen.View()
	case AppScreenPresets:
		screenView = m.presetScreen.View()
	case AppScreenCompare:
		screenView = m.compareScreen.View()
	case AppScreenStats:
		screenView = m.statsScreen.View()
	case AppScreenReset:
		screenView = m.resetScreen.View()
	case AppScreenShellExport:
		screenView = m.shellExportScreen.View()
	case AppScreenCommandPalette:
		screenView = m.commandPalette.View()
	}

	if m.saveError != "" {
		errorArea = lipgloss.PlaceHorizontal(
			m.width-docStyle.GetHorizontalFrameSize(),
			lipgloss.Center,
			errorMessageStyle.Render(m.saveError),
		)
	}

	status := m.getStatusInfo()
	statusLine := status.style.Render(status.text)
	helpLine := helpTextStyle.Render(listHelpText)

	switch m.currentScreen {
	case AppScreenList:
		// Full status bar (already set above)
	default:
		// Show screen-specific help text, keep status messages visible
		helpLine = helpTextStyle.Render("(Esc: back)")
		if m.successMessage == "" && m.saveError == "" && !m.isSaving && !m.isReloading {
			statusLine = status.style.Render(m.currentScreen.String())
		}
		// statusLine keeps its value when there are active messages
	}

	statusWidth := clampNonNegative(m.width - docStyle.GetHorizontalFrameSize() - lipgloss.Width(helpLine))

	footer = statusBarStyle.
		Width(m.width - docStyle.GetHorizontalFrameSize()).
		Render(
			lipgloss.JoinHorizontal(lipgloss.Top,
				statusLine,
				lipgloss.PlaceHorizontal(statusWidth-lipgloss.Width(statusLine), lipgloss.Right, helpLine),
			),
		)

	content := lipgloss.JoinVertical(lipgloss.Left,
		screenView,
		errorArea,
		footer,
	)

	v := tea.NewView(content)
	v.AltScreen = true
	v.WindowTitle = "goenv"
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// selectedEnvItem returns the currently selected item from the list screen, if any.
func (m *mainModel) selectedEnvItem() (envItem, bool) {
	if selectedItem := m.listScreen.list.SelectedItem(); selectedItem != nil {
		ei, ok := selectedItem.(envItem)
		return ei, ok
	}
	return envItem{}, false
}

// saveConfigAsync persists config in a background goroutine.
func (m *mainModel) saveConfigAsync() {
	go func(cfg persist.Config) {
		_ = persist.SaveConfig(cfg)
	}(m.cfg)
}

// saveHistoryAsync persists history in a background goroutine.
func (m *mainModel) saveHistoryAsync() {
	go func(records []persist.EditRecord, idx, maxHist int) {
		_ = persist.SaveHistory(persist.SessionHistory{Records: records, CurrentIdx: idx}, maxHist)
	}(m.editHistory, m.historyIdx, m.cfg.MaxHistory)
}

// currentEnvVarsForScreens returns items for screens that need []goenv.EnvVar.
func (m *mainModel) currentEnvVarsForScreens() []goenv.EnvVar {
	return m.items
}

// syncFavoritesToConfig updates the config with current favorites.
func (m *mainModel) syncFavoritesToConfig() {
	m.cfg.Favorites = slices.Collect(maps.Keys(m.favorites))
}

// viewableItems builds the list of items to show, applying category filter.
func (m *mainModel) viewableItems() []goenv.EnvVar {
	if m.categoryFilter != "" {
		return m.filteredItemsByCategory()
	}
	return m.items
}

// formatEnvVarPair returns KEY=VALUE formatted string.
func formatEnvVarPair(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

// envVarSliceContains checks if any envVar has the given key.
func envVarSliceContains(items []goenv.EnvVar, key string) bool {
	for _, item := range items {
		if item.Key == key {
			return true
		}
	}
	return false
}

// joinNonEmpty joins non-empty strings with separator.
func joinNonEmpty(parts []string, sep string) string {
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return strings.Join(nonEmpty, sep)
}
