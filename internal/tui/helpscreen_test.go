package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestRegression_HelpScreenKeyFilter verifies that only Esc, q, and ? keys
// close the help screen. Other keys should be ignored to prevent accidental
// navigation away from the help screen.
func TestRegression_HelpScreenKeyFilter(t *testing.T) {
	// Keys that SHOULD close help
	closeTests := []struct {
		name string
		msg  tea.KeyPressMsg
	}{
		{"esc", tea.KeyPressMsg{Code: tea.KeyEscape}},
		{"q", tea.KeyPressMsg{Code: 'q', Text: "q"}},
		{"?", tea.KeyPressMsg{Code: '?', Text: "?"}},
	}

	for _, tc := range closeTests {
		t.Run("close_"+tc.name, func(t *testing.T) {
			m := NewHelpScreenModel()
			cmd := m.Update(tc.msg)

			if cmd == nil {
				t.Errorf("key %q should close help screen", tc.name)
				return
			}

			// Execute the command and verify it returns changeScreenMsg
			resultMsg := cmd()
			if changeMsg, ok := resultMsg.(changeScreenMsg); !ok {
				t.Errorf("key %q should return changeScreenMsg, got %T", tc.name, resultMsg)
			} else if AppScreen(changeMsg) != AppScreenList {
				t.Errorf("key %q should return AppScreenList, got %v", tc.name, changeMsg)
			}
		})
	}

	// Keys that should NOT close help
	ignoreTests := []struct {
		name string
		msg  tea.KeyPressMsg
	}{
		{"a", tea.KeyPressMsg{Code: 'a', Text: "a"}},
		{"enter", tea.KeyPressMsg{Code: tea.KeyEnter}},
		{"space", tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}},
		{"j", tea.KeyPressMsg{Code: 'j', Text: "j"}},
		{"k", tea.KeyPressMsg{Code: 'k', Text: "k"}},
		{"s", tea.KeyPressMsg{Code: 's', Text: "s"}},
		{"r", tea.KeyPressMsg{Code: 'r', Text: "r"}},
	}

	for _, tc := range ignoreTests {
		t.Run("ignore_"+tc.name, func(t *testing.T) {
			m := NewHelpScreenModel()
			cmd := m.Update(tc.msg)

			if cmd != nil {
				t.Errorf("key %q should NOT close help screen, but returned a command", tc.name)
			}
		})
	}
}
