package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guilhermezuriel/git-resume/internal/tui/components"
)

var (
	menuCursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	menuSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	menuNormalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	menuDimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// MainMenu is the root menu screen.
type MainMenu struct {
	repoName string
	items    []string
	cursor   int
}

func NewMainMenu(repoName string) MainMenu {
	return MainMenu{
		repoName: repoName,
		items: []string{
			"Generate new summary",
			"View summaries",
			"Browse all repositories",
			"Settings",
			"Exit",
		},
	}
}

func (m MainMenu) Init() tea.Cmd { return nil }

func (m MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter", " ":
			return m, func() tea.Msg {
				return MenuSelectedMsg{Index: m.cursor, Item: m.items[m.cursor]}
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MainMenu) View() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(components.Header("git-resume v3.0.0", m.repoName))
	sb.WriteString("\n\n")

	for i, item := range m.items {
		// Thin separator before the last item to group it as a distinct action.
		if i == len(m.items)-1 {
			sb.WriteString(menuDimStyle.Render("  " + strings.Repeat("─", 28)))
			sb.WriteString("\n")
		}

		if i == m.cursor {
			sb.WriteString(menuCursorStyle.Render("  ▸ "))
			sb.WriteString(menuSelectedStyle.Render(item))
		} else {
			sb.WriteString(menuNormalStyle.Render("    " + item))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(menuDimStyle.Render("  ↑↓ move  ·  enter select  ·  q quit"))
	sb.WriteString("\n")

	return sb.String()
}

// MenuSelectedMsg is sent when the user selects a menu item.
type MenuSelectedMsg struct {
	Index int
	Item  string
}
