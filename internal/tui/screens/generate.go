package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guilhermezuriel/git-resume/internal/claude"
	"github.com/guilhermezuriel/git-resume/internal/config"
	gitpkg "github.com/guilhermezuriel/git-resume/internal/git"
	"github.com/guilhermezuriel/git-resume/internal/report"
	"github.com/guilhermezuriel/git-resume/internal/storage"
	"github.com/guilhermezuriel/git-resume/internal/tui/components"
)

var (
	genCursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	genSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	genNormalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	genDimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	genSuccessStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	genErrorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	genLabelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	genValueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	genSectionStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	genActiveStep    = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	genDoneStep      = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	genFutureStep    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type genStep int

const (
	stepDate genStep = iota
	stepAuthor
	stepMode
	stepLang
	stepConfirm
	stepProcessing
	stepResult
)

// GenerateFlow manages the multi-step generation wizard.
type GenerateFlow struct {
	repoInfo *gitpkg.RepoInfo
	step     genStep

	// Step options
	dateOptions   []string
	authorOptions []string
	modeOptions   []string
	langOptions   []string

	// Cursors per step
	cursors [6]int

	// Custom input fields
	customDateInput   textinput.Model
	customAuthorInput textinput.Model
	customLangInput   textinput.Model

	// Collected config
	cfg config.RunConfig

	// Processing
	spinner spinner.Model

	// Result
	resultContent string
	resultPath    string
	resultErr     error

	// Custom input mode flags
	enteringCustomDate   bool
	enteringCustomAuthor bool
	enteringCustomLang   bool
}

func NewGenerateFlow(repoInfo *gitpkg.RepoInfo) GenerateFlow {
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	gitName := ""
	if name, err := gitpkg.GetHostAuthor(); err == nil {
		gitName = name
	}

	ti := textinput.New()
	ti.Placeholder = "YYYY-MM-DD"
	ti.CharLimit = 10

	tai := textinput.New()
	tai.Placeholder = "Author name or email"
	tai.CharLimit = 100

	tli := textinput.New()
	tli.Placeholder = "Language code (e.g., ja, ko, zh)"
	tli.CharLimit = 20

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	authorLabel := "My commits only"
	if gitName != "" {
		authorLabel = fmt.Sprintf("My commits only (%s)", gitName)
	}

	return GenerateFlow{
		repoInfo:          repoInfo,
		step:              stepDate,
		dateOptions:       []string{fmt.Sprintf("Today (%s)", today), fmt.Sprintf("Yesterday (%s)", yesterday), "Custom date"},
		authorOptions:     []string{"All commits", authorLabel, "Specific author"},
		modeOptions:       []string{"Simple (fast)", "AI-enriched (requires Claude CLI)"},
		langOptions:       []string{"pt (Portuguese)", "en (English)", "es (Spanish)", "fr (French)", "de (German)", "Other"},
		customDateInput:   ti,
		customAuthorInput: tai,
		customLangInput:   tli,
		spinner:           sp,
	}
}

func (g GenerateFlow) Init() tea.Cmd {
	return nil
}

func (g GenerateFlow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if g.step == stepProcessing {
			var cmd tea.Cmd
			g.spinner, cmd = g.spinner.Update(msg)
			return g, cmd
		}

	case GenerationDoneMsg:
		g.resultContent = msg.Content
		g.resultPath = msg.Path
		g.resultErr = msg.Err
		g.step = stepResult
		return g, nil

	case tea.KeyMsg:
		// Handle custom input modes first
		if g.enteringCustomDate {
			return g.handleCustomDateInput(msg)
		}
		if g.enteringCustomAuthor {
			return g.handleCustomAuthorInput(msg)
		}
		if g.enteringCustomLang {
			return g.handleCustomLangInput(msg)
		}

		switch msg.String() {
		case "esc":
			if g.step == stepDate || g.step == stepResult {
				return g, func() tea.Msg { return NavigateMsg{To: ScreenMenu} }
			}
			if g.step > stepDate && g.step < stepProcessing {
				g.step--
				return g, nil
			}

		case "up", "k":
			if g.step < stepConfirm {
				idx := int(g.step)
				if g.cursors[idx] > 0 {
					g.cursors[idx]--
				}
			}

		case "down", "j":
			if g.step < stepConfirm {
				idx := int(g.step)
				opts := g.optionsForStep(g.step)
				if g.cursors[idx] < len(opts)-1 {
					g.cursors[idx]++
				}
			}

		case "enter", " ":
			return g.handleEnter()

		case "q", "ctrl+c":
			if g.step != stepProcessing {
				return g, tea.Quit
			}
		}
	}

	// Pass input updates to active text inputs
	if g.enteringCustomDate {
		var cmd tea.Cmd
		g.customDateInput, cmd = g.customDateInput.Update(msg)
		return g, cmd
	}
	if g.enteringCustomAuthor {
		var cmd tea.Cmd
		g.customAuthorInput, cmd = g.customAuthorInput.Update(msg)
		return g, cmd
	}
	if g.enteringCustomLang {
		var cmd tea.Cmd
		g.customLangInput, cmd = g.customLangInput.Update(msg)
		return g, cmd
	}

	return g, nil
}

