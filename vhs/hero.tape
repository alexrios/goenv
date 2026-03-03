# Hero demo: Launch goenv, browse variables, edit one, save
Output docs/gifs/hero.gif

Set Shell "bash"
Set FontSize 16
Set Width 1000
Set Height 600
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms

Hide
Type "go build -o goenv ./cmd/goenv && ./goenv"
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

# Save
Enter
Sleep 2s

# Undo so we don't leave a dirty env
Ctrl+Z
Sleep 1s

# Hide before quit to avoid shell artifact
Hide
Type "q"
Sleep 500ms
