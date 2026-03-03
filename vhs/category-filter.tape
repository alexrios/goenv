# Category Filter demo: Cycle through category filters with C
Output docs/gifs/category-filter.gif

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

# Let TUI fully render - showing All
Sleep 2s

# Cycle to General
Type "C"
Sleep 1.5s

# Cycle to Modules
Type "C"
Sleep 1.5s

# Cycle to Proxy/Network
Type "C"
Sleep 1.5s

# Cycle to Build
Type "C"
Sleep 1.5s

# Cycle to CGO
Type "C"
Sleep 1.5s

# Cycle to Architecture
Type "C"
Sleep 1.5s

# Cycle to Tooling
Type "C"
Sleep 1.5s

# Cycle to Debug/Telemetry
Type "C"
Sleep 1.5s

# Back to All
Type "C"
Sleep 2s

Hide
Type "q"
Sleep 500ms