func (g GenerateFlow) handleEnter() (tea.Model, tea.Cmd) {
	switch g.step {
	case stepDate:
		cur := g.cursors[0]
		switch cur {
		case 0:
			g.cfg.Date = time.Now().Format("2006-01-02")
			g.step = stepAuthor
		case 1:
			g.cfg.Date = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			g.step = stepAuthor
		case 2:
			g.enteringCustomDate = true
			g.customDateInput.SetValue("")
			cmd := g.customDateInput.Focus()
			return g, cmd
		}

	case stepAuthor:
		cur := g.cursors[1]
		switch cur {
		case 0:
			g.cfg.Author = ""
			g.step = stepMode
		case 1:
			name, _ := gitpkg.GetHostAuthor()
			g.cfg.Author = name
			g.step = stepMode
		case 2:
			g.enteringCustomAuthor = true
			g.customAuthorInput.SetValue("")
			cmd := g.customAuthorInput.Focus()
			return g, cmd
		}

	case stepMode:
		cur := g.cursors[2]
		if cur == 0 {
			g.cfg.Enrich = false
			g.step = stepConfirm
		} else {
			g.cfg.Enrich = true
			g.step = stepLang
		}

	case stepLang:
		cur := g.cursors[3]
		if cur == len(g.langOptions)-1 {
			// "Other"
			g.enteringCustomLang = true
			g.customLangInput.SetValue("")
			cmd := g.customLangInput.Focus()
			return g, cmd
		}
		// Extract code from "xx (Language)" format
		g.cfg.LangCode = strings.SplitN(g.langOptions[cur], " ", 2)[0]
		g.step = stepConfirm

	case stepConfirm:
		// Start generation
		g.step = stepProcessing
		cfg := g.cfg
		info := g.repoInfo
		return g, tea.Batch(
			g.spinner.Tick,
			runGeneration(cfg, info),
		)

	case stepResult:
		return g, func() tea.Msg { return NavigateMsg{To: ScreenMenu} }
	}

	return g, nil
}

func (g GenerateFlow) handleCustomDateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		val := strings.TrimSpace(g.customDateInput.Value())
		if val != "" {
			g.cfg.Date = val
		} else {
			g.cfg.Date = time.Now().Format("2006-01-02")
		}
		g.enteringCustomDate = false
		g.customDateInput.Blur()
		g.step = stepAuthor
		return g, nil
	case "esc":
		g.enteringCustomDate = false
		g.customDateInput.Blur()
		return g, nil
	}
	var cmd tea.Cmd
	g.customDateInput, cmd = g.customDateInput.Update(msg)
	return g, cmd
}

func (g GenerateFlow) handleCustomAuthorInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		g.cfg.Author = strings.TrimSpace(g.customAuthorInput.Value())
		g.enteringCustomAuthor = false
		g.customAuthorInput.Blur()
		g.step = stepMode
		return g, nil
	case "esc":
		g.enteringCustomAuthor = false
		g.customAuthorInput.Blur()
		return g, nil
	}
	var cmd tea.Cmd
	g.customAuthorInput, cmd = g.customAuthorInput.Update(msg)
	return g, cmd
}

func (g GenerateFlow) handleCustomLangInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		g.cfg.LangCode = strings.TrimSpace(g.customLangInput.Value())
		g.enteringCustomLang = false
		g.customLangInput.Blur()
		g.step = stepConfirm
		return g, nil
	case "esc":
		g.enteringCustomLang = false
		g.customLangInput.Blur()
		return g, nil
	}
	var cmd tea.Cmd
	g.customLangInput, cmd = g.customLangInput.Update(msg)
	return g, cmd
}

