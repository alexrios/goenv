package goenv

import (
	"strings"
	"testing"
)

func TestShellType_String(t *testing.T) {
	tests := []struct {
		shell    ShellType
		expected string
	}{
		{ShellBash, "Bash/Zsh"},
		{ShellFish, "Fish"},
		{ShellPowerShell, "PowerShell"},
		{ShellJSON, "JSON"},
		{ShellType(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.shell.String()
		if got != tt.expected {
			t.Errorf("ShellType(%d).String() = %q, want %q", tt.shell, got, tt.expected)
		}
	}
}

func TestShellType_Next(t *testing.T) {
	tests := []struct {
		shell    ShellType
		expected ShellType
	}{
		{ShellBash, ShellFish},
		{ShellFish, ShellPowerShell},
		{ShellPowerShell, ShellJSON},
		{ShellJSON, ShellBash},
		{ShellType(99), ShellBash},
	}

	for _, tt := range tests {
		got := tt.shell.Next()
		if got != tt.expected {
			t.Errorf("ShellType(%d).Next() = %v, want %v", tt.shell, got, tt.expected)
		}
	}
}

func TestExportFilter_String(t *testing.T) {
	tests := []struct {
		filter   ExportFilter
		expected string
	}{
		{ExportAll, "All variables"},
		{ExportModified, "Modified only"},
		{ExportFilter(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.filter.String()
		if got != tt.expected {
			t.Errorf("ExportFilter(%d).String() = %q, want %q", tt.filter, got, tt.expected)
		}
	}
}

func TestExportFilter_Next(t *testing.T) {
	tests := []struct {
		filter   ExportFilter
		expected ExportFilter
	}{
		{ExportAll, ExportModified},
		{ExportModified, ExportAll},
		{ExportFilter(99), ExportAll},
	}

	for _, tt := range tests {
		got := tt.filter.Next()
		if got != tt.expected {
			t.Errorf("ExportFilter(%d).Next() = %v, want %v", tt.filter, got, tt.expected)
		}
	}
}

func TestGenerateShellExport_Bash(t *testing.T) {
	vars := []EnvVar{
		{Key: "GOPATH", Value: "/home/user/go", Changed: true},
		{Key: "GOPROXY", Value: "https://proxy.golang.org,direct", Changed: false},
	}

	output := GenerateShellExport(vars, ShellBash, ExportAll)

	if !strings.Contains(output, "export GOPATH=/home/user/go") {
		t.Errorf("Expected GOPATH export in output, got:\n%s", output)
	}
	if !strings.Contains(output, "export GOPROXY=") {
		t.Errorf("Expected GOPROXY export in output, got:\n%s", output)
	}
}

func TestGenerateShellExport_Fish(t *testing.T) {
	vars := []EnvVar{
		{Key: "GOPATH", Value: "/home/user/go", Changed: true},
	}

	output := GenerateShellExport(vars, ShellFish, ExportAll)

	if !strings.Contains(output, "set -x GOPATH /home/user/go") {
		t.Errorf("Expected Fish set command in output, got:\n%s", output)
	}
}

func TestGenerateShellExport_PowerShell(t *testing.T) {
	vars := []EnvVar{
		{Key: "GOPATH", Value: "/home/user/go", Changed: true},
	}

	output := GenerateShellExport(vars, ShellPowerShell, ExportAll)

	if !strings.Contains(output, "$env:GOPATH = /home/user/go") {
		t.Errorf("Expected PowerShell export in output, got:\n%s", output)
	}
	if !strings.Contains(output, "# Go environment variables (PowerShell)") {
		t.Errorf("Expected PowerShell comment header, got:\n%s", output)
	}
}

func TestGenerateShellExport_JSON(t *testing.T) {
	vars := []EnvVar{
		{Key: "GOPATH", Value: "/home/user/go", Changed: true},
		{Key: "GOBIN", Value: "", Changed: true}, // Empty, should be skipped
	}

	output := GenerateShellExport(vars, ShellJSON, ExportAll)

	if !strings.Contains(output, `"GOPATH"`) {
		t.Errorf("Expected GOPATH key in JSON output, got:\n%s", output)
	}
	if !strings.Contains(output, `"/home/user/go"`) {
		t.Errorf("Expected GOPATH value in JSON output, got:\n%s", output)
	}
	if strings.Contains(output, `"GOBIN"`) {
		t.Errorf("Did not expect GOBIN (empty) in JSON output, got:\n%s", output)
	}
}

func TestGenerateShellExport_ModifiedOnly(t *testing.T) {
	vars := []EnvVar{
		{Key: "GOPATH", Value: "/home/user/go", Changed: true},
		{Key: "GOPROXY", Value: "https://proxy.golang.org", Changed: false},
	}

	output := GenerateShellExport(vars, ShellBash, ExportModified)

	if !strings.Contains(output, "export GOPATH=") {
		t.Errorf("Expected GOPATH (modified) in output, got:\n%s", output)
	}
	if strings.Contains(output, "GOPROXY") {
		t.Errorf("Did not expect GOPROXY (not modified) in output, got:\n%s", output)
	}
}

func TestGenerateShellExport_SkipsEmptyValues(t *testing.T) {
	vars := []EnvVar{
		{Key: "GOPATH", Value: "/home/user/go", Changed: true},
		{Key: "GOBIN", Value: "", Changed: true},
	}

	output := GenerateShellExport(vars, ShellBash, ExportAll)

	if !strings.Contains(output, "GOPATH") {
		t.Errorf("Expected GOPATH in output, got:\n%s", output)
	}
	if strings.Contains(output, "GOBIN") {
		t.Errorf("Did not expect GOBIN (empty) in output, got:\n%s", output)
	}
}

func TestEscapeShellValue_SimpleValue(t *testing.T) {
	// Simple value without special chars
	got := escapeShellValue("/home/user/go", ShellBash)
	if got != "/home/user/go" {
		t.Errorf("Simple value should not be quoted, got: %s", got)
	}
}

func TestEscapeShellValue_ValueWithSpaces(t *testing.T) {
	got := escapeShellValue("/home/my user/go", ShellBash)
	if got != `"/home/my user/go"` {
		t.Errorf("Value with spaces should be quoted, got: %s", got)
	}
}

func TestEscapeShellValue_ValueWithDollarSign(t *testing.T) {
	got := escapeShellValue("$HOME/go", ShellBash)
	if got != `"\$HOME/go"` {
		t.Errorf("Dollar sign should be escaped, got: %s", got)
	}
}

func TestCountExportableVars(t *testing.T) {
	vars := []EnvVar{
		{Key: "GOPATH", Value: "/home/user/go", Changed: true},
		{Key: "GOPROXY", Value: "https://proxy.golang.org", Changed: false},
		{Key: "GOBIN", Value: "", Changed: true}, // Empty, should not count
	}

	// All filter
	countAll := CountExportableVars(vars, ExportAll)
	if countAll != 2 {
		t.Errorf("ExportAll: expected 2, got %d", countAll)
	}

	// Modified only filter
	countModified := CountExportableVars(vars, ExportModified)
	if countModified != 1 {
		t.Errorf("ExportModified: expected 1, got %d", countModified)
	}
}
