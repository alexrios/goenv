package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/alexrios/goenv/internal/goenv"
)

// goPath holds the cached path to the go executable.
var goPath string

// InitGoPath looks up the go executable path and caches it for subsequent commands.
// Must be called before any other command functions.
func InitGoPath() error {
	path, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("go executable not found: %w", err)
	}
	goPath = path
	return nil
}

// runGoCommand executes a go command with the given arguments.
// Returns stdout on success, or an error with stderr content on failure.
func runGoCommand(args ...string) ([]byte, error) {
	cmd := exec.Command(goPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return nil, fmt.Errorf("%s: %w", stderrStr, err)
		}
		return nil, err
	}
	return stdout.Bytes(), nil
}

// validEnvKeyRegex matches valid Go environment variable names.
// Go env keys are uppercase letters, digits, and underscores, starting with a letter.
var validEnvKeyRegex = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

// ValidateEnvKey checks if a key is a valid Go environment variable name.
// Valid keys consist of uppercase letters, digits, and underscores, starting with a letter.
// Returns an error if the key is invalid, which prevents command injection attacks.
func ValidateEnvKey(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	if !validEnvKeyRegex.MatchString(key) {
		return fmt.Errorf("invalid key %q: must be uppercase letters, digits, and underscores, starting with a letter", key)
	}
	return nil
}

// ReloadEnv fetches all Go environment variables and identifies which ones have been modified.
// Uses 'go env -json' to get all variables and 'go env -changed -json' to detect modifications.
// Returns a slice of EnvVar.
func ReloadEnv() ([]goenv.EnvVar, error) {
	// Get all env vars
	output, err := runGoCommand("env", "-json")
	if err != nil {
		return nil, fmt.Errorf("failed to run 'go env -json': %w", err)
	}

	var envVars map[string]string
	err = json.Unmarshal(output, &envVars)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal 'go env -json' output: %w", err)
	}

	// Get changed env vars
	outputChanged, err := runGoCommand("env", "-changed", "-json")
	if err != nil {
		// -changed might not be supported or fail in old go versions or odd environments.
		// Treat as no changed vars if this specifically fails.
		outputChanged = []byte("{}")
	}

	var changedVars map[string]string
	err = json.Unmarshal(outputChanged, &changedVars)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal 'go env -changed -json' output: %w", err)
	}

	items := make([]goenv.EnvVar, 0)
	// Iterate through all vars, check if they were in the changed list
	for k, v := range envVars {
		if v != "" {
			_, changed := changedVars[k]
			items = append(items, goenv.EnvVar{
				Key:     k,
				Value:   v,
				Changed: changed,
			})
		}
	}

	return items, nil
}

// GetGoVersion returns the current Go version (e.g., "go1.21.0").
func GetGoVersion() (string, error) {
	output, err := runGoCommand("version")
	if err != nil {
		return "", err
	}

	// Parse "go version go1.21.0 linux/amd64" -> "go1.21.0"
	parts := strings.Fields(string(output))
	if len(parts) >= 3 {
		return parts[2], nil
	}
	return strings.TrimSpace(string(output)), nil
}

// SetEnvVar persists an environment variable using 'go env -w KEY=VALUE'.
// Validates the key to prevent injection attacks and properly escapes special characters.
func SetEnvVar(ev goenv.EnvVar) error {
	// Validate key to prevent injection attacks
	if err := ValidateEnvKey(ev.Key); err != nil {
		return fmt.Errorf("invalid environment variable: %w", err)
	}

	// Construct the key=value argument. exec.Command passes args directly
	// via syscall (no shell), so no quoting or escaping is needed.
	kvArg := ev.Key + "=" + ev.Value
	_, err := runGoCommand("env", "-w", kvArg)
	if err != nil {
		return fmt.Errorf("failed to run 'go env -w %s': %w", ev.Key, err)
	}

	return nil
}

// UnsetEnvVar resets an environment variable to its default value using 'go env -u KEY'.
// Validates the key to prevent injection attacks.
func UnsetEnvVar(key string) error {
	// Validate key to prevent injection attacks
	if err := ValidateEnvKey(key); err != nil {
		return fmt.Errorf("invalid environment variable: %w", err)
	}

	_, err := runGoCommand("env", "-u", key)
	if err != nil {
		return fmt.Errorf("failed to run 'go env -u %s': %w", key, err)
	}

	return nil
}
