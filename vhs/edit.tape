# Edit demo: Validation and suggestion cycling on GOOS
Output docs/gifs/edit.gif

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

# Filter to GOOS
Type "/"
Sleep 1s
Type "GOOS"
Sleep 1.5s
Enter
Sleep 500ms

# Select and edit GOOS
Enter
Sleep 1.5s

# Clear existing value
Ctrl+A
Ctrl+K
Sleep 300ms

# Type an invalid value to trigger validation
Type "invalid-os"
Sleep 500ms
Enter
Sleep 1.5s

# Clear and cycle through suggestions with Tab
Ctrl+A
Ctrl+K
Sleep 500ms
Tab
Sleep 600ms
Tab
Sleep 600ms
Tab
Sleep 600ms
Tab
Sleep 800ms

# Cancel editing (don't save - avoids env corruption)
Escape
Sleep 1s

Hide
Type "q"
Sleep 500ms
