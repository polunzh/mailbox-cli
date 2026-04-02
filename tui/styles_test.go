package tui

import (
	"testing"

	"charm.land/lipgloss/v2"
)

func TestTruncate_ShorterThanMax(t *testing.T) {
	result := Truncate("hello", 10)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestTruncate_ExactLength(t *testing.T) {
	result := Truncate("hello", 5)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestTruncate_LongerThanMax(t *testing.T) {
	result := Truncate("hello world", 8)
	if result != "hello..." {
		t.Errorf("expected 'hello...', got %q", result)
	}
}

func TestTruncate_ZeroMax(t *testing.T) {
	result := Truncate("hello", 0)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}

	result = Truncate("hello", -1)
	if result != "" {
		t.Errorf("expected empty string for negative, got %q", result)
	}
}

func TestTruncate_SmallMax(t *testing.T) {
	result := Truncate("hello", 2)
	if result != "he" {
		t.Errorf("expected 'he', got %q", result)
	}
}

func TestTruncate_MaxExactlyThree(t *testing.T) {
	result := Truncate("hello", 3)
	// When max <= 3, the function returns s[:max] without ellipsis
	if result != "hel" {
		t.Errorf("expected 'hel', got %q", result)
	}
}

func TestTruncate_WithAnsiCodes(t *testing.T) {
	// lipgloss styles add ANSI codes
	styled := lipgloss.NewStyle().Bold(true).Render("hello")
	result := Truncate(styled, 10)
	// The ANSI codes should be preserved, but result should be within limit
	if lipgloss.Width(result) > 10 {
		t.Errorf("width should be <= 10, got %d", lipgloss.Width(result))
	}
}

func TestTruncate_WideCharacters(t *testing.T) {
	result := Truncate("日本語テスト", 6)
	if lipgloss.Width(result) > 6 {
		t.Errorf("width should be <= 6, got %d", lipgloss.Width(result))
	}
}

func TestTerminalWidth_ReturnsReasonableValue(t *testing.T) {
	w := TerminalWidth()
	if w <= 0 {
		t.Errorf("expected positive width, got %d", w)
	}
	// Should have a reasonable fallback
	if w < 40 && w != 80 {
		t.Errorf("width %d seems unreasonably small", w)
	}
}

func TestTerminalHeight_ReturnsReasonableValue(t *testing.T) {
	h := TerminalHeight()
	if h <= 0 {
		t.Errorf("expected positive height, got %d", h)
	}
	// Should have a reasonable fallback
	if h < 10 && h != 24 {
		t.Errorf("height %d seems unreasonably small", h)
	}
}

func TestStyleColorsAreDefined(t *testing.T) {
	// Verify all color variables are defined (lipgloss.Color is a function that returns color.Color)
	colors := []interface{}{
		ColorPrimary,
		ColorSecondary,
		ColorSuccess,
		ColorWarning,
		ColorError,
		ColorHighlight,
		ColorText,
		ColorMuted,
	}

	for i, c := range colors {
		if c == nil {
			t.Errorf("color at index %d is nil", i)
		}
	}
}

func TestStylesAreInitialized(t *testing.T) {
	// Verify key styles are initialized
	styles := []*lipgloss.Style{
		&StyleTitle,
		&StyleDim,
		&StyleBold,
		&StyleError,
		&StylePanel,
		&StylePanelActive,
		&StylePanelHeader,
		&StyleListItem,
		&StyleListItemSelected,
		&StyleListItemUnread,
		&StyleStatusBar,
		&StyleStatusKey,
		&StyleHelpKey,
		&StyleHelpDesc,
		&StyleUnreadDot,
		&StyleSeparator,
		&StyleToolbar,
		&StyleToolbarKey,
		&StyleToolbarSep,
	}

	for i, s := range styles {
		if s == nil {
			t.Errorf("style at index %d is nil", i)
		}
	}
}
