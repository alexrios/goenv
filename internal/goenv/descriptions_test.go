package goenv

import "testing"

// --- FilterSuggestions Tests ---

func TestFilterSuggestions_KnownKey_EmptyPrefix(t *testing.T) {
	// GOOS has known values
	suggestions := FilterSuggestions("GOOS", "")

	if suggestions == nil {
		t.Fatal("expected suggestions for GOOS, got nil")
	}
	if len(suggestions) == 0 {
		t.Error("expected non-empty suggestions for GOOS")
	}

	// Should contain common OS values
	found := false
	for _, s := range suggestions {
		if s == "linux" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'linux' in GOOS suggestions")
	}
}

func TestFilterSuggestions_KnownKey_PartialPrefix(t *testing.T) {
	// Filter GOOS values starting with "l"
	suggestions := FilterSuggestions("GOOS", "l")

	if suggestions == nil {
		t.Fatal("expected suggestions, got nil")
	}

	for _, s := range suggestions {
		if s[0] != 'l' {
			t.Errorf("suggestion %q should start with 'l'", s)
		}
	}

	// Should include linux
	found := false
	for _, s := range suggestions {
		if s == "linux" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'linux' in filtered suggestions")
	}
}

func TestFilterSuggestions_KnownKey_FullMatch(t *testing.T) {
	suggestions := FilterSuggestions("GOOS", "linux")

	if suggestions == nil {
		t.Fatal("expected suggestions, got nil")
	}
	if len(suggestions) != 1 {
		t.Errorf("expected 1 suggestion for exact match, got %d", len(suggestions))
	}
	if suggestions[0] != "linux" {
		t.Errorf("suggestion = %q, want 'linux'", suggestions[0])
	}
}

func TestFilterSuggestions_KnownKey_NoMatches(t *testing.T) {
	suggestions := FilterSuggestions("GOOS", "xyz")

	// FilterSuggestions returns nil when no matches found (not empty slice)
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions, got %d", len(suggestions))
	}
}

func TestFilterSuggestions_UnknownKey(t *testing.T) {
	suggestions := FilterSuggestions("UNKNOWN_VAR", "")

	if suggestions != nil {
		t.Errorf("expected nil for unknown key, got %v", suggestions)
	}
}

func TestFilterSuggestions_CGO_ENABLED(t *testing.T) {
	suggestions := FilterSuggestions("CGO_ENABLED", "")

	if suggestions == nil {
		t.Fatal("expected suggestions for CGO_ENABLED, got nil")
	}
	if len(suggestions) != 2 {
		t.Errorf("expected 2 suggestions (0, 1), got %d", len(suggestions))
	}

	// Verify both values present
	has0, has1 := false, false
	for _, s := range suggestions {
		if s == "0" {
			has0 = true
		}
		if s == "1" {
			has1 = true
		}
	}
	if !has0 || !has1 {
		t.Error("expected both '0' and '1' in CGO_ENABLED suggestions")
	}
}

func TestFilterSuggestions_GO111MODULE(t *testing.T) {
	suggestions := FilterSuggestions("GO111MODULE", "")

	if suggestions == nil {
		t.Fatal("expected suggestions for GO111MODULE, got nil")
	}
	if len(suggestions) != 3 {
		t.Errorf("expected 3 suggestions (on, off, auto), got %d", len(suggestions))
	}
}

func TestFilterSuggestions_PrefixFilter_GOARCH(t *testing.T) {
	suggestions := FilterSuggestions("GOARCH", "arm")

	if suggestions == nil {
		t.Fatal("expected suggestions, got nil")
	}

	// Should match arm and arm64
	for _, s := range suggestions {
		if len(s) < 3 || s[:3] != "arm" {
			t.Errorf("suggestion %q should start with 'arm'", s)
		}
	}
}

// --- FilterKnownValuesForVersion Tests ---

func TestFilterKnownValuesForVersion_ZeroVersionReturnsAll(t *testing.T) {
	// Zero version should return all values (no filtering)
	values := FilterKnownValuesForVersion("GOOS", GoVersion{})
	allValues := GetEnvVarKnownValues("GOOS")

	if len(values) != len(allValues) {
		t.Errorf("zero version: got %d values, want %d", len(values), len(allValues))
	}
}

func TestFilterKnownValuesForVersion_OldGoExcludesWasip1(t *testing.T) {
	// Go 1.20 should not include wasip1 in GOOS
	values := FilterKnownValuesForVersion("GOOS", GoVersion{1, 20, 0})

	for _, v := range values {
		if v == "wasip1" {
			t.Error("Go 1.20 should not include wasip1 in GOOS values")
		}
	}
	// But should still include linux
	found := false
	for _, v := range values {
		if v == "linux" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Go 1.20 should still include linux in GOOS values")
	}
}

