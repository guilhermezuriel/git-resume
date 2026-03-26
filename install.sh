#!/bin/bash

# git-resume installer
# Supports: macOS, Ubuntu/Debian
# Usage: curl -fsSL https://raw.githubusercontent.com/.../install.sh | bash

set -e

# ============================================================================
# CONFIGURATION
# ============================================================================

VERSION="2.0.0"
INSTALL_DIR="/usr/local/bin"
SCRIPT_NAME="git-resume"
REPO_URL="https://github.com/guilhermezuriel/git-resume"
RAW_URL="https://raw.githubusercontent.com/guilhermezuriel/git-resume/main/git-resume"

# For local installation (when script is bundled)
BUNDLED_INSTALL=false

# ============================================================================
# COLORS
# ============================================================================

setup_colors() {
    if [[ -t 1 ]]; then
        BOLD='\033[1m'
        DIM='\033[2m'
        RED='\033[0;31m'
        GREEN='\033[0;32m'
        YELLOW='\033[0;33m'
        BLUE='\033[0;34m'
        CYAN='\033[0;36m'
        NC='\033[0m'
    else
        BOLD='' DIM='' RED='' GREEN='' YELLOW='' BLUE='' CYAN='' NC=''
    fi
}

setup_colors

# ============================================================================
# UI FUNCTIONS
# ============================================================================

H_LINE="─"
TL_CORNER="┌"
TR_CORNER="┐"
BL_CORNER="└"
BR_CORNER="┘"
V_LINE="│"

print_header() {
    local title="$1"
    local width=50
    local padding=$(( (width - ${#title} - 2) / 2 ))
    
    echo ""
    echo -e "${CYAN}${TL_CORNER}$(printf "${H_LINE}%.0s" $(seq 1 $width))${TR_CORNER}${NC}"
    echo -e "${CYAN}${V_LINE}${NC}$(printf ' %.0s' $(seq 1 $padding))${BOLD}${title}${NC}$(printf ' %.0s' $(seq 1 $((width - padding - ${#title}))))${CYAN}${V_LINE}${NC}"
    echo -e "${CYAN}${BL_CORNER}$(printf "${H_LINE}%.0s" $(seq 1 $width))${BR_CORNER}${NC}"
    echo ""
}

print_step() {
    echo -e "${BLUE}[*]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_info() {
    echo -e "${DIM}    $1${NC}"
}

print_section() {
    local title="$1"
    echo ""
    echo -e "${BOLD}${title}${NC}"
    echo -e "${DIM}$(printf "${H_LINE}%.0s" $(seq 1 ${#title}))${NC}"
}

# ============================================================================
# SYSTEM DETECTION
# ============================================================================

detect_os() {
    local os=""
    local version=""
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        os="macos"
        version=$(sw_vers -productVersion 2>/dev/null || echo "unknown")
    elif [[ -f /etc/os-release ]]; then
        . /etc/os-release
        case "$ID" in
            ubuntu|debian|linuxmint|pop)
                os="ubuntu"
                version="$VERSION_ID"
                ;;
            *)
                os="unsupported"
                version="$ID"
                ;;
        esac
    else
        os="unsupported"
    fi
    
    echo "$os|$version"
}

detect_arch() {
    local arch=$(uname -m)
    case "$arch" in
        x86_64|amd64)
            echo "amd64"
            ;;
        arm64|aarch64)
            echo "arm64"
            ;;
        *)
            echo "$arch"
            ;;
    esac
}

# ============================================================================
# DEPENDENCY MANAGEMENT
# ============================================================================

check_command() {
    command -v "$1" &>/dev/null
}

install_homebrew() {
    if check_command brew; then
        print_success "Homebrew already installed"
        return 0
    fi
    
    print_step "Installing Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    
    # Add to PATH for Apple Silicon
    if [[ -f "/opt/homebrew/bin/brew" ]]; then
        eval "$(/opt/homebrew/bin/brew shellenv)"
    fi
    
    print_success "Homebrew installed"
}

install_gum_macos() {
    if check_command gum; then
        print_success "gum already installed"
        return 0
    fi
    
    print_step "Installing gum..."
    brew install gum
    print_success "gum installed"
}

install_gum_ubuntu() {
    if check_command gum; then
        print_success "gum already installed"
        return 0
    fi
    
    print_step "Installing gum..."
    
    # Add Charm repository
    sudo mkdir -p /etc/apt/keyrings
    curl -fsSL https://repo.charm.sh/apt/gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/charm.gpg 2>/dev/null || true
    echo "deb [signed-by=/etc/apt/keyrings/charm.gpg] https://repo.charm.sh/apt/ * *" | sudo tee /etc/apt/sources.list.d/charm.list > /dev/null
    
    sudo apt update -qq
    sudo apt install -y gum
    
    print_success "gum installed"
}

install_git() {
    if check_command git; then
        print_success "git already installed"
        return 0
    fi
    
    print_step "Installing git..."
    
    case "$1" in
        macos)
            brew install git
            ;;
        ubuntu)
            sudo apt update -qq
            sudo apt install -y git
            ;;
    esac
    
    print_success "git installed"
}

