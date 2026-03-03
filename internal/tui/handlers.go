package tui

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

// handleKeyPress handles all key press messages, delegating to the current screen.
func (m mainModel) handleKeyPress(msg tea.KeyPressMsg) (mainModel, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		// Same exit-confirmation logic as 'q' when on the list screen
		if m.currentScreen == AppScreenList && m.hasEditsThisSession && !m.confirmQuit {
			m.confirmQuit = true
			m.successMessage = "You have unsaved session edits. Press q or Ctrl+C again to quit."
			return m, clearMessageAfter(messageTimeoutError)
		}
		return m, tea.Quit
	case "q":
		// Only quit from list screen; other screens use 'q' to go back
		if m.currentScreen == AppScreenList {
			if m.hasEditsThisSession && !m.confirmQuit {
				m.confirmQuit = true
				m.successMessage = "You have unsaved session edits. Press q again to quit."
				return m, clearMessageAfter(messageTimeoutError)
			}
			return m, tea.Quit
		}
	}

	// Reset confirmQuit on any non-q key press on list screen
	if m.currentScreen == AppScreenList && m.confirmQuit {
		m.confirmQuit = false
	}

	var cmd tea.Cmd
	switch m.currentScreen {
	case AppScreenList:
		cmd = m.listScreen.Update(msg)
	case AppScreenEdit:
		cmd = m.editScreen.Update(msg)
	case AppScreenHelp:
		cmd = m.helpScreen.Update(msg)
	case AppScreenExport:
		cmd = m.exportScreen.Update(msg)
	case AppScreenImport:
		cmd = m.importScreen.Update(msg)
	case AppScreenPresets:
		cmd = m.presetScreen.Update(msg)
	case AppScreenCompare:
		cmd = m.compareScreen.Update(msg)
	case AppScreenStats:
		cmd = m.statsScreen.Update(msg)
	case AppScreenReset:
		cmd = m.resetScreen.Update(msg)
	case AppScreenShellExport:
		cmd = m.shellExportScreen.Update(msg)
	case AppScreenCommandPalette:
		cmd = m.commandPalette.Update(msg)
	}
	return m, cmd
}

// handleWindowSize handles terminal resize events.
func (m mainModel) handleWindowSize(msg tea.WindowSizeMsg) (mainModel, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.calculateScreenSizes()
	return m, nil
}

// handleChangeScreen handles screen transition requests.
func (m mainModel) handleChangeScreen(msg changeScreenMsg) (mainModel, tea.Cmd) {
	m.currentScreen = AppScreen(msg)
	m.clearTempMessages()
	m.calculateScreenSizes()

	switch m.currentScreen {
	case AppScreenList:
		cmds := []tea.Cmd{m.refreshListScreen()}
		// Restart watch tick chain if watch mode is active (it dies on screen transitions)
		if m.watchMode {
			cmds = append(cmds, tea.Tick(m.watchInterval, func(time.Time) tea.Msg { return watchTickMsg{} }))
		}
		return m, tea.Batch(cmds...)
	case AppScreenEdit:
		return m, m.editScreen.Init()
	}
	return m, nil
}

// handleStartEdit handles transitioning to edit mode for a specific variable.
func (m mainModel) handleStartEdit(msg startEditMsg) (mainModel, tea.Cmd) {
	itemToEdit := goenv.EnvVar(msg)
	m.editScreen.editingKey = itemToEdit.Key
	m.editScreen.originalValue = itemToEdit.Value
	m.editScreen.textInput.SetValue(itemToEdit.Value)
	m.currentScreen = AppScreenEdit
	m.clearTempMessages()
	m.calculateScreenSizes()
	return m, m.editScreen.Init()
}

// handleSaveEnvRequest handles the request to save an environment variable.
func (m mainModel) handleSaveEnvRequest(msg saveEnvRequestMsg) (mainModel, tea.Cmd) {
	m.isSaving = true
	m.spinnerActive = true
	m.successMessage = fmt.Sprintf("Saving %s...", msg.Env.Key)
	return m, tea.Batch(performSetEnvCmd(msg.Env, msg.OriginalValue), m.spinner.Tick)
}

