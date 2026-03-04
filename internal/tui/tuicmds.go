package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/commands"
	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

// performSetEnvCmd creates a command that runs SetEnvVar asynchronously.
// Used for regular edits; includes original value for history recording.
func performSetEnvCmd(e goenv.EnvVar, originalValue string) tea.Cmd {
	return func() tea.Msg {
		err := commands.SetEnvVar(e)
		if err != nil {
			return envSetErrorMsg{Key: e.Key, OriginalValue: originalValue, Err: err}
		}
		return envSetSuccessMsg{Env: e, OriginalValue: originalValue}
	}
}

// performUndoRedoSetEnvCmd creates a command that runs SetEnvVar asynchronously.
// Used for undo/redo operations; marks the result to prevent history recording.
func performUndoRedoSetEnvCmd(e goenv.EnvVar) tea.Cmd {
	return func() tea.Msg {
		err := commands.SetEnvVar(e)
		if err != nil {
			return envSetErrorMsg{Key: e.Key, IsUndoRedo: true, Err: err}
		}
		return envSetSuccessMsg{Env: e, IsUndoRedo: true}
	}
}

// clearMessageAfter creates a command that sends clearMessageMsg after the specified delay.
// Used to auto-dismiss success and error messages from the status bar.
func clearMessageAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clearMessageMsg{}
	})
}

// performReloadEnvCmd creates a command that reloads environment variables asynchronously.
// Sends reloadEnvSuccessMsg on success or reloadEnvErrorMsg on failure.
func performReloadEnvCmd() tea.Cmd {
	return func() tea.Msg {
		items, err := commands.ReloadEnv()
		if err != nil {
			return reloadEnvErrorMsg{Err: err}
		}
		return reloadEnvSuccessMsg{Items: items}
	}
}

// performCopyCmd creates a command that copies text to the clipboard using
// Bubble Tea's native OSC52 clipboard support (works over SSH).
// If includeKey is true, copies "KEY=VALUE"; otherwise copies just the value.
// Sends clipboardResultMsg with the operation result.
func performCopyCmd(key, value string, includeKey bool) tea.Cmd {
	text := value
	msg := fmt.Sprintf("Copied value of %s", key)
	if includeKey {
		text = fmt.Sprintf("%s=%s", key, value)
		msg = fmt.Sprintf("Copied %s=...", key)
	}
	return tea.Batch(
		tea.SetClipboard(text),
		sendMsg(clipboardResultMsg{Success: true, Message: msg}),
	)
}

// performUnsetEnvCmd creates a command that runs UnsetEnvVar asynchronously.
func performUnsetEnvCmd(key string) tea.Cmd {
	return func() tea.Msg {
		err := commands.UnsetEnvVar(key)
		if err != nil {
			return envUnsetErrorMsg{Key: key, Err: err}
		}
		return envUnsetSuccessMsg{Key: key}
	}
}

// performBatchUnsetEnvCmd creates a command that unsets multiple variables.
func performBatchUnsetEnvCmd(keys []string) tea.Cmd {
	return func() tea.Msg {
		var successCount int
		var failedKeys []string
		for _, key := range keys {
			if err := commands.UnsetEnvVar(key); err != nil {
				failedKeys = append(failedKeys, key)
			} else {
				successCount++
			}
		}
		if len(failedKeys) > 0 && successCount == 0 {
			return envUnsetErrorMsg{Key: "batch", Err: fmt.Errorf("failed to reset: %s", strings.Join(failedKeys, ", "))}
		}
		return batchUnsetSuccessMsg{Count: successCount, FailedCount: len(failedKeys)}
	}
}

// performExportCmd creates a command that exports a snapshot.
func performExportCmd(name, description string, items []goenv.EnvVar) tea.Cmd {
	return func() tea.Msg {
		goVersion, _ := commands.GetGoVersion()
		snapshot := persist.NewSnapshot(name, items, goVersion)
		snapshot.Description = description

		dir, err := persist.DefaultSnapshotDir()
		if err != nil {
			return exportErrorMsg{Err: err}
		}

		base := persist.SanitizeFilename(name)
		filePath := persist.UniqueFilePath(dir, base, ".json")

		if err := persist.ExportSnapshot(snapshot, filePath); err != nil {
			return exportErrorMsg{Err: err}
		}

		return exportSuccessMsg{FilePath: filePath}
	}
}

// performImportApplyCmd creates a command that applies imported variables.
// Read-only variables are silently skipped.
func performImportApplyCmd(variables []goenv.EnvVar) tea.Cmd {
	return func() tea.Msg {
		var successCount, failedCount int
		for _, ev := range variables {
			if goenv.IsReadOnly(ev.Key) {
				continue
			}
			if err := commands.SetEnvVar(ev); err != nil {
				failedCount++
			} else {
				successCount++
			}
		}
		return importSuccessMsg{Count: successCount, FailedCount: failedCount}
	}
}
