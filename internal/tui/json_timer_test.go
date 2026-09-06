package tui

import (
	"strings"
	"testing"
	"testing/synctest"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

func TestShellExportJSONError(t *testing.T) {
	m := NewShellExportScreenModel()
	m.SetItems([]goenv.EnvVar{{Key: "GOPATH", Value: "\xff", Changed: true}})
	m.shell = goenv.ShellJSON
	if cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter}); cmd != nil {
		t.Fatal("unexpected command on error")
	}
	if m.phase != shellExportPhaseOptions || len(m.previewLines) != 0 || m.exportError == "" {
		t.Fatalf("invalid export state: %#v", m)
	}
	if !strings.Contains(m.View(), "failed to marshal JSON export") {
		t.Fatal("error missing from options")
	}
	if cmd := m.Update(tea.KeyPressMsg{Code: 'c', Text: "c"}); cmd != nil || m.copied {
		t.Fatal("copied invalid export")
	}
	m.items[0].Value = "/go"
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.phase != shellExportPhasePreview || m.exportError != "" || !strings.Contains(m.View(), "/go") {
		t.Fatal("valid retry did not open preview")
	}
}

func TestClearMessageTimer(t *testing.T) {
	for _, delay := range []time.Duration{messageTimeoutInfo, messageTimeoutSuccess, messageTimeoutError} {
		t.Run(delay.String(), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				assertTimerDeadline(t, clearMessageAfter(delay), delay, clearMessageMsg{})
			})
		})
	}
}

func TestWatchTimerDeadlines(t *testing.T) {
	for _, mode := range []string{"enable", "repeat"} {
		t.Run(mode, func(t *testing.T) {
			m := NewMainModel(nil, nil, persist.Config{WatchInterval: 7})
			synctest.Test(t, func(t *testing.T) {
				var cmd tea.Cmd
				var tick tea.Cmd
				if mode == "enable" {
					_, cmd = m.handleToggleWatch()
					tick = cmd().(tea.BatchMsg)[0]
				} else {
					m.watchMode = true
					_, cmd = m.handleWatchTick()
					// Do not execute the reload command: it performs real I/O.
					tick = cmd().(tea.BatchMsg)[1]
				}
				assertTimerDeadline(t, tick, 7*time.Second, watchTickMsg{})
			})
		})
	}
}

func assertTimerDeadline(t *testing.T, cmd tea.Cmd, delay time.Duration, want tea.Msg) {
	t.Helper()
	delivered := make(chan tea.Msg, 1)
	go func() { delivered <- cmd() }()
	synctest.Sleep(delay - time.Nanosecond)
	select {
	case msg := <-delivered:
		t.Fatalf("early message: %T", msg)
	default:
	}
	synctest.Sleep(time.Nanosecond)
	select {
	case msg := <-delivered:
		switch want.(type) {
		case clearMessageMsg:
			if _, ok := msg.(clearMessageMsg); !ok {
				t.Fatalf("got %T", msg)
			}
		case watchTickMsg:
			if _, ok := msg.(watchTickMsg); !ok {
				t.Fatalf("got %T", msg)
			}
		}
	default:
		t.Fatal("message not delivered at deadline")
	}
}

func TestExperimentSuggestionsAndFreeEntry(t *testing.T) {
	for _, ver := range []string{"go1.25.0", "go1.26.0", "go1.27.0", "unknown"} {
		t.Run(ver, func(t *testing.T) {
			m := NewEditScreenModel(goenv.ParseGoVersion(ver))
			m.editingKey = "GOEXPERIMENT"
			m.textInput.SetValue("")
			m.updateSuggestions()
			rendered := m.renderSuggestions()
			if !strings.Contains(rendered, "jsonv2") {
				t.Fatal("JSON experiment not rendered")
			}
			if (ver == "go1.26.0" || ver == "go1.27.0") && (strings.Contains(rendered, "rangefunc") || strings.Contains(rendered, "cacheprog")) {
				t.Fatal("removed experiment rendered")
			}
			m.textInput.SetValue("customexperiment")
			m.updateSuggestions()
			cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			if cmd == nil {
				t.Fatal("free-form experiment blocked")
			}
			if _, ok := cmd().(saveEnvRequestMsg); !ok {
				t.Fatal("free-form experiment not submitted")
			}
		})
	}
}