// handleEnvSetSuccess handles successful environment variable save.
func (m mainModel) handleEnvSetSuccess(msg envSetSuccessMsg) (mainModel, tea.Cmd) {
	updatedEnv := msg.Env

	// Only record history for regular edits, not undo/redo
	if !msg.IsUndoRedo && msg.OriginalValue != updatedEnv.Value {
		m.hasEditsThisSession = true
		m.editHistory = m.editHistory[:m.historyIdx]
		m.editHistory = append(m.editHistory, persist.EditRecord{
			Key:      updatedEnv.Key,
			OldValue: msg.OriginalValue,
			NewValue: updatedEnv.Value,
		})
		m.historyIdx = len(m.editHistory)

		// Save history to disk (fire-and-forget)
		m.saveHistoryAsync()
	}

	m.updateItemByKey(updatedEnv.Key, updatedEnv)

	m.clearTempMessages()
	m.isSaving = false
	m.spinnerActive = false
	if msg.IsUndoRedo {
		m.successMessage = fmt.Sprintf("Applied change to '%s'", updatedEnv.Key)
	} else {
		m.successMessage = fmt.Sprintf("Saved '%s' successfully!", updatedEnv.Key)
	}
	m.currentScreen = AppScreenList
	return m, tea.Batch(m.refreshListScreen(), clearMessageAfter(messageTimeoutSuccess))
}

// handleEnvSetError handles failed environment variable save.
func (m mainModel) handleEnvSetError(msg envSetErrorMsg) (mainModel, tea.Cmd) {
	m.isSaving = false
	m.spinnerActive = false
	errMsg := fmt.Sprintf("Error setting %s: %v", msg.Key, msg.Err)
	if !m.parsedVersion.IsZero() && !m.parsedVersion.AtLeast(goenv.MinVersionEnvWrite) {
		errMsg = fmt.Sprintf("go env -w requires Go 1.13+ (detected %s): %v", m.parsedVersion, msg.Err)
	}
	m.saveError = errMsg
	m.successMessage = ""
	m.calculateScreenSizes()
	return m, clearMessageAfter(messageTimeoutError)
}

// handleClearMessage handles clearing temporary messages.
func (m mainModel) handleClearMessage() (mainModel, tea.Cmd) {
	m.saveError = ""
	m.successMessage = ""
	m.confirmQuit = false
	m.calculateScreenSizes()
	return m, nil
}

// handleToggleSort handles toggling the sort mode.
func (m mainModel) handleToggleSort() (mainModel, tea.Cmd) {
	m.sortMode = m.sortMode.Next()
	m.successMessage = "Sort: " + m.sortMode.String()
	m.sortItems()
	// Update config synchronously, then persist from value-copied struct
	m.cfg.SortMode = persist.SortModeToString(m.sortMode)
	m.saveConfigAsync()
	return m, tea.Batch(m.refreshListScreen(), clearMessageAfter(messageTimeoutInfo))
}

// handleReloadEnv handles the request to reload environment variables.
func (m mainModel) handleReloadEnv() (mainModel, tea.Cmd) {
	m.isReloading = true
	m.spinnerActive = true
	m.successMessage = "Reloading..."
	return m, tea.Batch(performReloadEnvCmd(), m.spinner.Tick)
}

