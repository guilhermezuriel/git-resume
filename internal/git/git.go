package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Commit represents a single git commit entry.
type Commit struct {
	Hash    string
	Message string
	Author  string
	Date    string
}

// RepoInfo holds basic repository metadata.
type RepoInfo struct {
	Name   string
	Path   string
	Remote string
	ID     string
}

// GetRepoRoot returns the root directory of the current git repo.
func GetRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

// IsGitRepo returns true if the current directory is inside a git repository.
func IsGitRepo() bool {
	err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Run()
	return err == nil
}

// GetRepoInfo returns metadata for the current git repository.
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

// GetHostAuthor returns the configured git user name or email.
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

// GetCommits returns commits for the given date and optional author filter.
// date must be in YYYY-MM-DD format.
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
		"--pretty=format:%h|%s|%an|%ad",
		"--date=short",
	}

	if author != "" {
		args = append(args, fmt.Sprintf("--author=%s", author))
	}

	out, err := exec.Command("git", args...).Output()
	if err != nil {
		// git log exits 128 if not in repo; exit 0 for empty
		return nil, nil
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}

	var commits []Commit
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 4)
		if len(parts) < 4 {
			continue
		}
		commits = append(commits, Commit{
			Hash:    parts[0],
			Message: parts[1],
			Author:  parts[2],
			Date:    parts[3],
		})
	}
	return commits, nil
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
