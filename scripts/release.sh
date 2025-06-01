#!/bin/bash

# ========================================
# Jumpstart CLI Release Helper Script
# ========================================
# This script helps create GitHub releases with proper validation

set -euo pipefail

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly CYAN='\033[0;36m'
readonly BOLD='\033[1m'
readonly NC='\033[0m' # No Color

# Emojis for visual enhancement
readonly SUCCESS="✅"
readonly ERROR="❌"
readonly WARNING="⚠️"
readonly INFO="ℹ️"
readonly ROCKET="🚀"
readonly TAG="🏷️"
readonly BUILD="🔨"
readonly TEST="🧪"

# Helper functions
log_info() {
    echo -e "${CYAN}${INFO} $1${NC}"
}

log_success() {
    echo -e "${GREEN}${SUCCESS} $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}${WARNING} $1${NC}"
}

log_error() {
    echo -e "${RED}${ERROR} $1${NC}"
}

log_header() {
    echo -e "\n${BOLD}${PURPLE}=== $1 ===${NC}\n"
}

# Function to validate version format
validate_version() {
    local version="$1"
    
    # Remove 'v' prefix if present
    version=${version#v}
    
    # Check if version follows semantic versioning (with optional pre-release)
    if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?$ ]]; then
        log_error "Invalid version format: $1"
        log_info "Version should follow semantic versioning: v1.0.0, v1.2.3-alpha, v2.0.0-beta.1, etc."
        return 1
    fi
    
    return 0
}

# Function to check if tag already exists
check_tag_exists() {
    local version="$1"
    
    if git rev-parse "refs/tags/$version" >/dev/null 2>&1; then
        log_error "Tag $version already exists!"
        log_info "Use 'git tag -d $version' to delete locally and 'git push origin :refs/tags/$version' to delete remotely"
        return 1
    fi
    
    return 0
}

# Function to validate repository state
validate_repo_state() {
    log_header "Repository State Validation"
    
    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "Not in a git repository!"
        return 1
    fi
    
    # Check if working directory is clean
    if ! git diff-index --quiet HEAD --; then
        log_error "Working directory is not clean!"
        log_info "Please commit or stash your changes before creating a release"
        return 1
    fi
    
    # Check if we're on the correct branch
    local current_branch=$(git branch --show-current)
    if [ "$current_branch" != "main" ] && [ "$current_branch" != "master" ]; then
        log_warning "You're on branch '$current_branch' instead of main/master"
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Aborted by user"
            return 1
        fi
    fi
    
    # Check if remote is up to date
    log_info "Fetching latest changes..."
    git fetch origin
    
    local local_commit=$(git rev-parse HEAD)
    local remote_commit=$(git rev-parse origin/$(git branch --show-current))
    
    if [ "$local_commit" != "$remote_commit" ]; then
        log_error "Local branch is not up to date with remote!"
        log_info "Please pull the latest changes: git pull origin $(git branch --show-current)"
        return 1
    fi
    
    log_success "Repository state is valid"
    return 0
}

# Function to run tests
run_tests() {
    log_header "Running Tests"
    
    log_info "Installing dependencies..."
    go mod download
    
    log_info "Running tests..."
    if go test -v ./...; then
        log_success "All tests passed!"
        return 0
    else
        log_error "Tests failed!"
        return 1
    fi
}

# Function to build and test binaries
build_and_test() {
    local version="$1"
    
    log_header "Building and Testing Binaries"
    
    # Test builds for different platforms
    local platforms=(
        "linux amd64"
        "darwin amd64"
        "windows amd64"
    )
    
    for platform in "${platforms[@]}"; do
        IFS=' ' read -r os arch <<< "$platform"
        log_info "Testing build for $os/$arch..."
        
        if GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build \
            -ldflags "-X jumpstartcli/internal/utils.CliVersion=$version" \
            -o "test-binary-$os-$arch" . >/dev/null 2>&1; then
            log_success "Build successful for $os/$arch"
            rm -f "test-binary-$os-$arch"*
        else
            log_error "Build failed for $os/$arch"
            return 1
        fi
    done
    
    return 0
}

