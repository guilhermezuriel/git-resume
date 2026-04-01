package screens

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guilhermezuriel/git-resume/internal/storage"
)

// previewLoadedMsg carries a freshly-read file content for the preview panel.
type previewLoadedMsg struct {
	content string
}

func loadPreview(path string) tea.Cmd {
	return func() tea.Msg {
		data, err := os.ReadFile(path)
		if err != nil {
			return previewLoadedMsg{content: fmt.Sprintf("Error reading file:\n%v", err)}
		}
		return previewLoadedMsg{content: string(data)}
	}
}

// SummaryList shows all resume files for a repo with an inline split-pane preview.
type SummaryList struct {
	repoName string
	repoDir  string
	files    []storage.ResumeFile
	cursor   int
	// backScreen is where Esc navigates back to.
	backScreen Screen

	// layout
	width      int
	height     int
	focusPanel int // 0 = list, 1 = preview

	// preview panel
	preview        viewport.Model
	previewContent string
	previewReady   bool
	previewLoading bool

	// copy feedback
	copyMsg string
	copyOK  bool
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

func (s SummaryList) Init() tea.Cmd {
	if len(s.files) > 0 {
		s.previewLoading = true
		return loadPreview(s.files[0].Path)
	}
	return nil
}

func (s SummaryList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.preview.Width = s.previewWidth() - 4
		s.preview.Height = s.contentHeight() - 4
		if s.previewContent != "" {
			s.preview.SetContent(s.previewContent)
		}
		return s, nil

	case CopyDoneMsg:
		if msg.Err != nil {
			s.copyMsg = fmt.Sprintf("Copy failed: %v", msg.Err)
			s.copyOK = false
		} else {
			s.copyMsg = "Copied to clipboard!"
			s.copyOK = true
		}
		return s, nil

	case previewLoadedMsg:
		s.previewContent = msg.content
		s.previewLoading = false
		s.previewReady = true
		s.preview = viewport.New(s.previewWidth()-4, s.contentHeight()-4)
		s.preview.SetContent(msg.content)
		return s, nil

	case tea.MouseMsg:
		if s.focusPanel == 1 {
			var cmd tea.Cmd
			s.preview, cmd = s.preview.Update(msg)
			return s, cmd
		}

	case tea.KeyMsg:
		s.copyMsg = ""
		switch msg.String() {
		case "tab":
			s.focusPanel = (s.focusPanel + 1) % 2
			return s, nil

		case "left", "h":
			s.focusPanel = 0
			return s, nil

		case "right", "l":
			if s.previewReady {
				s.focusPanel = 1
			}
			return s, nil

		case "up", "k":
			if s.focusPanel == 0 {
				if s.cursor > 0 {
					s.cursor--
					s.previewLoading = true
					s.previewReady = false
					return s, loadPreview(s.files[s.cursor].Path)
				}
			} else {
				var cmd tea.Cmd
				s.preview, cmd = s.preview.Update(msg)
				return s, cmd
			}

		case "down", "j":
			if s.focusPanel == 0 {
				if s.cursor < len(s.files)-1 {
					s.cursor++
					s.previewLoading = true
					s.previewReady = false
					return s, loadPreview(s.files[s.cursor].Path)
				}
			} else {
				var cmd tea.Cmd
				s.preview, cmd = s.preview.Update(msg)
				return s, cmd
			}

		case "pgup", "u":
			if s.focusPanel == 1 {
				var cmd tea.Cmd
				s.preview, cmd = s.preview.Update(msg)
				return s, cmd
			}

		case "pgdown":
			if s.focusPanel == 1 {
				var cmd tea.Cmd
				s.preview, cmd = s.preview.Update(msg)
				return s, cmd
			}

		case "enter", " ":
			if len(s.files) > 0 && s.focusPanel == 0 && s.previewReady {
				s.focusPanel = 1
				return s, nil
			}

		case "c", "C":
			if s.previewReady && s.previewContent != "" {
				return s, copyToClipboard(s.previewContent)
			}

		case "n":
			return s, func() tea.Msg { return NavigateMsg{To: ScreenGenerate} }

		case "esc", "q":
			return s, func() tea.Msg { return NavigateMsg{To: s.backScreen} }

		case "ctrl+c":
			return s, tea.Quit
		}
	}

	return s, nil
}

