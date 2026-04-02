package tui

import (
	"os"

	"charm.land/lipgloss/v2"
	"golang.org/x/term"
)

// Color palette matching popular TUI tools
var (
	ColorPrimary   = lipgloss.Color("12") // Blue
	ColorSecondary = lipgloss.Color("8")  // Gray
	ColorSuccess   = lipgloss.Color("10") // Green
	ColorWarning   = lipgloss.Color("11") // Yellow
	ColorError     = lipgloss.Color("9")  // Red
	ColorHighlight = lipgloss.Color("14") // Cyan
	ColorText      = lipgloss.Color("15") // White
	ColorMuted     = lipgloss.Color("7")  // Light gray
)

// Base styles
var (
	StyleTitle = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	StyleDim   = lipgloss.NewStyle().Foreground(ColorSecondary)
	StyleBold  = lipgloss.NewStyle().Bold(true)
	StyleError = lipgloss.NewStyle().Foreground(ColorError)

	// Panel styles
	StylePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSecondary).
			Padding(0, 1)

	StylePanelActive = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1)

	StylePanelHeader = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorText).
				Background(ColorSecondary).
				Padding(0, 1)

	// List styles - improved highlighting, single line only
	StyleListItem = lipgloss.NewStyle().
			PaddingLeft(1).
			MaxHeight(1).
			Height(1)

	StyleListItemSelected = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorText).
				Background(ColorPrimary).
				PaddingLeft(1).
				MaxHeight(1).
				Height(1)

	StyleListItemUnread = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorText).
				MaxHeight(1).
				Height(1)

	// Status bar
	StyleStatusBar = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorSecondary)

	StyleStatusKey = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Background(ColorSecondary)

	// Help
	StyleHelpKey = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorHighlight)

	StyleHelpDesc = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// Unread indicator
	StyleUnreadDot = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	// Separator
	StyleSeparator = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	// Toolbar (bottom shortcut bar)
	StyleToolbar = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(lipgloss.Color("237"))

	StyleToolbarKey = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorHighlight).
			Background(lipgloss.Color("237"))

	StyleToolbarSep = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Background(lipgloss.Color("237"))
)

// TerminalWidth returns the terminal width, or 80 as fallback.
func TerminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

// TerminalHeight returns the terminal height, or 24 as fallback.
func TerminalHeight() int {
	_, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || h <= 0 {
		return 24
	}
	return h
}

// Truncate string to max width with ellipsis
func Truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
