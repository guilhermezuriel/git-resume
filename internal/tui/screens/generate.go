package screens

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	internalai "github.com/guilhermezuriel/git-resume/internal/ai"
	"github.com/guilhermezuriel/git-resume/internal/claude"
	"github.com/guilhermezuriel/git-resume/internal/config"
	gitpkg "github.com/guilhermezuriel/git-resume/internal/git"
	"github.com/guilhermezuriel/git-resume/internal/report"
	"github.com/guilhermezuriel/git-resume/internal/storage"
	"github.com/guilhermezuriel/git-resume/internal/tui/components"
)

var (
	// Adaptive palette for the generate wizard
	genPrimary   = lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}
	genSuccess   = lipgloss.AdaptiveColor{Light: "#10B981", Dark: "#34D399"}
	genErrColor  = lipgloss.AdaptiveColor{Light: "#EF4444", Dark: "#F87171"}
	genText      = lipgloss.AdaptiveColor{Light: "#1E293B", Dark: "#CDD6F4"}
	genMuted     = lipgloss.AdaptiveColor{Light: "#94A3B8", Dark: "#6C7086"}
	genSecondary = lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#A6ADC8"}

	genCursorStyle   = lipgloss.NewStyle().Foreground(genPrimary).Bold(true)
	genSelectedStyle = lipgloss.NewStyle().Foreground(genPrimary).Bold(true)
	genNormalStyle   = lipgloss.NewStyle().Foreground(genText)
	genDimStyle      = lipgloss.NewStyle().Foreground(genMuted)
	genSuccessStyle  = lipgloss.NewStyle().Foreground(genSuccess).Bold(true)
	genErrorStyle    = lipgloss.NewStyle().Foreground(genErrColor).Bold(true)
	genLabelStyle    = lipgloss.NewStyle().Foreground(genSecondary)
	genValueStyle    = lipgloss.NewStyle().Foreground(genText)
	genSectionStyle  = lipgloss.NewStyle().Foreground(genPrimary).Bold(true)
	genActiveStep    = lipgloss.NewStyle().Foreground(genPrimary).Bold(true)
	genDoneStep      = lipgloss.NewStyle().Foreground(genSuccess)
	genFutureStep    = lipgloss.NewStyle().Foreground(genMuted)
)

type genStep int

