package tui

import (
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

func TestWatchMode_ToggleOn(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())

	msg := toggleWatchMsg{}
	newModel, cmd := m.Update(msg)
	m = newModel.(mainModel)

	if !m.watchMode {
		t.Error("watchMode should be true after toggle on")
	}
	if cmd == nil {
		t.Error("toggle on should return a command (tick + clear message)")
	}
}

func TestWatchMode_ToggleOff(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.watchMode = true

	msg := toggleWatchMsg{}
	newModel, cmd := m.Update(msg)
	m = newModel.(mainModel)

	if m.watchMode {
		t.Error("watchMode should be false after toggle off")
	}
	if cmd == nil {
		t.Error("toggle off should return a clear message command")
	}
}

func TestWatchMode_TickTriggersReload(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.watchMode = true
	m.currentScreen = AppScreenList

	msg := watchTickMsg{}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("watchTickMsg should return a batch command when watching")
	}
}

func TestWatchMode_TickNoOpWhenNotWatching(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.watchMode = false

	msg := watchTickMsg{}
	_, cmd := m.Update(msg)

	if cmd != nil {
		t.Error("watchTickMsg should return nil when not watching")
	}
}

func TestWatchMode_TickNoOpOnNonListScreen(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	m := NewMainModel(items, nil, persist.DefaultConfig())
	m.watchMode = true
	m.currentScreen = AppScreenEdit

	msg := watchTickMsg{}
	_, cmd := m.Update(msg)

	if cmd != nil {
		t.Error("watchTickMsg should return nil on non-list screen")
	}
}

func TestWatchIntervalFromConfig(t *testing.T) {
	tests := []struct {
		input    int
		expected time.Duration
	}{
		{0, 5 * time.Second},          // zero defaults to 5s
		{-1, 5 * time.Second},         // negative defaults to 5s
		{-999, 5 * time.Second},       // large negative defaults to 5s
		{1, 1 * time.Second},          // minimum valid
		{10, 10 * time.Second},        // custom value
		{3, 3 * time.Second},          // custom value
		{86400, 86400 * time.Second},  // max (24h)
		{86401, 86400 * time.Second},  // over max clamped to 24h
		{999999, 86400 * time.Second}, // very large clamped to 24h
	}

	for _, tt := range tests {
		got := watchIntervalFromConfig(tt.input)
		if got != tt.expected {
			t.Errorf("watchIntervalFromConfig(%d) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestNewMainModel_WatchInterval_DefaultsTo5s(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	cfg := persist.Config{} // zero value, simulating omitempty decode
	m := NewMainModel(items, nil, cfg)

	if m.watchInterval != 5*time.Second {
		t.Errorf("default watchInterval = %v, want 5s", m.watchInterval)
	}
}

func TestNewMainModel_WatchInterval_FromConfig(t *testing.T) {
	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	cfg := persist.DefaultConfig()
	cfg.WatchInterval = 30
	m := NewMainModel(items, nil, cfg)

	if m.watchInterval != 30*time.Second {
		t.Errorf("watchInterval = %v, want 30s", m.watchInterval)
	}
}

func TestWatchMode_KeyBinding(t *testing.T) {
	items := []list.Item{envItem{goenv.EnvVar{Key: "GOPATH", Value: "/go"}}}
	m := NewListScreenModel(items)

	msg := tea.KeyPressMsg{Code: 'w', Text: "w"}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected command from 'w' key, got nil")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(toggleWatchMsg); !ok {
		t.Errorf("expected toggleWatchMsg, got %T", resultMsg)
	}
}
