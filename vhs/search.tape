# Search demo: Progressive filtering with /
Output docs/gifs/search.gif

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

# First search: proxy-related variables
Type "/"
Sleep 1s
Type "proxy"
Sleep 2s

# Clear search
Escape
Sleep 1s

# Second search: CGO variables
Type "/"
Sleep 1s
Type "cgo"
Sleep 2s

# Clear
Escape
Sleep 1s

# Third search: specific variable
Type "/"
Sleep 1s
Type "GOBIN"
Sleep 2s

Escape
Sleep 500ms

Hide
Type "q"
Sleep 500ms
