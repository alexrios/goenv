package goenv

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Category represents a group of related environment variables.
type Category string

// Environment variable categories.
const (
	CategoryGeneral Category = "General"
	CategoryCGO     Category = "CGO"
	CategoryArch    Category = "Architecture"
	CategoryProxy   Category = "Proxy/Network"
	CategoryBuild   Category = "Build"
	CategoryDebug   Category = "Debug/Telemetry"
	CategoryModule  Category = "Modules"
	CategoryTooling Category = "Tooling"
)

// CategoryOrder defines the display order of categories.
var CategoryOrder = []Category{
	CategoryGeneral,
	CategoryModule,
	CategoryProxy,
	CategoryBuild,
	CategoryCGO,
	CategoryArch,
	CategoryTooling,
	CategoryDebug,
}

// NextCategory returns the next category in the cycle.
// Empty string (All) -> first category -> ... -> last category -> empty string (All).
func NextCategory(current Category) Category {
	if current == "" {
		return CategoryOrder[0]
	}
	for i, c := range CategoryOrder {
		if c == current {
			if i+1 < len(CategoryOrder) {
				return CategoryOrder[i+1]
			}
			return "" // wrap back to "All"
		}
	}
	return "" // unknown category, reset to All
}

// readOnlyVariables lists environment variables that are computed by Go and
// cannot be set with 'go env -w'.
var readOnlyVariables = map[string]struct{}{
	"GOVERSION":  {},
	"GOTOOLDIR":  {},
	"GOMOD":      {},
	"GOHOSTOS":   {},
	"GOHOSTARCH": {},
	"GOEXE":      {},
	"GOGCCFLAGS": {},
	"GOENV":      {}, // path to config file; cannot be changed via go env -w
}

// IsReadOnly returns true if the given environment variable is computed by Go
// and cannot be set with 'go env -w'.
func IsReadOnly(key string) bool {
	_, ok := readOnlyVariables[key]
	return ok
}

// envVarCategories maps environment variable names to their categories.
var envVarCategories = map[string]Category{
	// General
	"GOROOT":    CategoryGeneral,
	"GOPATH":    CategoryGeneral,
	"GOBIN":     CategoryGeneral,
	"GOCACHE":   CategoryGeneral,
	"GOENV":     CategoryGeneral,
	"GOTMPDIR":  CategoryGeneral,
	"GOFLAGS":   CategoryGeneral,
	"GOVERSION": CategoryGeneral,

	// Modules
	"GO111MODULE": CategoryModule,
	"GOMODCACHE":  CategoryModule,
	"GOWORK":      CategoryModule,
	"GOPRIVATE":   CategoryModule,
	"GONOPROXY":   CategoryModule,
	"GONOSUMDB":   CategoryModule,
	"GOINSECURE":  CategoryModule,

	// Proxy/Network
	"GOPROXY": CategoryProxy,
	"GOSUMDB": CategoryProxy,
	"GOVCS":   CategoryProxy,

	// Build
	"GOEXE":        CategoryBuild,
	"GOEXPERIMENT": CategoryBuild,
	"GOGCCFLAGS":   CategoryBuild,
	"GOTOOLCHAIN":  CategoryBuild,

	// CGO
	"CGO_ENABLED":  CategoryCGO,
	"CGO_CFLAGS":   CategoryCGO,
	"CGO_CPPFLAGS": CategoryCGO,
	"CGO_CXXFLAGS": CategoryCGO,
	"CGO_FFLAGS":   CategoryCGO,
	"CGO_LDFLAGS":  CategoryCGO,
	"CC":           CategoryCGO,
	"CXX":          CategoryCGO,
	"FC":           CategoryCGO,
	"AR":           CategoryCGO,
	"PKG_CONFIG":   CategoryCGO,
	"GCCGO":        CategoryCGO,

	// Architecture
	"GOOS":       CategoryArch,
	"GOARCH":     CategoryArch,
	"GOHOSTOS":   CategoryArch,
	"GOHOSTARCH": CategoryArch,
	"GOAMD64":    CategoryArch,
	"GOARM":      CategoryArch,
	"GOARM64":    CategoryArch,
	"GO386":      CategoryArch,
	"GOMIPS":     CategoryArch,
	"GOMIPS64":   CategoryArch,
	"GOPPC64":    CategoryArch,
	"GORISCV64":  CategoryArch,
	"GOWASM":     CategoryArch,

	// Tooling
	"GOTOOLDIR":  CategoryTooling,
	"GOCOVERDIR": CategoryTooling,

	// Debug/Telemetry
	"GODEBUG":        CategoryDebug,
	"GOTELEMETRY":    CategoryDebug,
	"GOTELEMETRYDIR": CategoryDebug,
}

