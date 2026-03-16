package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#7C3AED") // violet
	successColor   = lipgloss.Color("#10B981") // green
	errorColor     = lipgloss.Color("#EF4444") // red
	warningColor   = lipgloss.Color("#F59E0B") // amber
	mutedColor     = lipgloss.Color("#6B7280") // gray
	accentColor    = lipgloss.Color("#3B82F6") // blue
	highlightColor = lipgloss.Color("#E0E7FF") // light indigo

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	promptStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	menuItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	menuSelectedStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(primaryColor).
				Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	infoBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor).
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(highlightColor).
			Background(primaryColor).
			Padding(0, 2).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true).
			Width(22)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)
)
