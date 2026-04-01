package components

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	primary    = lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}
	textMuted  = lipgloss.AdaptiveColor{Light: "#94A3B8", Dark: "#6C7086"}
	textSecond = lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#A6ADC8"}

	headerLogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primary).
			Padding(0, 2)

	headerSepStyle = lipgloss.NewStyle().
			Foreground(textMuted)

	headerSubStyle = lipgloss.NewStyle().
			Foreground(textSecond)
)

// Header renders a compact header bar with logo badge and optional subtitle.
func Header(title, subtitle string) string {
	logo := headerLogoStyle.Render(" " + title + " ")
	if subtitle == "" {
		return logo
	}
	sep := headerSepStyle.Render("  ·  ")
	sub := headerSubStyle.Render(subtitle)
	return lipgloss.JoinHorizontal(lipgloss.Center, logo, sep, sub)
}
