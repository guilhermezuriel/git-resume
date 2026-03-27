package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/guilhermezuriel/git-resume/internal/git"
)

const divider = "============================================================"
const thinDivider = "------------------------------------------------------------"
const branchDivider = "──────────────────────────────"


func GenerateSimple(commits []git.Commit, repoName, period, author string) string {
	var sb strings.Builder

	sb.WriteString(divider + "\n")
	sb.WriteString("COMMIT SUMMARY\n")
	sb.WriteString(divider + "\n")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Repository:    %s\n", repoName))
	sb.WriteString(fmt.Sprintf("Period:        %s\n", period))
	if author != "" {
		sb.WriteString(fmt.Sprintf("Author:        %s\n", author))
	}
	sb.WriteString(fmt.Sprintf("Generated:     %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString("Mode:          Simple\n")
	sb.WriteString("\n")
	sb.WriteString(divider + "\n")
	sb.WriteString("\n")

	if len(commits) == 0 {
		sb.WriteString("No commits found for this date.\n")
	} else {
		sb.WriteString("COMMITS:\n")
		sb.WriteString("--------\n")
		for _, c := range commits {
			sb.WriteString("\n")
			sb.WriteString(fmt.Sprintf("  [%s] %s\n", c.Hash, c.Message))
			if author == "" {
				sb.WriteString(fmt.Sprintf("          Author: %s\n", c.Author))
			}
		}
	}

	sb.WriteString("\n")
	sb.WriteString(thinDivider + "\n")
	sb.WriteString(fmt.Sprintf("Total commits: %d\n", len(commits)))

	return sb.String()
}


type EnrichedMeta struct {
	ClaudeResponse string
	LangCode       string
	CommitCount    int
	ProcessingTime int // seconds
	InputTokens    int
	OutputTokens   int

	ModelID string
	Period string
}

func GenerateEnriched(meta EnrichedMeta, repoName, period, author string) string {
	var sb strings.Builder

	sb.WriteString(divider + "\n")
	sb.WriteString("ACTIVITY SUMMARY\n")
	sb.WriteString(divider + "\n")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Repository:    %s\n", repoName))
	sb.WriteString(fmt.Sprintf("Period:        %s\n", period))
	if author != "" {
		sb.WriteString(fmt.Sprintf("Author:        %s\n", author))
	}
	sb.WriteString(fmt.Sprintf("Generated:     %s\n", time.Now().Format("2006-01-02 15:04:05")))
	modeLabel := "Claude"
	if meta.ModelID != "" {
		modeLabel = meta.ModelID
	}
	sb.WriteString(fmt.Sprintf("Mode:          AI-enriched (%s)\n", modeLabel))
	if meta.LangCode != "" {
		sb.WriteString(fmt.Sprintf("Language:      %s\n", meta.LangCode))
	}
	sb.WriteString("\n")
	sb.WriteString(divider + "\n")
	sb.WriteString("\n")
	sb.WriteString(meta.ClaudeResponse + "\n")
	sb.WriteString("\n")
	sb.WriteString(thinDivider + "\n")
	sb.WriteString("STATISTICS\n")
	sb.WriteString(thinDivider + "\n")
	sb.WriteString(fmt.Sprintf("Commits analyzed:    %d\n", meta.CommitCount))
	sb.WriteString(fmt.Sprintf("Processing time:     %ds\n", meta.ProcessingTime))
	sb.WriteString("\n")
	sb.WriteString("TOKEN USAGE (estimated):\n")
	sb.WriteString(fmt.Sprintf("  Input tokens:      ~%d\n", meta.InputTokens))
	sb.WriteString(fmt.Sprintf("  Output tokens:     ~%d\n", meta.OutputTokens))
	sb.WriteString(fmt.Sprintf("  Total tokens:      ~%d\n", meta.InputTokens+meta.OutputTokens))
	sb.WriteString(thinDivider + "\n")

	return sb.String()
}


func GenerateMultiBranch(branches []git.BranchCommits, repoName, period, author string) string {
	var sb strings.Builder

	authorLabel := "All authors"
	if author != "" {
		authorLabel = author
	}

	totalCommits := 0
	for _, b := range branches {
		totalCommits += len(b.Commits)
	}

	sb.WriteString(divider + "\n")
	sb.WriteString("COMMIT SUMMARY (ALL BRANCHES)\n")
	sb.WriteString(divider + "\n")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Repository:    %s\n", repoName))
	sb.WriteString(fmt.Sprintf("Period:        %s\n", period))
	sb.WriteString(fmt.Sprintf("Author:        %s\n", authorLabel))
	sb.WriteString(fmt.Sprintf("Generated:     %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString("Mode:          All branches\n")
	sb.WriteString("\n")
	sb.WriteString(divider + "\n")
	sb.WriteString("\n")

	if len(branches) == 0 {
		sb.WriteString("No commits found for this period.\n")
	} else {
		for _, bc := range branches {
			sb.WriteString(fmt.Sprintf("BRANCH: %s (%d commits)\n", bc.Branch, len(bc.Commits)))
			sb.WriteString(branchDivider + "\n")
			for _, c := range bc.Commits {
				sb.WriteString(fmt.Sprintf("  [%s] %s  %s\n", c.Hash, c.Date, c.Message))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString(strings.Repeat("─", 58) + "\n")
	sb.WriteString(fmt.Sprintf("Total branches with commits: %d\n", len(branches)))
	sb.WriteString(fmt.Sprintf("Total unique commits: %d\n", totalCommits))
	sb.WriteString(strings.Repeat("─", 58) + "\n")

	return sb.String()
}


type EnrichedBranchResult struct {
	BranchName     string
	AIResponse     string
	CommitCount    int
	ProcessingTime int
	InputTokens    int
	OutputTokens   int
}


func GenerateEnrichedMultiBranch(branches []EnrichedBranchResult, modelID, repoName, period, author, langCode string) string {
	var sb strings.Builder

	authorLabel := "All authors"
	if author != "" {
		authorLabel = author
	}

	modeLabel := "Claude"
	if modelID != "" {
		modeLabel = modelID
	}

	totalCommits := 0
	totalInputTokens := 0
	totalOutputTokens := 0
	for _, b := range branches {
		totalCommits += b.CommitCount
		totalInputTokens += b.InputTokens
		totalOutputTokens += b.OutputTokens
	}

	sb.WriteString(divider + "\n")
	sb.WriteString("ACTIVITY SUMMARY (ALL BRANCHES)\n")
	sb.WriteString(divider + "\n")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Repository:    %s\n", repoName))
	sb.WriteString(fmt.Sprintf("Period:        %s\n", period))
	sb.WriteString(fmt.Sprintf("Author:        %s\n", authorLabel))
	sb.WriteString(fmt.Sprintf("Generated:     %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Mode:          AI-enriched (%s) · All branches\n", modeLabel))
	if langCode != "" {
		sb.WriteString(fmt.Sprintf("Language:      %s\n", langCode))
	}
	sb.WriteString("\n")
	sb.WriteString(divider + "\n")
	sb.WriteString("\n")

	for _, br := range branches {
		sb.WriteString(fmt.Sprintf("BRANCH: %s (%d commits)\n", br.BranchName, br.CommitCount))
		sb.WriteString(branchDivider + "\n")
		sb.WriteString(br.AIResponse + "\n")
		sb.WriteString("\n")
	}

	sb.WriteString(thinDivider + "\n")
	sb.WriteString("STATISTICS\n")
	sb.WriteString(thinDivider + "\n")
	sb.WriteString(fmt.Sprintf("Branches analyzed:   %d\n", len(branches)))
	sb.WriteString(fmt.Sprintf("Commits analyzed:    %d\n", totalCommits))
	sb.WriteString("\n")
	sb.WriteString("TOKEN USAGE (estimated):\n")
	sb.WriteString(fmt.Sprintf("  Input tokens:      ~%d\n", totalInputTokens))
	sb.WriteString(fmt.Sprintf("  Output tokens:     ~%d\n", totalOutputTokens))
	sb.WriteString(fmt.Sprintf("  Total tokens:      ~%d\n", totalInputTokens+totalOutputTokens))
	sb.WriteString(thinDivider + "\n")

	return sb.String()
}


func FormatReport(content string) string {
	return content
}
