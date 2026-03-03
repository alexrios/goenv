package tui

import (
	"slices"
	"testing"
)

func TestGetTheme_Default(t *testing.T) {
	theme := GetTheme(ThemeDefault)
	if theme.Name != ThemeDefault {
		t.Errorf("GetTheme(%q).Name = %q, want %q", ThemeDefault, theme.Name, ThemeDefault)
	}
}

func TestGetTheme_Nord(t *testing.T) {
	theme := GetTheme(ThemeNord)
	if theme.Name != ThemeNord {
		t.Errorf("GetTheme(%q).Name = %q, want %q", ThemeNord, theme.Name, ThemeNord)
	}
}

func TestGetTheme_Dracula(t *testing.T) {
	theme := GetTheme(ThemeDracula)
	if theme.Name != ThemeDracula {
		t.Errorf("GetTheme(%q).Name = %q, want %q", ThemeDracula, theme.Name, ThemeDracula)
	}
}

func TestGetTheme_UnknownFallsBackToDefault(t *testing.T) {
	theme := GetTheme("nonexistent")
	if theme.Name != ThemeDefault {
		t.Errorf("GetTheme(unknown).Name = %q, want %q", theme.Name, ThemeDefault)
	}
}

func TestNextTheme_FullCycle(t *testing.T) {
	// default -> nord -> dracula -> default
	got := NextTheme(ThemeDefault)
	if got != ThemeNord {
		t.Errorf("NextTheme(default) = %q, want %q", got, ThemeNord)
	}
	got = NextTheme(ThemeNord)
	if got != ThemeDracula {
		t.Errorf("NextTheme(nord) = %q, want %q", got, ThemeDracula)
	}
	got = NextTheme(ThemeDracula)
	if got != ThemeDefault {
		t.Errorf("NextTheme(dracula) = %q, want %q", got, ThemeDefault)
	}
}

func TestNextTheme_UnknownReturnsDefault(t *testing.T) {
	got := NextTheme("unknown")
	if got != ThemeDefault {
		t.Errorf("NextTheme(unknown) = %q, want %q", got, ThemeDefault)
	}
}

func TestAvailableThemes_ContainsAllThemes(t *testing.T) {
	themes := AvailableThemes()
	if len(themes) != 3 {
		t.Fatalf("AvailableThemes() returned %d themes, want 3", len(themes))
	}
	for _, name := range []string{ThemeDefault, ThemeNord, ThemeDracula} {
		if !slices.Contains(themes, name) {
			t.Errorf("AvailableThemes() missing %q", name)
		}
	}
}
