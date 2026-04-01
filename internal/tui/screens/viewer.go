package screens

import (
	"fmt"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SummaryViewer shows the content of a single resume file (full-screen).
type SummaryViewer struct {
	repoName string
	filename string
	path     string
	content  string
	viewport viewport.Model
	width    int
	height   int
	ready    bool
	copyMsg  string
	copyOK   bool
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
		vw := v.width
		if vw == 0 {
			vw = 80
		}
		vh := v.height
		if vh == 0 {
			vh = 24
		}
		v.viewport = viewport.New(vw-4, vh-8)
		v.viewport.MouseWheelEnabled = true
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
		v.width = msg.Width
		v.height = msg.Height
		if v.ready {
			v.viewport.Width = msg.Width - 4
			v.viewport.Height = msg.Height - 8
		}
		return v, nil

	case tea.MouseMsg:
		if v.ready {
			var cmd tea.Cmd
			v.viewport, cmd = v.viewport.Update(msg)
			return v, cmd
		}

	case tea.KeyMsg:
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
	primary := lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}
	colorSuccess := lipgloss.AdaptiveColor{Light: "#10B981", Dark: "#34D399"}
	colorError := lipgloss.AdaptiveColor{Light: "#EF4444", Dark: "#F87171"}
	border := lipgloss.AdaptiveColor{Light: "#E2E8F0", Dark: "#45475A"}
	textMuted := lipgloss.AdaptiveColor{Light: "#94A3B8", Dark: "#6C7086"}
	textSecondary := lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#A6ADC8"}

	// Header bar
	logoStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(primary).
		Padding(0, 2)

	repoStyle := lipgloss.NewStyle().Foreground(textSecondary)
	header := lipgloss.JoinHorizontal(lipgloss.Center,
		logoStyle.Render(" git-resume "),
		repoStyle.Render("  · "+v.repoName),
	)

	if !v.ready {
		loadStyle := lipgloss.NewStyle().Foreground(textSecondary)
		return header + "\n\n" + loadStyle.Render("  Loading…") + "\n"
	}

	// Viewport in a card
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(0, 1)

	viewportBlock := cardStyle.Render(v.viewport.View())

	pct := int(v.viewport.ScrollPercent() * 100)

	// Status / help bar
	var statusBar string
	if v.copyMsg != "" {
		left := lipgloss.NewStyle().Foreground(textMuted).
			Render(fmt.Sprintf("  %s  ·  %d%%  ·  ", v.filename, pct))
		var right string
		if v.copyOK {
			right = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true).Render("✓ " + v.copyMsg)
		} else {
			right = lipgloss.NewStyle().Foreground(colorError).Bold(true).Render("✗ " + v.copyMsg)
		}
		statusBar = left + right
	} else {
		statusBar = lipgloss.NewStyle().Foreground(textMuted).Render(
			fmt.Sprintf("  %s  ·  %d%%  ·  ↑↓/mouse scroll  ·  c copy  ·  esc back", v.filename, pct),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		viewportBlock,
		statusBar,
	)
}

// CopyDoneMsg is sent when the clipboard write completes.
type CopyDoneMsg struct{ Err error }

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
