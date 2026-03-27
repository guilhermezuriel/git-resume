package screens

import (
	"fmt"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guilhermezuriel/git-resume/internal/tui/components"
)

var (
	viewerStatusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	viewerCopiedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	viewerCopyErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	viewerBorderStyle  = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, false).
				BorderForeground(lipgloss.Color("240"))
)

// SummaryViewer shows the content of a single resume file.
type SummaryViewer struct {
	repoName string
	filename string
	path     string
	content  string // raw text kept for clipboard copy
	viewport viewport.Model
	ready    bool
	copyMsg  string // non-empty while feedback is shown ("Copied!" or error)
	copyOK   bool   // true = success, false = error
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
		v.content = msg.Content
		v.viewport = viewport.New(80, 24)
		v.viewport.SetContent(msg.Content)
		v.ready = true
		return v, nil

	case CopyDoneMsg:
		if msg.Err != nil {
			v.copyMsg = fmt.Sprintf("Copy failed: %v", msg.Err)
			v.copyOK = false
		} else {
			v.copyMsg = "Copied to clipboard!"
			v.copyOK = true
		}
		return v, nil

	case tea.WindowSizeMsg:
		if v.ready {
			v.viewport.Width = msg.Width - 4
			v.viewport.Height = msg.Height - 8
		}
		return v, nil

	case tea.KeyMsg:
		// Any keypress clears the copy feedback.
		v.copyMsg = ""

		switch msg.String() {
		case "q", "esc":
			return v, func() tea.Msg { return NavigateMsg{To: ScreenSummaries} }
		case "ctrl+c":
			return v, tea.Quit
		case "c", "C":
			if v.ready && v.content != "" {
				return v, copyToClipboard(v.content)
			}
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

	pct := int(v.viewport.ScrollPercent() * 100)

	// Status bar: show copy feedback when present, otherwise show keybindings.
	if v.copyMsg != "" {
		left := viewerStatusStyle.Render(fmt.Sprintf("  %s  ·  %d%%  ·  ", v.filename, pct))
		var right string
		if v.copyOK {
			right = viewerCopiedStyle.Render("✓ " + v.copyMsg)
		} else {
			right = viewerCopyErrStyle.Render("✗ " + v.copyMsg)
		}
		sb.WriteString(left + right)
	} else {
		sb.WriteString(viewerStatusStyle.Render(
			fmt.Sprintf("  %s  ·  %d%%  ·  ↑↓ scroll  ·  c copy  ·  esc back", v.filename, pct),
		))
	}

	sb.WriteString("\n")
	return sb.String()
}

// CopyDoneMsg is sent when the clipboard write completes.
type CopyDoneMsg struct{ Err error }

// copyToClipboard writes text to the system clipboard asynchronously.
func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		return CopyDoneMsg{Err: clipboard.WriteAll(text)}
	}
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