// handleReloadEnvSuccess handles successful environment reload.
func (m mainModel) handleReloadEnvSuccess(msg reloadEnvSuccessMsg) (mainModel, tea.Cmd) {
	m.isReloading = false
	m.spinnerActive = false

	// In watch mode, detect changed values and track them
	if m.watchMode {
		oldValues := make(map[string]string)
		for _, ev := range m.items {
			oldValues[ev.Key] = ev.Value
		}
		now := time.Now()
		for _, ev := range msg.Items {
			if old, ok := oldValues[ev.Key]; ok && old != ev.Value {
				m.recentChanges[ev.Key] = now
			}
		}
		// Clean up stale entries older than 10 seconds
		for k, t := range m.recentChanges {
			if time.Since(t) >= 10*time.Second {
				delete(m.recentChanges, k)
			}
		}
	}

	m.items = msg.Items
	m.sortItems()
	m.successMessage = fmt.Sprintf("Reloaded %d variables", len(m.items))
	return m, tea.Batch(m.refreshListScreen(), clearMessageAfter(messageTimeoutInfo))
}

// handleReloadEnvError handles failed environment reload.
func (m mainModel) handleReloadEnvError(msg reloadEnvErrorMsg) (mainModel, tea.Cmd) {
	m.isReloading = false
	m.spinnerActive = false
	errMsg := fmt.Sprintf("Reload failed: %v", msg.Err)
	if !m.parsedVersion.IsZero() && !m.parsedVersion.AtLeast(goenv.MinVersionEnvJSON) {
		errMsg = fmt.Sprintf("go env -json requires Go 1.9+ (detected %s): %v", m.parsedVersion, msg.Err)
	}
	m.saveError = errMsg
	m.calculateScreenSizes()
	return m, clearMessageAfter(messageTimeoutError)
}

// handleCopyValue handles copying a variable's value to clipboard.
func (m mainModel) handleCopyValue(msg copyValueMsg) (mainModel, tea.Cmd) {
	return m, performCopyCmd(msg.Key, msg.Value, false)
}

// handleCopyKeyValue handles copying KEY=VALUE to clipboard.
func (m mainModel) handleCopyKeyValue(msg copyKeyValueMsg) (mainModel, tea.Cmd) {
	return m, performCopyCmd(msg.Key, msg.Value, true)
}

// handleClipboardResult handles clipboard operation result.
func (m mainModel) handleClipboardResult(msg clipboardResultMsg) (mainModel, tea.Cmd) {
	if msg.Success {
		m.successMessage = msg.Message
	} else {
		m.saveError = msg.Message
	}
	return m, clearMessageAfter(messageTimeoutInfo)
}

// handleShowHelp handles transitioning to the help screen.
func (m mainModel) handleShowHelp() (mainModel, tea.Cmd) {
	m.currentScreen = AppScreenHelp
	return m, m.helpScreen.Init()
}

// handleUndo handles undoing the last change.
func (m mainModel) handleUndo() (mainModel, tea.Cmd) {
	if m.isSaving {
		m.successMessage = "Please wait, saving in progress..."
		return m, clearMessageAfter(messageTimeoutInfo)
	}
	if m.historyIdx <= 0 {
		m.successMessage = "Nothing to undo"
		return m, clearMessageAfter(messageTimeoutInfo)
	}
	m.historyIdx--
	record := m.editHistory[m.historyIdx]
	undoEnv := goenv.EnvVar{Key: record.Key, Value: record.OldValue, Changed: true}
	m.updateItemByKey(record.Key, undoEnv)
	m.isSaving = true
	m.spinnerActive = true
	m.successMessage = fmt.Sprintf("Undoing %s...", record.Key)

	// Save updated history index (fire-and-forget)
	m.saveHistoryAsync()

	return m, tea.Batch(performUndoRedoSetEnvCmd(undoEnv), m.refreshListScreen(), m.spinner.Tick)
}

// handleRedo handles redoing the last undone change.
func (m mainModel) handleRedo() (mainModel, tea.Cmd) {
	if m.isSaving {
		m.successMessage = "Please wait, saving in progress..."
		return m, clearMessageAfter(messageTimeoutInfo)
	}
	if m.historyIdx >= len(m.editHistory) {
		m.successMessage = "Nothing to redo"
		return m, clearMessageAfter(messageTimeoutInfo)
	}
	record := m.editHistory[m.historyIdx]
	m.historyIdx++
	redoEnv := goenv.EnvVar{Key: record.Key, Value: record.NewValue, Changed: true}
	m.updateItemByKey(record.Key, redoEnv)
	m.isSaving = true
	m.spinnerActive = true
	m.successMessage = fmt.Sprintf("Redoing %s...", record.Key)

	// Save updated history index (fire-and-forget)
	m.saveHistoryAsync()

	return m, tea.Batch(performUndoRedoSetEnvCmd(redoEnv), m.refreshListScreen(), m.spinner.Tick)
}

