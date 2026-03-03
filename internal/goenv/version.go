package goenv

import (
	"fmt"
	"strconv"
	"strings"
)

// GoVersion represents a parsed Go version with major, minor, and patch components.
type GoVersion struct {
	Major int
	Minor int
	Patch int
}

// Minimum Go version requirements for specific features.
var (
	MinVersionEnvJSON     = GoVersion{1, 9, 0}  // go env -json
	MinVersionEnvWrite    = GoVersion{1, 13, 0}  // go env -w / go env -u
	MinVersionChangedFlag = GoVersion{1, 21, 0}  // go env -changed
	MinVersionTelemetry   = GoVersion{1, 21, 0}  // GOTELEMETRY variables
	MinVersionToolchain   = GoVersion{1, 21, 0}  // GOTOOLCHAIN
	MinVersionWasip1      = GoVersion{1, 21, 0}  // wasip1 GOOS target
	MinVersionLoong64     = GoVersion{1, 19, 0}  // loong64 GOARCH target
	MinVersionRangefunc   = GoVersion{1, 22, 0}  // rangefunc experiment
	MinVersionCacheprog   = GoVersion{1, 24, 0}  // cacheprog experiment
)

// ParseGoVersion parses a Go version string like "go1.21.3" into a GoVersion.
// Handles formats: "go1.21.3", "go1.21", "1.21.3", "go1.21rc1", "go1.21beta1".
// Returns a zero GoVersion if parsing fails.
func ParseGoVersion(s string) GoVersion {
	s = strings.TrimPrefix(s, "go")
	if s == "" {
		return GoVersion{}
	}

	// Strip prerelease suffixes (rc1, beta1, etc.)
	// Find the first character that isn't a digit or dot
	cleanEnd := len(s)
	for i, c := range s {
		if c != '.' && (c < '0' || c > '9') {
			cleanEnd = i
			break
		}
	}
	s = s[:cleanEnd]

	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return GoVersion{}
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return GoVersion{}
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return GoVersion{}
	}

	var patch int
	if len(parts) >= 3 {
		patch, _ = strconv.Atoi(parts[2])
	}

	return GoVersion{Major: major, Minor: minor, Patch: patch}
}

// AtLeast returns true if v is greater than or equal to the given version.
func (v GoVersion) AtLeast(other GoVersion) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch >= other.Patch
}

// IsZero returns true if the version was not successfully parsed.
func (v GoVersion) IsZero() bool {
	return v.Major == 0 && v.Minor == 0 && v.Patch == 0
}

// String returns the version in "go1.21.3" format.
func (v GoVersion) String() string {
	if v.IsZero() {
		return "unknown"
	}
	return fmt.Sprintf("go%d.%d.%d", v.Major, v.Minor, v.Patch)
}