// GetEnvVarCategory returns the category for a given environment variable.
// Returns CategoryGeneral if no specific category is defined.
func GetEnvVarCategory(key string) Category {
	if cat, ok := envVarCategories[key]; ok {
		return cat
	}
	return CategoryGeneral
}

// CategoryIndex returns the sort order index for a category.
func CategoryIndex(cat Category) int {
	if i := slices.Index(CategoryOrder, cat); i >= 0 {
		return i
	}
	return len(CategoryOrder) // Unknown categories at end
}

// envVarDescriptions provides brief descriptions for common Go environment variables.
// Source: go help environment
var envVarDescriptions = map[string]string{
	// General behavior
	"GO111MODULE":    "Controls module mode (on, off, or auto)",
	"GCCGO":          "gccgo compiler command to run",
	"GOARCH":         "Target architecture (e.g., amd64, arm64)",
	"GOBIN":          "Directory for installed binaries",
	"GOCACHE":        "Directory for build cache",
	"GOENV":          "Location of Go environment config file",
	"GOEXE":          "Executable file suffix (.exe on Windows)",
	"GOEXPERIMENT":   "Comma-separated list of enabled experiments",
	"GOFLAGS":        "Default flags for go commands",
	"GOGCCFLAGS":     "Flags for the C compiler (gcc)",
	"GOHOSTARCH":     "Host machine architecture",
	"GOHOSTOS":       "Host machine operating system",
	"GOINSECURE":     "Glob patterns for insecure module fetch",
	"GOMODCACHE":     "Directory for downloaded modules",
	"GONOPROXY":      "Glob patterns to fetch directly, not via proxy",
	"GONOSUMDB":      "Glob patterns to skip checksum verification",
	"GOOS":           "Target operating system",
	"GOPATH":         "Workspace directory for GOPATH mode",
	"GOPRIVATE":      "Glob patterns for private module paths",
	"GOPROXY":        "Module proxy URL(s)",
	"GOROOT":         "Root directory of Go installation",
	"GOSUMDB":        "Checksum database for module verification",
	"GOTMPDIR":       "Directory for temporary files",
	"GOTOOLCHAIN":    "Go toolchain to use (e.g., go1.21.0, auto)",
	"GOTOOLDIR":      "Directory containing go tool executables",
	"GOVCS":          "Version control systems allowed for modules",
	"GOVERSION":      "Go version that built this binary",
	"GOWORK":         "Path to go.work file or 'off'",

	// CGO
	"AR":           "Archive program for buildmode=c-archive",
	"CC":           "C compiler command",
	"CGO_ENABLED":  "Whether cgo is enabled (0 or 1)",
	"CGO_CFLAGS":   "Flags passed to C compiler",
	"CGO_CPPFLAGS": "Flags passed to C preprocessor",
	"CGO_CXXFLAGS": "Flags passed to C++ compiler",
	"CGO_FFLAGS":   "Flags passed to Fortran compiler",
	"CGO_LDFLAGS":  "Flags passed to C linker",
	"CXX":          "C++ compiler command",
	"FC":           "Fortran compiler command",
	"PKG_CONFIG":   "Path to pkg-config tool",

	// Architecture-specific
	"GOAMD64":   "AMD64 microarchitecture level (v1-v4)",
	"GOARM":     "ARM architecture version (5, 6, 7)",
	"GOARM64":   "ARM64 architecture features",
	"GO386":     "x86 floating point instruction set",
	"GOMIPS":    "MIPS floating point instructions",
	"GOMIPS64":  "MIPS64 floating point instructions",
	"GOPPC64":   "PowerPC64 minimum instruction set",
	"GORISCV64": "RISC-V 64-bit instruction set extensions",
	"GOWASM":    "WebAssembly features to use",

	// Coverage and telemetry
	"GOCOVERDIR":    "Directory for coverage data files",
	"GOTELEMETRY":   "Telemetry mode (off, local, on)",
	"GOTELEMETRYDIR": "Directory for telemetry data",

	// Debugging
	"GODEBUG": "Comma-separated debugging variables",
}