// handleToggleFavorite handles toggling favorite status for a variable.
func (m mainModel) handleToggleFavorite(msg toggleFavoriteMsg) (mainModel, tea.Cmd) {
	if m.favorites[msg.Key] {
		delete(m.favorites, msg.Key)
		m.successMessage = fmt.Sprintf("Removed '%s' from favorites", msg.Key)
	} else {
		m.favorites[msg.Key] = true
		m.successMessage = fmt.Sprintf("Added '%s' to favorites", msg.Key)
	}

	// Update config synchronously, then persist from a snapshot
	m.syncFavoritesToConfig()
	m.saveConfigAsync()

	m.sortItems()
	return m, tea.Batch(m.refreshListScreen(), clearMessageAfter(messageTimeoutInfo))
}

// handleShowExport handles transitioning to the export screen.
func (m mainModel) handleShowExport() (mainModel, tea.Cmd) {
	m.exportScreen.SetItems(m.items)
	m.currentScreen = AppScreenExport
	return m, m.exportScreen.Init()
}

// handleShowImport handles transitioning to the import screen.
func (m mainModel) handleShowImport() (mainModel, tea.Cmd) {
	m.importScreen.SetItems(m.items)
	m.currentScreen = AppScreenImport
	return m, m.importScreen.Init()
}

// handleShowPresets handles transitioning to the presets screen.
func (m mainModel) handleShowPresets() (mainModel, tea.Cmd) {
	m.presetScreen.SetItems(m.items)
	m.currentScreen = AppScreenPresets
	return m, m.presetScreen.Init()
}

// handleShowCompare handles transitioning to the compare screen.
func (m mainModel) handleShowCompare() (mainModel, tea.Cmd) {
	m.compareScreen.SetItems(m.items)
	m.currentScreen = AppScreenCompare
	return m, m.compareScreen.Init()
}

// handleShowStats handles transitioning to the stats screen.
func (m mainModel) handleShowStats() (mainModel, tea.Cmd) {
	m.statsScreen.SetData(m.items, m.favorites, m.editHistory)
	m.currentScreen = AppScreenStats
	return m, m.statsScreen.Init()
}

// handleExportSnapshot handles the export snapshot request.
func (m mainModel) handleExportSnapshot(msg exportSnapshotMsg) (mainModel, tea.Cmd) {
	m.isSaving = true
	m.spinnerActive = true
	m.successMessage = "Exporting..."
	return m, tea.Batch(performExportCmd(msg.Name, msg.Description, msg.Items), m.spinner.Tick)
}

// handleExportSuccess handles successful snapshot export.
func (m mainModel) handleExportSuccess(msg exportSuccessMsg) (mainModel, tea.Cmd) {
	m.clearTempMessages()
	m.isSaving = false
	m.spinnerActive = false
	m.successMessage = fmt.Sprintf("Exported to %s", msg.FilePath)
	m.currentScreen = AppScreenList
	return m, tea.Batch(m.refreshListScreen(), clearMessageAfter(messageTimeoutSuccess))
}

// handleExportError handles failed snapshot export.
func (m mainModel) handleExportError(msg exportErrorMsg) (mainModel, tea.Cmd) {
	m.isSaving = false
	m.spinnerActive = false
	m.saveError = fmt.Sprintf("Export failed: %v", msg.Err)
	m.calculateScreenSizes()
	return m, clearMessageAfter(messageTimeoutError)
}