check_claude_cli() {
    if check_command claude; then
        print_success "Claude CLI found"
        return 0
    fi
    
    print_warning "Claude CLI not found (optional, needed for --enrich)"
    print_info "Install later with: npm install -g @anthropic-ai/claude-code"
    return 1
}

# ============================================================================
# INSTALLATION
# ============================================================================

install_git_resume() {
    local temp_file=$(mktemp)
    
    print_step "Downloading git-resume..."
    
    # Check if we're doing a bundled install (script in same directory)
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local bundled_script="${script_dir}/git-resume"
    
    if [[ -f "$bundled_script" && "$BUNDLED_INSTALL" == true ]]; then
        cp "$bundled_script" "$temp_file"
        print_info "Using bundled script"
    else
        # Download from repository
        if ! curl -fsSL "$RAW_URL" -o "$temp_file" 2>/dev/null; then
            # Fallback: try to use embedded script
            print_warning "Could not download from repository"
            print_info "Using embedded script..."
            
            # Write embedded script
            write_embedded_script "$temp_file"
        fi
    fi
    
    # Make executable
    chmod +x "$temp_file"
    
    # Install to system
    print_step "Installing to ${INSTALL_DIR}..."
    
    if [[ -w "$INSTALL_DIR" ]]; then
        mv "$temp_file" "${INSTALL_DIR}/${SCRIPT_NAME}"
    else
        sudo mv "$temp_file" "${INSTALL_DIR}/${SCRIPT_NAME}"
    fi
    
    print_success "git-resume installed to ${INSTALL_DIR}/${SCRIPT_NAME}"
}