func TestFilterKnownValuesForVersion_NewGoIncludesWasip1(t *testing.T) {
	values := FilterKnownValuesForVersion("GOOS", GoVersion{1, 21, 0})

	found := false
	for _, v := range values {
		if v == "wasip1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Go 1.21 should include wasip1 in GOOS values")
	}
}

func TestFilterKnownValuesForVersion_OldGoExcludesLoong64(t *testing.T) {
	values := FilterKnownValuesForVersion("GOARCH", GoVersion{1, 18, 0})

	for _, v := range values {
		if v == "loong64" {
			t.Error("Go 1.18 should not include loong64 in GOARCH values")
		}
	}
}

func TestFilterKnownValuesForVersion_WholeVarGating_Telemetry(t *testing.T) {
	// Go 1.20 should return nil for GOTELEMETRY (entire var gated)
	values := FilterKnownValuesForVersion("GOTELEMETRY", GoVersion{1, 20, 0})
	if values != nil {
		t.Errorf("Go 1.20: GOTELEMETRY should return nil, got %v", values)
	}

	// Go 1.21 should return the full list
	values = FilterKnownValuesForVersion("GOTELEMETRY", GoVersion{1, 21, 0})
	if values == nil {
		t.Error("Go 1.21: GOTELEMETRY should return values")
	}
}

func TestFilterKnownValuesForVersion_WholeVarGating_Toolchain(t *testing.T) {
	values := FilterKnownValuesForVersion("GOTOOLCHAIN", GoVersion{1, 20, 0})
	if values != nil {
		t.Errorf("Go 1.20: GOTOOLCHAIN should return nil, got %v", values)
	}

	values = FilterKnownValuesForVersion("GOTOOLCHAIN", GoVersion{1, 21, 0})
	if values == nil {
		t.Error("Go 1.21: GOTOOLCHAIN should return values")
	}
}

func TestFilterKnownValuesForVersion_UnknownKey(t *testing.T) {
	values := FilterKnownValuesForVersion("UNKNOWN_VAR", GoVersion{1, 21, 0})
	if values != nil {
		t.Errorf("unknown key should return nil, got %v", values)
	}
}

func TestFilterKnownValuesForVersion_NoGatedValues(t *testing.T) {
	// CGO_ENABLED has no version-gated values, should always return full list
	values := FilterKnownValuesForVersion("CGO_ENABLED", GoVersion{1, 10, 0})
	if len(values) != 2 {
		t.Errorf("CGO_ENABLED should return 2 values regardless of version, got %d", len(values))
	}
}

// --- FilterSuggestionsForVersion Tests ---

func TestFilterSuggestionsForVersion_ExcludesVersionGated(t *testing.T) {
	// Go 1.20 should not suggest wasip1
	suggestions := FilterSuggestionsForVersion("GOOS", "wasi", GoVersion{1, 20, 0})
	for _, s := range suggestions {
		if s == "wasip1" {
			t.Error("Go 1.20 should not suggest wasip1")
		}
	}

	// Go 1.21 should suggest wasip1
	suggestions = FilterSuggestionsForVersion("GOOS", "wasi", GoVersion{1, 21, 0})
	found := false
	for _, s := range suggestions {
		if s == "wasip1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Go 1.21 should suggest wasip1")
	}
}

func TestFilterSuggestionsForVersion_WholeVarGated(t *testing.T) {
	// GOTELEMETRY on Go 1.20 should return nil
	suggestions := FilterSuggestionsForVersion("GOTELEMETRY", "", GoVersion{1, 20, 0})
	if suggestions != nil {
		t.Errorf("Go 1.20: GOTELEMETRY suggestions should be nil, got %v", suggestions)
	}
}

// --- GetEnvVarKnownValues Tests ---

func TestGetEnvVarKnownValues_KnownKey(t *testing.T) {
	values := GetEnvVarKnownValues("GOOS")

	if values == nil {
		t.Error("expected values for GOOS, got nil")
	}
	if len(values) == 0 {
		t.Error("expected non-empty values for GOOS")
	}
}

func TestGetEnvVarKnownValues_UnknownKey(t *testing.T) {
	values := GetEnvVarKnownValues("NONEXISTENT_VAR")

	if values != nil {
		t.Errorf("expected nil for unknown key, got %v", values)
	}
}

// --- GetEnvVarCategory Tests ---

func TestGetEnvVarCategory_KnownKeys(t *testing.T) {
	tests := []struct {
		key      string
		expected Category
	}{
		{"GOPATH", CategoryGeneral},
		{"GOROOT", CategoryGeneral},
		{"CGO_ENABLED", CategoryCGO},
		{"CC", CategoryCGO},
		{"GOOS", CategoryArch},
		{"GOARCH", CategoryArch},
		{"GOPROXY", CategoryProxy},
		{"GO111MODULE", CategoryModule},
		{"GOTOOLCHAIN", CategoryBuild},
		{"GODEBUG", CategoryDebug},
		{"GOTOOLDIR", CategoryTooling},
	}

	for _, tt := range tests {
		got := GetEnvVarCategory(tt.key)
		if got != tt.expected {
			t.Errorf("GetEnvVarCategory(%q) = %q, want %q", tt.key, got, tt.expected)
		}
	}
}

