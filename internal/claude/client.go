package claude

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/guilhermezuriel/git-resume/internal/i18n"
)

// CheckClaude returns true if the claude CLI is available on PATH.
func CheckClaude() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

// EstimateTokens estimates the token count for text using ~4 chars per token.
func EstimateTokens(text string) int {
	return (len(text) + 3) / 4
}

// BuildPrompt constructs the Claude prompt from commit messages and language code.
func BuildPrompt(commitMessages []string, langCode string) string {
	langInstruction := i18n.LangInstruction(langCode)

	var sb strings.Builder
	for _, m := range commitMessages {
		sb.WriteString("- ")
		sb.WriteString(m)
		sb.WriteString("\n")
	}
	commitsBlock := strings.TrimRight(sb.String(), "\n")

	return fmt.Sprintf(`You are a senior software engineer generating a development activity summary.
Analyze the commits below and produce an intelligent summary.

GOAL:
Generate a report suitable for:
- Daily standup meetings
- Timesheet entries
- Version changelog

RULES:
- %s
- Maximum 5 bullet points
- Group commits by functional context (e.g., authentication, payments, UI)
- Ignore irrelevant commits (merge, wip, update, typo, lint, etc.)
- Deduplicate redundant information
- Infer context even from poorly written messages
- Prioritize impact (what changed in the system)

FORMAT:
- Clear and objective bullet points
- Start with a verb (e.g., "Implemented...", "Fixed...", "Refactored...")
- Do not mention commit hashes
- Do not cite authors

IF POSSIBLE:
- Identify affected system areas (e.g., auth, API, frontend)
- Combine multiple commits into a single cohesive description

COMMITS:
%s`, langInstruction, commitsBlock)
}

// RunClaude pipes the prompt to the claude CLI and returns its output.
func RunClaude(prompt string) (string, error) {
	cmd := exec.Command("claude", "--print")
	cmd.Stdin = strings.NewReader(prompt)

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("claude CLI execution failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
