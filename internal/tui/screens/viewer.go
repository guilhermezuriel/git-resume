package screens

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guilhermezuriel/git-resume/internal/tui/components"
)

var (
	viewerStatusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	viewerBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, false).
				BorderForeground(lipgloss.Color("240"))
)

// SummaryViewer shows the content of a single resume file.
type SummaryViewer struct {
	repoName string
	filename string
	path     string
	viewport viewport.Model
	ready    bool
}

func NewSummaryViewer(repoName, path string) SummaryViewer {
	parts := strings.Split(path, string(os.PathSeparator))
	filename := parts[len(parts)-1]
	return SummaryViewer{
		repoName: repoName,
		filename: filename,
		path:     path,
	}
}

func (v SummaryViewer) Init() tea.Cmd {
	return loadFile(v.path)
}

func (v SummaryViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FileLoadedMsg:
		v.viewport = viewport.New(80, 24)
		v.viewport.SetContent(msg.Content)
		v.ready = true
		return v, nil

	case tea.WindowSizeMsg:
		if v.ready {
			v.viewport.Width = msg.Width - 4
			v.viewport.Height = msg.Height - 8
		}
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return v, func() tea.Msg { return NavigateMsg{To: ScreenSummaries} }
		case "ctrl+c":
			return v, tea.Quit
		}
		if v.ready {
			var cmd tea.Cmd
			v.viewport, cmd = v.viewport.Update(msg)
			return v, cmd
		}
	}
	return v, nil
}

func (v SummaryViewer) View() string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(components.Header("git-resume v3.0.0", v.repoName))
	sb.WriteString("\n\n")

	if !v.ready {
		sb.WriteString(viewerStatusStyle.Render("  Loading..."))
		return sb.String()
	}

	sb.WriteString(viewerBorderStyle.Render(v.viewport.View()))
	sb.WriteString("\n")
	sb.WriteString(viewerStatusStyle.Render(
		fmt.Sprintf("  %s  ·  %d%%  ·  ↑↓ scroll  ·  esc back",
			v.filename,
			int(v.viewport.ScrollPercent()*100),
		),
	))
	sb.WriteString("\n")
	return sb.String()
}

// loadFile reads a file and returns a FileLoadedMsg.
func loadFile(path string) tea.Cmd {
	return func() tea.Msg {
		data, err := os.ReadFile(path)
		if err != nil {
			return FileLoadedMsg{Content: fmt.Sprintf("Error reading file: %v", err), Path: path}
		}
		return FileLoadedMsg{Content: string(data), Path: path}
	}
}

// FileLoadedMsg is sent when a file has been loaded into memory.
type FileLoadedMsg struct {
	Content string
	Path    string
}
