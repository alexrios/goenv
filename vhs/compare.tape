# Compare demo: Diff current environment with a snapshot
Output docs/gifs/compare.gif

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

# Open compare screen
Type "c"
Sleep 2s

# Show source selection, navigate
Type "j"
Sleep 500ms
Type "k"
Sleep 500ms

# Select source and view diff
Enter
Sleep 2s

# Scroll through the diff
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "k"
Sleep 400ms
Type "k"
Sleep 800ms

# Go back
Escape
Sleep 500ms
Escape
Sleep 500ms

Hide
Type "q"
Sleep 500ms
