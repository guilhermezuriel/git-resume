package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/guilhermezuriel/git-resume/internal/claude"
	"github.com/guilhermezuriel/git-resume/internal/config"
	gitpkg "github.com/guilhermezuriel/git-resume/internal/git"
	"github.com/guilhermezuriel/git-resume/internal/report"
	"github.com/guilhermezuriel/git-resume/internal/storage"
	"github.com/guilhermezuriel/git-resume/internal/tui"
)

const version = "3.0.0"

var (
	flagEnrich  bool
	flagLang    string
	flagDate    string
	flagAuthor  string
	flagHost    bool
	flagVersion bool
	flagUpdate  bool
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

	// Resolve --host to author name.
	author := flagAuthor
	if flagHost {
		name, err := gitpkg.GetHostAuthor()
		if err != nil {
			return fmt.Errorf("%w\nConfigure with:\n  git config user.name \"Your Name\"\n  git config user.email \"your@email.com\"", err)
		}
		author = name
	}

	// Resolve date.
	date := flagDate
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	cfg := config.RunConfig{
		Date:     date,
		Author:   author,
		Enrich:   flagEnrich,
		LangCode: flagLang,
	}

	return generate(cfg)
}

func generate(cfg config.RunConfig) error {
	info, err := gitpkg.GetRepoInfo()
	if err != nil {
		return err
	}

	commits, err := gitpkg.GetCommits(cfg.Date, cfg.Author)
	if err != nil {
		return err
	}

	repoDir, err := storage.InitRepo(info.ID, info.Name, info.Path, info.Remote)
	if err != nil {
		return err
	}

	var content string

	if cfg.Enrich {
		if !claude.CheckClaude() {
			fmt.Fprintln(os.Stderr, "  [WARN] Claude CLI not found — falling back to simple format")
			fmt.Fprintln(os.Stderr, "  Install with: npm install -g @anthropic-ai/claude-code")
			cfg.Enrich = false
		}
	}

	if cfg.Enrich {
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

			fmt.Fprintf(os.Stderr, "  [INFO] Processing %d commits with Claude...\n", len(commits))
			start := time.Now()
			resp, err := claude.RunClaude(prompt)
			elapsed := int(time.Since(start).Seconds())
			if err != nil {
				fmt.Fprintf(os.Stderr, "  [WARN] Claude CLI failed: %v — falling back to simple format\n", err)
				content = report.GenerateSimple(commits, info.Name, cfg.Date, cfg.Author)
			} else {
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
		}
	} else {
		content = report.GenerateSimple(commits, info.Name, cfg.Date, cfg.Author)
	}

	path, err := storage.WriteReport(repoDir, cfg.Date, cfg.Author, cfg.Enrich, content)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "  [OK] Report saved: %s\n\n", path)
	fmt.Print(content)
	return nil
}
