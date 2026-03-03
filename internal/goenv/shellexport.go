package goenv

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ShellType represents the shell format for export.
type ShellType int

const (
	ShellBash       ShellType = iota // Bash/Zsh (export KEY="value")
	ShellFish                        // Fish (set -x KEY "value")
	ShellPowerShell                  // PowerShell ($env:KEY = "VALUE")
	ShellJSON                        // JSON ({"KEY": "VALUE", ...})
)

// String returns the display name for the shell type.
func (s ShellType) String() string {
	switch s {
	case ShellBash:
		return "Bash/Zsh"
	case ShellFish:
		return "Fish"
	case ShellPowerShell:
		return "PowerShell"
	case ShellJSON:
		return "JSON"
	default:
		return "Unknown"
	}
}

// Next returns the next shell type in the cycle.
func (s ShellType) Next() ShellType {
	switch s {
	case ShellBash:
		return ShellFish
	case ShellFish:
		return ShellPowerShell
	case ShellPowerShell:
		return ShellJSON
	case ShellJSON:
		return ShellBash
	default:
		return ShellBash
	}
}

// ExportFilter represents which variables to export.
type ExportFilter int

const (
	ExportAll      ExportFilter = iota // Export all variables
	ExportModified                     // Export only modified variables
)

// String returns the display name for the export filter.
func (f ExportFilter) String() string {
	switch f {
	case ExportAll:
		return "All variables"
	case ExportModified:
		return "Modified only"
	default:
		return "Unknown"
	}
}

// Next returns the next filter in the cycle.
func (f ExportFilter) Next() ExportFilter {
	switch f {
	case ExportAll:
		return ExportModified
	case ExportModified:
		return ExportAll
	default:
		return ExportAll
	}
}

// GenerateShellExport generates shell export commands for the given variables.
func GenerateShellExport(vars []EnvVar, shell ShellType, filter ExportFilter) string {
	// JSON format is special-cased
	if shell == ShellJSON {
		return generateJSONExport(vars, filter)
	}

	var lines []string
	var comment string

	switch shell {
	case ShellBash:
		comment = "# Go environment variables"
	case ShellFish:
		comment = "# Go environment variables (Fish shell)"
	case ShellPowerShell:
		comment = "# Go environment variables (PowerShell)"
	}
	lines = append(lines, comment)
	lines = append(lines, "")

	for _, v := range vars {
		// Skip if filter is modified only and variable is not changed
		if filter == ExportModified && !v.Changed {
			continue
		}

		// Skip empty values
		if v.Value == "" {
			continue
		}

		lines = append(lines, formatExportLine(v.Key, v.Value, shell))
	}

	return strings.Join(lines, "\n")
}

// generateJSONExport produces a JSON object of variables.
func generateJSONExport(vars []EnvVar, filter ExportFilter) string {
	m := make(map[string]string)
	for _, v := range vars {
		if filter == ExportModified && !v.Changed {
			continue
		}
		if v.Value == "" {
			continue
		}
		m[v.Key] = v.Value
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}

// formatExportLine formats a single export line for the given shell.
func formatExportLine(key, value string, shell ShellType) string {
	// Escape special characters in the value
	escapedValue := escapeShellValue(value, shell)

	switch shell {
	case ShellBash:
		return fmt.Sprintf("export %s=%s", key, escapedValue)
	case ShellFish:
		return fmt.Sprintf("set -x %s %s", key, escapedValue)
	case ShellPowerShell:
		return fmt.Sprintf("$env:%s = %s", key, escapedValue)
	default:
		return fmt.Sprintf("export %s=%s", key, escapedValue)
	}
}

// escapeShellValue escapes special characters in a value for the given shell.
func escapeShellValue(value string, shell ShellType) string {
	// Check if value needs quoting
	needsQuotes := strings.ContainsAny(value, " \t\n'\"\\$`!()[]{}|&;<>?*#~")

	if !needsQuotes {
		return value
	}

	// Use double quotes and escape special characters
	switch shell {
	case ShellBash:
		// In bash, escape $, `, \, ", and !
		escaped := value
		escaped = strings.ReplaceAll(escaped, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		escaped = strings.ReplaceAll(escaped, "$", "\\$")
		escaped = strings.ReplaceAll(escaped, "`", "\\`")
		escaped = strings.ReplaceAll(escaped, "!", "\\!")
		return "\"" + escaped + "\""
	case ShellFish:
		// In fish, escape $ and "
		escaped := value
		escaped = strings.ReplaceAll(escaped, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		escaped = strings.ReplaceAll(escaped, "$", "\\$")
		return "\"" + escaped + "\""
	case ShellPowerShell:
		// In PowerShell, escape ", `, and $
		escaped := value
		escaped = strings.ReplaceAll(escaped, "\"", "`\"")
		escaped = strings.ReplaceAll(escaped, "`", "``")
		escaped = strings.ReplaceAll(escaped, "$", "`$")
		return "\"" + escaped + "\""
	default:
		return "\"" + value + "\""
	}
}

// CountExportableVars counts how many variables would be exported with the given filter.
func CountExportableVars(vars []EnvVar, filter ExportFilter) int {
	count := 0
	for _, v := range vars {
		if filter == ExportModified && !v.Changed {
			continue
		}
		if v.Value != "" {
			count++
		}
	}
	return count
}
