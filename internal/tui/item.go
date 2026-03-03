package tui

import (
	"strings"

	"charm.land/bubbles/v2/list"

	"github.com/alexrios/goenv/internal/goenv"
)

// envItem wraps goenv.EnvVar and implements list.Item for the TUI list component.
type envItem struct {
	goenv.EnvVar
}

// Title returns the variable name for list display, with star if favorited.
func (i envItem) Title() string {
	if i.Favorite {
		return IconFavorite + " " + i.Key
	}
	return i.Key
}

// Description returns the variable value for list display.
func (i envItem) Description() string { return i.Value }

// FilterValue returns a combined string of key, value, and description for fuzzy filtering.
func (i envItem) FilterValue() string {
	parts := []string{i.Key, i.Value}
	if desc := goenv.GetEnvVarDescription(i.Key); desc != "" {
		parts = append(parts, desc)
	}
	return strings.Join(parts, " ")
}

// envVarsToListItems converts a slice of EnvVar to a slice of list.Item.
func envVarsToListItems(vars []goenv.EnvVar) []list.Item {
	items := make([]list.Item, len(vars))
	for i, ev := range vars {
		items[i] = envItem{ev}
	}
	return items
}

// EnvVarsFromListItems extracts EnvVar from list items, skipping non-envItem items.
func EnvVarsFromListItems(items []list.Item) []goenv.EnvVar {
	vars := make([]goenv.EnvVar, 0, len(items))
	for _, item := range items {
		if ei, ok := item.(envItem); ok {
			vars = append(vars, ei.EnvVar)
		}
	}
	return vars
}
