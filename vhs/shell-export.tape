# Shell Export demo: Generate shell export commands
Output docs/gifs/shell-export.gif

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

# Brief list view
Sleep 1s

# Open shell export screen
Type "x"
Sleep 2s

# Toggle shell format (bash -> fish)
Type "s"
Sleep 1.5s

# Toggle shell format (fish -> powershell)
Type "s"
Sleep 1.5s

# Toggle filter
Type "f"
Sleep 1s

# Preview the export
Enter
Sleep 2s

# Scroll through preview
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 600ms

# Go back
Type "b"
Sleep 500ms
Escape
Sleep 500ms

Hide
Type "q"
Sleep 500ms
