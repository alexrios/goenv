package persist

import "testing"

func TestBuiltinPresets_Returns5Presets(t *testing.T) {
	presets := BuiltinPresets()
	if len(presets) != 5 {
		t.Errorf("BuiltinPresets() returned %d presets, want 5", len(presets))
	}
}

func TestBuiltinPresets_AllMarkedBuiltIn(t *testing.T) {
	for _, p := range BuiltinPresets() {
		if !p.BuiltIn {
			t.Errorf("preset %q should be marked as BuiltIn", p.Name)
		}
	}
}

func TestBuiltinPresets_AllHaveVariables(t *testing.T) {
	for _, p := range BuiltinPresets() {
		if len(p.Variables) == 0 {
			t.Errorf("preset %q should have at least one variable", p.Name)
		}
	}
}

func TestBuiltinPresets_AllHaveNames(t *testing.T) {
	for _, p := range BuiltinPresets() {
		if p.Name == "" {
			t.Error("all built-in presets should have a name")
		}
	}
}