func (g GenerateFlow) optionsForStep(s genStep) []string {
	switch s {
	case stepDate:
		return g.dateOptions
	case stepAuthor:
		return g.authorOptions
	case stepMode:
		return g.modeOptions
	case stepLang:
		return g.langOptions
	}
	return nil
}

func (g GenerateFlow) View() string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(components.Header("git-resume v3.0.0", g.repoInfo.Name))
	sb.WriteString("\n\n")

	switch g.step {
	case stepDate:
		sb.WriteString(renderStepProgress(stepDate) + "\n")
		sb.WriteString(genSectionStyle.Render("  Select Date") + "\n\n")
		if g.enteringCustomDate {
			sb.WriteString(genDimStyle.Render("  Date: "))
			sb.WriteString(g.customDateInput.View())
			sb.WriteString("\n\n")
			sb.WriteString(genDimStyle.Render("  enter confirm  ·  esc cancel"))
		} else {
			sb.WriteString(renderOptions(g.dateOptions, g.cursors[0]))
			sb.WriteString(genDimStyle.Render("\n  ↑↓ move  ·  enter select  ·  esc back"))
		}

	case stepAuthor:
		sb.WriteString(renderStepProgress(stepAuthor) + "\n")
		sb.WriteString(genSectionStyle.Render("  Author Filter") + "\n\n")
		if g.enteringCustomAuthor {
			sb.WriteString(genDimStyle.Render("  Author: "))
			sb.WriteString(g.customAuthorInput.View())
			sb.WriteString("\n\n")
			sb.WriteString(genDimStyle.Render("  enter confirm  ·  esc cancel"))
		} else {
			sb.WriteString(renderOptions(g.authorOptions, g.cursors[1]))
			sb.WriteString(genDimStyle.Render("\n  ↑↓ move  ·  enter select  ·  esc back"))
		}

	case stepMode:
		sb.WriteString(renderStepProgress(stepMode) + "\n")
		sb.WriteString(genSectionStyle.Render("  Summary Mode") + "\n\n")
		sb.WriteString(renderOptions(g.modeOptions, g.cursors[2]))
		sb.WriteString(genDimStyle.Render("\n  ↑↓ move  ·  enter select  ·  esc back"))

	case stepLang:
		sb.WriteString(renderStepProgress(stepLang) + "\n")
		sb.WriteString(genSectionStyle.Render("  Output Language") + "\n\n")
		if g.enteringCustomLang {
			sb.WriteString(genDimStyle.Render("  Language code: "))
			sb.WriteString(g.customLangInput.View())
			sb.WriteString("\n\n")
			sb.WriteString(genDimStyle.Render("  enter confirm  ·  esc cancel"))
		} else {
			sb.WriteString(renderOptions(g.langOptions, g.cursors[3]))
			sb.WriteString(genDimStyle.Render("\n  ↑↓ move  ·  enter select  ·  esc back"))
		}

	case stepConfirm:
		sb.WriteString(genSectionStyle.Render("  Confirm Configuration") + "\n\n")

		modeLabel := "Simple"
		if g.cfg.Enrich {
			modeLabel = "AI-enriched (Claude)"
		}
		authorLabel := "All commits"
		if g.cfg.Author != "" {
			authorLabel = g.cfg.Author
		}

		sb.WriteString(genDimStyle.Render("  " + strings.Repeat("─", 32)) + "\n")
		sb.WriteString(fmt.Sprintf("  %-10s %s\n", genLabelStyle.Render("Date"), genValueStyle.Render(g.cfg.Date)))
		sb.WriteString(fmt.Sprintf("  %-10s %s\n", genLabelStyle.Render("Author"), genValueStyle.Render(authorLabel)))
		sb.WriteString(fmt.Sprintf("  %-10s %s\n", genLabelStyle.Render("Mode"), genValueStyle.Render(modeLabel)))
		if g.cfg.Enrich && g.cfg.LangCode != "" {
			sb.WriteString(fmt.Sprintf("  %-10s %s\n", genLabelStyle.Render("Language"), genValueStyle.Render(g.cfg.LangCode)))
		}
		sb.WriteString(genDimStyle.Render("  " + strings.Repeat("─", 32)) + "\n")
		sb.WriteString("\n")
		sb.WriteString(genDimStyle.Render("  enter generate  ·  esc back"))

	case stepProcessing:
		sb.WriteString("\n\n")
		sb.WriteString(fmt.Sprintf("  %s  Generating summary...\n\n", g.spinner.View()))
		sb.WriteString(genDimStyle.Render("  Analyzing git commits, please wait"))

	case stepResult:
		if g.resultErr != nil {
			sb.WriteString(genErrorStyle.Render("  Generation failed") + "\n\n")
			sb.WriteString(genDimStyle.Render("  "+g.resultErr.Error()) + "\n")
		} else {
			sb.WriteString(genSuccessStyle.Render("  Summary generated successfully!") + "\n\n")
			sb.WriteString(fmt.Sprintf("  %s  %s\n", genLabelStyle.Render("Saved to"), genValueStyle.Render(g.resultPath)))
		}
		sb.WriteString("\n")
		sb.WriteString(genDimStyle.Render("  press any key to return to menu"))
	}

	sb.WriteString("\n")
	return sb.String()
}

