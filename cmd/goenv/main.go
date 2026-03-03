package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/alexrios/goenv/internal/cli"
	"github.com/alexrios/goenv/internal/commands"
	"github.com/alexrios/goenv/internal/persist"
	"github.com/alexrios/goenv/internal/tui"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "\ngoenv crashed unexpectedly: %v\n", r)
			os.Exit(1)
		}
	}()

	// Cache go executable path at startup
	if err := commands.InitGoPath(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Parse CLI flags
	opts := cli.ParseCLI(os.Args[1:])
	if cli.RunCLI(opts) {
		return
	}

	// TUI mode
	var dump *os.File
	if _, ok := os.LookupEnv("DEBUG"); ok {
		var err error
		dump, err = os.OpenFile("messages.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open debug log file: %v\n", err)
			os.Exit(1)
		}
		defer dump.Close()
	}

	cfg, err := persist.LoadConfig()
	if err != nil {
		cfg = persist.DefaultConfig()
	}

	items, err := commands.ReloadEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load environment variables: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(tui.NewMainModel(items, dump, cfg))

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