func TestGetEnvVarCategory_UnknownKey_DefaultsToGeneral(t *testing.T) {
	got := GetEnvVarCategory("UNKNOWN_VAR")

	if got != CategoryGeneral {
		t.Errorf("GetEnvVarCategory(unknown) = %q, want %q", got, CategoryGeneral)
	}
}

// --- CategoryIndex Tests ---

func TestCategoryIndex_KnownCategories(t *testing.T) {
	// Verify General comes before Debug in order
	generalIdx := CategoryIndex(CategoryGeneral)
	debugIdx := CategoryIndex(CategoryDebug)

	if generalIdx >= debugIdx {
		t.Errorf("CategoryGeneral index (%d) should be less than CategoryDebug index (%d)",
			generalIdx, debugIdx)
	}
}

func TestCategoryIndex_UnknownCategory_AtEnd(t *testing.T) {
	unknownIdx := CategoryIndex(Category("Unknown"))
	generalIdx := CategoryIndex(CategoryGeneral)

	if unknownIdx <= generalIdx {
		t.Error("unknown category should have higher index than known categories")
	}
}

// --- GetEnvVarDescription Tests ---

func TestGetEnvVarDescription_KnownKeys(t *testing.T) {
	tests := []string{
		"GOPATH", "GOROOT", "GOPROXY", "CGO_ENABLED", "GOOS", "GOARCH",
	}

	for _, key := range tests {
		desc := GetEnvVarDescription(key)
		if desc == "" {
			t.Errorf("GetEnvVarDescription(%q) = empty, want description", key)
		}
	}
}

func TestGetEnvVarDescription_UnknownKey(t *testing.T) {
	desc := GetEnvVarDescription("NONEXISTENT_VAR")

	if desc != "" {
		t.Errorf("GetEnvVarDescription(unknown) = %q, want empty", desc)
	}
}

// --- GetPathSuggestions Tests ---

func TestGetPathSuggestions_KnownPathVar(t *testing.T) {
	suggestions := GetPathSuggestions("GOPATH")
	if suggestions == nil {
		t.Fatal("expected path suggestions for GOPATH, got nil")
	}
	if len(suggestions) == 0 {
		t.Error("expected at least one path suggestion for GOPATH")
	}
	// Should not contain ~ (should be expanded)
	for _, s := range suggestions {
		if s[0] == '~' {
			t.Errorf("path suggestion %q should have ~ expanded", s)
		}
	}
}

func TestGetPathSuggestions_UnknownVar(t *testing.T) {
	suggestions := GetPathSuggestions("GOOS")
	if suggestions != nil {
		t.Errorf("expected nil for non-path variable, got %v", suggestions)
	}
}

// --- IsCSVVariable Tests ---

func TestIsCSVVariable_Known(t *testing.T) {
	csvVars := []string{"GOEXPERIMENT", "GODEBUG", "GOWASM"}
	for _, v := range csvVars {
		if !IsCSVVariable(v) {
			t.Errorf("IsCSVVariable(%q) = false, want true", v)
		}
	}
}

func TestIsCSVVariable_Unknown(t *testing.T) {
	if IsCSVVariable("GOPATH") {
		t.Error("IsCSVVariable(GOPATH) = true, want false")
	}
}

// --- GetExtendedDoc Tests ---

func TestGetExtendedDoc_KnownVars(t *testing.T) {
	vars := []string{"GOOS", "GOARCH", "GOPROXY", "GODEBUG", "CGO_ENABLED"}
	for _, v := range vars {
		doc := GetExtendedDoc(v)
		if doc == "" {
			t.Errorf("GetExtendedDoc(%q) = empty, want documentation", v)
		}
	}
}

func TestGetExtendedDoc_UnknownVar(t *testing.T) {
	doc := GetExtendedDoc("NONEXISTENT_VAR")
	if doc != "" {
		t.Errorf("GetExtendedDoc(unknown) = %q, want empty", doc)
	}
}

// --- NextCategory Tests ---

func TestNextCategory_CyclesAll(t *testing.T) {
	cat := Category("")
	seen := make(map[Category]bool)
	// Cycle through all categories and back to ""
	for range len(CategoryOrder) + 1 {
		cat = NextCategory(cat)
		if seen[cat] {
			t.Fatalf("duplicate category in cycle: %q", cat)
		}
		seen[cat] = true
	}
	// After full cycle, should be back to ""
	if cat != "" {
		t.Errorf("expected empty string after full cycle, got %q", cat)
	}
}

func TestGetEnvVarCategory_DefaultsToGeneral(t *testing.T) {
	cat := GetEnvVarCategory("SOME_UNKNOWN_VAR")
	if cat != CategoryGeneral {
		t.Errorf("unknown var category = %q, want General (default)", cat)
	}
}
