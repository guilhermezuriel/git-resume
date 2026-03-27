package claude

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/guilhermezuriel/git-resume/internal/i18n"
)

// ClaudeModel describes a Claude model accessible via the CLI.
type ClaudeModel struct {
	ID    string // passed to --model; "claude" means no flag (CLI default)
	Label string
}

// CLIModels is the catalogue of Claude models available via the CLI.
// Aliases (sonnet, opus, haiku) always resolve to the latest version.
// No ANTHROPIC_API_KEY is required — the CLI uses its own auth.
var CLIModels = []ClaudeModel{
	{ID: "claude", Label: "Claude · CLI default"},
	{ID: "sonnet", Label: "Claude Sonnet · latest"},
	{ID: "opus", Label: "Claude Opus · latest"},
	{ID: "haiku", Label: "Claude Haiku · latest"},
	{ID: "claude-sonnet-4-6", Label: "Claude Sonnet 4.6 · pinned"},
	{ID: "claude-opus-4-6", Label: "Claude Opus 4.6 · pinned"},
	{ID: "claude-haiku-4-5-20251001", Label: "Claude Haiku 4.5 · pinned"},
}

// ListCLIModels returns the catalogue of Claude CLI models.
func ListCLIModels() []ClaudeModel {
	return CLIModels
}

// IsCLIModel reports whether modelID should be routed through the Claude CLI.
func IsCLIModel(modelID string) bool {
	for _, m := range CLIModels {
		if m.ID == modelID {
			return true
		}
	}
	return false
}

// CheckClaude returns true if the claude CLI is available on PATH.
func CheckClaude() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}


func EstimateTokens(text string) int {
	return (len(text) + 3) / 4
}

type PromptContext struct {
	Period      string 
	RepoName    string
	Author      string
	Branch      string 
	BranchCount int    
}

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


func BuildPromptWithContext(msgs []string, langCode string, ctx PromptContext) string {
	langInstruction := i18n.LangInstruction(langCode)

	var sb strings.Builder
	for _, m := range msgs {
		sb.WriteString("- ")
		sb.WriteString(m)
		sb.WriteString("\n")
	}
	commitsBlock := strings.TrimRight(sb.String(), "\n")

	var ctxLines strings.Builder
	ctxLines.WriteString("CONTEXT:\n")
	if ctx.RepoName != "" {
		ctxLines.WriteString(fmt.Sprintf("- Repository: %s\n", ctx.RepoName))
	}
	if ctx.Period != "" {
		ctxLines.WriteString(fmt.Sprintf("- Period: %s\n", ctx.Period))
	}
	if ctx.Author != "" {
		ctxLines.WriteString(fmt.Sprintf("- Author: %s\n", ctx.Author))
	}
	if ctx.Branch != "" {
		ctxLines.WriteString(fmt.Sprintf("- Branch: %s\n", ctx.Branch))
	}
	if ctx.BranchCount > 0 {
		ctxLines.WriteString(fmt.Sprintf("- Branches analyzed: %d\n", ctx.BranchCount))
	}

	return fmt.Sprintf(`You are a senior software engineer generating a development activity summary.
Analyze the commits below and produce an intelligent summary.

GOAL:
Generate a report suitable for:
- Daily standup meetings
- Timesheet entries
- Version changelog

%s
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
%s`, ctxLines.String(), langInstruction, commitsBlock)
}

// RunClaude pipes prompt to the claude CLI and returns its output.
// modelID selects the model via --model; pass "claude" or "" to use the CLI default.
func RunClaude(prompt, modelID string) (string, error) {
	args := []string{"--print"}
	if modelID != "" && modelID != "claude" {
		args = append(args, "--model", modelID)
	}
	cmd := exec.Command("claude", args...)
	cmd.Stdin = strings.NewReader(prompt)

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("claude CLI execution failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
