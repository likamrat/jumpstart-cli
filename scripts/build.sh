#!/bin/bash

# Build script for Jumpstart CLI

set -e  # Exit on any error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_header() {
    echo -e "${CYAN}=== $1 ===${NC}"
}

# Check Go version
print_header "Go Version Check"
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
print_status "Go version: ${GO_VERSION}"

# Check if Go version is at least 1.19 (adjust as needed)
REQUIRED_VERSION="1.19"
if ! go version | grep -q "go1\.\([2-9][0-9]\|1[9-9]\)"; then
    print_warning "Go version ${GO_VERSION} detected. Consider upgrading to ${REQUIRED_VERSION}+ for best compatibility"
fi

print_header "Building Jumpstart CLI"

# Clean previous builds
if [ -f "js" ]; then
    rm js
    print_status "Cleaned previous build"
fi

# Build the Go binary
print_status "Compiling Go binary..."
go build -o js

# Make the binary executable
chmod +x js

# Report binary size
if [ -f "js" ]; then
    BINARY_SIZE=$(du -h js | cut -f1)
    print_status "Binary size: ${BINARY_SIZE}"
else
    print_error "Binary was not created successfully"
    exit 1
fi

# Copy binary to /usr/local/bin for system-wide access
print_status "Installing binary to /usr/local/bin..."
sudo cp js /usr/local/bin/

print_header "Build Summary"
print_success "Build completed successfully!"
print_success "Binary created: js (${BINARY_SIZE})"
print_success "Binary installed to: /usr/local/bin/js"
print_status "You can now run 'js' from anywhere in your terminal"
