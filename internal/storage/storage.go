package storage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const storageBase = ".git-resumes"

// RepoEntry describes a repository stored in ~/.git-resumes.
type RepoEntry struct {
	Name    string
	Path    string
	Remote  string
	Created string
	Dir     string
	Count   int
}

// ResumeFile describes a single resume text file.
type ResumeFile struct {
	Name    string
	Path    string
	ModTime time.Time
}

// BaseDir returns the absolute path to ~/.git-resumes.
func BaseDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, storageBase)
}

// RepoDirFor returns the storage directory path for the given repo ID.
func RepoDirFor(repoID string) string {
	return filepath.Join(BaseDir(), repoID)
}

// InitRepo creates the repo storage directory and writes/updates .metadata.
func InitRepo(repoID, name, path, remote string) (string, error) {
	dir := RepoDirFor(repoID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create storage dir: %w", err)
	}

	metaPath := filepath.Join(dir, ".metadata")
	created := time.Now().Format(time.RFC3339)

	// Preserve existing created timestamp if metadata already exists.
	if existing, err := readMetadata(metaPath); err == nil {
		if v, ok := existing["created"]; ok && v != "" {
			created = v
		}
	}

	content := fmt.Sprintf("name=%s\npath=%s\nremote=%s\ncreated=%s\n",
		name, path, remote, created)
	if err := os.WriteFile(metaPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write metadata: %w", err)
	}

	return dir, nil
}

// WriteReport writes content to a new file in the repo's storage directory and
// returns the full path.
func WriteReport(repoDir, date, author string, enriched bool, content string) (string, error) {
	filename := buildFilename(date, author, enriched)
	fullPath := filepath.Join(repoDir, filename)
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write report: %w", err)
	}
	return fullPath, nil
}

// ListRepos returns all repo entries found in ~/.git-resumes.
func ListRepos() ([]RepoEntry, error) {
	base := BaseDir()
	entries, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var repos []RepoEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(base, e.Name())
		meta, err := readMetadata(filepath.Join(dir, ".metadata"))
		if err != nil {
			continue
		}

		files, _ := ListResumes(dir)
		repos = append(repos, RepoEntry{
			Name:    meta["name"],
			Path:    meta["path"],
			Remote:  meta["remote"],
			Created: meta["created"],
			Dir:     dir,
			Count:   len(files),
		})
	}
	return repos, nil
}

// ListResumes returns all .txt resume files in a repo directory, sorted newest first.
func ListResumes(repoDir string) ([]ResumeFile, error) {
	entries, err := os.ReadDir(repoDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var files []ResumeFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, ResumeFile{
			Name:    e.Name(),
			Path:    filepath.Join(repoDir, e.Name()),
			ModTime: info.ModTime(),
		})
	}

	// Sort newest first.
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})
	return files, nil
}

// StorageStats returns total repos, total files, and disk usage string.
func StorageStats() (repos, files int, size string) {
	base := BaseDir()
	entries, err := ListRepos()
	if err == nil {
		repos = len(entries)
	}

	_ = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".txt") {
			files++
		}
		return nil
	})

	// Approximate disk usage.
	var totalBytes int64
	_ = filepath.Walk(base, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		totalBytes += info.Size()
		return nil
	})
	size = formatBytes(totalBytes)
	return
}

// ClearAll removes the entire ~/.git-resumes directory.
func ClearAll() error {
	return os.RemoveAll(BaseDir())
}

// --- helpers ----------------------------------------------------------------

func buildFilename(date, author string, enriched bool) string {
	name := "resume_" + date

	if author != "" {
		slug := strings.ToLower(author)
		slug = strings.ReplaceAll(slug, " ", "_")
		// Remove non-alphanumeric/underscore chars.
		var b strings.Builder
		for _, r := range slug {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
				b.WriteRune(r)
			}
		}
		if s := b.String(); s != "" {
			name += "_" + s
		}
	}

	if enriched {
		name += "_enriched"
	}

	name += "_" + time.Now().Format("150405")
	return name + ".txt"
}

func readMetadata(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		m[line[:idx]] = line[idx+1:]
	}
	return m, scanner.Err()
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
