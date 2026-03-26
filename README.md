# git-resume

A CLI tool that generates daily commit summaries from any Git repository. Supports plain output and AI-powered summaries via Claude CLI, with an interactive TUI mode.

## Features

- Summarize commits for any date with a single command
- AI-enriched summaries using [Claude CLI](https://claude.ai/code) (`--enrich`)
- Multi-language output (`--lang=pt`, `--lang=en`, `--lang=es`, ...)
- Filter by author or your local git user
- Interactive menu (requires [gum](https://github.com/charmbracelet/gum))
- Summaries saved locally in `~/.git-resumes` per repository

## Requirements

- `bash` 4+
- `git`
- **Optional:** [gum](https://github.com/charmbracelet/gum) — for the interactive menu (`git-resume init`)
- **Optional:** [Claude CLI](https://claude.ai/code) (`claude`) — for AI-powered summaries (`--enrich`)

## Installation

### One-liner (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/YOUR_USERNAME/git-resume/main/install.sh | bash
```

### Manual

```bash
curl -fsSL https://raw.githubusercontent.com/YOUR_USERNAME/git-resume/main/git-resume \
  -o /usr/local/bin/git-resume && chmod +x /usr/local/bin/git-resume
```

### Clone & install

```bash
git clone https://github.com/YOUR_USERNAME/git-resume
cd git-resume
bash install.sh
```

## Usage

```
git-resume [command] [options]
```

### Commands

| Command   | Description                          |
|-----------|--------------------------------------|
| _(none)_  | Generate summary for today           |
| `init`    | Open interactive menu (requires gum) |
| `list`    | List all saved summaries for current repo |
| `history` | List all repos with stored summaries |

### Options

| Option             | Description                                      |
|--------------------|--------------------------------------------------|
| `--enrich`         | Generate AI summary using Claude CLI             |
| `--lang=CODE`      | Output language — requires `--enrich` (e.g. `pt`, `en`, `es`) |
| `--date=YYYY-MM-DD`| Target date (default: today)                     |
| `--author=NAME`    | Filter commits by author name or email           |
| `--host`           | Filter commits by your local git user            |
| `-h`, `--help`     | Show help                                        |
| `-v`, `--version`  | Show version                                     |

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

# Open interactive menu
git-resume init

# List all saved summaries for this repo
git-resume list

# Show summaries across all repos
git-resume history
```

## Storage

Summaries are stored locally at:

```
~/.git-resumes/<repo-id>/
```

Each summary is a plain `.txt` file named by date, organized per repository.

## Installing optional dependencies

### gum (interactive menus)

```bash
# macOS
brew install gum

# Ubuntu/Debian
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://repo.charm.sh/apt/gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/charm.gpg
echo "deb [signed-by=/etc/apt/keyrings/charm.gpg] https://repo.charm.sh/apt/ * *" | sudo tee /etc/apt/sources.list.d/charm.list
sudo apt update && sudo apt install gum

# Arch
pacman -S gum
```

### Claude CLI (AI summaries)

Follow the official instructions at [claude.ai/code](https://claude.ai/code).

## License

[MIT](LICENSE)
