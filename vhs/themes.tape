# Themes demo: Show default, nord, and dracula themes
Output docs/gifs/themes.gif

Set Shell "bash"
Set FontSize 16
Set Width 1000
Set Height 600
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms

# Setup: ensure config dir, build binary, backup config
Hide
Type "mkdir -p ~/.config/goenv && go build -o goenv ./cmd/goenv"
Enter
Sleep 1s
Type "cp ~/.config/goenv/config.json ~/.config/goenv/config.json.bak 2>/dev/null; true"
Enter
Sleep 500ms

# === DEFAULT THEME ===
Type `echo '{"theme":"default"}' > ~/.config/goenv/config.json`
Enter
Sleep 500ms
Type "./goenv"
Enter
Sleep 2s
Show

Sleep 1s
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "k"
Sleep 300ms
Type "k"
Sleep 1.5s

# === NORD THEME ===
Hide
Type "q"
Sleep 500ms
Type `echo '{"theme":"nord"}' > ~/.config/goenv/config.json`
Enter
Sleep 500ms
Type "./goenv"
Enter
Sleep 2s
Show

Sleep 1s
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "k"
Sleep 300ms
Type "k"
Sleep 1.5s

# === DRACULA THEME ===
Hide
Type "q"
Sleep 500ms
Type `echo '{"theme":"dracula"}' > ~/.config/goenv/config.json`
Enter
Sleep 500ms
Type "./goenv"
Enter
Sleep 2s
Show

Sleep 1s
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "k"
Sleep 300ms
Type "k"
Sleep 1.5s

# Cleanup: restore original config
Hide
Type "q"
Sleep 500ms
Type "cp ~/.config/goenv/config.json.bak ~/.config/goenv/config.json 2>/dev/null; true"
Enter
Sleep 500ms
