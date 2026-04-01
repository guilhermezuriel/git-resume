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
	repoPrimary  = lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}
	repoText     = lipgloss.AdaptiveColor{Light: "#1E293B", Dark: "#CDD6F4"}
	repoMuted    = lipgloss.AdaptiveColor{Light: "#94A3B8", Dark: "#6C7086"}
	repoSecond   = lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#A6ADC8"}
	repoSecColor = lipgloss.AdaptiveColor{Light: "#0EA5E9", Dark: "#38BDF8"}

	repoCursorStyle   = lipgloss.NewStyle().Foreground(repoPrimary).Bold(true)
	repoSelectedStyle = lipgloss.NewStyle().Foreground(repoPrimary).Bold(true)
	repoNormalStyle   = lipgloss.NewStyle().Foreground(repoText)
	repoDimStyle      = lipgloss.NewStyle().Foreground(repoMuted)
	repoInfoStyle     = lipgloss.NewStyle().Foreground(repoSecond)
	repoBadgeStyle    = lipgloss.NewStyle().Foreground(repoSecColor)
)

// RepoBrowser shows all repos in ~/.git-resumes.
type RepoBrowser struct {
	repos  []storage.RepoEntry
	cursor int
}

func NewRepoBrowser() RepoBrowser {
	repos, _ := storage.ListRepos()
	return RepoBrowser{repos: repos}
}

func (r RepoBrowser) Init() tea.Cmd { return nil }

func (r RepoBrowser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if r.cursor > 0 {
				r.cursor--
			}
		case "down", "j":
			if r.cursor < len(r.repos)-1 {
				r.cursor++
			}
		case "enter", " ":
			if len(r.repos) > 0 {
				repo := r.repos[r.cursor]
				return r, func() tea.Msg {
					return OpenRepoDirMsg{RepoName: repo.Name, RepoDir: repo.Dir}
				}
			}
		case "esc", "q":
			return r, func() tea.Msg { return NavigateMsg{To: ScreenMenu} }
		case "ctrl+c":
			return r, tea.Quit
		}
	}
	return r, nil
}

func (r RepoBrowser) View() string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(components.Header("git-resume", "All Repositories"))
	sb.WriteString("\n\n")

	if len(r.repos) == 0 {
		sb.WriteString(repoInfoStyle.Render("  No repositories found."))
		sb.WriteString("\n\n")
		sb.WriteString(repoDimStyle.Render("  Generate summaries in a git repository first."))
		sb.WriteString("\n\n")
		sb.WriteString(repoDimStyle.Render("  esc back"))
		return sb.String()
	}

	sb.WriteString(repoDimStyle.Render(fmt.Sprintf("  %d repositories\n\n", len(r.repos))))

	for i, repo := range r.repos {
		shortPath := repo.Path
		if len(shortPath) > 40 {
			shortPath = "..." + shortPath[len(shortPath)-37:]
		}
		count := fmt.Sprintf("%d", repo.Count)

		if i == r.cursor {
			sb.WriteString(repoCursorStyle.Render("  ▸ "))
			sb.WriteString(repoSelectedStyle.Render(fmt.Sprintf("%-20s", repo.Name)))
			sb.WriteString(repoDimStyle.Render("  " + fmt.Sprintf("%-42s", shortPath) + "  "))
			sb.WriteString(repoBadgeStyle.Render(count + " summaries"))
		} else {
			sb.WriteString(repoNormalStyle.Render(fmt.Sprintf("    %-20s", repo.Name)))
			sb.WriteString(repoDimStyle.Render("  " + fmt.Sprintf("%-42s", shortPath) + "  " + count + " summaries"))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(repoDimStyle.Render("  ↑↓ move  ·  enter open  ·  esc back"))
	sb.WriteString("\n")
	return sb.String()
}

// OpenRepoDirMsg is sent when user selects a repo from the browser.
type OpenRepoDirMsg struct {
	RepoName string
	RepoDir  string
}
