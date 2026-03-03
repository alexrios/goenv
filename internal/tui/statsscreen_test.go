package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

// --- StatsScreenModel Tests ---

func TestStatsScreen_NewModel(t *testing.T) {
	m := NewStatsScreenModel()

	if m.items != nil && len(m.items) != 0 {
		t.Error("NewModel: items should be empty/nil")
	}
}

func TestStatsScreen_Init(t *testing.T) {
	m := NewStatsScreenModel()
	m.SetData(
		[]goenv.EnvVar{{Key: "GOPATH", Value: "/go"}},
		map[string]bool{},
		[]persist.EditRecord{},
	)

	cmd := m.Init()

	if cmd != nil {
		t.Error("Init should return nil command")
	}
	if m.stats.TotalVars != 1 {
		t.Errorf("Init: TotalVars = %d, want 1", m.stats.TotalVars)
	}
}

func TestStatsScreen_ComputeStats_TotalVars(t *testing.T) {
	m := NewStatsScreenModel()
	m.SetData(
		[]goenv.EnvVar{
			{Key: "GOPATH", Value: "/go"},
			{Key: "GOROOT", Value: "/usr/local/go"},
			{Key: "GOBIN", Value: "/go/bin"},
		},
		map[string]bool{},
		[]persist.EditRecord{},
	)
	m.computeStats()

	if m.stats.TotalVars != 3 {
		t.Errorf("TotalVars = %d, want 3", m.stats.TotalVars)
	}
}

func TestStatsScreen_ComputeStats_ModifiedCount(t *testing.T) {
	m := NewStatsScreenModel()
	m.SetData(
		[]goenv.EnvVar{
			{Key: "GOPATH", Value: "/go", Changed: true},
			{Key: "GOROOT", Value: "/usr/local/go", Changed: false},
			{Key: "GOBIN", Value: "/go/bin", Changed: true},
		},
		map[string]bool{},
		[]persist.EditRecord{},
	)
	m.computeStats()

	if m.stats.ModifiedCount != 2 {
		t.Errorf("ModifiedCount = %d, want 2", m.stats.ModifiedCount)
	}
}

func TestStatsScreen_ComputeStats_FavoritesCount(t *testing.T) {
	m := NewStatsScreenModel()
	m.SetData(
		[]goenv.EnvVar{
			{Key: "GOPATH", Value: "/go"},
			{Key: "GOROOT", Value: "/usr/local/go"},
			{Key: "GOBIN", Value: "/go/bin"},
		},
		map[string]bool{"GOPATH": true, "GOBIN": true},
		[]persist.EditRecord{},
	)
	m.computeStats()

	if m.stats.FavoritesCount != 2 {
		t.Errorf("FavoritesCount = %d, want 2", m.stats.FavoritesCount)
	}
}

func TestStatsScreen_ComputeStats_CategoryBreakdown(t *testing.T) {
	m := NewStatsScreenModel()
	m.SetData(
		[]goenv.EnvVar{
			{Key: "GOPATH", Value: "/go"},       // CategoryGeneral
			{Key: "GOROOT", Value: "/usr"},       // CategoryGeneral
			{Key: "CGO_ENABLED", Value: "1"},     // CategoryCGO
			{Key: "GOOS", Value: "linux"},        // CategoryArch
			{Key: "GOARCH", Value: "amd64"},      // CategoryArch
			{Key: "GOPROXY", Value: "direct"},    // CategoryProxy
		},
		map[string]bool{},
		[]persist.EditRecord{},
	)
	m.computeStats()

	if m.stats.ByCategory[goenv.CategoryGeneral] != 2 {
		t.Errorf("CategoryGeneral count = %d, want 2", m.stats.ByCategory[goenv.CategoryGeneral])
	}
	if m.stats.ByCategory[goenv.CategoryCGO] != 1 {
		t.Errorf("CategoryCGO count = %d, want 1", m.stats.ByCategory[goenv.CategoryCGO])
	}
	if m.stats.ByCategory[goenv.CategoryArch] != 2 {
		t.Errorf("CategoryArch count = %d, want 2", m.stats.ByCategory[goenv.CategoryArch])
	}
	if m.stats.ByCategory[goenv.CategoryProxy] != 1 {
		t.Errorf("CategoryProxy count = %d, want 1", m.stats.ByCategory[goenv.CategoryProxy])
	}
}

