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
	sumCursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	sumSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	sumNormalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	sumDimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	sumInfoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
)

// SummaryList shows all resume files for a repo.
type SummaryList struct {
	repoName string
	repoDir  string
	files    []storage.ResumeFile
	cursor   int
	// backScreen is where Esc navigates back to.
	backScreen Screen
}

func NewSummaryList(repoName, repoDir string, back Screen) SummaryList {
	files, _ := storage.ListResumes(repoDir)
	return SummaryList{
		repoName:   repoName,
		repoDir:    repoDir,
		files:      files,
		backScreen: back,
	}
}

func (s SummaryList) Init() tea.Cmd { return nil }

func (s SummaryList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < len(s.files)-1 {
				s.cursor++
			}
		case "enter", " ":
			if len(s.files) > 0 {
				f := s.files[s.cursor]
				return s, func() tea.Msg {
					return OpenFileMsg{Path: f.Path, RepoName: s.repoName}
				}
			}
		case "esc", "q":
			return s, func() tea.Msg { return NavigateMsg{To: s.backScreen} }
		case "ctrl+c":
			return s, tea.Quit
		}
	}
	return s, nil
}

func (s SummaryList) View() string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(components.Header("git-resume v3.0.0", s.repoName))
	sb.WriteString("\n\n")

	if len(s.files) == 0 {
		sb.WriteString(sumInfoStyle.Render("  No summaries found for this repository."))
		sb.WriteString("\n\n")
		sb.WriteString(sumDimStyle.Render("  Generate your first summary from the main menu."))
		sb.WriteString("\n\n")
		sb.WriteString(sumDimStyle.Render("  esc back"))
		return sb.String()
	}

	sb.WriteString(sumDimStyle.Render(fmt.Sprintf("  %d summaries\n\n", len(s.files))))

	for i, f := range s.files {
		modTime := f.ModTime.Format("2006-01-02  15:04")
		if i == s.cursor {
			sb.WriteString(sumCursorStyle.Render("  ▸ "))
			sb.WriteString(sumSelectedStyle.Render(fmt.Sprintf("%-44s", f.Name)))
			sb.WriteString(sumDimStyle.Render("  " + modTime))
		} else {
			sb.WriteString(sumNormalStyle.Render(fmt.Sprintf("    %-44s", f.Name)))
			sb.WriteString(sumDimStyle.Render("  " + modTime))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(sumDimStyle.Render("  ↑↓ move  ·  enter view  ·  esc back"))
	sb.WriteString("\n")
	return sb.String()
}

// OpenFileMsg is sent when the user wants to open a file in the viewer.
type OpenFileMsg struct {
	Path     string
	RepoName string
}
