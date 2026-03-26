# git-resume

A CLI tool that generates daily commit summaries from any Git repository. Supports plain output and AI-powered summaries via Claude CLI, with an interactive TUI mode.

## Features

- Summarize commits for any date with a single command
- AI-enriched summaries using [Claude CLI](https://claude.ai/code) (`--enrich`)
- Multi-language output (`--lang=pt`, `--lang=en`, `--lang=es`, ...)
- Filter by author or your local git user
- Interactive TUI (`git-resume init`)
- Summaries saved locally in `~/.git-resumes` per repository
- Self-updating via `git-resume --update`

## Requirements

- `git`
- **Optional:** [Claude CLI](https://claude.ai/code) (`claude`) — for AI-powered summaries (`--enrich`)

## Installation

### One-liner (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/guilhermezuriel/git-resume/main/install.sh | bash
```

The installer detects your OS and architecture, downloads the correct pre-built binary from GitHub Releases, and places it in `/usr/local/bin`.

### Manual

Download the binary for your platform from [Releases](https://github.com/guilhermezuriel/git-resume/releases/latest):

| Platform       | Binary                          |
|----------------|---------------------------------|
| macOS (Apple Silicon) | `git-resume_darwin_arm64` |
| macOS (Intel)  | `git-resume_darwin_amd64`       |
| Linux (x86_64) | `git-resume_linux_amd64`        |
| Linux (ARM64)  | `git-resume_linux_arm64`        |

```bash
# Example for macOS Apple Silicon
curl -fsSL https://github.com/guilhermezuriel/git-resume/releases/latest/download/git-resume_darwin_arm64 \
  -o /usr/local/bin/git-resume && chmod +x /usr/local/bin/git-resume
```

## Updating

```bash
git-resume --update
```

Fetches the latest release from GitHub and replaces the current binary automatically.

## Usage

```
git-resume [command] [options]
```

### Commands

| Command   | Description                                |
|-----------|--------------------------------------------|
| _(none)_  | Generate summary for today                 |
| `init`    | Open the interactive TUI                   |
| `list`    | List all saved summaries for current repo  |
| `history` | List all repos with stored summaries       |

### Options

| Option              | Description                                            |
|---------------------|--------------------------------------------------------|
| `--enrich`          | Generate AI summary using Claude CLI                   |
| `--lang=CODE`       | Output language — requires `--enrich` (e.g. `pt`, `en`, `es`) |
| `--date=YYYY-MM-DD` | Target date (default: today)                           |
| `--author=NAME`     | Filter commits by author name or email                 |
| `--host`            | Filter commits by your local git user                  |
| `--update`          | Update git-resume to the latest release                |
| `-v`, `--version`   | Show version                                           |
| `-h`, `--help`      | Show help                                              |

### Examples

```bash
# Today's commits, simple format
git-resume

# Today's commits with AI summary
git-resume --enrich

# AI summary in Portuguese
git-resume --enrich --lang=pt

# Summary for a specific date
git-resume --date=2026-03-20

# Only my commits, with AI summary
git-resume --host --enrich

# Filter by a specific author
git-resume --author="john@example.com"

# Open interactive TUI
git-resume init

# List all saved summaries for this repo
git-resume list

# Show summaries across all repos
git-resume history

# Update to latest version
git-resume --update
```

## Storage

Summaries are stored locally at:

```
~/.git-resumes/<repo-id>/
```

Each summary is a plain `.txt` file named by date, organized per repository.

## Claude CLI (AI summaries)

Follow the official instructions at [claude.ai/code](https://claude.ai/code).

## License

[MIT](LICENSE)
