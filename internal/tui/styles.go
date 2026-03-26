package tui

import "github.com/charmbracelet/lipgloss"

// Color palette — terminal ANSI colors for wide compatibility.
var (
	colorCyan   = lipgloss.Color("6")
	colorDim    = lipgloss.Color("240")
	colorGreen  = lipgloss.Color("2")
	colorRed    = lipgloss.Color("1")
	colorYellow = lipgloss.Color("3")
	colorBlue   = lipgloss.Color("4")
	colorWhite  = lipgloss.Color("15")
)

var (
	// Header: rounded border in cyan, with horizontal padding.
	headerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorCyan).
			Padding(0, 2)

	headerTitleStyle = lipgloss.NewStyle().
				Foreground(colorCyan).
				Bold(true)

	// Status bar at the bottom.
	statusStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	// Cursor indicator for list items.
	cursorStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(colorCyan)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	dimItemStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	successStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(colorYellow)

	infoStyle = lipgloss.NewStyle().
			Foreground(colorBlue)

	labelStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	valueStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	sectionStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)
)
