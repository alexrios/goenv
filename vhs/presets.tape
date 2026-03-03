# Presets demo: Create a new preset
Output docs/gifs/presets.gif

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

# Open presets screen
Type "p"
Sleep 2s

# Create a new preset
Type "n"
Sleep 2s

# Type preset name
Type "cgo-enabled"
Sleep 1s
Enter
Sleep 2s

# Return to list
Escape
Sleep 1s

Hide
Type "q"
Sleep 500ms