// GetEnvVarDescription returns the description for a given environment variable key.
// Returns empty string if no description is available.
func GetEnvVarDescription(key string) string {
	return envVarDescriptions[key]
}

// EnvVarKnownValues provides known valid values for environment variables.
// Used for autocomplete suggestions in the edit screen.
var EnvVarKnownValues = map[string][]string{
	"GOOS": {
		"aix", "android", "darwin", "dragonfly", "freebsd", "illumos",
		"ios", "js", "linux", "netbsd", "openbsd", "plan9", "solaris",
		"wasip1", "windows",
	},
	"GOARCH": {
		"386", "amd64", "arm", "arm64", "loong64", "mips", "mips64",
		"mips64le", "mipsle", "ppc64", "ppc64le", "riscv64", "s390x", "wasm",
	},
	"CGO_ENABLED":  {"0", "1"},
	"GO111MODULE":  {"on", "off", "auto"},
	"GOTELEMETRY":  {"off", "local", "on"},
	"GOAMD64":      {"v1", "v2", "v3", "v4"},
	"GOARM":        {"5", "6", "7"},
	"GOARM64":      {"v8.0", "v8.1", "v8.2", "v8.3", "v8.4", "v8.5", "v8.6", "v8.7", "v8.8", "v8.9", "v9.0"},
	"GO386":        {"sse2", "softfloat"},
	"GOMIPS":       {"hardfloat", "softfloat"},
	"GOMIPS64":     {"hardfloat", "softfloat"},
	"GOPPC64":      {"power8", "power9", "power10"},
	"GORISCV64":    {"rva20u64", "rva22u64"},
	"GOWASM":       {"satconv", "signext"},
	"GOTOOLCHAIN":  {"auto", "local", "path"},
	"GOEXPERIMENT": {"arenas", "boringcrypto", "cacheprog", "newinliner", "rangefunc"},
}

// GetEnvVarKnownValues returns the list of known values for a given environment variable.
// Returns nil if no known values are defined.
func GetEnvVarKnownValues(key string) []string {
	return EnvVarKnownValues[key]
}

// versionedValue represents a known value that requires a minimum Go version.
type versionedValue struct {
	Value      string
	MinVersion GoVersion
}

// versionGatedValues maps env var keys to values that require specific Go versions.
// Values NOT listed here are assumed to be available in all versions.
var versionGatedValues = map[string][]versionedValue{
	"GOOS": {
		{Value: "wasip1", MinVersion: MinVersionWasip1},
	},
	"GOARCH": {
		{Value: "loong64", MinVersion: MinVersionLoong64},
	},
	"GOEXPERIMENT": {
		{Value: "rangefunc", MinVersion: MinVersionRangefunc},
		{Value: "cacheprog", MinVersion: MinVersionCacheprog},
	},
}

// wholeVarMinVersions maps env var keys that require a minimum Go version entirely.
var wholeVarMinVersions = map[string]GoVersion{
	"GOTOOLCHAIN":    MinVersionToolchain,
	"GOTELEMETRY":    MinVersionTelemetry,
	"GOTELEMETRYDIR": MinVersionTelemetry,
}

