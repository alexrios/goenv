package cli

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/alexrios/goenv/internal/commands"
	"github.com/alexrios/goenv/internal/goenv"
)

// AppVersion is the application version, set at build time via ldflags.
var AppVersion = "dev"

// action represents what the CLI should do after parsing flags.
type action int

const (
	actionTUI     action = iota // Launch TUI (default)
	actionVersion               // Print version
	actionList                   // List variables to stdout
	actionExport                 // Export shell commands to stdout
	actionSet                    // Set a variable headlessly
	actionGet                    // Get a single variable value
)

// Options holds parsed CLI flags.
type Options struct {
	Action       action
	ExportShell  string // "bash", "fish", "powershell", "json"
	ExportFilter string // "all", "modified"
	SetPairs     []string
	GetKey       string
}

// ParseCLI parses command-line arguments and returns the action to take.
// Returns actionTUI if no flags are provided (default behavior).
func ParseCLI(args []string) Options {
	opts := Options{Action: actionTUI}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--version" || arg == "-v":
			opts.Action = actionVersion
			return opts

		case arg == "--list" || arg == "-l":
			opts.Action = actionList
			return opts

		case arg == "--export":
			opts.Action = actionExport
			opts.ExportShell = "bash"
			opts.ExportFilter = "all"
			shellSet := false
			// Consume remaining args: shell format and --modified in any order
			for j := i + 1; j < len(args); j++ {
				if args[j] == "--modified" {
					opts.ExportFilter = "modified"
				} else if !strings.HasPrefix(args[j], "-") {
					if shellSet {
						fmt.Fprintf(os.Stderr, "Multiple shell formats specified. Run 'goenv --help' for usage.\n")
						os.Exit(1)
					}
					opts.ExportShell = strings.ToLower(args[j])
					shellSet = true
				}
			}
			return opts

		case arg == "--set":
			opts.Action = actionSet
			for j := i + 1; j < len(args); j++ {
				if strings.HasPrefix(args[j], "-") {
					break
				}
				opts.SetPairs = append(opts.SetPairs, args[j])
			}
			return opts

		case arg == "--get":
			opts.Action = actionGet
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				opts.GetKey = args[i]
			}
			return opts

		case arg == "--help" || arg == "-h":
			printUsage()
			os.Exit(0)

		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(os.Stderr, "Unknown flag: %s\nRun 'goenv --help' for usage.\n", arg)
			os.Exit(1)
		}
	}

	return opts
}

// RunCLI executes the CLI action. Returns true if the action was handled
// (program should exit), false if TUI should launch.
func RunCLI(opts Options) bool {
	switch opts.Action {
	case actionTUI:
		return false

	case actionVersion:
		goVer, _ := commands.GetGoVersion()
		if goVer == "" {
			goVer = "unknown"
		}
		fmt.Printf("goenv %s (Go %s)\n", AppVersion, goVer)
		return true

	case actionList:
		return runCLIList()

	case actionExport:
		return runCLIExport(opts)

	case actionSet:
		return runCLISet(opts)

	case actionGet:
		return runCLIGet(opts)
	}

	return false
}

func runCLIList() bool {
	items, err := commands.ReloadEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	sorted := sortItemsForCLI(items)
	for _, ev := range sorted {
		marker := " "
		if ev.Changed {
			marker = "*"
		}
		fmt.Printf("%s %s=%s\n", marker, ev.Key, ev.Value)
	}
	return true
}

func runCLIExport(opts Options) bool {
	shell := parseShellFlag(opts.ExportShell)
	filter := goenv.ExportAll
	if opts.ExportFilter == "modified" {
		filter = goenv.ExportModified
	}

	items, err := commands.ReloadEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	output := goenv.GenerateShellExport(items, shell, filter)
	fmt.Println(output)
	return true
}

func runCLISet(opts Options) bool {
	if len(opts.SetPairs) == 0 {
		fmt.Fprintln(os.Stderr, "Error: --set requires KEY=VALUE arguments")
		os.Exit(1)
	}

	for _, pair := range opts.SetPairs {
		key, value, ok := strings.Cut(pair, "=")
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: invalid format %q (expected KEY=VALUE)\n", pair)
			os.Exit(1)
		}
		if err := commands.ValidateEnvKey(key); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if goenv.IsReadOnly(key) {
			fmt.Fprintf(os.Stderr, "Error: %s is read-only and cannot be set\n", key)
			os.Exit(1)
		}
		if valErr := goenv.ValidateEnvValue(key, value); valErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", valErr.Message)
		}
		if err := commands.SetEnvVar(goenv.EnvVar{Key: key, Value: value, Changed: true}); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting %s: %v\n", key, err)
			os.Exit(1)
		}
		fmt.Printf("Set %s=%s\n", key, value)
	}
	return true
}

func runCLIGet(opts Options) bool {
	if opts.GetKey == "" {
		fmt.Fprintln(os.Stderr, "Error: --get requires a KEY argument")
		os.Exit(1)
	}

	items, err := commands.ReloadEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for _, ev := range items {
		if ev.Key == opts.GetKey {
			fmt.Println(ev.Value)
			return true
		}
	}

	fmt.Fprintf(os.Stderr, "Error: variable %q not found\n", opts.GetKey)
	os.Exit(1)
	return true
}

func parseShellFlag(s string) goenv.ShellType {
	switch strings.ToLower(s) {
	case "fish":
		return goenv.ShellFish
	case "powershell", "pwsh":
		return goenv.ShellPowerShell
	case "json":
		return goenv.ShellJSON
	default:
		return goenv.ShellBash
	}
}

func printUsage() {
	fmt.Print(`goenv - Terminal UI for Go environment variables

Usage:
  goenv                       Launch interactive TUI
  goenv --version             Show version information
  goenv --list                List all variables (pipe-friendly)
  goenv --get KEY             Get a single variable value
  goenv --set KEY=VAL ...     Set one or more variables
  goenv --export [FORMAT]     Export shell commands to stdout
                              Formats: bash (default), fish, powershell, json
  goenv --export bash --modified   Export only modified variables

Options:
  -v, --version    Show version
  -l, --list       List variables
  -h, --help       Show this help

Items marked with * in --list output are user-modified (via go env -w).
`)
}

// sortItemsForCLI sorts a slice of EnvVar by key name for consistent CLI output.
func sortItemsForCLI(items []goenv.EnvVar) []goenv.EnvVar {
	sorted := slices.Clone(items)
	slices.SortFunc(sorted, func(a, b goenv.EnvVar) int {
		return cmp.Compare(a.Key, b.Key)
	})
	return sorted
}