// handleImportApply handles applying imported variables.
func (m mainModel) handleImportApply(msg importApplyMsg) (mainModel, tea.Cmd) {
	m.isSaving = true
	m.spinnerActive = true
	m.successMessage = fmt.Sprintf("Applying %d variables...", len(msg.Variables))
	return m, tea.Batch(performImportApplyCmd(msg.Variables), m.spinner.Tick)
}

// handleImportSuccess handles successful import application.
func (m mainModel) handleImportSuccess(msg importSuccessMsg) (mainModel, tea.Cmd) {
	m.clearTempMessages()
	m.isSaving = false
	m.spinnerActive = false
	if msg.FailedCount > 0 {
		m.successMessage = fmt.Sprintf("Applied %d variable(s), %d failed", msg.Count, msg.FailedCount)
	} else {
		m.successMessage = fmt.Sprintf("Applied %d variables", msg.Count)
	}
	m.currentScreen = AppScreenList
	return m, tea.Batch(performReloadEnvCmd(), clearMessageAfter(messageTimeoutSuccess))
}

// handleImportError handles failed import application.
func (m mainModel) handleImportError(msg importErrorMsg) (mainModel, tea.Cmd) {
	m.isSaving = false
	m.spinnerActive = false
	m.saveError = fmt.Sprintf("Import failed: %v", msg.Err)
	m.calculateScreenSizes()
	return m, clearMessageAfter(messageTimeoutError)
}

// handlePresetApply handles applying a preset.
func (m mainModel) handlePresetApply(msg presetApplyMsg) (mainModel, tea.Cmd) {
	m.isSaving = true
	m.spinnerActive = true
	m.successMessage = fmt.Sprintf("Applying preset '%s'...", msg.Preset.Name)

	var varsToApply []goenv.EnvVar
	for key, val := range msg.Preset.Variables {
		varsToApply = append(varsToApply, goenv.EnvVar{Key: key, Value: val, Changed: true})
	}

	return m, tea.Batch(performImportApplyCmd(varsToApply), m.spinner.Tick)
}

// handlePresetCreateSuccess handles successful preset creation.
func (m mainModel) handlePresetCreateSuccess(msg presetCreateSuccessMsg) (mainModel, tea.Cmd) {
	m.successMessage = fmt.Sprintf("Created preset '%s'", msg.Name)
	m.presetScreen.phase = presetPhaseList
	return m, tea.Batch(m.presetScreen.loadPresets(), clearMessageAfter(messageTimeoutSuccess))
}

// handlePresetDeleteSuccess handles successful preset deletion.
func (m mainModel) handlePresetDeleteSuccess(msg presetDeleteSuccessMsg) (mainModel, tea.Cmd) {
	m.successMessage = fmt.Sprintf("Deleted preset '%s'", msg.Name)
	m.presetScreen.phase = presetPhaseList
	return m, tea.Batch(m.presetScreen.loadPresets(), clearMessageAfter(messageTimeoutSuccess))
}

// handleDefault delegates unhandled messages to the current screen.
func (m mainModel) handleDefault(msg tea.Msg) (mainModel, tea.Cmd) {
	var cmd tea.Cmd
	switch m.currentScreen {
	case AppScreenList:
		cmd = m.listScreen.Update(msg)
	case AppScreenEdit:
		cmd = m.editScreen.Update(msg)
	case AppScreenHelp:
		cmd = m.helpScreen.Update(msg)
	case AppScreenExport:
		cmd = m.exportScreen.Update(msg)
	case AppScreenImport:
		cmd = m.importScreen.Update(msg)
	case AppScreenPresets:
		cmd = m.presetScreen.Update(msg)
	case AppScreenCompare:
		cmd = m.compareScreen.Update(msg)
	case AppScreenStats:
		cmd = m.statsScreen.Update(msg)
	case AppScreenReset:
		cmd = m.resetScreen.Update(msg)
	case AppScreenShellExport:
		cmd = m.shellExportScreen.Update(msg)
	case AppScreenCommandPalette:
		cmd = m.commandPalette.Update(msg)
	}
	return m, cmd
}

