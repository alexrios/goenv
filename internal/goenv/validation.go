package goenv

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// ValidationError represents a validation failure for an environment variable value.
type ValidationError struct {
	Key      string   // The variable name
	Value    string   // The invalid value
	Message  string   // Human-readable error message
	Expected []string // List of valid values, if applicable
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return e.Message
}

// ValidateEnvValue validates a value for a given environment variable key.
// Returns nil if the value is valid, or a ValidationError if invalid.
// Uses the zero GoVersion (no version filtering) for backward compatibility.
func ValidateEnvValue(key, value string) *ValidationError {
	return ValidateEnvValueForVersion(key, value, GoVersion{})
}

// ValidateEnvValueForVersion validates a value, using version-filtered known values.
// If ver is zero, all known values are considered valid (no version filtering).
func ValidateEnvValueForVersion(key, value string, ver GoVersion) *ValidationError {
	// Empty values are allowed (to clear/unset)
	if value == "" {
		return nil
	}

	switch key {
	case "GOOS":
		return validateFromList(key, value, FilterKnownValuesForVersion("GOOS", ver))
	case "GOARCH":
		return validateFromList(key, value, FilterKnownValuesForVersion("GOARCH", ver))
	case "CGO_ENABLED":
		return validateFromList(key, value, EnvVarKnownValues["CGO_ENABLED"])
	case "GO111MODULE":
		return validateFromList(key, value, EnvVarKnownValues["GO111MODULE"])
	case "GOTELEMETRY":
		values := FilterKnownValuesForVersion("GOTELEMETRY", ver)
		if values == nil {
			// Variable doesn't exist in this Go version; skip validation
			return nil
		}
		return validateFromList(key, value, values)
	case "GOAMD64":
		return validateFromList(key, value, EnvVarKnownValues["GOAMD64"])
	case "GOARM":
		return validateFromList(key, value, EnvVarKnownValues["GOARM"])
	case "GO386":
		return validateFromList(key, value, EnvVarKnownValues["GO386"])
	case "GOMIPS", "GOMIPS64":
		return validateFromList(key, value, EnvVarKnownValues["GOMIPS"])
	case "GOPPC64":
		return validateFromList(key, value, EnvVarKnownValues["GOPPC64"])
	case "GOTOOLCHAIN":
		// GOTOOLCHAIN accepts go1.X.Y format in addition to known shortcuts; skip strict validation
		return nil
	case "GOEXPERIMENT":
		// GOEXPERIMENT accepts comma-separated values; skip strict validation
		return nil

	// Path variables
	case "GOROOT", "GOPATH", "GOBIN", "GOCACHE", "GOMODCACHE", "GOTMPDIR", "GOCOVERDIR", "GOTELEMETRYDIR":
		return validatePath(key, value)

	// URL/proxy variables
	case "GOPROXY":
		return validateProxy(key, value)
	}

	// No validation rules for this variable
	return nil
}

// validateFromList checks if value is in the allowed list.
func validateFromList(key, value string, allowed []string) *ValidationError {
	if allowed == nil {
		return nil
	}
	if slices.Contains(allowed, value) {
		return nil
	}
	return &ValidationError{
		Key:      key,
		Value:    value,
		Message:  "invalid value for " + key,
		Expected: allowed,
	}
}

// validatePath checks if a value looks like a valid path.
// It doesn't require the path to exist, just that it's plausible.
func validatePath(key, value string) *ValidationError {
	// Check for obvious issues
	if strings.ContainsAny(value, "\x00") {
		return &ValidationError{
			Key:     key,
			Value:   value,
			Message: "path contains invalid characters",
		}
	}

	// Expand ~ to home directory for validation
	expandedPath := value
	if strings.HasPrefix(value, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			expandedPath = filepath.Join(home, value[2:])
		}
	}

	// Check if it looks like an absolute path or starts with common patterns
	if !filepath.IsAbs(expandedPath) && !strings.HasPrefix(value, "~/") && !strings.HasPrefix(value, "$") {
		return &ValidationError{
			Key:     key,
			Value:   value,
			Message: "path should be absolute (e.g., /home/user/go or ~/go)",
		}
	}

	return nil
}

// validateProxy validates GOPROXY format.
// Valid formats: URL, "direct", "off", or comma-separated list.
func validateProxy(key, value string) *ValidationError {
	// Split by comma for multiple proxies
	for part := range strings.SplitSeq(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// Valid special values
		if part == "direct" || part == "off" {
			continue
		}
		// Should look like a URL
		if !strings.HasPrefix(part, "http://") && !strings.HasPrefix(part, "https://") && !strings.HasPrefix(part, "file://") {
			return &ValidationError{
				Key:      key,
				Value:    value,
				Message:  "proxy should be a URL, 'direct', or 'off'",
				Expected: []string{"https://proxy.golang.org", "direct", "off"},
			}
		}
	}
	return nil
}

// FormatValidationError formats a ValidationError for display.
// Returns a user-friendly message with suggestions if available.
func FormatValidationError(err *ValidationError) string {
	if err == nil {
		return ""
	}

	msg := err.Message
	if len(err.Expected) > 0 {
		// Show up to 5 valid values
		showCount := len(err.Expected)
		if showCount > 5 {
			showCount = 5
		}
		msg += ". Valid: " + strings.Join(err.Expected[:showCount], ", ")
		if len(err.Expected) > 5 {
			msg += ", ..."
		}
	}
	return msg
}
