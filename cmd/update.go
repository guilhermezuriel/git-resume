package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

const (
	githubAPI  = "https://api.github.com/repos/guilhermezuriel/git-resume/releases/latest"
	githubRepo = "https://github.com/guilhermezuriel/git-resume"
)

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func runUpdate() error {
	fmt.Fprintln(os.Stderr, "  [*] Checking for updates...")

	release, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to fetch release info: %w", err)
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := version

	if latest == current {
		fmt.Fprintf(os.Stderr, "  [OK] Already up to date (v%s)\n", current)
		return nil
	}

	fmt.Fprintf(os.Stderr, "  [*] Updating v%s → v%s\n", current, latest)

	assetName := binaryName()
	url := ""
	for _, a := range release.Assets {
		if a.Name == assetName {
			url = a.BrowserDownloadURL
			break
		}
	}

	if url == "" {
		return fmt.Errorf("no binary found for %s in release %s\n    Available at: %s/releases", assetName, release.TagName, githubRepo)
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %w", err)
	}

	if err := downloadBinary(url, execPath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "  [OK] Updated to v%s\n", latest)
	return nil
}

func fetchLatestRelease() (*githubRelease, error) {
	req, err := http.NewRequest("GET", githubAPI, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "git-resume/"+version)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func downloadBinary(url, dest string) error {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "git-resume-update-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return err
	}

	return replaceExecutable(tmpPath, dest)
}

// replaceExecutable swaps the running binary with the new one.
// On Unix we rename the temp file over the destination.
// If that fails (cross-device), we fall back to copy + chmod.
func replaceExecutable(src, dest string) error {
	if err := os.Rename(src, dest); err == nil {
		return nil
	}

	// Fallback: copy contents
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		// Try with sudo hint
		return fmt.Errorf("%w\n    Hint: try running with sudo", err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func binaryName() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	name := fmt.Sprintf("git-resume_%s_%s", goos, goarch)
	if goos == "windows" {
		name += ".exe"
	}
	return name
}