// handleResetEnv handles the request to reset a single variable.
func (m mainModel) handleResetEnv(msg resetEnvMsg) (mainModel, tea.Cmd) {
	m.resetScreen.SetSingleReset(msg.Key)
	m.currentScreen = AppScreenReset
	return m, m.resetScreen.Init()
}

// handleConfirmReset handles the confirmed reset of a single variable.
func (m mainModel) handleConfirmReset(msg confirmResetMsg) (mainModel, tea.Cmd) {
	m.isSaving = true
	m.successMessage = fmt.Sprintf("Resetting %s...", msg.Key)
	return m, performUnsetEnvCmd(msg.Key)
}

// handleEnvUnsetSuccess handles successful variable reset.
func (m mainModel) handleEnvUnsetSuccess(msg envUnsetSuccessMsg) (mainModel, tea.Cmd) {
	m.clearTempMessages()
	m.isSaving = false
	m.successMessage = fmt.Sprintf("Reset '%s' to default", msg.Key)
	m.currentScreen = AppScreenList
	return m, tea.Batch(performReloadEnvCmd(), clearMessageAfter(messageTimeoutSuccess))
}

// handleEnvUnsetError handles failed variable reset.
func (m mainModel) handleEnvUnsetError(msg envUnsetErrorMsg) (mainModel, tea.Cmd) {
	m.isSaving = false
	errMsg := fmt.Sprintf("Error resetting %s: %v", msg.Key, msg.Err)
	if !m.parsedVersion.IsZero() && !m.parsedVersion.AtLeast(goenv.MinVersionEnvWrite) {
		errMsg = fmt.Sprintf("go env -u requires Go 1.13+ (detected %s): %v", m.parsedVersion, msg.Err)
	}
	m.saveError = errMsg
	m.currentScreen = AppScreenList
	m.calculateScreenSizes()
	return m, clearMessageAfter(messageTimeoutError)
}

// handleShowResetScreen handles transitioning to the batch reset screen.
func (m mainModel) handleShowResetScreen() (mainModel, tea.Cmd) {
	m.resetScreen.SetItems(m.items)
	m.currentScreen = AppScreenReset
	return m, m.resetScreen.Init()
}

// handleBatchUnsetSuccess handles successful batch reset.
func (m mainModel) handleBatchUnsetSuccess(msg batchUnsetSuccessMsg) (mainModel, tea.Cmd) {
	m.clearTempMessages()
	m.isSaving = false
	if msg.FailedCount > 0 {
		m.successMessage = fmt.Sprintf("Reset %d variable(s), %d failed", msg.Count, msg.FailedCount)
	} else {
		m.successMessage = fmt.Sprintf("Reset %d variable(s) to defaults", msg.Count)
	}
	m.currentScreen = AppScreenList
	return m, tea.Batch(performReloadEnvCmd(), clearMessageAfter(messageTimeoutSuccess))
}

// handleShowShellExport handles transitioning to the shell export screen.
func (m mainModel) handleShowShellExport() (mainModel, tea.Cmd) {
	m.shellExportScreen.SetItems(m.items)
	m.currentScreen = AppScreenShellExport
	return m, m.shellExportScreen.Init()
}

// handleShowCommandPalette handles transitioning to the command palette.
func (m mainModel) handleShowCommandPalette() (mainModel, tea.Cmd) {
	m.currentScreen = AppScreenCommandPalette
	return m, m.commandPalette.Init()
}

// handleCommandPaletteEdit handles the "edit" command from the palette.
func (m mainModel) handleCommandPaletteEdit() (mainModel, tea.Cmd) {
	m.currentScreen = AppScreenList
	if ei, ok := m.selectedEnvItem(); ok {
		return m, sendMsg(startEditMsg(ei.EnvVar))
	}
	return m, m.refreshListScreen()
}

