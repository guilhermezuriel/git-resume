package components

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	colorCyan  = lipgloss.Color("6")
	colorDim   = lipgloss.Color("240")
	colorWhite = lipgloss.Color("15")

	headerBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorCyan).
				Padding(0, 2)

	headerTitleStyle = lipgloss.NewStyle().
				Foreground(colorCyan).
				Bold(true)

	headerSepStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	headerSubStyle = lipgloss.NewStyle().
			Foreground(colorWhite)
)

// Header renders a rounded-bordered header bar with title and optional subtitle.
func Header(title, subtitle string) string {
	content := headerTitleStyle.Render(title)
	if subtitle != "" {
		content += headerSepStyle.Render("  ·  ") + headerSubStyle.Render(subtitle)
	}
	return headerBorderStyle.Render(content)
}
