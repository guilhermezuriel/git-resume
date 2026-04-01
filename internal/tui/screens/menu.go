package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MainMenu is the root menu screen.
type MainMenu struct {
	repoName string
	items    []menuItem
	cursor   int
	width    int
	height   int
}

type menuItem struct {
	key   string
	label string
	desc  string
}

func NewMainMenu(repoName string) MainMenu {
	return MainMenu{
		repoName: repoName,
		items: []menuItem{
			{"n", "New Summary", "Generate a new commit summary"},
			{"s", "View Summaries", "Browse and preview saved summaries"},
			{"b", "Browse Repos", "View all repositories"},
			{"⚙", "Settings", "Configure preferences"},
			{"q", "Quit", "Exit git-resume"},
		},
	}
}

func (m MainMenu) Init() tea.Cmd { return nil }

func (m MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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
				return MenuSelectedMsg{Index: m.cursor, Item: m.items[m.cursor].label}
			}
		case "n":
			return m, func() tea.Msg { return MenuSelectedMsg{Index: 0, Item: "New Summary"} }
		case "s":
			return m, func() tea.Msg { return MenuSelectedMsg{Index: 1, Item: "View Summaries"} }
		case "b":
			return m, func() tea.Msg { return MenuSelectedMsg{Index: 2, Item: "Browse Repos"} }
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MainMenu) View() string {
	// Adaptive colors
	primary := lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}
	textPrimary := lipgloss.AdaptiveColor{Light: "#1E293B", Dark: "#CDD6F4"}
	textMuted := lipgloss.AdaptiveColor{Light: "#94A3B8", Dark: "#6C7086"}
	border := lipgloss.AdaptiveColor{Light: "#E2E8F0", Dark: "#45475A"}

	// Logo box
	logo := lipgloss.NewStyle().
		Foreground(primary).
		Bold(true).
		Render(
			  "   ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓\n" +
				"   ┃       git-resume            ┃\n" +
				"   ┃   Daily Commit Summaries    ┃\n" +
				"   ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛",
		)

	// Repo info below logo
	repoLine := lipgloss.NewStyle().
		Foreground(textMuted).
		Render(fmt.Sprintf("   Repository: %s", m.repoName))

	// Menu items
	var menuLines strings.Builder
	menuLines.WriteString("\n")

	sepStyle := lipgloss.NewStyle().Foreground(border)

	for i, item := range m.items {
		// Separator before quit
		if i == len(m.items)-1 {
			menuLines.WriteString("   " + sepStyle.Render(strings.Repeat("─", 32)) + "\n")
		}

		keyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primary).
			Padding(0, 1).
			Bold(true)

		labelStyle := lipgloss.NewStyle().
			Foreground(textPrimary).
			Bold(i == m.cursor).
			MarginLeft(2)

		descStyle := lipgloss.NewStyle().
			Foreground(textMuted).
			MarginLeft(2)

		cursor := "  "
		if i == m.cursor {
			cursor = lipgloss.NewStyle().Foreground(primary).Bold(true).Render("▸ ")
		}

		line := cursor + keyStyle.Render(item.key) + labelStyle.Render(" "+item.label)
		menuLines.WriteString("  " + line + "\n")
		menuLines.WriteString(descStyle.Render("        "+item.desc) + "\n\n")
	}

	// Footer hint
	helpStyle := lipgloss.NewStyle().Foreground(textMuted)
	footer := helpStyle.Render("   ↑↓ navigate  ·  enter select  ·  q quit")

	return "\n" + logo + "\n" + repoLine + "\n" + menuLines.String() + footer + "\n"
}

// MenuSelectedMsg is sent when the user selects a menu item.
type MenuSelectedMsg struct {
	Index int
	Item  string
}
