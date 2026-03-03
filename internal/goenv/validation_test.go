package goenv

import (
	"testing"
)

func TestValidateEnvValue_GOOS(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{"linux", false},
		{"darwin", false},
		{"windows", false},
		{"invalid", true},
		{"LINUX", true}, // case sensitive
		{"", false},     // empty is allowed
	}

	for _, tt := range tests {
		err := ValidateEnvValue("GOOS", tt.value)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateEnvValue(GOOS, %q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
		}
	}
}

func TestValidateEnvValue_GOARCH(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{"amd64", false},
		{"arm64", false},
		{"386", false},
		{"invalid", true},
		{"", false},
	}

	for _, tt := range tests {
		err := ValidateEnvValue("GOARCH", tt.value)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateEnvValue(GOARCH, %q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
		}
	}
}

func TestValidateEnvValue_CGO_ENABLED(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{"0", false},
		{"1", false},
		{"2", true},
		{"true", true},
		{"false", true},
		{"", false},
	}

	for _, tt := range tests {
		err := ValidateEnvValue("CGO_ENABLED", tt.value)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateEnvValue(CGO_ENABLED, %q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
		}
	}
}

func TestValidateEnvValue_GO111MODULE(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{"on", false},
		{"off", false},
		{"auto", false},
		{"true", true},
		{"1", true},
		{"", false},
	}

	for _, tt := range tests {
		err := ValidateEnvValue("GO111MODULE", tt.value)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateEnvValue(GO111MODULE, %q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
		}
	}
}

func TestValidateEnvValue_Path(t *testing.T) {
	tests := []struct {
		key     string
		value   string
		wantErr bool
	}{
		{"GOPATH", "/home/user/go", false},
		{"GOPATH", "~/go", false},
		{"GOPATH", "$HOME/go", false},
		{"GOPATH", "relative/path", true},
		{"GOROOT", "/usr/local/go", false},
		{"GOBIN", "/home/user/go/bin", false},
		{"GOCACHE", "/tmp/gocache", false},
		{"GOPATH", "", false}, // empty is allowed
	}

	for _, tt := range tests {
		err := ValidateEnvValue(tt.key, tt.value)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateEnvValue(%s, %q) error = %v, wantErr %v", tt.key, tt.value, err, tt.wantErr)
		}
	}
}

func TestValidateEnvValue_GOPROXY(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{"https://proxy.golang.org", false},
		{"https://proxy.golang.org,direct", false},
		{"direct", false},
		{"off", false},
		{"https://proxy.golang.org,https://goproxy.cn,direct", false},
		{"invalid", true},
		{"", false},
	}

	for _, tt := range tests {
		err := ValidateEnvValue("GOPROXY", tt.value)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateEnvValue(GOPROXY, %q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
		}
	}
}

func TestValidateEnvValue_UnknownVariable(t *testing.T) {
	// Unknown variables should pass validation (no rules)
	err := ValidateEnvValue("UNKNOWN_VAR", "any value")
	if err != nil {
		t.Errorf("ValidateEnvValue(UNKNOWN_VAR, ...) should not return error, got %v", err)
	}
}

// --- ValidateEnvValueForVersion Tests ---

func TestValidateEnvValueForVersion_Wasip1_OldGo(t *testing.T) {
	// wasip1 should be rejected on Go 1.20 (not in the filtered list)
	err := ValidateEnvValueForVersion("GOOS", "wasip1", GoVersion{1, 20, 0})
	if err == nil {
		t.Error("wasip1 should be rejected on Go 1.20")
	}
}

func TestValidateEnvValueForVersion_Wasip1_NewGo(t *testing.T) {
	// wasip1 should be accepted on Go 1.21+
	err := ValidateEnvValueForVersion("GOOS", "wasip1", GoVersion{1, 21, 0})
	if err != nil {
		t.Errorf("wasip1 should be accepted on Go 1.21, got %v", err)
	}
}

func TestValidateEnvValueForVersion_Telemetry_OldGo(t *testing.T) {
	// GOTELEMETRY values should be accepted on old Go (variable doesn't exist, skip validation)
	err := ValidateEnvValueForVersion("GOTELEMETRY", "on", GoVersion{1, 20, 0})
	if err != nil {
		t.Errorf("GOTELEMETRY should skip validation on old Go, got %v", err)
	}
}

func TestValidateEnvValueForVersion_ZeroVersion_BackwardCompat(t *testing.T) {
	// Zero version should behave exactly like ValidateEnvValue
	err := ValidateEnvValueForVersion("GOOS", "linux", GoVersion{})
	if err != nil {
		t.Errorf("zero version should accept linux, got %v", err)
	}
	err = ValidateEnvValueForVersion("GOOS", "invalid", GoVersion{})
	if err == nil {
		t.Error("zero version should reject invalid GOOS value")
	}
}

func TestFormatValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name: "simple error",
			err: &ValidationError{
				Key:     "GOOS",
				Value:   "invalid",
				Message: "invalid value for GOOS",
			},
			expected: "invalid value for GOOS",
		},
		{
			name: "error with expected values",
			err: &ValidationError{
				Key:      "CGO_ENABLED",
				Value:    "2",
				Message:  "invalid value for CGO_ENABLED",
				Expected: []string{"0", "1"},
			},
			expected: "invalid value for CGO_ENABLED. Valid: 0, 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatValidationError(tt.err)
			if got != tt.expected {
				t.Errorf("FormatValidationError() = %q, want %q", got, tt.expected)
			}
		})
	}
}
