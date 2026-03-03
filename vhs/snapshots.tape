# Snapshots demo: Export a snapshot
Output docs/gifs/snapshots.gif

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

# Open export screen
Type "e"
Sleep 1.5s

# Enter snapshot name
Type "my-dev-setup"
Sleep 1s
Enter
Sleep 1s

# Enter description
Type "Development environment snapshot"
Sleep 1s
Enter
Sleep 2s

Hide
Type "q"
Sleep 500ms
