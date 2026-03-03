package cli

import (
	"testing"

	"github.com/alexrios/goenv/internal/goenv"
)

func TestParseCLI_NoArgs(t *testing.T) {
	opts := ParseCLI(nil)
	if opts.Action != actionTUI {
		t.Errorf("no args should default to TUI, got %d", opts.Action)
	}
}

func TestParseCLI_Version(t *testing.T) {
	for _, flag := range []string{"--version", "-v"} {
		opts := ParseCLI([]string{flag})
		if opts.Action != actionVersion {
			t.Errorf("%s should set action to Version, got %d", flag, opts.Action)
		}
	}
}

func TestParseCLI_List(t *testing.T) {
	for _, flag := range []string{"--list", "-l"} {
		opts := ParseCLI([]string{flag})
		if opts.Action != actionList {
			t.Errorf("%s should set action to List, got %d", flag, opts.Action)
		}
	}
}

func TestParseCLI_Export_DefaultBash(t *testing.T) {
	opts := ParseCLI([]string{"--export"})
	if opts.Action != actionExport {
		t.Fatalf("--export should set action to Export, got %d", opts.Action)
	}
	if opts.ExportShell != "bash" {
		t.Errorf("default shell should be bash, got %q", opts.ExportShell)
	}
	if opts.ExportFilter != "all" {
		t.Errorf("default filter should be all, got %q", opts.ExportFilter)
	}
}

func TestParseCLI_Export_WithFormat(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"--export", "fish"}, "fish"},
		{[]string{"--export", "powershell"}, "powershell"},
		{[]string{"--export", "json"}, "json"},
		{[]string{"--export", "bash"}, "bash"},
	}
	for _, tt := range tests {
		opts := ParseCLI(tt.args)
		if opts.ExportShell != tt.want {
			t.Errorf("ParseCLI(%v) shell = %q, want %q", tt.args, opts.ExportShell, tt.want)
		}
	}
}

func TestParseCLI_Export_Modified(t *testing.T) {
	// --modified after format
	opts := ParseCLI([]string{"--export", "bash", "--modified"})
	if opts.ExportFilter != "modified" {
		t.Errorf("expected filter 'modified', got %q", opts.ExportFilter)
	}
}

func TestParseCLI_Export_Modified_BeforeFormat(t *testing.T) {
	// --modified before format
	opts := ParseCLI([]string{"--export", "--modified", "fish"})
	if opts.ExportFilter != "modified" {
		t.Errorf("expected filter 'modified', got %q", opts.ExportFilter)
	}
	if opts.ExportShell != "fish" {
		t.Errorf("expected shell 'fish', got %q", opts.ExportShell)
	}
}

func TestParseCLI_Set(t *testing.T) {
	opts := ParseCLI([]string{"--set", "GOOS=linux", "GOARCH=amd64"})
	if opts.Action != actionSet {
		t.Fatalf("--set should set action to Set, got %d", opts.Action)
	}
	if len(opts.SetPairs) != 2 {
		t.Fatalf("expected 2 set pairs, got %d", len(opts.SetPairs))
	}
	if opts.SetPairs[0] != "GOOS=linux" {
		t.Errorf("first pair = %q, want GOOS=linux", opts.SetPairs[0])
	}
	if opts.SetPairs[1] != "GOARCH=amd64" {
		t.Errorf("second pair = %q, want GOARCH=amd64", opts.SetPairs[1])
	}
}

func TestParseCLI_Get(t *testing.T) {
	opts := ParseCLI([]string{"--get", "GOPATH"})
	if opts.Action != actionGet {
		t.Fatalf("--get should set action to Get, got %d", opts.Action)
	}
	if opts.GetKey != "GOPATH" {
		t.Errorf("key = %q, want GOPATH", opts.GetKey)
	}
}

func TestParseShellFlag(t *testing.T) {
	tests := []struct {
		input string
		want  goenv.ShellType
	}{
		{"bash", goenv.ShellBash},
		{"fish", goenv.ShellFish},
		{"powershell", goenv.ShellPowerShell},
		{"pwsh", goenv.ShellPowerShell},
		{"json", goenv.ShellJSON},
		{"BASH", goenv.ShellBash},
		{"Fish", goenv.ShellFish},
		{"unknown", goenv.ShellBash}, // default
		{"", goenv.ShellBash},        // default
	}
	for _, tt := range tests {
		got := parseShellFlag(tt.input)
		if got != tt.want {
			t.Errorf("parseShellFlag(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestSortItemsForCLI(t *testing.T) {
	items := []goenv.EnvVar{
		{Key: "GOPATH", Value: "/go"},
		{Key: "GOARCH", Value: "amd64"},
		{Key: "GOOS", Value: "linux"},
	}
	sorted := sortItemsForCLI(items)
	if len(sorted) != 3 {
		t.Fatalf("expected 3 items, got %d", len(sorted))
	}
	keys := make([]string, len(sorted))
	for i, ev := range sorted {
		keys[i] = ev.Key
	}
	if keys[0] != "GOARCH" || keys[1] != "GOOS" || keys[2] != "GOPATH" {
		t.Errorf("sort order wrong: %v", keys)
	}
}
