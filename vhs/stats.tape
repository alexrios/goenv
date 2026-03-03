# Stats demo: View environment statistics
Output docs/gifs/stats.gif

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

# Open statistics screen
Type "S"
Sleep 4s

# Return to list
Escape
Sleep 500ms

Hide
Type "q"
Sleep 500ms