# Function to create and push tag
create_tag() {
    local version="$1"
    local message="$2"
    
    log_header "Creating Git Tag"
    
    log_info "Creating tag $version..."
    if git tag -a "$version" -m "$message"; then
        log_success "Tag created successfully"
    else
        log_error "Failed to create tag"
        return 1
    fi
    
    log_info "Pushing tag to remote..."
    if git push origin "$version"; then
        log_success "Tag pushed successfully"
        log_info "GitHub Actions will now start building the release"
        log_info "Monitor progress at: https://github.com/likamrat/jumpstart-cli/actions"
    else
        log_error "Failed to push tag"
        log_warning "You may need to delete the local tag: git tag -d $version"
        return 1
    fi
    
    return 0
}

# Function to show release URL
show_release_info() {
    local version="$1"
    
    log_header "Release Information"
    
    echo -e "${BOLD}${ROCKET} Release $version has been triggered!${NC}"
    echo
    echo -e "${BOLD}Monitor the release process:${NC}"
    echo -e "${BLUE}https://github.com/likamrat/jumpstart-cli/actions${NC}"
    echo
    echo -e "${BOLD}Once complete, the release will be available at:${NC}"
    echo -e "${BLUE}https://github.com/likamrat/jumpstart-cli/releases/tag/$version${NC}"
    echo
    echo -e "${BOLD}Users can upgrade using:${NC}"
    echo -e "${CYAN}js upgrade${NC}"
}

# Main function
main() {
    echo -e "${BOLD}${PURPLE}"
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║                🚀 Jumpstart CLI Release Helper           ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    # Parse arguments
    local version=""
    local message=""
    local skip_tests=false
    local force=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                version="$2"
                shift 2
                ;;
            -m|--message)
                message="$2"
                shift 2
                ;;
            --skip-tests)
                skip_tests=true
                shift
                ;;
            --force)
                force=true
                shift
                ;;
            -h|--help)
                echo "Usage: $0 -v <version> [-m <message>] [--skip-tests] [--force]"
                echo
                echo "Options:"
                echo "  -v, --version     Release version (e.g., v1.0.0, v1.2.3-alpha)"
                echo "  -m, --message     Release message (optional)"
                echo "  --skip-tests      Skip running tests"
                echo "  --force           Force release even if tag exists"
                echo "  -h, --help        Show this help message"
                echo
                echo "Examples:"
                echo "  $0 -v v1.0.0"
                echo "  $0 -v v1.2.3-beta -m \"Beta release with new features\""
                echo "  $0 -v v2.0.0 --skip-tests"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                echo "Use -h or --help for usage information"
                exit 1
                ;;
        esac
    done
    
    # Validate required arguments
    if [ -z "$version" ]; then
        log_error "Version is required!"
        echo "Usage: $0 -v <version> [-m <message>] [--skip-tests] [--force]"
        exit 1
    fi
    
    # Add 'v' prefix if not present
    if [[ ! "$version" =~ ^v ]]; then
        version="v$version"
    fi
    
    # Set default message if not provided
    if [ -z "$message" ]; then
        message="Release $version"
    fi
    
    # Validate version format
    if ! validate_version "$version"; then
        exit 1
    fi
    
    # Check if tag exists (unless forced)
    if [ "$force" = false ] && ! check_tag_exists "$version"; then
        exit 1
    fi
    
    # Validate repository state
    if ! validate_repo_state; then
        exit 1
    fi
    
    # Run tests (unless skipped)
    if [ "$skip_tests" = false ]; then
        if ! run_tests; then
            exit 1
        fi
    else
        log_warning "Skipping tests as requested"
    fi
    
    # Build and test binaries
    if ! build_and_test "$version"; then
        exit 1
    fi
    
    # Final confirmation
    echo
    log_warning "Ready to create release $version"
    echo -e "${BOLD}Release details:${NC}"
    echo -e "  ${BOLD}Version:${NC} $version"
    echo -e "  ${BOLD}Message:${NC} $message"
    echo -e "  ${BOLD}Repository:${NC} likamrat/jumpstart-cli"
    echo
    
    if [ "$force" = false ]; then
        read -p "Continue with release? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Release cancelled by user"
            exit 0
        fi
    fi
    
    # Create and push tag
    if ! create_tag "$version" "$message"; then
        exit 1
    fi
    
    # Show release information
    show_release_info "$version"
    
    log_success "Release process initiated successfully!"
    echo -e "\n${BOLD}${GREEN}🎉 Happy releasing! 🎉${NC}"
}

# Run main function with all arguments
main "$@"