func TestStatsScreen_ComputeStats_RecentEdits(t *testing.T) {
	m := NewStatsScreenModel()
	history := []persist.EditRecord{
		{Key: "A", OldValue: "1", NewValue: "2"},
		{Key: "B", OldValue: "3", NewValue: "4"},
		{Key: "C", OldValue: "5", NewValue: "6"},
		{Key: "D", OldValue: "7", NewValue: "8"},
		{Key: "E", OldValue: "9", NewValue: "10"},
		{Key: "F", OldValue: "11", NewValue: "12"}, // 6th entry
	}
	m.SetData(
		[]goenv.EnvVar{{Key: "GOPATH", Value: "/go"}},
		map[string]bool{},
		history,
	)
	m.computeStats()

	// Should only keep last 5
	if len(m.stats.RecentEdits) != 5 {
		t.Errorf("RecentEdits length = %d, want 5", len(m.stats.RecentEdits))
	}

	// Should be in reverse order (most recent first)
	if m.stats.RecentEdits[0].Key != "F" {
		t.Errorf("RecentEdits[0].Key = %q, want 'F' (most recent)", m.stats.RecentEdits[0].Key)
	}
	if m.stats.RecentEdits[4].Key != "B" {
		t.Errorf("RecentEdits[4].Key = %q, want 'B' (oldest of last 5)", m.stats.RecentEdits[4].Key)
	}
}

func TestStatsScreen_ComputeStats_RecentEdits_LessThanMax(t *testing.T) {
	m := NewStatsScreenModel()
	history := []persist.EditRecord{
		{Key: "A", OldValue: "1", NewValue: "2"},
		{Key: "B", OldValue: "3", NewValue: "4"},
	}
	m.SetData(
		[]goenv.EnvVar{{Key: "GOPATH", Value: "/go"}},
		map[string]bool{},
		history,
	)
	m.computeStats()

	if len(m.stats.RecentEdits) != 2 {
		t.Errorf("RecentEdits length = %d, want 2", len(m.stats.RecentEdits))
	}

	// Should be in reverse order
	if m.stats.RecentEdits[0].Key != "B" {
		t.Errorf("RecentEdits[0].Key = %q, want 'B'", m.stats.RecentEdits[0].Key)
	}
}

func TestStatsScreen_ComputeStats_EmptyHistory(t *testing.T) {
	m := NewStatsScreenModel()
	m.SetData(
		[]goenv.EnvVar{{Key: "GOPATH", Value: "/go"}},
		map[string]bool{},
		[]persist.EditRecord{},
	)
	m.computeStats()

	if len(m.stats.RecentEdits) != 0 {
		t.Errorf("RecentEdits length = %d, want 0", len(m.stats.RecentEdits))
	}
}

func TestStatsScreen_EscReturnsToList(t *testing.T) {
	m := NewStatsScreenModel()

	msg := tea.KeyPressMsg{Code: tea.KeyEscape}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("Esc should return command")
	}

	resultMsg := cmd()
	changeMsg, ok := resultMsg.(changeScreenMsg)
	if !ok {
		t.Fatalf("expected changeScreenMsg, got %T", resultMsg)
	}
	if AppScreen(changeMsg) != AppScreenList {
		t.Errorf("changeScreenMsg = %v, want AppScreenList", changeMsg)
	}
}

func TestStatsScreen_QReturnsToList(t *testing.T) {
	m := NewStatsScreenModel()

	msg := tea.KeyPressMsg{Code: 'q', Text: "q"}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("q should return command")
	}

	resultMsg := cmd()
	changeMsg, ok := resultMsg.(changeScreenMsg)
	if !ok {
		t.Fatalf("expected changeScreenMsg, got %T", resultMsg)
	}
	if AppScreen(changeMsg) != AppScreenList {
		t.Errorf("changeScreenMsg = %v, want AppScreenList", changeMsg)
	}
}

func TestStatsScreen_SReturnsToList(t *testing.T) {
	m := NewStatsScreenModel()

	msg := tea.KeyPressMsg{Code: 'S', Text: "S"}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("S should return command")
	}

	resultMsg := cmd()
	changeMsg, ok := resultMsg.(changeScreenMsg)
	if !ok {
		t.Fatalf("expected changeScreenMsg, got %T", resultMsg)
	}
	if AppScreen(changeMsg) != AppScreenList {
		t.Errorf("changeScreenMsg = %v, want AppScreenList", changeMsg)
	}
}

func TestStatsScreen_View(t *testing.T) {
	m := NewStatsScreenModel()
	m.SetData(
		[]goenv.EnvVar{
			{Key: "GOPATH", Value: "/go", Changed: true},
		},
		map[string]bool{"GOPATH": true},
		[]persist.EditRecord{{Key: "GOPATH", OldValue: "/old", NewValue: "/go"}},
	)
	m.computeStats()

	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
	// Check for expected sections
	if !strings.Contains(view, "Statistics") {
		t.Error("View should contain 'Statistics'")
	}
}

func TestStatsScreen_SetData(t *testing.T) {
	m := NewStatsScreenModel()

	items := []goenv.EnvVar{{Key: "GOPATH", Value: "/go"}}
	favorites := map[string]bool{"GOPATH": true}
	history := []persist.EditRecord{{Key: "GOPATH", OldValue: "/old", NewValue: "/go"}}

	m.SetData(items, favorites, history)

	if len(m.items) != 1 {
		t.Errorf("items length = %d, want 1", len(m.items))
	}
	if !m.favorites["GOPATH"] {
		t.Error("GOPATH should be in favorites")
	}
	if len(m.editHistory) != 1 {
		t.Errorf("editHistory length = %d, want 1", len(m.editHistory))
	}
}