// renderStepProgress renders a minimal 1 ─ 2 ─ 3 ─ 4 progress indicator.
func renderStepProgress(current genStep) string {
	steps := []genStep{stepDate, stepAuthor, stepMode, stepLang}
	labels := []string{"Date", "Author", "Mode", "Lang"}
	var parts []string
	for i, s := range steps {
		var label string
		switch {
		case s < current:
			label = genDoneStep.Render(labels[i])
		case s == current:
			label = genActiveStep.Render(labels[i])
		default:
			label = genFutureStep.Render(labels[i])
		}
		parts = append(parts, label)
	}
	sep := genDimStyle.Render("  ─  ")
	return "  " + strings.Join(parts, sep)
}

// renderOptions renders a list with cursor indicator.
func renderOptions(opts []string, cursor int) string {
	var sb strings.Builder
	for i, opt := range opts {
		if i == cursor {
			sb.WriteString(genCursorStyle.Render("  ▸ "))
			sb.WriteString(genSelectedStyle.Render(opt))
		} else {
			sb.WriteString(genNormalStyle.Render("    " + opt))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// runGeneration is the async tea.Cmd that performs the actual work.
func runGeneration(cfg config.RunConfig, info *gitpkg.RepoInfo) tea.Cmd {
	return func() tea.Msg {
		commits, err := gitpkg.GetCommits(cfg.Date, cfg.Author)
		if err != nil {
			return GenerationDoneMsg{Err: err}
		}

		repoDir, err := storage.InitRepo(info.ID, info.Name, info.Path, info.Remote)
		if err != nil {
			return GenerationDoneMsg{Err: err}
		}

		var content string
		if cfg.Enrich {
			if !claude.CheckClaude() {
				return GenerationDoneMsg{Err: fmt.Errorf("claude CLI not found — install with: npm install -g @anthropic-ai/claude-code")}
			}
			if len(commits) == 0 {
				meta := report.EnrichedMeta{
					ClaudeResponse: "No activity recorded for this date.",
					LangCode:       cfg.LangCode,
				}
				content = report.GenerateEnriched(meta, info.Name, cfg.Date, cfg.Author)
			} else {
				var msgs []string
				for _, c := range commits {
					msgs = append(msgs, c.Message)
				}
				prompt := claude.BuildPrompt(msgs, cfg.LangCode)
				inputTokens := claude.EstimateTokens(prompt)

				start := time.Now()
				resp, err := claude.RunClaude(prompt)
				elapsed := int(time.Since(start).Seconds())
				if err != nil {
					return GenerationDoneMsg{Err: err}
				}
				outputTokens := claude.EstimateTokens(resp)

				meta := report.EnrichedMeta{
					ClaudeResponse: resp,
					LangCode:       cfg.LangCode,
					CommitCount:    len(commits),
					ProcessingTime: elapsed,
					InputTokens:    inputTokens,
					OutputTokens:   outputTokens,
				}
				content = report.GenerateEnriched(meta, info.Name, cfg.Date, cfg.Author)
			}
		} else {
			content = report.GenerateSimple(commits, info.Name, cfg.Date, cfg.Author)
		}

		path, err := storage.WriteReport(repoDir, cfg.Date, cfg.Author, cfg.Enrich, content)
		if err != nil {
			return GenerationDoneMsg{Err: err}
		}
		return GenerationDoneMsg{Content: content, Path: path}
	}
}

// GenerationDoneMsg is sent when the generation goroutine completes.
type GenerationDoneMsg struct {
	Content string
	Path    string
	Err     error
}
