package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guilhermezuriel/git-resume/internal/storage"
	"github.com/guilhermezuriel/git-resume/internal/tui/components"
)

var (
	settingsPrimary  = lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}
	settingsSucc     = lipgloss.AdaptiveColor{Light: "#10B981", Dark: "#34D399"}
	settingsWarn     = lipgloss.AdaptiveColor{Light: "#F59E0B", Dark: "#FBBF24"}
	settingsTextMain = lipgloss.AdaptiveColor{Light: "#1E293B", Dark: "#CDD6F4"}
	settingsTextMut  = lipgloss.AdaptiveColor{Light: "#94A3B8", Dark: "#6C7086"}
	settingsTextSec  = lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#A6ADC8"}
	settingsSecColor = lipgloss.AdaptiveColor{Light: "#0EA5E9", Dark: "#38BDF8"}

	settingsCursorStyle   = lipgloss.NewStyle().Foreground(settingsPrimary).Bold(true)
	settingsSelectedStyle = lipgloss.NewStyle().Foreground(settingsPrimary).Bold(true)
	settingsNormalStyle   = lipgloss.NewStyle().Foreground(settingsTextMain)
	settingsDimStyle      = lipgloss.NewStyle().Foreground(settingsTextMut)
	settingsSuccessStyle  = lipgloss.NewStyle().Foreground(settingsSucc).Bold(true)
	settingsWarningStyle  = lipgloss.NewStyle().Foreground(settingsWarn)
	settingsLabelStyle    = lipgloss.NewStyle().Foreground(settingsTextSec)
	settingsValueStyle    = lipgloss.NewStyle().Foreground(settingsTextMain)
	settingsAccentStyle   = lipgloss.NewStyle().Foreground(settingsSecColor)
)

type settingsView int

const (
	settingsMain settingsView = iota
	settingsConfirmDelete
	settingsDeleted
)

// Settings shows storage statistics and allows clearing all summaries.
type Settings struct {
	items   []string
	cursor  int
	view    settingsView
	message string

	// Stats (loaded on init)
	repos int
	files int
	size  string
}

func NewSettings() Settings {
	repos, files, size := storage.StorageStats()
	return Settings{
		items:  []string{"View storage stats", "Clear all summaries", "< Back"},
		repos:  repos,
		files:  files,
		size:   size,
	}
}

func (s Settings) Init() tea.Cmd { return nil }

func (s Settings) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch s.view {
		case settingsMain:
			return s.updateMain(msg)
		case settingsConfirmDelete:
			return s.updateConfirm(msg)
		case settingsDeleted:
			s.view = settingsMain
			return s, nil
		}
	}
	return s, nil
}

func (s Settings) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
		}
	case "down", "j":
		if s.cursor < len(s.items)-1 {
			s.cursor++
		}
	case "enter", " ":
		switch s.cursor {
		case 0:
			// Stats are already loaded; just display them (handled in View)
			s.message = fmt.Sprintf(
				"Repos: %d  |  Summaries: %d  |  Size: %s",
				s.repos, s.files, s.size,
			)
		case 1:
			s.view = settingsConfirmDelete
		case 2:
			return s, func() tea.Msg { return NavigateMsg{To: ScreenMenu} }
		}
	case "esc", "q":
		return s, func() tea.Msg { return NavigateMsg{To: ScreenMenu} }
	case "ctrl+c":
		return s, tea.Quit
	}
	return s, nil
}

func (s Settings) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		if err := storage.ClearAll(); err != nil {
			s.message = "Error: " + err.Error()
		} else {
			s.message = "All summaries deleted."
			s.repos = 0
			s.files = 0
			s.size = "0 B"
		}
		s.view = settingsDeleted
	case "n", "N", "esc", "q":
		s.view = settingsMain
	case "ctrl+c":
		return s, tea.Quit
	}
	return s, nil
}

func (s Settings) View() string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(components.Header("git-resume", "Settings"))
	sb.WriteString("\n\n")

	switch s.view {
	case settingsMain:
		sb.WriteString(settingsDimStyle.Render(fmt.Sprintf("  %s\n\n", storage.BaseDir())))

		sb.WriteString(settingsDimStyle.Render("  " + strings.Repeat("─", 32)) + "\n")
		sb.WriteString(fmt.Sprintf("  %-14s %s\n",
			settingsLabelStyle.Render("Repositories"),
			settingsAccentStyle.Render(fmt.Sprintf("%d", s.repos)),
		))
		sb.WriteString(fmt.Sprintf("  %-14s %s\n",
			settingsLabelStyle.Render("Summaries"),
			settingsAccentStyle.Render(fmt.Sprintf("%d", s.files)),
		))
		sb.WriteString(fmt.Sprintf("  %-14s %s\n",
			settingsLabelStyle.Render("Disk usage"),
			settingsValueStyle.Render(s.size),
		))
		sb.WriteString(settingsDimStyle.Render("  "+strings.Repeat("─", 32)) + "\n\n")

		for i, item := range s.items {
			if i == s.cursor {
				sb.WriteString(settingsCursorStyle.Render("  ▸ "))
				sb.WriteString(settingsSelectedStyle.Render(item))
			} else {
				sb.WriteString(settingsNormalStyle.Render("    " + item))
			}
			sb.WriteString("\n")
		}

		if s.message != "" {
			sb.WriteString("\n")
			sb.WriteString(settingsSuccessStyle.Render("  " + s.message))
			sb.WriteString("\n")
		}

		sb.WriteString("\n")
		sb.WriteString(settingsDimStyle.Render("  ↑↓ move  ·  enter select  ·  esc back"))

	case settingsConfirmDelete:
		sb.WriteString(settingsWarningStyle.Render("  Delete ALL summaries? This cannot be undone."))
		sb.WriteString("\n\n")
		sb.WriteString(settingsDimStyle.Render("  "))
		sb.WriteString(settingsSelectedStyle.Render("y"))
		sb.WriteString(settingsDimStyle.Render(" confirm  ·  "))
		sb.WriteString(settingsSelectedStyle.Render("n"))
		sb.WriteString(settingsDimStyle.Render(" / "))
		sb.WriteString(settingsSelectedStyle.Render("esc"))
		sb.WriteString(settingsDimStyle.Render(" cancel"))

	case settingsDeleted:
		sb.WriteString(settingsSuccessStyle.Render("  " + s.message))
		sb.WriteString("\n\n")
		sb.WriteString(settingsDimStyle.Render("  press any key to continue"))
	}

	sb.WriteString("\n")
	return strings.TrimRight(sb.String(), " ")
}