// FilterKnownValuesForVersion returns known values for key, excluding those
// that require a newer Go version than ver. If ver is zero, returns all values.
func FilterKnownValuesForVersion(key string, ver GoVersion) []string {
	values := GetEnvVarKnownValues(key)
	if values == nil {
		return nil
	}

	// If version is unknown, return all values (no filtering)
	if ver.IsZero() {
		return values
	}

	// If the entire variable requires a newer version, return nil
	if minVer, ok := wholeVarMinVersions[key]; ok && !ver.AtLeast(minVer) {
		return nil
	}

	gated := versionGatedValues[key]
	if len(gated) == 0 {
		return values
	}

	var filtered []string
	for _, v := range values {
		include := true
		for _, gv := range gated {
			if gv.Value == v && !ver.AtLeast(gv.MinVersion) {
				include = false
				break
			}
		}
		if include {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// pathVarDefaults maps path-type variables to common default paths.
// ~ is expanded to $HOME at runtime.
var pathVarDefaults = map[string][]string{
	"GOROOT":    {"~/sdk/go", "/usr/local/go", "/usr/lib/go"},
	"GOPATH":    {"~/go"},
	"GOBIN":     {"~/go/bin"},
	"GOCACHE":   {"~/.cache/go-build"},
	"GOMODCACHE": {"~/go/pkg/mod"},
	"GOTMPDIR":  {"/tmp"},
}

// GetPathSuggestions returns expanded path suggestions for a variable.
// Returns nil if the variable is not a path variable.
func GetPathSuggestions(key string) []string {
	templates, ok := pathVarDefaults[key]
	if !ok {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}

	var paths []string
	for _, tmpl := range templates {
		if strings.HasPrefix(tmpl, "~/") && home != "" {
			paths = append(paths, filepath.Join(home, tmpl[2:]))
		} else {
			paths = append(paths, tmpl)
		}
	}
	return paths
}

// csvVariables is the set of variables that contain comma-separated values.
var csvVariables = map[string]bool{
	"GOEXPERIMENT": true,
	"GODEBUG":     true,
	"GOWASM":      true,
}

// IsCSVVariable returns true if the given variable stores comma-separated values.
func IsCSVVariable(key string) bool {
	return csvVariables[key]
}

// envVarExtendedDocs provides longer documentation with examples for key variables.
var envVarExtendedDocs = map[string]string{
	"GOOS":         "Examples: linux, darwin, windows, js, wasip1",
	"GOARCH":       "Examples: amd64, arm64, 386, wasm",
	"GOPROXY":      "Comma-separated list of proxy URLs. Use 'direct' to skip. Default: https://proxy.golang.org,direct",
	"GOPRIVATE":    "Comma-separated glob patterns (e.g. github.com/myorg/*). Also sets GONOPROXY and GONOSUMDB.",
	"GODEBUG":      "Comma-separated key=value pairs. Examples: gctrace=1, schedtrace=1000, http2debug=1",
	"GOEXPERIMENT": "Comma-separated experiment names. Prefix with 'no' to disable (e.g. norangefunc).",
	"GOFLAGS":      "Space-separated flags applied to every go command. Example: -mod=vendor -count=1",
	"GOTOOLCHAIN":  "Controls toolchain selection. 'auto' downloads as needed. 'local' uses only installed version.",
	"CGO_ENABLED":  "Set to 0 for static builds without C dependencies. Set to 1 to enable CGO linking.",
	"GOPATH":       "Legacy workspace root. Modules are stored in $GOPATH/pkg/mod by default.",
	"GOROOT":       "Root of the Go SDK installation. Usually auto-detected.",
	"GOBIN":        "Install directory for 'go install'. Defaults to $GOPATH/bin or $HOME/go/bin.",
	"GO111MODULE":  "Module mode control. 'on' (always), 'off' (GOPATH mode), 'auto' (detect go.mod).",
	"GONOSUMDB":    "Comma-separated glob patterns of modules to skip checksum verification for.",
}

// GetExtendedDoc returns extended documentation for a variable, or empty string.
func GetExtendedDoc(key string) string {
	return envVarExtendedDocs[key]
}

// FilterSuggestions returns suggestions that start with the given prefix.
// Returns all known values if prefix is empty.
func FilterSuggestions(key, prefix string) []string {
	return FilterSuggestionsForVersion(key, prefix, GoVersion{})
}

// FilterSuggestionsForVersion returns suggestions filtered by prefix and Go version.
// If ver is zero, no version filtering is applied.
func FilterSuggestionsForVersion(key, prefix string, ver GoVersion) []string {
	values := FilterKnownValuesForVersion(key, ver)
	if values == nil {
		return nil
	}
	if prefix == "" {
		return values
	}

	var filtered []string
	for _, v := range values {
		if strings.HasPrefix(v, prefix) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}
