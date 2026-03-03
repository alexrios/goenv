package persist

import "testing"

// --- SortMode Conversion Tests ---

func TestSortModeFromString_AllModes(t *testing.T) {
	tests := []struct {
		input    string
		expected SortMode
	}{
		{"alpha", SortAlpha},
		{"modified_first", SortModifiedFirst},
		{"category", SortCategory},
	}

	for _, tt := range tests {
		got := SortModeFromString(tt.input)
		if got != tt.expected {
			t.Errorf("SortModeFromString(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestSortModeFromString_InvalidDefaults(t *testing.T) {
	tests := []string{
		"invalid",
		"",
		"ALPHA",
		"alphabetical",
		"modified",
	}

	for _, input := range tests {
		got := SortModeFromString(input)
		if got != SortAlpha {
			t.Errorf("SortModeFromString(%q) = %v, want SortAlpha (default)", input, got)
		}
	}
}

func TestSortModeToString_AllModes(t *testing.T) {
	tests := []struct {
		input    SortMode
		expected string
	}{
		{SortAlpha, "alpha"},
		{SortModifiedFirst, "modified_first"},
		{SortCategory, "category"},
	}

	for _, tt := range tests {
		got := SortModeToString(tt.input)
		if got != tt.expected {
			t.Errorf("SortModeToString(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSortModeToString_InvalidDefaults(t *testing.T) {
	// Invalid SortMode value should default to "alpha"
	got := SortModeToString(SortMode(999))
	if got != "alpha" {
		t.Errorf("SortModeToString(invalid) = %q, want 'alpha'", got)
	}
}

func TestSortMode_RoundTrip(t *testing.T) {
	modes := []SortMode{SortAlpha, SortModifiedFirst, SortCategory}

	for _, mode := range modes {
		str := SortModeToString(mode)
		roundTrip := SortModeFromString(str)
		if roundTrip != mode {
			t.Errorf("round-trip failed for %v: FromString(ToString(%v)) = %v", mode, mode, roundTrip)
		}
	}
}

// --- DefaultConfig Tests ---

func TestDefaultConfig_Values(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.SortMode != "alpha" {
		t.Errorf("DefaultConfig().SortMode = %q, want 'alpha'", cfg.SortMode)
	}
	if cfg.Theme != ThemeDefault {
		t.Errorf("DefaultConfig().Theme = %q, want %q", cfg.Theme, ThemeDefault)
	}
	if cfg.Favorites == nil {
		t.Error("DefaultConfig().Favorites should not be nil")
	}
	if len(cfg.Favorites) != 0 {
		t.Errorf("DefaultConfig().Favorites length = %d, want 0", len(cfg.Favorites))
	}
}

// --- SortMode Methods Tests ---

func TestSortMode_String(t *testing.T) {
	tests := []struct {
		mode     SortMode
		expected string
	}{
		{SortAlpha, "Alphabetical"},
		{SortModifiedFirst, "Modified first"},
		{SortCategory, "By category"},
	}

	for _, tt := range tests {
		got := tt.mode.String()
		if got != tt.expected {
			t.Errorf("%v.String() = %q, want %q", tt.mode, got, tt.expected)
		}
	}
}

func TestSortMode_String_Unknown(t *testing.T) {
	got := SortMode(999).String()
	if got != "Unknown" {
		t.Errorf("SortMode(999).String() = %q, want 'Unknown'", got)
	}
}

func TestSortMode_Next(t *testing.T) {
	tests := []struct {
		current  SortMode
		expected SortMode
	}{
		{SortAlpha, SortModifiedFirst},
		{SortModifiedFirst, SortCategory},
		{SortCategory, SortAlpha},
	}

	for _, tt := range tests {
		got := tt.current.Next()
		if got != tt.expected {
			t.Errorf("%v.Next() = %v, want %v", tt.current, got, tt.expected)
		}
	}
}

func TestSortMode_Next_FullCycle(t *testing.T) {
	mode := SortAlpha

	// Cycle through all modes and back
	mode = mode.Next() // SortModifiedFirst
	if mode != SortModifiedFirst {
		t.Errorf("first Next() = %v, want SortModifiedFirst", mode)
	}

	mode = mode.Next() // SortCategory
	if mode != SortCategory {
		t.Errorf("second Next() = %v, want SortCategory", mode)
	}

	mode = mode.Next() // SortAlpha
	if mode != SortAlpha {
		t.Errorf("third Next() = %v, want SortAlpha (full cycle)", mode)
	}
}

func TestSortMode_Next_Unknown(t *testing.T) {
	got := SortMode(999).Next()
	if got != SortAlpha {
		t.Errorf("SortMode(999).Next() = %v, want SortAlpha", got)
	}
}
