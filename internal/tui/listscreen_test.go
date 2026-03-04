package tui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

func testItems() []list.Item {
	return []list.Item{
		envItem{goenv.EnvVar{Key: "GOPATH", Value: "/go"}},
		envItem{goenv.EnvVar{Key: "GOOS", Value: "linux"}},
		envItem{goenv.EnvVar{Key: "GOARCH", Value: "amd64"}},
	}
}

func TestNewListScreenModel_Basic(t *testing.T) {
	m := NewListScreenModel(testItems())
	if len(m.list.Items()) != 3 {
		t.Errorf("item count = %d, want 3", len(m.list.Items()))
	}
}

func TestNewListScreenModel_TitleIncludesVersion(t *testing.T) {
	m := NewListScreenModelWithVersion(testItems(), "go1.21.0")
	if !strings.Contains(m.list.Title, "go1.21.0") {
		t.Errorf("title %q should contain version", m.list.Title)
	}
}

func TestNewListScreenModel_ReadOnlyWarning(t *testing.T) {
	// Go 1.12 (before go env -w was added in 1.13)
	m := NewListScreenModelWithParsedVersion(testItems(), "go1.12.0", goenv.GoVersion{Major: 1, Minor: 12, Patch: 0})
	if !strings.Contains(m.list.Title, "read-only") {
		t.Errorf("title %q should contain read-only warning for Go 1.12", m.list.Title)
	}
}

func TestNewListScreenModel_NoChangeDetectionWarning(t *testing.T) {
	// Go 1.14 (before -changed flag added in 1.15)
	m := NewListScreenModelWithParsedVersion(testItems(), "go1.14.0", goenv.GoVersion{Major: 1, Minor: 14, Patch: 0})
	if !strings.Contains(m.list.Title, "no change detection") {
		t.Errorf("title %q should contain change detection warning for Go 1.14", m.list.Title)
	}
}

func TestNewListScreenModel_ModernVersionNoWarning(t *testing.T) {
	m := NewListScreenModelWithParsedVersion(testItems(), "go1.21.0", goenv.GoVersion{Major: 1, Minor: 21, Patch: 0})
	if strings.Contains(m.list.Title, "read-only") || strings.Contains(m.list.Title, "no change") {
		t.Errorf("title %q should not contain warnings for Go 1.21", m.list.Title)
	}
}

func TestListScreen_EnterSendsStartEditMsg(t *testing.T) {
	m := NewListScreenModelWithSortMode(testItems(), "go1.21.0", goenv.GoVersion{Major: 1, Minor: 21, Patch: 0}, persist.SortAlpha, nil)
	m.list.SetSize(80, 24)
	m.list.Select(0)

	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command from Enter key")
	}
	msg := cmd()
	if _, ok := msg.(startEditMsg); !ok {
		t.Errorf("expected startEditMsg, got %T", msg)
	}
}

func TestListScreen_QuestionMarkShowsHelp(t *testing.T) {
	m := NewListScreenModelWithSortMode(testItems(), "", goenv.GoVersion{}, persist.SortAlpha, nil)
	m.list.SetSize(80, 24)

	cmd := m.Update(tea.KeyPressMsg{Code: '?'})
	if cmd == nil {
		t.Fatal("expected command from ? key")
	}
	msg := cmd()
	if _, ok := msg.(showHelpMsg); !ok {
		t.Errorf("expected showHelpMsg, got %T", msg)
	}
}

func TestListScreen_SSendsToggleSort(t *testing.T) {
	m := NewListScreenModelWithSortMode(testItems(), "", goenv.GoVersion{}, persist.SortAlpha, nil)
	m.list.SetSize(80, 24)

	cmd := m.Update(tea.KeyPressMsg{Code: 's'})
	if cmd == nil {
		t.Fatal("expected command from s key")
	}
	msg := cmd()
	if _, ok := msg.(toggleSortMsg); !ok {
		t.Errorf("expected toggleSortMsg, got %T", msg)
	}
}

func TestListScreen_ColonOpensCommandPalette(t *testing.T) {
	m := NewListScreenModelWithSortMode(testItems(), "", goenv.GoVersion{}, persist.SortAlpha, nil)
	m.list.SetSize(80, 24)

	cmd := m.Update(tea.KeyPressMsg{Code: ':'})
	if cmd == nil {
		t.Fatal("expected command from : key")
	}
	msg := cmd()
	if _, ok := msg.(showCommandPaletteMsg); !ok {
		t.Errorf("expected showCommandPaletteMsg, got %T", msg)
	}
}

func TestListScreen_ViewNotEmpty(t *testing.T) {
	m := NewListScreenModelWithSortMode(testItems(), "go1.21.0", goenv.GoVersion{Major: 1, Minor: 21, Patch: 0}, persist.SortAlpha, nil)
	m.list.SetSize(80, 24)

	view := m.View()
	if view == "" {
		t.Error("view should not be empty")
	}
}

func TestListScreen_InitReturnsNil(t *testing.T) {
	m := NewListScreenModel(testItems())
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestListScreen_EnterOnReadOnlySendsShowDetailMsg(t *testing.T) {
	items := []list.Item{
		envItem{goenv.EnvVar{Key: "GOVERSION", Value: "go1.21.0", ReadOnly: true}},
		envItem{goenv.EnvVar{Key: "GOPATH", Value: "/go"}},
	}
	m := NewListScreenModelWithSortMode(items, "go1.21.0", goenv.GoVersion{Major: 1, Minor: 21, Patch: 0}, persist.SortAlpha, nil)
	m.list.SetSize(80, 24)
	m.list.Select(0)

	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command from Enter key on read-only item")
	}
	msg := cmd()
	if _, ok := msg.(showDetailMsg); !ok {
		t.Errorf("expected showDetailMsg for read-only item, got %T", msg)
	}
}

func TestListScreen_EnterOnSettableSendsStartEditMsg(t *testing.T) {
	items := []list.Item{
		envItem{goenv.EnvVar{Key: "GOPATH", Value: "/go", ReadOnly: false}},
	}
	m := NewListScreenModelWithSortMode(items, "go1.21.0", goenv.GoVersion{Major: 1, Minor: 21, Patch: 0}, persist.SortAlpha, nil)
	m.list.SetSize(80, 24)
	m.list.Select(0)

	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command from Enter key on settable item")
	}
	msg := cmd()
	if _, ok := msg.(startEditMsg); !ok {
		t.Errorf("expected startEditMsg for settable item, got %T", msg)
	}
}
