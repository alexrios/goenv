# Command Palette demo: Fuzzy command search with :
Output docs/gifs/command-palette.gif

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
Sleep 1.5s

# Open command palette
Type ":"
Sleep 1s

# Type character-by-character to show live filtering
Type "e"
Sleep 400ms
Type "x"
Sleep 400ms
Type "p"
Sleep 1s

# Navigate through filtered results
Down
Sleep 500ms
Down
Sleep 500ms

# Clear and try a different filter
Ctrl+A
Ctrl+K
Sleep 400ms
Type "th"
Sleep 500ms
Type "e"
Sleep 500ms
Type "m"
Sleep 1s

# Execute the theme command to show it actually works
Enter
Sleep 1.5s

# Open palette again and cancel with Esc
Type ":"
Sleep 1s
Type "wa"
Sleep 800ms

Escape
Sleep 1s

Hide
Type "q"
Sleep 500ms
