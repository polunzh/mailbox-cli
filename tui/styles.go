package tui

import (
	lipgloss "charm.land/lipgloss/v2"
)

// Style tokens shared across all TUI views.
var (
	StyleTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	StyleDim   = lipgloss.NewStyle().Faint(true)
	StyleBold  = lipgloss.NewStyle().Bold(true)
	StyleError = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)
