package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)


type Commit struct {
	FullHash string
	Hash     string
	Message  string
	Author   string
	Date     string
}


type RepoInfo struct {
	Name   string
	Path   string
	Remote string
	ID     string
}


type DateRange struct {
	From time.Time
	To   time.Time
}


func (dr DateRange) Label() string {
	from := dr.From.Format("2006-01-02")
	to := dr.To.Format("2006-01-02")
	if from == to {
		return from
	}
	return from + " to " + to
}


type BranchCommits struct {
	Branch  string
	Commits []Commit
}


func GetRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}


func IsGitRepo() bool {
	err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Run()
	return err == nil
}


func GetRepoInfo() (*RepoInfo, error) {
	root, err := GetRepoRoot()
	if err != nil {
		return nil, err
	}

	name := filepath.Base(root)

	remote := ""
	if out, err := exec.Command("git", "config", "--get", "remote.origin.url").Output(); err == nil {
		remote = strings.TrimSpace(string(out))
	}

	id := makeRepoID(remote, root)

	return &RepoInfo{
		Name:   name,
		Path:   root,
		Remote: remote,
		ID:     id,
	}, nil
}


func GetHostAuthor() (string, error) {
	name := ""
	email := ""

	if out, err := exec.Command("git", "config", "user.name").Output(); err == nil {
		name = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("git", "config", "user.email").Output(); err == nil {
		email = strings.TrimSpace(string(out))
	}

	if name == "" && email == "" {
		return "", fmt.Errorf("no git user configured")
	}
	if name != "" {
		return name, nil
	}
	return email, nil
}


func GetCommits(date, author string) ([]Commit, error) {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
	}
	nextDate := parsed.AddDate(0, 0, 1).Format("2006-01-02")

	args := []string{
		"log",
		fmt.Sprintf("--after=%s 00:00:00", date),
		fmt.Sprintf("--before=%s 00:00:00", nextDate),
		"--pretty=format:%H|%h|%s|%an|%ad",
		"--date=short",
		"--no-merges",
	}

	if author != "" {
		args = append(args, fmt.Sprintf("--author=%s", author))
	}

	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, nil
	}

	return parseCommitLines(string(out)), nil
}


func GetCommitsRange(dr DateRange, author string) ([]Commit, error) {
	afterDate := dr.From.AddDate(0, 0, -1).Format("2006-01-02")
	beforeDate := dr.To.Format("2006-01-02")

	args := []string{
		"log",
		"HEAD",
		fmt.Sprintf("--after=%s", afterDate),
		fmt.Sprintf("--before=%s 23:59:59", beforeDate),
		"--pretty=format:%H|%h|%s|%an|%ad",
		"--date=short",
		"--no-merges",
	}

	if author != "" {
		args = append(args, fmt.Sprintf("--author=%s", author))
	}

	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, nil
	}

	return parseCommitLines(string(out)), nil
}


func FetchAll() error {
	return exec.Command("git", "fetch", "--all", "--quiet").Run()
}


func GetAllBranches() ([]string, error) {
	out, err := exec.Command("git", "branch", "-a", "--format=%(refname:short)").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}

	var branches []string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "HEAD") {
			continue
		}
		branches = append(branches, line)
	}
	return branches, nil
}


func GetCommitsForBranch(branch string, dr DateRange, author string) ([]Commit, error) {
	afterDate := dr.From.AddDate(0, 0, -1).Format("2006-01-02")
	beforeDate := dr.To.Format("2006-01-02")

	args := []string{
		"log",
		branch,
		fmt.Sprintf("--after=%s", afterDate),
		fmt.Sprintf("--before=%s 23:59:59", beforeDate),
		"--pretty=format:%H|%h|%s|%an|%ad",
		"--date=short",
		"--no-merges",
	}

	if author != "" {
		args = append(args, fmt.Sprintf("--author=%s", author))
	}

	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, nil
	}

	return parseCommitLines(string(out)), nil
}

// GetCommitsAllBranches fetches commits from all branches, deduplicating by full hash.
func GetCommitsAllBranches(dr DateRange, author string) ([]BranchCommits, error) {
	branches, err := GetAllBranches()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var result []BranchCommits

	for _, branch := range branches {
		commits, err := GetCommitsForBranch(branch, dr, author)
		if err != nil {
			continue
		}
		unique := deduplicateByHash(commits, seen)
		if len(unique) > 0 {
			result = append(result, BranchCommits{Branch: branch, Commits: unique})
		}
	}

	return result, nil
}


func deduplicateByHash(commits []Commit, seen map[string]bool) []Commit {
	var result []Commit
	for _, c := range commits {
		if !seen[c.FullHash] {
			seen[c.FullHash] = true
			result = append(result, c)
		}
	}
	return result
}


func parseCommitLines(raw string) []Commit {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var commits []Commit
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}
		commits = append(commits, Commit{
			FullHash: parts[0],
			Hash:     parts[1],
			Message:  parts[2],
			Author:   parts[3],
			Date:     parts[4],
		})
	}
	return commits
}

// makeRepoID creates a unique, filesystem-safe identifier from remote URL or path.
func makeRepoID(remote, path string) string {
	src := remote
	if src == "" {
		src = path
	}

	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	id := re.ReplaceAllString(src, "_")

	// Trim leading/trailing underscores and collapse multiples
	re2 := regexp.MustCompile(`_+`)
	id = re2.ReplaceAllString(id, "_")
	id = strings.Trim(id, "_")

	return id
}
