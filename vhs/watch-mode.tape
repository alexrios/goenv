# Watch Mode demo: Auto-reload with w
Output docs/gifs/watch-mode.gif

Set Shell "bash"
Set FontSize 16
Set Width 1000
Set Height 600
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms

Hide
# Schedule a background change that fires while watch mode is active
Type "go build -o goenv ./cmd/goenv && (sleep 10 && go env -w GOFLAGS=-v) & ./goenv"
Enter
Sleep 2s
Show

# Let TUI fully render
Sleep 2s

# Enable watch mode
Type "w"
Sleep 2s

# Browse while watching
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms

# Wait for the background go env -w to fire and watch to catch it
Sleep 8s

# Disable watch mode
Type "w"
Sleep 2s

# Undo the GOFLAGS change
Hide
Type "q"
Sleep 500ms
Type "go env -w GOFLAGS="
Enter
Sleep 500ms
