package tui

import (
	tea "charm.land/bubbletea/v2"
)

// sendMsg creates a command that returns the given message.
// Reduces boilerplate in Update methods.
func sendMsg(m tea.Msg) tea.Cmd {
	return func() tea.Msg { return m }
}

// clampNonNegative ensures a value is at least 0.
func clampNonNegative(val int) int {
	return max(val, 0)
}
