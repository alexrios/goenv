# Light theme demo: Show goenv with the light color theme
Output docs/gifs/light-theme.gif

Set Shell "bash"
Set FontSize 16
Set Width 1000
Set Height 600
Set Padding 20
Set Theme "Catppuccin Latte"
Set TypingSpeed 60ms

# Setup: build binary, backup config, set light theme
Hide
Type "mkdir -p ~/.config/goenv && go build -o goenv ./cmd/goenv"
Enter
Sleep 1s
Type "cp ~/.config/goenv/config.json ~/.config/goenv/config.json.bak 2>/dev/null; true"
Enter
Sleep 500ms
Type `echo '{"theme":"light"}' > ~/.config/goenv/config.json`
Enter
Sleep 500ms
Type "./goenv"
Enter
Sleep 2s
Show

# Let TUI fully render
Sleep 2s

# Browse the list
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "j"
Sleep 500ms
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "j"
Sleep 1s

# Search for GOFLAGS (safe to edit, empty by default)
Type "/"
Sleep 1s
Type "GOFLAGS"
Sleep 1.5s
Enter
Sleep 500ms

# Select and edit
Enter
Sleep 1.5s

# Type a value
Type "-v -count=1"
Sleep 1s

# Cancel editing (don't save)
Escape
Sleep 1.5s

# Cleanup: restore original config
Hide
Type "q"
Sleep 500ms
Type "cp ~/.config/goenv/config.json.bak ~/.config/goenv/config.json 2>/dev/null; true"
Enter
Sleep 500ms