write_embedded_script() {
    local output_file="$1"
    
    # This is the embedded git-resume script
    cat > "$output_file" << 'EMBEDDED_SCRIPT'
#!/bin/bash

# git-resume - Daily commit summary generator with interactive UI
# Version 2.0.0

set -e

# ============================================================================
# CONFIGURATION
# ============================================================================

VERSION="2.0.0"
STORAGE_DIR="$HOME/.git-resumes"
ENRICH=false
TARGET_DATE=$(date +%Y-%m-%d)
AUTHOR_FILTER=""
USE_HOST_AUTHOR=false
LANG_CODE=""
INTERACTIVE_MODE=false

# ============================================================================
# COLORS & STYLES
# ============================================================================

setup_colors() {
    if [[ -t 1 ]]; then
        BOLD='\033[1m'
        DIM='\033[2m'
        ITALIC='\033[3m'
        RED='\033[0;31m'
        GREEN='\033[0;32m'
        YELLOW='\033[0;33m'
        BLUE='\033[0;34m'
        MAGENTA='\033[0;35m'
        CYAN='\033[0;36m'
        WHITE='\033[0;37m'
        NC='\033[0m'
    else
        BOLD='' DIM='' ITALIC='' RED='' GREEN='' YELLOW='' 
        BLUE='' MAGENTA='' CYAN='' WHITE='' NC=''
    fi
}

H_LINE="─"
V_LINE="│"
TL_CORNER="┌"
TR_CORNER="┐"
BL_CORNER="└"
BR_CORNER="┘"
T_DOWN="┬"
T_UP="┴"
T_RIGHT="├"
T_LEFT="┤"
CROSS="┼"

setup_colors

# ============================================================================
# UI FUNCTIONS
# ============================================================================

print_header() {
    local title="$1"
    local width=60
    local padding=$(( (width - ${#title} - 2) / 2 ))
    
    echo ""
    echo -e "${CYAN}${TL_CORNER}$(printf "${H_LINE}%.0s" $(seq 1 $width))${TR_CORNER}${NC}"
    echo -e "${CYAN}${V_LINE}${NC}$(printf ' %.0s' $(seq 1 $padding))${BOLD}${title}${NC}$(printf ' %.0s' $(seq 1 $((width - padding - ${#title}))))${CYAN}${V_LINE}${NC}"
    echo -e "${CYAN}${BL_CORNER}$(printf "${H_LINE}%.0s" $(seq 1 $width))${BR_CORNER}${NC}"
}

print_section() {
    local title="$1"
    echo ""
    echo -e "${BOLD}${title}${NC}"
    echo -e "${DIM}$(printf "${H_LINE}%.0s" $(seq 1 ${#title}))${NC}"
}

print_key_value() {
    local key="$1"
    local value="$2"
    printf "  ${DIM}%-14s${NC} %s\n" "$key:" "$value"
}

print_success() {
    echo -e "  ${GREEN}[OK]${NC} $1"
}

print_error() {
    echo -e "  ${RED}[ERROR]${NC} $1" >&2
}

print_warning() {
    echo -e "  ${YELLOW}[WARN]${NC} $1"
}

print_info() {
    echo -e "  ${BLUE}[INFO]${NC} $1"
}

# ============================================================================
# DEPENDENCY CHECK
# ============================================================================

check_gum() {
    command -v gum &>/dev/null
}

install_gum_instructions() {
    echo ""
    echo -e "  ${BOLD}Install gum for interactive menus:${NC}"
    echo ""
    echo "  macOS:"
    echo "    brew install gum"
    echo ""
    echo "  Linux (Debian/Ubuntu):"
    echo "    sudo mkdir -p /etc/apt/keyrings"
    echo "    curl -fsSL https://repo.charm.sh/apt/gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/charm.gpg"
    echo "    echo \"deb [signed-by=/etc/apt/keyrings/charm.gpg] https://repo.charm.sh/apt/ * *\" | sudo tee /etc/apt/sources.list.d/charm.list"
    echo "    sudo apt update && sudo apt install gum"
    echo ""
}

check_claude_cli() {
    command -v claude &>/dev/null
}

# ============================================================================
# STORAGE MANAGEMENT
# ============================================================================

get_repo_id() {
    local remote=$(git config --get remote.origin.url 2>/dev/null || echo "")
    if [[ -n "$remote" ]]; then
        echo "$remote" | sed 's/[^a-zA-Z0-9]/_/g' | sed 's/__*/_/g' | sed 's/^_//;s/_$//'
    else
        local repo_path=$(git rev-parse --show-toplevel 2>/dev/null)
        echo "$repo_path" | sed 's/[^a-zA-Z0-9]/_/g' | sed 's/__*/_/g' | sed 's/^_//;s/_$//'
    fi
}

get_repo_name() {
    basename "$(git rev-parse --show-toplevel 2>/dev/null)"
}

get_repo_storage_dir() {
    local repo_id=$(get_repo_id)
    echo "${STORAGE_DIR}/${repo_id}"
}

init_storage() {
    local repo_dir=$(get_repo_storage_dir)
    local repo_name=$(get_repo_name)
    
    mkdir -p "$repo_dir"
    
    cat > "${repo_dir}/.metadata" <<EOF
name=${repo_name}
path=$(git rev-parse --show-toplevel)
remote=$(git config --get remote.origin.url 2>/dev/null || echo "local")
created=$(date -Iseconds)
EOF
    
    echo "$repo_dir"
}

list_all_repos() {
    [[ ! -d "$STORAGE_DIR" ]] && return
    
    for repo_dir in "$STORAGE_DIR"/*/; do
        [[ -d "$repo_dir" ]] || continue
        local metadata="${repo_dir}.metadata"
        if [[ -f "$metadata" ]]; then
            local name=$(grep "^name=" "$metadata" | cut -d= -f2-)
            local path=$(grep "^path=" "$metadata" | cut -d= -f2-)
            local count=$(find "$repo_dir" -name "*.txt" -type f 2>/dev/null | wc -l | tr -d ' ')
            echo "${name}|${path}|${count}|${repo_dir}"
        fi
    done
}

list_repo_resumes() {
    local repo_dir="$1"
    [[ ! -d "$repo_dir" ]] && return
    
    find "$repo_dir" -name "*.txt" -type f -printf "%T@ %p\n" 2>/dev/null | \
        sort -rn | cut -d' ' -f2-
}

# ============================================================================
# TOKEN COUNTING
# ============================================================================

estimate_tokens() {
    local text="$1"
    local char_count=${#text}
    echo $(( (char_count + 3) / 4 ))
}

# ============================================================================
# HELP
# ============================================================================

show_help() {
    print_header "git-resume v${VERSION}"
    
    echo ""
    echo -e "  ${BOLD}USAGE${NC}"
    echo "    git-resume [command] [options]"
    
    echo ""
    echo -e "  ${BOLD}COMMANDS${NC}"
    echo "    (none)            Generate summary for today"
    echo "    init              Open interactive menu"
    echo "    list              List all summaries for current repo"
    echo "    history           List all repos with summaries"
    
    echo ""
    echo -e "  ${BOLD}OPTIONS${NC}"
    echo "    --enrich          Generate AI-powered summary using Claude CLI"
    echo "    --lang=CODE       Output language (requires --enrich)"
    echo "                      Examples: en, pt, es, fr, de, ja, zh"
    echo "    --date=YYYY-MM-DD Target date (default: today)"
    echo "    --author=NAME     Filter by author name or email"
    echo "    --host            Filter by local git user"
    echo "    -h, --help        Show this help"
    echo "    -v, --version     Show version"
    
    echo ""
    echo -e "  ${BOLD}EXAMPLES${NC}"
    echo "    git-resume                         Today's commits, simple format"
    echo "    git-resume --enrich                Today's commits, AI summary"
    echo "    git-resume --enrich --lang=en      AI summary in English"
    echo "    git-resume init                    Open interactive menu"
    echo "    git-resume list                    Show all summaries"
    
    echo ""
    echo -e "  ${BOLD}STORAGE${NC}"
    echo "    Summaries are stored in: ${STORAGE_DIR}"
    echo ""
}

show_version() {
    echo "git-resume v${VERSION}"
}

# ============================================================================
# VALIDATION
# ============================================================================

check_git_repo() {
    if ! git rev-parse --is-inside-work-tree &>/dev/null; then
        print_error "Not inside a Git repository"
        exit 1
    fi
}

# ============================================================================
# ARGUMENT PARSING
# ============================================================================

parse_args() {
    case "${1:-}" in
        init)
            INTERACTIVE_MODE=true
            shift
            ;;
        list)
            cmd_list
            exit 0
            ;;
        history)
            cmd_history
            exit 0
            ;;
    esac
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --enrich)
                ENRICH=true
                shift
                ;;
            --lang=*)
                LANG_CODE="${1#*=}"
                shift
                ;;
            --date=*)
                TARGET_DATE="${1#*=}"
                shift
                ;;
            --date)
                if [[ -n "$2" && "$2" != --* ]]; then
                    TARGET_DATE="$2"
                    shift 2
                else
                    print_error "--date requires a value (YYYY-MM-DD)"
                    exit 1
                fi
                ;;
            --author=*)
                AUTHOR_FILTER="${1#*=}"
                shift
                ;;
            --author)
                if [[ -n "$2" && "$2" != --* ]]; then
                    AUTHOR_FILTER="$2"
                    shift 2
                else
                    print_error "--author requires a name or email"
                    exit 1
                fi
                ;;
            --host)
                USE_HOST_AUTHOR=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            -v|--version)
                show_version
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                echo "  Run 'git-resume --help' for usage"
                exit 1
                ;;
        esac
    done
    
    if [[ -n "$LANG_CODE" && "$ENRICH" != true ]]; then
        print_error "--lang requires --enrich flag"
        exit 1
    fi
    
    if [[ "$USE_HOST_AUTHOR" == true ]]; then
        local git_name=$(git config user.name 2>/dev/null)
        local git_email=$(git config user.email 2>/dev/null)
        
        if [[ -z "$git_name" && -z "$git_email" ]]; then
            print_error "No git user configured"
            echo "  Configure with:"
            echo "    git config user.name \"Your Name\""
            echo "    git config user.email \"your@email.com\""
            exit 1
        fi
        
        AUTHOR_FILTER="${git_name:-$git_email}"
    fi
}

# ============================================================================
# GIT OPERATIONS
# ============================================================================

get_commits() {
    local date="$1"
    local author="$2"
    
    local next_date
    next_date=$(date -d "$date + 1 day" +%Y-%m-%d 2>/dev/null) || \
    next_date=$(date -v+1d -j -f "%Y-%m-%d" "$date" +%Y-%m-%d 2>/dev/null)
    
    local git_cmd="git log --after=\"$date 00:00:00\" --before=\"$next_date 00:00:00\" --pretty=format:\"%h|%s|%an|%ad\" --date=short"
    
    [[ -n "$author" ]] && git_cmd="$git_cmd --author=\"$author\""
    
    eval "$git_cmd" 2>/dev/null
}

# ============================================================================
# PROMPT BUILDER
# ============================================================================

get_language_instruction() {
    local lang="$1"
    
    case "$lang" in
        pt|pt-br|pt-BR) echo "Respond in Brazilian Portuguese (pt-BR)." ;;
        en|en-us|en-US) echo "Respond in English." ;;
        es) echo "Respond in Spanish." ;;
        fr) echo "Respond in French." ;;
        de) echo "Respond in German." ;;
        it) echo "Respond in Italian." ;;
        ja) echo "Respond in Japanese." ;;
        ko) echo "Respond in Korean." ;;
        zh) echo "Respond in Chinese (Simplified)." ;;
        ru) echo "Respond in Russian." ;;
        *) echo "Respond in $lang." ;;
    esac
}

build_prompt() {
    local commits="$1"
    local lang_instruction=""
    
    if [[ -n "$LANG_CODE" ]]; then
        lang_instruction=$(get_language_instruction "$LANG_CODE")
    else
        lang_instruction="Respond in Brazilian Portuguese (pt-BR)."
    fi
    
    cat <<EOF
You are a senior software engineer generating a development activity summary.
Analyze the commits below and produce an intelligent summary.

GOAL:
Generate a report suitable for:
- Daily standup meetings
- Timesheet entries
- Version changelog

RULES:
- ${lang_instruction}
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
${commits}
EOF
}

# ============================================================================
# REPORT GENERATION
# ============================================================================

generate_output_filename() {
    local date="$1"
    local author="$2"
    local enriched="$3"
    
    local filename="resume_${date}"
    
    if [[ -n "$author" ]]; then
        local author_slug=$(echo "$author" | tr ' ' '_' | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9_]//g')
        filename="${filename}_${author_slug}"
    fi
    
    [[ "$enriched" == true ]] && filename="${filename}_enriched"
    filename="${filename}_$(date +%H%M%S)"
    
    echo "${filename}.txt"
}

generate_simple_resume() {
    local commits="$1"
    local date="$2"
    local author="$3"
    local repo_name=$(get_repo_name)
    local repo_dir=$(init_storage)
    local filename=$(generate_output_filename "$date" "$author" false)
    local output_file="${repo_dir}/${filename}"
    local commit_count=0
    
    {
        echo "============================================================"
        echo "COMMIT SUMMARY"
        echo "============================================================"
        echo ""
        echo "Repository:    $repo_name"
        echo "Date:          $date"
        [[ -n "$author" ]] && echo "Author:        $author"
        echo "Generated:     $(date '+%Y-%m-%d %H:%M:%S')"
        echo "Mode:          Simple"
        echo ""
        echo "============================================================"
        echo ""
        
        if [[ -z "$commits" ]]; then
            echo "No commits found for this date."
        else
            echo "COMMITS:"
            echo "--------"
            while IFS='|' read -r hash msg author_commit date_commit; do
                [[ -z "$hash" ]] && continue
                ((commit_count++))
                echo ""
                echo "  [$hash] $msg"
                [[ -z "$author" ]] && echo "          Author: $author_commit"
            done <<< "$commits"
        fi
        
        echo ""
        echo "------------------------------------------------------------"
        echo "Total commits: $commit_count"
    } > "$output_file"
    
    print_success "Report saved: $output_file"
    echo ""
    cat "$output_file"
}

generate_enriched_resume() {
    local commits="$1"
    local date="$2"
    local author="$3"
    local repo_name=$(get_repo_name)
    local repo_dir=$(init_storage)
    local filename=$(generate_output_filename "$date" "$author" true)
    local output_file="${repo_dir}/${filename}"
    
    if ! check_claude_cli; then
        print_error "Claude CLI not found"
        echo ""
        echo "  Install with:"
        echo "    npm install -g @anthropic-ai/claude-code"
        echo ""
        print_warning "Falling back to simple format..."
        echo ""
        generate_simple_resume "$commits" "$date" "$author"
        return
    fi
    
    if [[ -z "$commits" ]]; then
        print_warning "No commits found for $date"
        {
            echo "============================================================"
            echo "ACTIVITY SUMMARY"
            echo "============================================================"
            echo ""
            echo "Repository:    $repo_name"
            echo "Date:          $date"
            [[ -n "$author" ]] && echo "Author:        $author"
            echo "Generated:     $(date '+%Y-%m-%d %H:%M:%S')"
            echo ""
            echo "No activity recorded for this date."
        } > "$output_file"
        print_success "Report saved: $output_file"
        return
    fi
    
    local commits_formatted=$(echo "$commits" | while IFS='|' read -r hash msg author_commit date_commit; do
        [[ -n "$msg" ]] && echo "- $msg"
    done)
    
    local commit_count=$(echo "$commits" | grep -c . 2>/dev/null || echo 0)
    
    print_info "Processing $commit_count commits with Claude..."
    
    local prompt=$(build_prompt "$commits_formatted")
    local input_tokens=$(estimate_tokens "$prompt")
    
    local claude_response
    local start_time=$(date +%s)
    
    claude_response=$(echo "$prompt" | claude --print 2>/dev/null) || {
        print_error "Claude CLI execution failed"
        print_warning "Falling back to simple format..."
        echo ""
        generate_simple_resume "$commits" "$date" "$author"
        return
    }
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    local output_tokens=$(estimate_tokens "$claude_response")
    local total_tokens=$((input_tokens + output_tokens))
    
    {
        echo "============================================================"
        echo "ACTIVITY SUMMARY"
        echo "============================================================"
        echo ""
        echo "Repository:    $repo_name"
        echo "Date:          $date"
        [[ -n "$author" ]] && echo "Author:        $author"
        echo "Generated:     $(date '+%Y-%m-%d %H:%M:%S')"
        echo "Mode:          AI-enriched (Claude)"
        [[ -n "$LANG_CODE" ]] && echo "Language:      $LANG_CODE"
        echo ""
        echo "============================================================"
        echo ""
        echo "$claude_response"
        echo ""
        echo "------------------------------------------------------------"
        echo "STATISTICS"
        echo "------------------------------------------------------------"
        echo "Commits analyzed:    $commit_count"
        echo "Processing time:     ${duration}s"
        echo ""
        echo "TOKEN USAGE (estimated):"
        echo "  Input tokens:      ~$input_tokens"
        echo "  Output tokens:     ~$output_tokens"
        echo "  Total tokens:      ~$total_tokens"
        echo "------------------------------------------------------------"
    } > "$output_file"
    
    print_success "Enriched report saved: $output_file"
    echo ""
    cat "$output_file"
}

# ============================================================================
# COMMANDS
# ============================================================================

cmd_list() {
    check_git_repo
    
    local repo_dir=$(get_repo_storage_dir)
    local repo_name=$(get_repo_name)
    
    print_header "Summaries: $repo_name"
    
    if [[ ! -d "$repo_dir" ]]; then
        print_info "No summaries found for this repository"
        echo ""
        echo "  Generate your first summary with:"
        echo "    git-resume"
        echo "    git-resume --enrich"
        echo ""
        return
    fi
    
    local files=$(list_repo_resumes "$repo_dir")
    
    if [[ -z "$files" ]]; then
        print_info "No summaries found"
        return
    fi
    
    echo ""
    printf "  ${BOLD}%-40s %s${NC}\n" "FILE" "DATE"
    echo "  $(printf "${H_LINE}%.0s" $(seq 1 55))"
    
    while IFS= read -r file; do
        [[ -z "$file" ]] && continue
        local filename=$(basename "$file")
        local modified=$(date -r "$file" '+%Y-%m-%d %H:%M' 2>/dev/null || stat -c '%y' "$file" 2>/dev/null | cut -d'.' -f1)
        printf "  %-40s %s\n" "$filename" "$modified"
    done <<< "$files"
    
    echo ""
}

cmd_history() {
    print_header "All Repositories"
    
    if [[ ! -d "$STORAGE_DIR" ]]; then
        print_info "No repositories found"
        echo ""
        echo "  Generate summaries in a git repository first."
        echo ""
        return
    fi
    
    local repos=$(list_all_repos)
    
    if [[ -z "$repos" ]]; then
        print_info "No repositories found"
        return
    fi
    
    echo ""
    printf "  ${BOLD}%-20s %-30s %s${NC}\n" "REPOSITORY" "PATH" "SUMMARIES"
    echo "  $(printf "${H_LINE}%.0s" $(seq 1 65))"
    
    while IFS='|' read -r name path count dir; do
        [[ -z "$name" ]] && continue
        local short_path="$path"
        [[ ${#path} -gt 28 ]] && short_path="...${path: -25}"
        printf "  %-20s %-30s %s\n" "$name" "$short_path" "$count"
    done <<< "$repos"
    
    echo ""
}

# ============================================================================
# INTERACTIVE MODE
# ============================================================================

interactive_menu() {
    if ! check_gum; then
        print_error "Interactive mode requires 'gum'"
        install_gum_instructions
        exit 1
    fi
    
    check_git_repo
    
    local repo_name=$(get_repo_name)
    
    while true; do
        clear
        echo ""
        gum style \
            --border rounded \
            --border-foreground 6 \
            --padding "0 2" \
            --margin "0 0 1 0" \
            "git-resume v${VERSION}  |  ${repo_name}"
        
        local choice=$(gum choose \
            --cursor.foreground 6 \
            --selected.foreground 6 \
            "Generate new summary" \
            "View summaries" \
            "Browse all repositories" \
            "Settings" \
            "Exit")
        
        case "$choice" in
            "Generate new summary") interactive_generate ;;
            "View summaries") interactive_view_summaries ;;
            "Browse all repositories") interactive_browse_repos ;;
            "Settings") interactive_settings ;;
            "Exit"|"") clear; echo "Goodbye!"; exit 0 ;;
        esac
    done
}

interactive_generate() {
    clear
    echo ""
    gum style --foreground 6 --bold "Generate New Summary"
    echo ""
    
    local date_choice=$(gum choose \
        --cursor.foreground 6 \
        "Today ($(date +%Y-%m-%d))" \
        "Yesterday ($(date -d 'yesterday' +%Y-%m-%d 2>/dev/null || date -v-1d +%Y-%m-%d))" \
        "Custom date")
    
    local target_date
    case "$date_choice" in
        "Today"*) target_date=$(date +%Y-%m-%d) ;;
        "Yesterday"*) target_date=$(date -d 'yesterday' +%Y-%m-%d 2>/dev/null || date -v-1d +%Y-%m-%d) ;;
        "Custom date") target_date=$(gum input --placeholder "YYYY-MM-DD" --value "$(date +%Y-%m-%d)") ;;
    esac
    
    echo ""
    local author_choice=$(gum choose \
        --cursor.foreground 6 \
        "All authors" \
        "My commits only ($(git config user.name 2>/dev/null || echo 'not configured'))" \
        "Specific author")
    
    local author_filter=""
    case "$author_choice" in
        "My commits only"*) author_filter=$(git config user.name 2>/dev/null) ;;
        "Specific author") author_filter=$(gum input --placeholder "Author name or email") ;;
    esac
    
    echo ""
    local mode_choice=$(gum choose \
        --cursor.foreground 6 \
        "Simple (fast)" \
        "AI-enriched (requires Claude CLI)")
    
    local enrich=false
    local lang=""
    
    if [[ "$mode_choice" == "AI-enriched"* ]]; then
        enrich=true
        echo ""
        lang=$(gum choose \
            --cursor.foreground 6 \
            "pt (Portuguese)" \
            "en (English)" \
            "es (Spanish)" \
            "fr (French)" \
            "de (German)" \
            "Other")
        
        lang=$(echo "$lang" | cut -d' ' -f1)
        [[ "$lang" == "Other" ]] && lang=$(gum input --placeholder "Language code (e.g., ja, ko, zh)")
    fi
    
    echo ""
    gum style --foreground 3 "Configuration:"
    echo "  Date:   $target_date"
    echo "  Author: ${author_filter:-All}"
    echo "  Mode:   $([ "$enrich" = true ] && echo "AI-enriched ($lang)" || echo "Simple")"
    echo ""
    
    if gum confirm "Generate summary?"; then
        echo ""
        TARGET_DATE="$target_date"
        AUTHOR_FILTER="$author_filter"
        ENRICH="$enrich"
        LANG_CODE="$lang"
        run_generation
        echo ""
        gum input --placeholder "Press Enter to continue..."
    fi
}

interactive_view_summaries() {
    local repo_dir=$(get_repo_storage_dir)
    
    if [[ ! -d "$repo_dir" ]]; then
        gum style --foreground 3 "No summaries found for this repository"
        sleep 2
        return
    fi
    
    local files=$(list_repo_resumes "$repo_dir")
    
    if [[ -z "$files" ]]; then
        gum style --foreground 3 "No summaries found"
        sleep 2
        return
    fi
    
    local options=()
    while IFS= read -r file; do
        [[ -z "$file" ]] && continue
        options+=("$(basename "$file")")
    done <<< "$files"
    options+=("< Back")
    
    while true; do
        clear
        echo ""
        gum style --foreground 6 --bold "Select Summary to View"
        echo ""
        
        local selected=$(gum choose --cursor.foreground 6 "${options[@]}")
        
        [[ "$selected" == "< Back" || -z "$selected" ]] && return
        
        local full_path="${repo_dir}/${selected}"
        [[ -f "$full_path" ]] && { clear; gum pager < "$full_path"; }
    done
}

interactive_browse_repos() {
    if [[ ! -d "$STORAGE_DIR" ]]; then
        gum style --foreground 3 "No repositories found"
        sleep 2
        return
    fi
    
    local repos=$(list_all_repos)
    
    if [[ -z "$repos" ]]; then
        gum style --foreground 3 "No repositories found"
        sleep 2
        return
    fi
    
    local options=()
    declare -A repo_dirs
    
    while IFS='|' read -r name path count dir; do
        [[ -z "$name" ]] && continue
        local label="$name ($count summaries)"
        options+=("$label")
        repo_dirs["$label"]="$dir"
    done <<< "$repos"
    options+=("< Back")
    
    clear
    echo ""
    gum style --foreground 6 --bold "Select Repository"
    echo ""
    
    local selected=$(gum choose --cursor.foreground 6 "${options[@]}")
    
    [[ "$selected" == "< Back" || -z "$selected" ]] && return
    
    local selected_dir="${repo_dirs[$selected]}"
    
    if [[ -d "$selected_dir" ]]; then
        local files=$(list_repo_resumes "$selected_dir")
        
        if [[ -n "$files" ]]; then
            local file_options=()
            while IFS= read -r file; do
                [[ -z "$file" ]] && continue
                file_options+=("$(basename "$file")")
            done <<< "$files"
            file_options+=("< Back")
            
            clear
            echo ""
            gum style --foreground 6 --bold "Summaries in $selected"
            echo ""
            
            local file_selected=$(gum choose --cursor.foreground 6 "${file_options[@]}")
            
            if [[ "$file_selected" != "< Back" && -n "$file_selected" ]]; then
                local full_path="${selected_dir}/${file_selected}"
                [[ -f "$full_path" ]] && { clear; gum pager < "$full_path"; }
            fi
        fi
    fi
}

interactive_settings() {
    clear
    echo ""
    gum style --foreground 6 --bold "Settings"
    echo ""
    gum style --foreground 8 "Storage directory: $STORAGE_DIR"
    echo ""
    
    local choice=$(gum choose \
        --cursor.foreground 6 \
        "View storage stats" \
        "Clear all summaries" \
        "< Back")
    
    case "$choice" in
        "View storage stats")
            echo ""
            local total_files=$(find "$STORAGE_DIR" -name "*.txt" -type f 2>/dev/null | wc -l | tr -d ' ')
            local total_size=$(du -sh "$STORAGE_DIR" 2>/dev/null | cut -f1 || echo "0")
            local repo_count=$(find "$STORAGE_DIR" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l | tr -d ' ')
            
            echo "  Total repositories: $repo_count"
            echo "  Total summaries:    $total_files"
            echo "  Storage size:       $total_size"
            echo ""
            gum input --placeholder "Press Enter to continue..."
            ;;
        "Clear all summaries")
            if gum confirm --negative "Cancel" "Delete ALL summaries? This cannot be undone."; then
                rm -rf "$STORAGE_DIR"
                gum style --foreground 2 "All summaries deleted"
                sleep 2
            fi
            ;;
    esac
}

# ============================================================================
# CORE GENERATION
# ============================================================================

run_generation() {
    local repo_name=$(get_repo_name)
    
    print_header "git-resume"
    
    print_section "Configuration"
    print_key_value "Repository" "$repo_name"
    print_key_value "Date" "$TARGET_DATE"
    [[ -n "$AUTHOR_FILTER" ]] && print_key_value "Author" "$AUTHOR_FILTER"
    [[ "$ENRICH" == true ]] && print_key_value "Mode" "AI-enriched"
    [[ -n "$LANG_CODE" ]] && print_key_value "Language" "$LANG_CODE"
    print_key_value "Storage" "$(get_repo_storage_dir)"
    
    print_section "Processing"
    
    local commits
    commits=$(get_commits "$TARGET_DATE" "$AUTHOR_FILTER")
    
    local commit_count=0
    [[ -n "$commits" ]] && commit_count=$(echo "$commits" | grep -c . 2>/dev/null || echo 0)
    
    print_info "Found $commit_count commits"
    
    print_section "Output"
    
    if [[ "$ENRICH" == true ]]; then
        generate_enriched_resume "$commits" "$TARGET_DATE" "$AUTHOR_FILTER"
    else
        generate_simple_resume "$commits" "$TARGET_DATE" "$AUTHOR_FILTER"
    fi
}

# ============================================================================
# MAIN
# ============================================================================

main() {
    parse_args "$@"
    
    if [[ "$INTERACTIVE_MODE" == true ]]; then
        interactive_menu
    else
        check_git_repo
        run_generation
    fi
}

main "$@"
EMBEDDED_SCRIPT
}

# ============================================================================
# VERIFICATION
# ============================================================================

verify_installation() {
    print_step "Verifying installation..."
    
    if ! check_command git-resume; then
        print_error "Installation verification failed"
        return 1
    fi
    
    local installed_version=$(git-resume --version 2>/dev/null || echo "unknown")
    print_success "Verified: $installed_version"
    return 0
}

# ============================================================================
# UNINSTALL
# ============================================================================

uninstall() {
    print_header "Uninstall git-resume"
    
    if [[ ! -f "${INSTALL_DIR}/${SCRIPT_NAME}" ]]; then
        print_warning "git-resume is not installed"
        exit 0
    fi
    
    print_step "Removing git-resume..."
    
    if [[ -w "${INSTALL_DIR}/${SCRIPT_NAME}" ]]; then
        rm "${INSTALL_DIR}/${SCRIPT_NAME}"
    else
        sudo rm "${INSTALL_DIR}/${SCRIPT_NAME}"
    fi
    
    print_success "git-resume removed"
    
    echo ""
    echo -e "  ${DIM}Note: Summaries in ~/.git-resumes were not removed.${NC}"
    echo -e "  ${DIM}To remove them: rm -rf ~/.git-resumes${NC}"
    echo ""
}

# ============================================================================
# MAIN INSTALLER
# ============================================================================

main() {
    # Check for uninstall flag
    if [[ "${1:-}" == "--uninstall" || "${1:-}" == "uninstall" ]]; then
        uninstall
        exit 0
    fi
    
    print_header "git-resume Installer v${VERSION}"
    
    # Detect system
    print_step "Detecting system..."
    
    local os_info=$(detect_os)
    local os=$(echo "$os_info" | cut -d'|' -f1)
    local os_version=$(echo "$os_info" | cut -d'|' -f2)
    local arch=$(detect_arch)
    
    print_success "Detected: $os $os_version ($arch)"
    
    # Check if supported
    if [[ "$os" == "unsupported" ]]; then
        print_error "Unsupported operating system: $os_version"
        echo ""
        echo "  Supported systems:"
        echo "    - macOS (10.15+)"
        echo "    - Ubuntu / Debian / Linux Mint / Pop!_OS"
        echo ""
        echo "  For other systems, install manually:"
        echo "    1. Download git-resume script"
        echo "    2. chmod +x git-resume"
        echo "    3. Move to /usr/local/bin/"
        echo "    4. Install gum: https://github.com/charmbracelet/gum"
        echo ""
        exit 1
    fi
    
    # Install dependencies based on OS
    print_section "Installing Dependencies"
    
    case "$os" in
        macos)
            install_homebrew
            install_git macos
            install_gum_macos
            ;;
        ubuntu)
            print_step "Updating package list..."
            sudo apt update -qq
            print_success "Package list updated"
            
            install_git ubuntu
            install_gum_ubuntu
            ;;
    esac
    
    # Check for Claude CLI (optional)
    check_claude_cli || true
    
    # Install git-resume
    print_section "Installing git-resume"
    install_git_resume
    
    # Verify installation
    print_section "Verification"
    verify_installation
    
    # Done
    print_section "Installation Complete"
    
    echo ""
    echo -e "  ${GREEN}git-resume has been installed successfully!${NC}"
    echo ""
    echo -e "  ${BOLD}Quick Start:${NC}"
    echo "    cd /path/to/your/git/repo"
    echo "    git-resume              # Simple summary"
    echo "    git-resume --enrich     # AI-powered summary"
    echo "    git-resume init         # Interactive menu"
    echo ""
    echo -e "  ${BOLD}Documentation:${NC}"
    echo "    git-resume --help"
    echo ""
    
    # Optional: Claude CLI reminder
    if ! check_command claude; then
        echo -e "  ${YELLOW}Optional:${NC} For AI-powered summaries, install Claude CLI:"
        echo "    npm install -g @anthropic-ai/claude-code"
        echo ""
    fi
}

main "$@"
