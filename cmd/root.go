package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	internalai "github.com/guilhermezuriel/git-resume/internal/ai"
	"github.com/guilhermezuriel/git-resume/internal/claude"
	"github.com/guilhermezuriel/git-resume/internal/config"
	gitpkg "github.com/guilhermezuriel/git-resume/internal/git"
	"github.com/guilhermezuriel/git-resume/internal/report"
	"github.com/guilhermezuriel/git-resume/internal/storage"
	"github.com/guilhermezuriel/git-resume/internal/tui"
)

const version = "3.0.0"

var (
	flagEnrich      bool
	flagLang        string
	flagDate        string
	flagAuthor      string
	flagHost        bool
	flagVersion     bool
	flagUpdate      bool
	flagModel       string
	flagFrom        string
	flagTo          string
	flagAllBranches bool
	flagConsolidate bool
)

var rootCmd = &cobra.Command{
	Use:   "git-resume",
	Short: "Daily commit summary generator",
	Long: `git-resume — Daily commit summary generator with optional AI enrichment.

Generate concise commit summaries for standup meetings, timesheets, and changelogs.
Summaries are stored in ~/.git-resumes/<repo-id>/.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagVersion {
			fmt.Printf("git-resume v%s\n", version)
			return nil
		}
		if flagUpdate {
			return runUpdate()
		}
		return runGenerate()
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Open the interactive TUI",
	Long:  "Launch the interactive terminal UI for git-resume.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !gitpkg.IsGitRepo() {
			return fmt.Errorf("not inside a git repository")
		}
		info, err := gitpkg.GetRepoInfo()
		if err != nil {
			return err
		}
		return tui.Run(info)
	},
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVar(&flagEnrich, "enrich", false, "Generate AI-powered summary using Claude CLI")
	rootCmd.Flags().StringVar(&flagLang, "lang", "", "Output language code (requires --enrich), e.g. en, pt, es")
	rootCmd.Flags().StringVar(&flagDate, "date", "", "Target date in YYYY-MM-DD format (default: today)")
	rootCmd.Flags().StringVar(&flagAuthor, "author", "", "Filter by author name or email")
	rootCmd.Flags().BoolVar(&flagHost, "host", false, "Filter by the local git user.name")
	rootCmd.Flags().BoolVarP(&flagVersion, "version", "v", false, "Print version and exit")
	rootCmd.Flags().BoolVar(&flagUpdate, "update", false, "Update git-resume to the latest release")
	rootCmd.Flags().StringVar(&flagModel, "model", "claude", `LLM to use with --enrich. "claude" uses the local Claude CLI; use "gpt-4o", "gpt-4o-mini", or "o3-mini" for direct API calls (requires OPENAI_API_KEY)`)
	rootCmd.Flags().StringVar(&flagFrom, "from", "", "Start date YYYY-MM-DD (inclusive); use with --to")
	rootCmd.Flags().StringVar(&flagTo, "to", "", "End date YYYY-MM-DD (inclusive); requires --from")
	rootCmd.Flags().BoolVarP(&flagAllBranches, "all-branches", "a", false, "Search commits in all branches (local and remote)")
	rootCmd.Flags().BoolVar(&flagConsolidate, "consolidate", false, "With --all-branches --enrich: generate one consolidated AI summary instead of per-branch")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(historyCmd)
}

func runGenerate() error {
	if !gitpkg.IsGitRepo() {
		return fmt.Errorf("not inside a git repository")
	}

	// Validate flag combinations.
	if flagLang != "" && !flagEnrich {
		return fmt.Errorf("--lang requires --enrich flag")
	}
	if flagDate != "" && (flagFrom != "" || flagTo != "") {
		return fmt.Errorf("--date and --from/--to are mutually exclusive")
	}
	if flagTo != "" && flagFrom == "" {
		return fmt.Errorf("--to requires --from")
	}
	if flagConsolidate && (!flagAllBranches || !flagEnrich) {
		return fmt.Errorf("--consolidate requires --all-branches and --enrich")
	}

	// Resolve --host to author name.
	author := flagAuthor
	if flagHost {
		name, err := gitpkg.GetHostAuthor()
		if err != nil {
			return fmt.Errorf("%w\nConfigure with:\n  git config user.name \"Your Name\"\n  git config user.email \"your@email.com\"", err)
		}
		author = name
	}

	// Resolve date range.
	var dateFrom, dateTo string
	today := time.Now().Format("2006-01-02")
	if flagDate != "" {
		dateFrom = flagDate
		dateTo = flagDate
	} else if flagFrom != "" {
		dateFrom = flagFrom
		if flagTo != "" {
			dateTo = flagTo
		} else {
			dateTo = today
		}
	} else {
		dateFrom = today
		dateTo = today
	}

	// Validate parsed dates.
	from, err := time.Parse("2006-01-02", dateFrom)
	if err != nil {
		return fmt.Errorf("invalid --from/--date format (expected YYYY-MM-DD): %w", err)
	}
	to, err := time.Parse("2006-01-02", dateTo)
	if err != nil {
		return fmt.Errorf("invalid --to format (expected YYYY-MM-DD): %w", err)
	}
	if from.After(to) {
		return fmt.Errorf("--from date must not be after --to date")
	}

	cfg := config.RunConfig{
		DateFrom:    dateFrom,
		DateTo:      dateTo,
		Author:      author,
		Enrich:      flagEnrich,
		LangCode:    flagLang,
		Model:       flagModel,
		AllBranches: flagAllBranches,
		Consolidate: flagConsolidate,
	}

	return generate(cfg)
}

// periodLabel returns a human-readable period string.
func periodLabel(dateFrom, dateTo string) string {
	if dateFrom == dateTo {
		return dateFrom
	}
	return dateFrom + " to " + dateTo
}

// filenameDateSlug returns a slug suitable for use in filenames.
func filenameDateSlug(dateFrom, dateTo string) string {
	if dateFrom == dateTo {
		return dateFrom
	}
	return dateFrom + "_" + dateTo
}

func generate(cfg config.RunConfig) error {
	info, err := gitpkg.GetRepoInfo()
	if err != nil {
		return err
	}

	from, err := time.Parse("2006-01-02", cfg.DateFrom)
	if err != nil {
		return fmt.Errorf("invalid DateFrom: %w", err)
	}
	to, err := time.Parse("2006-01-02", cfg.DateTo)
	if err != nil {
		return fmt.Errorf("invalid DateTo: %w", err)
	}

	period := periodLabel(cfg.DateFrom, cfg.DateTo)
	dr := gitpkg.DateRange{From: from, To: to}

	repoDir, err := storage.InitRepo(info.ID, info.Name, info.Path, info.Remote)
	if err != nil {
		return err
	}

	var content string

	if cfg.AllBranches {
		if err := gitpkg.FetchAll(); err != nil {
			fmt.Fprintf(os.Stderr, "  [WARN] git fetch --all failed: %v\n", err)
		}
		branchCommits, err := gitpkg.GetCommitsAllBranches(dr, cfg.Author)
		if err != nil {
			return err
		}

		if !cfg.Enrich {
			content = report.GenerateMultiBranch(branchCommits, info.Name, period, cfg.Author)
		} else if cfg.Consolidate {
			content = enrichAllBranchesConsolidated(cfg, branchCommits, info.Name, period)
		} else {
			content = enrichAllBranchesPerBranch(cfg, branchCommits, info.Name, period)
		}
	} else {
		commits, err := gitpkg.GetCommitsRange(dr, cfg.Author)
		if err != nil {
			return err
		}

		if cfg.Enrich {
			if len(commits) == 0 {
				meta := report.EnrichedMeta{
					ClaudeResponse: "No activity recorded for this period.",
					LangCode:       cfg.LangCode,
					ModelID:        cfg.Model,
				}
				content = report.GenerateEnriched(meta, info.Name, period, cfg.Author)
			} else if claude.IsCLIModel(cfg.Model) {
				content = enrichWithClaude(cfg, commits, info.Name, period)
			} else {
				content = enrichWithAPI(cfg, commits, info.Name, period)
			}
		} else {
			content = report.GenerateSimple(commits, info.Name, period, cfg.Author)
		}
	}

	dateSlug := filenameDateSlug(cfg.DateFrom, cfg.DateTo)
	path, err := storage.WriteReport(repoDir, dateSlug, cfg.Author, cfg.Enrich, content)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "  [OK] Report saved: %s\n\n", path)
	fmt.Print(content)
	return nil
}

// enrichWithClaude uses the local Claude CLI (original behaviour).
func enrichWithClaude(cfg config.RunConfig, commits []gitpkg.Commit, repoName, period string) string {
	if !claude.CheckClaude() {
		fmt.Fprintln(os.Stderr, "  [WARN] Claude CLI not found — falling back to simple format")
		fmt.Fprintln(os.Stderr, "  Install with: npm install -g @anthropic-ai/claude-code")
		return report.GenerateSimple(commits, repoName, period, cfg.Author)
	}

	var msgs []string
	for _, c := range commits {
		msgs = append(msgs, c.Message)
	}
	pctx := claude.PromptContext{
		Period:   period,
		RepoName: repoName,
		Author:   cfg.Author,
	}
	prompt := claude.BuildPromptWithContext(msgs, cfg.LangCode, pctx)

	fmt.Fprintf(os.Stderr, "  [INFO] Processing %d commits with %s...\n", len(commits), cfg.Model)
	start := time.Now()
	resp, err := claude.RunClaude(prompt, cfg.Model)
	elapsed := int(time.Since(start).Seconds())
	if err != nil {
		fmt.Fprintf(os.Stderr, "  [WARN] Claude CLI failed: %v — falling back to simple format\n", err)
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

// enrichWithAPI uses the Jetify AI SDK to call any supported LLM directly.
func enrichWithAPI(cfg config.RunConfig, commits []gitpkg.Commit, repoName, period string) string {
	model, err := internalai.NewModel(cfg.Model)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  [WARN] Could not init model %q: %v — falling back to simple format\n", cfg.Model, err)
		return report.GenerateSimple(commits, repoName, period, cfg.Author)
	}

	var msgs []string
	for _, c := range commits {
		msgs = append(msgs, c.Message)
	}
	pctx := claude.PromptContext{
		Period:   period,
		RepoName: repoName,
		Author:   cfg.Author,
	}
	prompt := claude.BuildPromptWithContext(msgs, cfg.LangCode, pctx)

	messages := []internalai.Message{
		{Role: "user", Content: prompt},
	}

	fmt.Fprintf(os.Stderr, "  [INFO] Processing %d commits with %s...\n", len(commits), cfg.Model)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	start := time.Now()
	result, err := internalai.ChatStream(ctx, model, messages, nil)
	elapsed := int(time.Since(start).Seconds())
	if err != nil {
		fmt.Fprintf(os.Stderr, "  [WARN] API call failed: %v — falling back to simple format\n", err)
		return report.GenerateSimple(commits, repoName, period, cfg.Author)
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
	return report.GenerateEnriched(meta, repoName, period, cfg.Author)
}

// enrichAllBranchesPerBranch generates per-branch AI summaries.
func enrichAllBranchesPerBranch(cfg config.RunConfig, branchCommits []gitpkg.BranchCommits, repoName, period string) string {
	var results []report.EnrichedBranchResult

	for _, bc := range branchCommits {
		var msgs []string
		for _, c := range bc.Commits {
			msgs = append(msgs, c.Message)
		}

		pctx := claude.PromptContext{
			Period:   period,
			RepoName: repoName,
			Author:   cfg.Author,
			Branch:   bc.Branch,
		}

		var aiResp string
		var inputTokens, outputTokens, elapsed int

		if claude.IsCLIModel(cfg.Model) {
			if !claude.CheckClaude() {
				aiResp = "(Claude CLI not available)"
			} else {
				prompt := claude.BuildPromptWithContext(msgs, cfg.LangCode, pctx)
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
				prompt := claude.BuildPromptWithContext(msgs, cfg.LangCode, pctx)
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

// enrichAllBranchesConsolidated flattens all branches into one AI summary.
func enrichAllBranchesConsolidated(cfg config.RunConfig, branchCommits []gitpkg.BranchCommits, repoName, period string) string {
	var allCommits []gitpkg.Commit
	for _, bc := range branchCommits {
		allCommits = append(allCommits, bc.Commits...)
	}

	pctx := claude.PromptContext{
		Period:      period,
		RepoName:    repoName,
		Author:      cfg.Author,
		BranchCount: len(branchCommits),
	}

	if claude.IsCLIModel(cfg.Model) {
		var msgs []string
		for _, c := range allCommits {
			msgs = append(msgs, c.Message)
		}
		if !claude.CheckClaude() {
			fmt.Fprintln(os.Stderr, "  [WARN] Claude CLI not found — falling back to simple format")
			return report.GenerateMultiBranch(branchCommits, repoName, period, cfg.Author)
		}
		prompt := claude.BuildPromptWithContext(msgs, cfg.LangCode, pctx)
		fmt.Fprintf(os.Stderr, "  [INFO] Processing %d commits (consolidated) with %s...\n", len(allCommits), cfg.Model)
		start := time.Now()
		resp, err := claude.RunClaude(prompt, cfg.Model)
		elapsed := int(time.Since(start).Seconds())
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [WARN] Claude CLI failed: %v — falling back to multi-branch simple\n", err)
			return report.GenerateMultiBranch(branchCommits, repoName, period, cfg.Author)
		}
		meta := report.EnrichedMeta{
			ClaudeResponse: resp,
			LangCode:       cfg.LangCode,
			CommitCount:    len(allCommits),
			ProcessingTime: elapsed,
			InputTokens:    claude.EstimateTokens(prompt),
			OutputTokens:   claude.EstimateTokens(resp),
			ModelID:        cfg.Model,
		}
		return report.GenerateEnriched(meta, repoName, period, cfg.Author)
	}

	// API model
	model, err := internalai.NewModel(cfg.Model)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  [WARN] Could not init model %q: %v — falling back to simple format\n", cfg.Model, err)
		return report.GenerateMultiBranch(branchCommits, repoName, period, cfg.Author)
	}
	var msgs []string
	for _, c := range allCommits {
		msgs = append(msgs, c.Message)
	}
	prompt := claude.BuildPromptWithContext(msgs, cfg.LangCode, pctx)
	messages := []internalai.Message{{Role: "user", Content: prompt}}
	fmt.Fprintf(os.Stderr, "  [INFO] Processing %d commits (consolidated) with %s...\n", len(allCommits), cfg.Model)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	start := time.Now()
	result, err := internalai.ChatStream(ctx, model, messages, nil)
	elapsed := int(time.Since(start).Seconds())
	if err != nil {
		fmt.Fprintf(os.Stderr, "  [WARN] API call failed: %v — falling back to simple format\n", err)
		return report.GenerateMultiBranch(branchCommits, repoName, period, cfg.Author)
	}
	meta := report.EnrichedMeta{
		ClaudeResponse: result.FullText,
		LangCode:       cfg.LangCode,
		CommitCount:    len(allCommits),
		ProcessingTime: elapsed,
		InputTokens:    result.InputTokens,
		OutputTokens:   result.OutputTokens,
		ModelID:        cfg.Model,
	}
	return report.GenerateEnriched(meta, repoName, period, cfg.Author)
}
