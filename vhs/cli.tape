# CLI Mode demo: Non-interactive usage for scripting
Output docs/gifs/cli.gif

Set Shell "bash"
Set FontSize 16
Set Width 1000
Set Height 600
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms

Hide
Type "go build -o goenv ./cmd/goenv"
Enter
Sleep 2s
# Clear screen so build output is not visible
Ctrl+L
Sleep 300ms
Show

Sleep 500ms

# Show version (short flag)
Type "./goenv -v"
Sleep 500ms
Enter
Sleep 2s

# List variables (short flag, pipe-friendly)
Type "./goenv -l | head -10"
Sleep 500ms
Enter
Sleep 2.5s

# Get a single variable
Type "./goenv --get GOPATH"
Sleep 500ms
Enter
Sleep 2s

# Set a variable then verify
Type "./goenv --set GOFLAGS=-v"
Sleep 500ms
Enter
Sleep 2s

Type "./goenv --get GOFLAGS"
Sleep 500ms
Enter
Sleep 2s

# Export modified variables as JSON
Type "./goenv --export json --modified"
Sleep 500ms
Enter
Sleep 2.5s

# Clean up: reset GOFLAGS
Hide
Type "./goenv --set GOFLAGS="
Enter
Sleep 1s