// handleCommandPaletteCopy handles the "copy" command from the palette.
func (m mainModel) handleCommandPaletteCopy() (mainModel, tea.Cmd) {
	m.currentScreen = AppScreenList
	if ei, ok := m.selectedEnvItem(); ok {
		return m, tea.Batch(m.refreshListScreen(), sendMsg(copyValueMsg{Key: ei.Key, Value: ei.Value}))
	}
	return m, m.refreshListScreen()
}

// handleCommandPaletteCopyKey handles the "copykey" command from the palette.
func (m mainModel) handleCommandPaletteCopyKey() (mainModel, tea.Cmd) {
	m.currentScreen = AppScreenList
	if ei, ok := m.selectedEnvItem(); ok {
		return m, tea.Batch(m.refreshListScreen(), sendMsg(copyKeyValueMsg{Key: ei.Key, Value: ei.Value}))
	}
	return m, m.refreshListScreen()
}

// handleCommandPaletteFavorite handles the "favorite" command from the palette.
func (m mainModel) handleCommandPaletteFavorite() (mainModel, tea.Cmd) {
	m.currentScreen = AppScreenList
	if ei, ok := m.selectedEnvItem(); ok {
		return m, tea.Batch(m.refreshListScreen(), sendMsg(toggleFavoriteMsg{Key: ei.Key}))
	}
	return m, m.refreshListScreen()
}

// handleCommandPaletteReset handles the "reset" command from the palette.
func (m mainModel) handleCommandPaletteReset() (mainModel, tea.Cmd) {
	m.currentScreen = AppScreenList
	if ei, ok := m.selectedEnvItem(); ok && ei.Changed {
		return m, sendMsg(resetEnvMsg{Key: ei.Key})
	}
	return m, m.refreshListScreen()
}

// handleCycleCategory handles cycling the category filter.
func (m mainModel) handleCycleCategory() (mainModel, tea.Cmd) {
	m.categoryFilter = goenv.NextCategory(m.categoryFilter)
	if m.categoryFilter == "" {
		m.successMessage = "Category: All"
	} else {
		m.successMessage = fmt.Sprintf("Category: %s", m.categoryFilter)
	}
	// Reset selection since item set changes with category
	m.listScreen.list.Select(0)
	return m, tea.Batch(m.refreshListScreen(), clearMessageAfter(messageTimeoutInfo))
}

// handleToggleTheme handles cycling to the next color theme.
func (m mainModel) handleToggleTheme() (mainModel, tea.Cmd) {
	next := NextTheme(m.cfg.Theme)
	m.cfg.Theme = next
	theme := GetTheme(next)
	SetActiveTheme(theme)
	m.successMessage = "Theme: " + theme.Name
	m.saveConfigAsync()
	return m, tea.Batch(m.refreshListScreen(), clearMessageAfter(messageTimeoutInfo))
}

// handleToggleWatch handles toggling watch mode on/off.
func (m mainModel) handleToggleWatch() (mainModel, tea.Cmd) {
	m.watchMode = !m.watchMode
	if m.watchMode {
		m.successMessage = fmt.Sprintf("Watch mode ON (every %ds)", int(m.watchInterval.Seconds()))
		return m, tea.Batch(
			tea.Tick(m.watchInterval, func(time.Time) tea.Msg { return watchTickMsg{} }),
			clearMessageAfter(messageTimeoutInfo),
		)
	}
	m.successMessage = "Watch mode OFF"
	return m, clearMessageAfter(messageTimeoutInfo)
}

// handleWatchTick handles the periodic watch tick — triggers reload if still watching.
func (m mainModel) handleWatchTick() (mainModel, tea.Cmd) {
	if !m.watchMode || m.currentScreen != AppScreenList {
		return m, nil
	}
	// Continue the tick chain and trigger a reload
	return m, tea.Batch(
		performReloadEnvCmd(),
		tea.Tick(m.watchInterval, func(time.Time) tea.Msg { return watchTickMsg{} }),
	)
}
