package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/guilhermezuriel/git-resume/internal/git"
)

const divider = "============================================================"
const thinDivider = "------------------------------------------------------------"

// GenerateSimple builds a simple commit summary report as a string.
func GenerateSimple(commits []git.Commit, repoName, date, author string) string {
	var sb strings.Builder

	sb.WriteString(divider + "\n")
	sb.WriteString("COMMIT SUMMARY\n")
	sb.WriteString(divider + "\n")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Repository:    %s\n", repoName))
	sb.WriteString(fmt.Sprintf("Date:          %s\n", date))
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

// EnrichedMeta carries statistics used in the enriched report footer.
type EnrichedMeta struct {
	ClaudeResponse  string
	LangCode        string
	CommitCount     int
	ProcessingTime  int // seconds
	InputTokens     int
	OutputTokens    int
}

// GenerateEnriched builds an AI-enriched activity summary report as a string.
func GenerateEnriched(meta EnrichedMeta, repoName, date, author string) string {
	var sb strings.Builder

	sb.WriteString(divider + "\n")
	sb.WriteString("ACTIVITY SUMMARY\n")
	sb.WriteString(divider + "\n")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Repository:    %s\n", repoName))
	sb.WriteString(fmt.Sprintf("Date:          %s\n", date))
	if author != "" {
		sb.WriteString(fmt.Sprintf("Author:        %s\n", author))
	}
	sb.WriteString(fmt.Sprintf("Generated:     %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString("Mode:          AI-enriched (Claude)\n")
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

// FormatReport is a convenience alias — just returns the content as-is,
// useful if callers want a single formatting pass in the future.
func FormatReport(content string) string {
	return content
}
