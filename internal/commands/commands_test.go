package commands

import (
	"os"
	"testing"

	"github.com/alexrios/goenv/internal/goenv"
)

// TestReloadEnv_ReturnsItems tests that ReloadEnv() returns valid items.
// This is an integration test that requires 'go' to be installed.
func TestReloadEnv_ReturnsItems(t *testing.T) {
	if err := InitGoPath(); err != nil {
		t.Fatalf("InitGoPath() error = %v", err)
	}
	items, err := ReloadEnv()
	if err != nil {
		t.Fatalf("ReloadEnv() error = %v", err)
	}

	if len(items) == 0 {
		t.Error("ReloadEnv() returned 0 items, expected at least some Go environment variables")
	}

	// Verify items have expected fields
	foundGOROOT := false
	for _, ev := range items {
		// Every item should have a non-empty key
		if ev.Key == "" {
			t.Error("found EnvVar with empty key")
		}

		// Check for common env vars
		if ev.Key == "GOROOT" {
			foundGOROOT = true
		}
	}

	if !foundGOROOT {
		t.Error("expected to find GOROOT in environment variables")
	}
}

// TestReloadEnv_ContainsExpectedVars verifies specific expected variables exist.
func TestReloadEnv_ContainsExpectedVars(t *testing.T) {
	if err := InitGoPath(); err != nil {
		t.Fatalf("InitGoPath() error = %v", err)
	}
	items, err := ReloadEnv()
	if err != nil {
		t.Fatalf("ReloadEnv() error = %v", err)
	}

	// These variables should always exist in a Go environment
	expectedVars := []string{"GOROOT", "GOPATH", "GOOS", "GOARCH"}

	varMap := make(map[string]bool)
	for _, ev := range items {
		varMap[ev.Key] = true
	}

	for _, expected := range expectedVars {
		if !varMap[expected] {
			t.Errorf("expected variable %s not found", expected)
		}
	}
}

// =============================================================================
// REGRESSION TESTS - Key Validation (Security)
// =============================================================================

// TestRegression_ValidateEnvKey_ValidKeys verifies that valid Go env keys pass validation.
func TestRegression_ValidateEnvKey_ValidKeys(t *testing.T) {
	validKeys := []string{
		"GOPATH",
		"GOROOT",
		"GOOS",
		"GOARCH",
		"GO111MODULE",
		"GOFLAGS",
		"GOPRIVATE",
		"GOPROXY",
		"A",
		"A1",
		"A_B",
		"GO_TEST_VAR",
	}

	for _, key := range validKeys {
		if err := ValidateEnvKey(key); err != nil {
			t.Errorf("ValidateEnvKey(%q) should pass, got error: %v", key, err)
		}
	}
}

// TestRegression_ValidateEnvKey_InvalidKeys verifies that invalid keys are rejected.
// This prevents potential command injection attacks.
func TestRegression_ValidateEnvKey_InvalidKeys(t *testing.T) {
	invalidKeys := []string{
		"",            // empty
		"lowercase",   // must be uppercase
		"Mixed_Case",  // must be all uppercase
		"123START",    // can't start with number
		"_UNDERSCORE", // can't start with underscore
		"HAS SPACE",   // no spaces
		"HAS-DASH",    // no dashes
		"HAS.DOT",     // no dots
		"HAS;SEMI",    // no semicolons (injection risk)
		"HAS&AMP",     // no ampersands (injection risk)
		"HAS|PIPE",    // no pipes (injection risk)
		"HAS$DOLLAR",  // no dollar signs (injection risk)
		"HAS`TICK",    // no backticks (injection risk)
		"HAS$(CMD)",   // no command substitution
		"HAS\nNEWLINE", // no newlines
	}

	for _, key := range invalidKeys {
		if err := ValidateEnvKey(key); err == nil {
			t.Errorf("ValidateEnvKey(%q) should fail, but passed", key)
		}
	}
}

// TestRegression_SetEnvVar_RejectsInvalidKeys verifies that SetEnvVar rejects
// invalid keys before executing any commands.
func TestRegression_SetEnvVar_RejectsInvalidKeys(t *testing.T) {
	// These are potentially dangerous keys that could be used for injection
	dangerousKeys := []string{
		"foo;rm -rf /",  // command injection attempt
		"$(whoami)",     // command substitution
		"`id`",          // backtick command substitution
		"a\nb",          // newline injection
		"key=value&cmd", // parameter injection
	}

	for _, key := range dangerousKeys {
		e := goenv.EnvVar{Key: key, Value: "value"}
		err := SetEnvVar(e)
		if err == nil {
			t.Errorf("SetEnvVar with key %q should fail validation", key)
		}
	}
}

// TestRegression_InitGoPathCachesPath verifies that InitGoPath() sets the
// package-level goPath variable to the cached path of the go executable.
func TestRegression_InitGoPathCachesPath(t *testing.T) {
	// Reset goPath to ensure test isolation
	goPath = ""

	err := InitGoPath()
	if err != nil {
		t.Fatalf("InitGoPath() error = %v", err)
	}

	if goPath == "" {
		t.Error("goPath should be set after InitGoPath()")
	}

	// Verify it's a valid path by checking the file exists
	if _, err := os.Stat(goPath); err != nil {
		t.Errorf("goPath %q is not a valid file path: %v", goPath, err)
	}
}