const (
	stepDate genStep = iota
	stepAuthor
	stepBranch
	stepMode
	stepModel // only visited when Enrich = true
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
	branchOptions []string
	modeOptions   []string
	modelOptions  []string // labels shown in the TUI
	modelIDs      []string // parallel slice with the raw IDs passed to NewModel
	langOptions   []string

	// Cursors per step (indices match step iota values 0-5).
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
	sp.Style = lipgloss.NewStyle().Foreground(genPrimary)

	authorLabel := "My commits only"
	if gitName != "" {
		authorLabel = fmt.Sprintf("My commits only (%s)", gitName)
	}

	// Build model list: Claude CLI models first (no API key needed), then OpenAI API models.
	var modelLabels []string
	var modelIDs []string
	for _, m := range claude.ListCLIModels() {
		modelLabels = append(modelLabels, m.Label)
		modelIDs = append(modelIDs, m.ID)
	}
	for _, m := range internalai.ListModels() {
		modelLabels = append(modelLabels, m.Label)
		modelIDs = append(modelIDs, m.ID)
	}

	return GenerateFlow{
		repoInfo:          repoInfo,
		step:              stepDate,
		dateOptions:       []string{fmt.Sprintf("Today (%s)", today), fmt.Sprintf("Yesterday (%s)", yesterday), "Custom date"},
		authorOptions:     []string{"All commits", authorLabel, "Specific author"},
		branchOptions:     []string{"Current branch only", "All branches"},
		modeOptions:       []string{"Simple (fast)", "AI-enriched"},
		modelOptions:      modelLabels,
		modelIDs:          modelIDs,
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
				// Skip stepModel when going back from stepLang if mode is simple.
				if g.step == stepModel && !g.cfg.Enrich {
					g.step--
				}
				return g, nil
			}

		case "up", "k":
			if g.step < stepConfirm {
				idx := cursorIdx(g.step)
				if g.cursors[idx] > 0 {
					g.cursors[idx]--
				}
			}

		case "down", "j":
			if g.step < stepConfirm {
				idx := cursorIdx(g.step)
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
		cur := g.cursors[cursorIdx(stepDate)]
		switch cur {
		case 0:
			d := time.Now().Format("2006-01-02")
			g.cfg.DateFrom = d
			g.cfg.DateTo = d
			g.step = stepAuthor
		case 1:
			d := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			g.cfg.DateFrom = d
			g.cfg.DateTo = d
			g.step = stepAuthor
		case 2:
			g.enteringCustomDate = true
			g.customDateInput.SetValue("")
			return g, g.customDateInput.Focus()
		}

	case stepAuthor:
		cur := g.cursors[cursorIdx(stepAuthor)]
		switch cur {
		case 0:
			g.cfg.Author = ""
			g.step = stepBranch
		case 1:
			name, _ := gitpkg.GetHostAuthor()
			g.cfg.Author = name
			g.step = stepBranch
		case 2:
			g.enteringCustomAuthor = true
			g.customAuthorInput.SetValue("")
			return g, g.customAuthorInput.Focus()
		}

	case stepBranch:
		cur := g.cursors[cursorIdx(stepBranch)]
		g.cfg.AllBranches = cur == 1
		g.step = stepMode

	case stepMode:
		cur := g.cursors[cursorIdx(stepMode)]
		if cur == 0 {
			g.cfg.Enrich = false
			g.cfg.Model = ""
			g.step = stepConfirm
		} else {
			g.cfg.Enrich = true
			g.step = stepModel
		}

	case stepModel:
		cur := g.cursors[cursorIdx(stepModel)]
		g.cfg.Model = g.modelIDs[cur]
		g.step = stepLang

	case stepLang:
		cur := g.cursors[cursorIdx(stepLang)]
		if cur == len(g.langOptions)-1 {
			g.enteringCustomLang = true
			g.customLangInput.SetValue("")
			return g, g.customLangInput.Focus()
		}
		g.cfg.LangCode = strings.SplitN(g.langOptions[cur], " ", 2)[0]
		g.step = stepConfirm

	case stepConfirm:
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
		if val == "" {
			val = time.Now().Format("2006-01-02")
		}
		g.cfg.DateFrom = val
		g.cfg.DateTo = val
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
		g.step = stepBranch
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

// cursorIdx maps a step to its slot in the cursors array.
// stepModel sits at index 3; stepLang shifts to 4.
func cursorIdx(s genStep) int {
	return int(s) // iota values already match 0-4 for Date..Lang
}

func (g GenerateFlow) optionsForStep(s genStep) []string {
	switch s {
	case stepDate:
		return g.dateOptions
	case stepAuthor:
		return g.authorOptions
	case stepBranch:
		return g.branchOptions
	case stepMode:
		return g.modeOptions
	case stepModel:
		return g.modelOptions
	case stepLang:
		return g.langOptions
	}
	return nil
}

func (g GenerateFlow) View() string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(components.Header("git-resume", g.repoInfo.Name))
	sb.WriteString("\n\n")

	switch g.step {
	case stepDate:
		sb.WriteString(renderStepProgress(g.cfg.Enrich, stepDate) + "\n")
		sb.WriteString(genSectionStyle.Render("  Select Date") + "\n\n")
		if g.enteringCustomDate {
			sb.WriteString(genDimStyle.Render("  Date: "))
			sb.WriteString(g.customDateInput.View())
			sb.WriteString("\n\n")
			sb.WriteString(genDimStyle.Render("  enter confirm  ·  esc cancel"))
		} else {
			sb.WriteString(renderOptions(g.dateOptions, g.cursors[cursorIdx(stepDate)]))
			sb.WriteString(genDimStyle.Render("\n  ↑↓ move  ·  enter select  ·  esc back"))
		}

	case stepAuthor:
		sb.WriteString(renderStepProgress(g.cfg.Enrich, stepAuthor) + "\n")
		sb.WriteString(genSectionStyle.Render("  Author Filter") + "\n\n")
		if g.enteringCustomAuthor {
			sb.WriteString(genDimStyle.Render("  Author: "))
			sb.WriteString(g.customAuthorInput.View())
			sb.WriteString("\n\n")
			sb.WriteString(genDimStyle.Render("  enter confirm  ·  esc cancel"))
		} else {
			sb.WriteString(renderOptions(g.authorOptions, g.cursors[cursorIdx(stepAuthor)]))
			sb.WriteString(genDimStyle.Render("\n  ↑↓ move  ·  enter select  ·  esc back"))
		}

	case stepBranch:
		sb.WriteString(renderStepProgress(g.cfg.Enrich, stepBranch) + "\n")
		sb.WriteString(genSectionStyle.Render("  Branch Scope") + "\n\n")
		sb.WriteString(renderOptions(g.branchOptions, g.cursors[cursorIdx(stepBranch)]))
		sb.WriteString(genDimStyle.Render("\n  ↑↓ move  ·  enter select  ·  esc back"))

	case stepMode:
		sb.WriteString(renderStepProgress(g.cfg.Enrich, stepMode) + "\n")
		sb.WriteString(genSectionStyle.Render("  Summary Mode") + "\n\n")
		sb.WriteString(renderOptions(g.modeOptions, g.cursors[cursorIdx(stepMode)]))
		sb.WriteString(genDimStyle.Render("\n  ↑↓ move  ·  enter select  ·  esc back"))

	case stepModel:
		sb.WriteString(renderStepProgress(g.cfg.Enrich, stepModel) + "\n")
		sb.WriteString(genSectionStyle.Render("  Select Model") + "\n\n")
		sb.WriteString(renderOptions(g.modelOptions, g.cursors[cursorIdx(stepModel)]))
		sb.WriteString(genDimStyle.Render("\n  ↑↓ move  ·  enter select  ·  esc back"))

	case stepLang:
		sb.WriteString(renderStepProgress(g.cfg.Enrich, stepLang) + "\n")
		sb.WriteString(genSectionStyle.Render("  Output Language") + "\n\n")
		if g.enteringCustomLang {
			sb.WriteString(genDimStyle.Render("  Language code: "))
			sb.WriteString(g.customLangInput.View())
			sb.WriteString("\n\n")
			sb.WriteString(genDimStyle.Render("  enter confirm  ·  esc cancel"))
		} else {
			sb.WriteString(renderOptions(g.langOptions, g.cursors[cursorIdx(stepLang)]))
			sb.WriteString(genDimStyle.Render("\n  ↑↓ move  ·  enter select  ·  esc back"))
		}

	case stepConfirm:
		sb.WriteString(genSectionStyle.Render("  Confirm Configuration") + "\n\n")

		modeLabel := "Simple"
		if g.cfg.Enrich {
			modelLabel := g.cfg.Model
			for i, id := range g.modelIDs {
				if id == g.cfg.Model {
					modelLabel = g.modelOptions[i]
					break
				}
			}
			modeLabel = "AI-enriched · " + modelLabel
		}
		authorLabel := "All commits"
		if g.cfg.Author != "" {
			authorLabel = g.cfg.Author
		}

		branchScopeLabel := "Current branch"
		if g.cfg.AllBranches {
			branchScopeLabel = "All branches"
		}

		sb.WriteString(genDimStyle.Render("  " + strings.Repeat("─", 32)) + "\n")
		sb.WriteString(fmt.Sprintf("  %-10s %s\n", genLabelStyle.Render("Date"), genValueStyle.Render(g.cfg.DateFrom)))
		sb.WriteString(fmt.Sprintf("  %-10s %s\n", genLabelStyle.Render("Author"), genValueStyle.Render(authorLabel)))
		sb.WriteString(fmt.Sprintf("  %-10s %s\n", genLabelStyle.Render("Branches"), genValueStyle.Render(branchScopeLabel)))
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

// renderStepProgress renders a progress bar that adapts to the current flow.
// When enriched is true the bar shows Model; otherwise it omits it.
func renderStepProgress(enriched bool, current genStep) string {
	type stepLabel struct {
		s genStep
		l string
	}
	all := []stepLabel{
		{stepDate, "Date"},
		{stepAuthor, "Author"},
		{stepBranch, "Branch"},
		{stepMode, "Mode"},
		{stepModel, "Model"},
		{stepLang, "Lang"},
	}

	var parts []string
	for _, sl := range all {
		// Hide the Model step from the bar when in Simple mode.
		if sl.s == stepModel && !enriched && current != stepModel {
			continue
		}
		var label string
		switch {
		case sl.s < current:
			label = genDoneStep.Render(sl.l)
		case sl.s == current:
			label = genActiveStep.Render(sl.l)
		default:
			label = genFutureStep.Render(sl.l)
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
		from, err := time.Parse("2006-01-02", cfg.DateFrom)
		if err != nil {
			return GenerationDoneMsg{Err: fmt.Errorf("invalid date: %w", err)}
		}
		to, err := time.Parse("2006-01-02", cfg.DateTo)
		if err != nil {
			return GenerationDoneMsg{Err: fmt.Errorf("invalid date: %w", err)}
		}

		dr := gitpkg.DateRange{From: from, To: to}

		repoDir, err := storage.InitRepo(info.ID, info.Name, info.Path, info.Remote)
		if err != nil {
			return GenerationDoneMsg{Err: err}
		}

		period := cfg.DateFrom
		if cfg.DateFrom != cfg.DateTo {
			period = cfg.DateFrom + " to " + cfg.DateTo
		}

		var content string

		if cfg.AllBranches {
			_ = gitpkg.FetchAll() // best effort; ignore fetch errors
			branchCommits, err := gitpkg.GetCommitsAllBranches(dr, cfg.Author)
			if err != nil {
				return GenerationDoneMsg{Err: err}
			}
			if !cfg.Enrich {
				content = report.GenerateMultiBranch(branchCommits, info.Name, period, cfg.Author)
			} else {
				content = tuiEnrichAllBranchesPerBranch(cfg, branchCommits, info.Name, period)
			}
		} else {
			commits, err := gitpkg.GetCommitsRange(dr, cfg.Author)
			if err != nil {
				return GenerationDoneMsg{Err: err}
			}
			if !cfg.Enrich {
				content = report.GenerateSimple(commits, info.Name, period, cfg.Author)
			} else if len(commits) == 0 {
				meta := report.EnrichedMeta{
					ClaudeResponse: "No activity recorded for this date.",
					LangCode:       cfg.LangCode,
					ModelID:        cfg.Model,
				}
				content = report.GenerateEnriched(meta, info.Name, period, cfg.Author)
			} else if claude.IsCLIModel(cfg.Model) {
				content = tuiEnrichWithClaude(cfg, commits, info.Name, period)
			} else {
				var genErr error
				content, genErr = tuiEnrichWithAPI(cfg, commits, info.Name, period)
				if genErr != nil {
					return GenerationDoneMsg{Err: genErr}
				}
			}
		}

		path, err := storage.WriteReport(repoDir, cfg.DateFrom, cfg.Author, cfg.Enrich, content)
		if err != nil {
			return GenerationDoneMsg{Err: err}
		}
		return GenerationDoneMsg{Content: content, Path: path}
	}
}

func tuiEnrichWithClaude(cfg config.RunConfig, commits []gitpkg.Commit, repoName, period string) string {
	if !claude.CheckClaude() {
		return report.GenerateSimple(commits, repoName, period, cfg.Author)
	}
	var msgs []string
	for _, c := range commits {
		msgs = append(msgs, c.Message)
	}
	prompt := claude.BuildPrompt(msgs, cfg.LangCode)
	start := time.Now()
	resp, err := claude.RunClaude(prompt, cfg.Model)
	elapsed := int(time.Since(start).Seconds())
	if err != nil {
		return report.GenerateSimple(commits, repoName, period, cfg.Author)
	}
	meta := report.EnrichedMeta{
		ClaudeResponse: resp,
		LangCode:       cfg.LangCode,
		CommitCount:    len(commits),
		ProcessingTime: elapsed,
		InputTokens:    claude.EstimateTokens(prompt),
		OutputTokens:   claude.EstimateTokens(resp),
		ModelID:        cfg.Model,
	}
	return report.GenerateEnriched(meta, repoName, period, cfg.Author)
}

func tuiEnrichWithAPI(cfg config.RunConfig, commits []gitpkg.Commit, repoName, period string) (string, error) {
	model, err := internalai.NewModel(cfg.Model)
	if err != nil {
		return "", err
	}
	var msgs []string
	for _, c := range commits {
		msgs = append(msgs, c.Message)
	}
	prompt := claude.BuildPrompt(msgs, cfg.LangCode)
	messages := []internalai.Message{{Role: "user", Content: prompt}}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	start := time.Now()
	result, err := internalai.ChatStream(ctx, model, messages, nil)
	elapsed := int(time.Since(start).Seconds())
	if err != nil {
		return "", err
	}
	meta := report.EnrichedMeta{
		ClaudeResponse: result.FullText,
		LangCode:       cfg.LangCode,
		CommitCount:    len(commits),
		ProcessingTime: elapsed,
		InputTokens:    result.InputTokens,
		OutputTokens:   result.OutputTokens,
		ModelID:        cfg.Model,
	}
	return report.GenerateEnriched(meta, repoName, period, cfg.Author), nil
}

func tuiEnrichAllBranchesPerBranch(cfg config.RunConfig, branchCommits []gitpkg.BranchCommits, repoName, period string) string {
	var results []report.EnrichedBranchResult

	for _, bc := range branchCommits {
		var msgs []string
		for _, c := range bc.Commits {
			msgs = append(msgs, c.Message)
		}
		prompt := claude.BuildPrompt(msgs, cfg.LangCode)

		var aiResp string
		var inputTokens, outputTokens, elapsed int

		if claude.IsCLIModel(cfg.Model) {
			if !claude.CheckClaude() {
				aiResp = "(Claude CLI not available)"
			} else {
				start := time.Now()
				resp, err := claude.RunClaude(prompt, cfg.Model)
				elapsed = int(time.Since(start).Seconds())
				if err != nil {
					aiResp = fmt.Sprintf("(error: %v)", err)
				} else {
					aiResp = resp
					inputTokens = claude.EstimateTokens(prompt)
					outputTokens = claude.EstimateTokens(resp)
				}
			}
		} else {
			model, err := internalai.NewModel(cfg.Model)
			if err != nil {
				aiResp = fmt.Sprintf("(model init error: %v)", err)
			} else {
				messages := []internalai.Message{{Role: "user", Content: prompt}}
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				start := time.Now()
				result, err := internalai.ChatStream(ctx, model, messages, nil)
				elapsed = int(time.Since(start).Seconds())
				cancel()
				if err != nil {
					aiResp = fmt.Sprintf("(api error: %v)", err)
				} else {
					aiResp = result.FullText
					inputTokens = result.InputTokens
					outputTokens = result.OutputTokens
				}
			}
		}

		results = append(results, report.EnrichedBranchResult{
			BranchName:     bc.Branch,
			AIResponse:     aiResp,
			CommitCount:    len(bc.Commits),
			ProcessingTime: elapsed,
			InputTokens:    inputTokens,
			OutputTokens:   outputTokens,
		})
	}

	return report.GenerateEnrichedMultiBranch(results, cfg.Model, repoName, period, cfg.Author, cfg.LangCode)
}

// GenerationDoneMsg is sent when the generation goroutine completes.
type GenerationDoneMsg struct {
	Content string
	Path    string
	Err     error
}
