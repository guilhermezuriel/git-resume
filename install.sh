#!/bin/bash

# git-resume installer
# Downloads the pre-built Go binary from GitHub Releases
# Supports: macOS (arm64/amd64), Linux (amd64/arm64)
# Usage: curl -fsSL https://raw.githubusercontent.com/guilhermezuriel/git-resume/main/install.sh | bash

set -e

# ============================================================================
# CONFIGURATION
# ============================================================================

REPO="guilhermezuriel/git-resume"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="git-resume"
GITHUB_API="https://api.github.com/repos/${REPO}/releases/latest"

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
    case "$(uname -s)" in
        Darwin) echo "darwin" ;;
        Linux)  echo "linux"  ;;
        *)
            print_error "Unsupported OS: $(uname -s)"
            print_info "Download manually from: https://github.com/${REPO}/releases"
            exit 1
            ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo "amd64" ;;
        arm64|aarch64)  echo "arm64" ;;
        *)
            print_error "Unsupported architecture: $(uname -m)"
            print_info "Download manually from: https://github.com/${REPO}/releases"
            exit 1
            ;;
    esac
}

# ============================================================================
# INSTALLATION
# ============================================================================

fetch_latest_version() {
    if command -v curl &>/dev/null; then
        curl -fsSL "$GITHUB_API" 2>/dev/null \
            | grep '"tag_name"' \
            | sed -E 's/.*"tag_name": *"v?([^"]+)".*/\1/'
    elif command -v wget &>/dev/null; then
        wget -qO- "$GITHUB_API" 2>/dev/null \
            | grep '"tag_name"' \
            | sed -E 's/.*"tag_name": *"v?([^"]+)".*/\1/'
    else
        print_error "curl or wget is required"
        exit 1
    fi
}

download_file() {
    local url="$1"
    local dest="$2"

    if command -v curl &>/dev/null; then
        curl -fsSL "$url" -o "$dest"
    else
        wget -qO "$dest" "$url"
    fi
}

install_git_resume() {
    local os arch version asset_name download_url temp_file

    os=$(detect_os)
    arch=$(detect_arch)

    print_step "Fetching latest release..."
    version=$(fetch_latest_version)

    if [[ -z "$version" ]]; then
        print_error "Could not determine latest version"
        print_info "Check releases at: https://github.com/${REPO}/releases"
        exit 1
    fi

    print_info "Latest version: v${version}"

    asset_name="${BINARY_NAME}_${os}_${arch}"
    download_url="https://github.com/${REPO}/releases/download/v${version}/${asset_name}"

    print_step "Downloading ${asset_name}..."
    temp_file=$(mktemp)

    if ! download_file "$download_url" "$temp_file"; then
        rm -f "$temp_file"
        print_error "Download failed: ${download_url}"
        print_info "Check releases at: https://github.com/${REPO}/releases"
        exit 1
    fi

    chmod +x "$temp_file"

    print_step "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."

    if [[ -w "$INSTALL_DIR" ]]; then
        mv "$temp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        sudo mv "$temp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    print_success "git-resume v${version} installed to ${INSTALL_DIR}/${BINARY_NAME}"
}

check_dependencies() {
    print_section "Checking dependencies"

    if ! command -v git &>/dev/null; then
        print_error "git is required but not installed"
        exit 1
    fi
    print_success "git found"

    if command -v claude &>/dev/null; then
        print_success "Claude CLI found (AI summaries enabled)"
    else
        print_warning "Claude CLI not found (optional, needed for --enrich)"
        print_info "Install with: npm install -g @anthropic-ai/claude-code"
    fi
}

print_next_steps() {
    print_section "Next steps"
    echo -e "  Run ${BOLD}git-resume${NC} inside any git repository"
    echo -e "  Run ${BOLD}git-resume init${NC} for the interactive TUI"
    echo -e "  Run ${BOLD}git-resume --update${NC} to update in the future"
    echo ""
}

# ============================================================================
# MAIN
# ============================================================================

main() {
    print_header "git-resume installer"
    check_dependencies
    print_section "Installing git-resume"
    install_git_resume
    print_next_steps
}

main
