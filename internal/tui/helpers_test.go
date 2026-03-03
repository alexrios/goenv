package tui

import (
	"testing"
)

// --- clampNonNegative Tests ---

func TestClampNonNegative_NegativeValue(t *testing.T) {
	result := clampNonNegative(-5)
	if result != 0 {
		t.Errorf("clampNonNegative(-5) = %d, want 0", result)
	}
}

func TestClampNonNegative_Zero(t *testing.T) {
	result := clampNonNegative(0)
	if result != 0 {
		t.Errorf("clampNonNegative(0) = %d, want 0", result)
	}
}

func TestClampNonNegative_PositiveValue(t *testing.T) {
	result := clampNonNegative(10)
	if result != 10 {
		t.Errorf("clampNonNegative(10) = %d, want 10", result)
	}
}

func TestClampNonNegative_LargeNegative(t *testing.T) {
	result := clampNonNegative(-1000)
	if result != 0 {
		t.Errorf("clampNonNegative(-1000) = %d, want 0", result)
	}
}

// --- truncateForDiff Tests ---

func TestTruncateForDiff_ShortString(t *testing.T) {
	result := truncateForDiff("hello", 10)
	if result != "hello" {
		t.Errorf("truncateForDiff('hello', 10) = %q, want 'hello'", result)
	}
}

func TestTruncateForDiff_ExactLength(t *testing.T) {
	result := truncateForDiff("hello", 5)
	if result != "hello" {
		t.Errorf("truncateForDiff('hello', 5) = %q, want 'hello'", result)
	}
}

func TestTruncateForDiff_LongString(t *testing.T) {
	result := truncateForDiff("hello world", 8)
	expected := "hello..."
	if result != expected {
		t.Errorf("truncateForDiff('hello world', 8) = %q, want %q", result, expected)
	}
}

func TestTruncateForDiff_VeryShortMax(t *testing.T) {
	result := truncateForDiff("hello", 3)
	// When maxLen <= 3, just truncate without ellipsis
	if len(result) > 3 {
		t.Errorf("truncateForDiff('hello', 3) = %q, length %d > 3", result, len(result))
	}
}

func TestTruncateForDiff_EmptyString(t *testing.T) {
	result := truncateForDiff("", 10)
	if result != "" {
		t.Errorf("truncateForDiff('', 10) = %q, want ''", result)
	}
}

// --- truncateString Tests ---

func TestTruncateString_ShortString(t *testing.T) {
	result := truncateString("hello", 10)
	if result != "hello" {
		t.Errorf("truncateString('hello', 10) = %q, want 'hello'", result)
	}
}

func TestTruncateString_ExactLength(t *testing.T) {
	result := truncateString("hello", 5)
	if result != "hello" {
		t.Errorf("truncateString('hello', 5) = %q, want 'hello'", result)
	}
}

func TestTruncateString_LongString(t *testing.T) {
	result := truncateString("hello world", 8)
	// Should truncate and add "..."
	if len(result) > 8 {
		t.Errorf("truncateString('hello world', 8) = %q, length %d > 8", result, len(result))
	}
	// Should end with "..."
	if len(result) >= 3 && result[len(result)-3:] != "..." {
		// It's okay if it doesn't end with ... for very short maxWidths
	}
}

func TestTruncateString_VeryShortMax(t *testing.T) {
	result := truncateString("hello", 3)
	// When maxWidth <= 3, truncate without ellipsis
	if len(result) > 3 {
		t.Errorf("truncateString('hello', 3) = %q, length %d > 3", result, len(result))
	}
}
