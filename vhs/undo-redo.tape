# Undo/Redo demo: Edit a variable, undo, then redo, then undo again to restore
Output docs/gifs/undo-redo.gif

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

# Filter to GOFLAGS
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
Type "-v"
Sleep 500ms
Enter
Sleep 2s

# Undo the change
Ctrl+Z
Sleep 2s

# Redo the change
Ctrl+Y
Sleep 2s

# Undo again to leave env clean
Hide
Ctrl+Z
Sleep 500ms
Type "q"
Sleep 500ms