func (s SummaryList) sidebarWidth() int {
	w := s.width / 3
	if w < 32 {
		w = 32
	}
	return w
}

func (s SummaryList) previewWidth() int {
	pw := s.width - s.sidebarWidth() - 2
	if pw < 20 {
		pw = 20
	}
	return pw
}

func (s SummaryList) contentHeight() int {
	h := s.height - 5 // header + footer
	if h < 5 {
		h = 5
	}
	return h
}

func (s SummaryList) View() string {
	// Colors
	primary := lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}
	secondary := lipgloss.AdaptiveColor{Light: "#0EA5E9", Dark: "#38BDF8"}
	colorSuccess := lipgloss.AdaptiveColor{Light: "#10B981", Dark: "#34D399"}
	border := lipgloss.AdaptiveColor{Light: "#E2E8F0", Dark: "#45475A"}
	borderFocus := lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}
	textPrimary := lipgloss.AdaptiveColor{Light: "#1E293B", Dark: "#CDD6F4"}
	textSecondary := lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#A6ADC8"}
	textMuted := lipgloss.AdaptiveColor{Light: "#94A3B8", Dark: "#6C7086"}

	// If terminal size not yet known, show minimal view.
	if s.width == 0 {
		return s.viewCompact()
	}

	sideW := s.sidebarWidth()
	prevW := s.previewWidth()
	contH := s.contentHeight()

	// ── Header ────────────────────────────────────────────────────────────────
	logoStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(primary).
		Padding(0, 2)

	repoStyle := lipgloss.NewStyle().Foreground(textSecondary)

	logo := logoStyle.Render(" git-resume ")
	repoInfo := repoStyle.Render(fmt.Sprintf("%s", s.repoName))
	headerSpacer := strings.Repeat(" ",
		max(0, s.width-lipgloss.Width(logo)-lipgloss.Width(repoInfo)-4),
	)
	header := lipgloss.JoinHorizontal(lipgloss.Center, logo, headerSpacer, repoInfo)

	// ── Sidebar (list panel) ──────────────────────────────────────────────────
	var sidePanel lipgloss.Style
	if s.focusPanel == 0 {
		sidePanel = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(borderFocus).
			Padding(0, 1).
			Width(sideW).
			Height(contH)
	} else {
		sidePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(border).
			Padding(0, 1).
			Width(sideW).
			Height(contH)
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(primary)
	badgeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(secondary).
		Padding(0, 1)

	sideTitle := lipgloss.JoinHorizontal(lipgloss.Center,
		titleStyle.Render("Summaries"),
		" ",
		badgeStyle.Render(fmt.Sprintf("%d", len(s.files))),
	)

	var listLines strings.Builder
	listLines.WriteString(sideTitle + "\n\n")

	if len(s.files) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(textMuted)
		listLines.WriteString(emptyStyle.Render("No summaries yet.\n\nPress 'n' to generate one."))
	} else {
		innerW := sideW - 4 // account for border + padding
		for i, f := range s.files {
			modTime := f.ModTime.Format("2006-01-02  15:04")
			name := f.Name
			if len(name) > innerW-18 {
				name = name[:innerW-19] + "…"
			}

			if i == s.cursor {
				selStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FFFFFF")).
					Background(primary).
					Bold(true)
				listLines.WriteString(selStyle.Render(fmt.Sprintf("▸ %-*s  %s", innerW-18, name, modTime)) + "\n")
			} else {
				nameStyle := lipgloss.NewStyle().Foreground(textPrimary)
				timeStyle := lipgloss.NewStyle().Foreground(textMuted)
				listLines.WriteString(
					nameStyle.Render(fmt.Sprintf("  %-*s", innerW-18, name)) +
						timeStyle.Render("  "+modTime) + "\n",
				)
			}
		}
	}

	sidebar := sidePanel.Render(listLines.String())

	// ── Preview panel ─────────────────────────────────────────────────────────
	var prevPanel lipgloss.Style
	if s.focusPanel == 1 {
		prevPanel = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(borderFocus).
			Padding(0, 1).
			Width(prevW).
			Height(contH)
	} else {
		prevPanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(border).
			Padding(0, 1).
			Width(prevW).
			Height(contH)
	}

	var previewContent string
	if len(s.files) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(textMuted).
			Align(lipgloss.Center).
			Width(prevW - 4)
		previewContent = titleStyle.Render("Preview") + "\n\n" +
			emptyStyle.Render("Select a summary to preview it here.")
	} else if s.previewLoading {
		loadStyle := lipgloss.NewStyle().Foreground(textSecondary)
		previewContent = titleStyle.Render("Preview") + "\n\n" +
			loadStyle.Render("  Loading…")
	} else if s.previewReady {
		// Metadata from filename
		f := s.files[s.cursor]
		isEnriched := strings.Contains(f.Name, "_enriched")
		var modeBadge string
		if isEnriched {
			modeBadge = " " + lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(colorSuccess).
				Padding(0, 1).
				Render("AI")
		} else {
			modeBadge = " " + lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(secondary).
				Padding(0, 1).
				Render("Simple")
		}

		metaStyle := lipgloss.NewStyle().Foreground(textSecondary)
		meta := metaStyle.Render(f.ModTime.Format("2006-01-02  15:04")) + modeBadge

		divider := lipgloss.NewStyle().
			Foreground(border).
			Render(strings.Repeat("─", prevW-6))

		scrollPct := ""
		if s.preview.TotalLineCount() > s.preview.Height {
			scrollPct = lipgloss.NewStyle().Foreground(textMuted).
				Render(fmt.Sprintf("  %.0f%%", s.preview.ScrollPercent()*100))
		}

		header2 := lipgloss.JoinHorizontal(lipgloss.Center,
			titleStyle.Render("Preview"),
			scrollPct,
		)

		previewContent = header2 + "\n" + meta + "\n" + divider + "\n" + s.preview.View()
	}

	previewPane := prevPanel.Render(previewContent)

	// ── Join panels horizontally ──────────────────────────────────────────────
	content := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, previewPane)

	// ── Footer ────────────────────────────────────────────────────────────────
	helpStyle := lipgloss.NewStyle().Foreground(textMuted)
	var footer string
	if s.copyMsg != "" {
		colorSuccess := lipgloss.AdaptiveColor{Light: "#10B981", Dark: "#34D399"}
		colorError := lipgloss.AdaptiveColor{Light: "#EF4444", Dark: "#F87171"}
		hints := helpStyle.Render("↑↓ navigate · tab/enter switch panel · c copy · n new · esc back · q quit  ")
		var status string
		if s.copyOK {
			status = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true).Render("✓ " + s.copyMsg)
		} else {
			status = lipgloss.NewStyle().Foreground(colorError).Bold(true).Render("✗ " + s.copyMsg)
		}
		footer = hints + status
	} else {
		hints := []string{"↑↓ navigate", "tab/enter switch panel", "c copy", "n new", "esc back", "q quit"}
		footer = helpStyle.Render(strings.Join(hints, " · "))
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		content,
		footer,
	)
}

// viewCompact is a fallback before WindowSizeMsg arrives.
func (s SummaryList) viewCompact() string {
	primary := lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}
	textMuted := lipgloss.AdaptiveColor{Light: "#94A3B8", Dark: "#6C7086"}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(primary).Render(" git-resume "))
	sb.WriteString("\n\n")

	if len(s.files) == 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(textMuted).Render("  No summaries found."))
		sb.WriteString("\n\n")
		sb.WriteString(lipgloss.NewStyle().Foreground(textMuted).Render("  esc back"))
		return sb.String()
	}

	for i, f := range s.files {
		if i == s.cursor {
			sb.WriteString(lipgloss.NewStyle().Foreground(primary).Bold(true).Render("  ▸ " + f.Name))
		} else {
			sb.WriteString("    " + f.Name)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(textMuted).Render("  ↑↓ navigate  ·  enter view  ·  esc back"))
	sb.WriteString("\n")
	return sb.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// OpenFileMsg is sent when the user wants to open a file in the viewer.
type OpenFileMsg struct {
	Path     string
	RepoName string
}
