package tui

import "github.com/charmbracelet/lipgloss"

// Adaptive color palette — works in both dark and light terminals.
var (
	// Brand
	Primary   = lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}
	Secondary = lipgloss.AdaptiveColor{Light: "#0EA5E9", Dark: "#38BDF8"}
	Accent    = lipgloss.AdaptiveColor{Light: "#10B981", Dark: "#34D399"}

	// Surfaces
	Border      = lipgloss.AdaptiveColor{Light: "#E2E8F0", Dark: "#45475A"}
	BorderFocus = lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}

	// Text
	TextPrimary   = lipgloss.AdaptiveColor{Light: "#1E293B", Dark: "#CDD6F4"}
	TextSecondary = lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#A6ADC8"}
	TextMuted     = lipgloss.AdaptiveColor{Light: "#94A3B8", Dark: "#6C7086"}

	// Status
	ColorSuccess = lipgloss.AdaptiveColor{Light: "#10B981", Dark: "#34D399"}
	ColorWarning = lipgloss.AdaptiveColor{Light: "#F59E0B", Dark: "#FBBF24"}
	ColorError   = lipgloss.AdaptiveColor{Light: "#EF4444", Dark: "#F87171"}
	ColorInfo    = lipgloss.AdaptiveColor{Light: "#3B82F6", Dark: "#60A5FA"}
)

// Shared base styles.
var (
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(0, 1)

	CardFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(BorderFocus).
				Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(TextSecondary)

	MutedStyle = lipgloss.NewStyle().
			Foreground(TextMuted)

	BadgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(Secondary).
			Padding(0, 1)

	BadgeSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(ColorSuccess).
				Padding(0, 1)

	BadgeErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorError).
			Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(TextMuted)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(Primary)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)
)
