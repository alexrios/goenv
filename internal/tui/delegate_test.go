package tui

import (
	"bytes"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/persist"
)

func TestCategoryTag_ShowsTagForFirstInGroup(t *testing.T) {
	items := []list.Item{
		envItem{goenv.EnvVar{Key: "GOROOT", Value: "/usr/local/go"}},  // General
		envItem{goenv.EnvVar{Key: "GOPATH", Value: "/home/user/go"}},  // General
		envItem{goenv.EnvVar{Key: "GOPROXY", Value: "direct"}},        // Proxy/Network
	}
	d := NewEnvVarDelegateWithSortMode(persist.SortCategory, nil)
	l := list.New(items, d, 80, 40)

	// First item (GOROOT) should have category tag
	tag := d.categoryTag(l, 0, items[0].(envItem))
	if tag == "" {
		t.Error("first item should have a category tag")
	}
	if !strings.Contains(tag, "General") {
		t.Errorf("expected [General] tag, got %q", tag)
	}

	// Second item (GOPATH) same category - no tag
	tag = d.categoryTag(l, 1, items[1].(envItem))
	if tag != "" {
		t.Errorf("second item in same category should have no tag, got %q", tag)
	}

	// Third item (GOPROXY) different category - should have tag
	tag = d.categoryTag(l, 2, items[2].(envItem))
	if tag == "" {
		t.Error("first item in new category should have a tag")
	}
	if !strings.Contains(tag, "Proxy") {
		t.Errorf("expected [Proxy/Network] tag, got %q", tag)
	}
}

func TestCategoryTag_NoTagWhenNotCategorySort(t *testing.T) {
	items := []list.Item{
		envItem{goenv.EnvVar{Key: "GOROOT", Value: "/usr/local/go"}},
	}
	d := NewEnvVarDelegateWithSortMode(persist.SortAlpha, nil)
	l := list.New(items, d, 80, 40)

	tag := d.categoryTag(l, 0, items[0].(envItem))
	if tag != "" {
		t.Errorf("should have no tag in alpha sort mode, got %q", tag)
	}
}

func TestCategoryTag_NoTagForFavorites(t *testing.T) {
	favorites := map[string]bool{"GOROOT": true}
	items := []list.Item{
		envItem{goenv.EnvVar{Key: "GOROOT", Value: "/usr/local/go", Favorite: true}},
	}
	d := NewEnvVarDelegateWithSortMode(persist.SortCategory, favorites)
	l := list.New(items, d, 80, 40)

	tag := d.categoryTag(l, 0, items[0].(envItem))
	if tag != "" {
		t.Errorf("favorited items should have no category tag, got %q", tag)
	}
}

func TestDelegate_RendersCategoryTagInOutput(t *testing.T) {
	items := []list.Item{
		envItem{goenv.EnvVar{Key: "GOROOT", Value: "/usr/local/go"}},  // General
		envItem{goenv.EnvVar{Key: "GOPROXY", Value: "direct"}},        // Proxy/Network
	}
	d := NewEnvVarDelegateWithSortMode(persist.SortCategory, nil)
	l := list.New(items, d, 80, 40)

	var buf bytes.Buffer
	d.Render(&buf, l, 0, items[0])
	output := buf.String()

	if !strings.Contains(output, "General") {
		t.Error("rendered output for first item should contain category tag")
	}
}
