package goenv

import "testing"

func TestParseGoVersion_Standard(t *testing.T) {
	tests := []struct {
		input string
		want  GoVersion
	}{
		{"go1.21.3", GoVersion{1, 21, 3}},
		{"go1.13.0", GoVersion{1, 13, 0}},
		{"go1.9.7", GoVersion{1, 9, 7}},
		{"go1.22.0", GoVersion{1, 22, 0}},
		{"go1.26.0", GoVersion{1, 26, 0}},
	}

	for _, tt := range tests {
		got := ParseGoVersion(tt.input)
		if got != tt.want {
			t.Errorf("ParseGoVersion(%q) = %+v, want %+v", tt.input, got, tt.want)
		}
	}
}

func TestParseGoVersion_NoPrefix(t *testing.T) {
	got := ParseGoVersion("1.21.3")
	want := GoVersion{1, 21, 3}
	if got != want {
		t.Errorf("ParseGoVersion('1.21.3') = %+v, want %+v", got, want)
	}
}

func TestParseGoVersion_NoPatch(t *testing.T) {
	got := ParseGoVersion("go1.21")
	want := GoVersion{1, 21, 0}
	if got != want {
		t.Errorf("ParseGoVersion('go1.21') = %+v, want %+v", got, want)
	}
}

func TestParseGoVersion_Prerelease(t *testing.T) {
	tests := []struct {
		input string
		want  GoVersion
	}{
		{"go1.21rc1", GoVersion{1, 21, 0}},
		{"go1.22beta1", GoVersion{1, 22, 0}},
		{"go1.21.0rc2", GoVersion{1, 21, 0}},
	}

	for _, tt := range tests {
		got := ParseGoVersion(tt.input)
		if got != tt.want {
			t.Errorf("ParseGoVersion(%q) = %+v, want %+v", tt.input, got, tt.want)
		}
	}
}

func TestParseGoVersion_Invalid(t *testing.T) {
	tests := []string{
		"",
		"go",
		"not-a-version",
		"go1",
	}

	for _, input := range tests {
		got := ParseGoVersion(input)
		if !got.IsZero() {
			t.Errorf("ParseGoVersion(%q) = %+v, want zero", input, got)
		}
	}
}

func TestGoVersion_AtLeast(t *testing.T) {
	tests := []struct {
		v     GoVersion
		other GoVersion
		want  bool
	}{
		{GoVersion{1, 21, 3}, GoVersion{1, 21, 0}, true},
		{GoVersion{1, 21, 0}, GoVersion{1, 21, 0}, true},
		{GoVersion{1, 22, 0}, GoVersion{1, 21, 0}, true},
		{GoVersion{1, 20, 0}, GoVersion{1, 21, 0}, false},
		{GoVersion{1, 21, 2}, GoVersion{1, 21, 3}, false},
		{GoVersion{2, 0, 0}, GoVersion{1, 99, 0}, true},
		{GoVersion{1, 13, 0}, GoVersion{1, 13, 0}, true},
		{GoVersion{1, 12, 9}, GoVersion{1, 13, 0}, false},
	}

	for _, tt := range tests {
		got := tt.v.AtLeast(tt.other)
		if got != tt.want {
			t.Errorf("%+v.AtLeast(%+v) = %v, want %v", tt.v, tt.other, got, tt.want)
		}
	}
}

func TestGoVersion_IsZero(t *testing.T) {
	if !(GoVersion{}).IsZero() {
		t.Error("zero GoVersion should be zero")
	}
	if (GoVersion{1, 21, 0}).IsZero() {
		t.Error("non-zero GoVersion should not be zero")
	}
}

func TestGoVersion_String(t *testing.T) {
	tests := []struct {
		v    GoVersion
		want string
	}{
		{GoVersion{1, 21, 3}, "go1.21.3"},
		{GoVersion{1, 13, 0}, "go1.13.0"},
		{GoVersion{}, "unknown"},
	}

	for _, tt := range tests {
		got := tt.v.String()
		if got != tt.want {
			t.Errorf("%+v.String() = %q, want %q", tt.v, got, tt.want)
		}
	}
}

func TestGoVersion_AtLeast_CapabilityConstants(t *testing.T) {
	v121 := GoVersion{1, 21, 0}
	v113 := GoVersion{1, 13, 0}
	v112 := GoVersion{1, 12, 0}
	v108 := GoVersion{1, 8, 0}

	// Go 1.21 has all features
	if !v121.AtLeast(MinVersionEnvWrite) {
		t.Error("Go 1.21 should support env write")
	}
	if !v121.AtLeast(MinVersionChangedFlag) {
		t.Error("Go 1.21 should support changed flag")
	}
	if !v121.AtLeast(MinVersionTelemetry) {
		t.Error("Go 1.21 should support telemetry")
	}

	// Go 1.13 has env write but not changed flag
	if !v113.AtLeast(MinVersionEnvWrite) {
		t.Error("Go 1.13 should support env write")
	}
	if v113.AtLeast(MinVersionChangedFlag) {
		t.Error("Go 1.13 should not support changed flag")
	}

	// Go 1.12 lacks env write
	if v112.AtLeast(MinVersionEnvWrite) {
		t.Error("Go 1.12 should not support env write")
	}

	// Go 1.8 lacks even env json
	if v108.AtLeast(MinVersionEnvJSON) {
		t.Error("Go 1.8 should not support env json")
	}
}
