# 🚀 GitHub Release Process Guide

This guide explains how to create releases for the Jumpstart CLI project using the automated GitHub Actions workflow.

## 📋 Table of Contents

- [Quick Start](#-quick-start)
- [Release Types](#-release-types)
- [Prerequisites](#-prerequisites)
- [Creating a Release](#-creating-a-release)
- [Manual Release Process](#-manual-release-process)
- [Automated Workflow](#-automated-workflow)
- [Troubleshooting](#-troubleshooting)
- [Upgrade System](#-upgrade-system)

## 🚀 Quick Start

The fastest way to create a release:

```bash
# Using the release helper script
./scripts/release.sh -v v1.0.0

# Or manually create a tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

## 📦 Release Types

### Stable Releases
- **Format**: `v1.0.0`, `v2.1.3`
- **Description**: Production-ready releases
- **Automated**: Full testing, cross-platform builds, GitHub release

### Pre-releases
- **Format**: `v1.0.0-alpha`, `v1.2.0-beta.1`, `v2.0.0-rc.1`
- **Description**: Testing and preview releases
- **Automated**: Same as stable but marked as pre-release

### Development Releases
- **Format**: `v1.0.0-dev`, `v1.0.0-dev.20250601`
- **Description**: Development snapshots
- **Automated**: Basic builds, marked as pre-release

## ✅ Prerequisites

### Repository Setup
- [x] Fork configured with upstream (`Azure/jumpstart-cli`)
- [x] Local repository up to date
- [x] Working directory clean (no uncommitted changes)
- [x] All tests passing

### Permissions
- [x] Write access to the repository
- [x] Ability to create tags and releases

### Tools Required
- Git
- Go 1.21+
- GitHub CLI (optional, for manual releases)

## 🔧 Creating a Release

### Method 1: Using the Release Script (Recommended)

The release script provides comprehensive validation and automation:

```bash
# Basic release
./scripts/release.sh -v v1.0.0

# Release with custom message
./scripts/release.sh -v v1.2.0 -m "Major feature update with new CLI commands"

# Pre-release
./scripts/release.sh -v v2.0.0-beta.1 -m "Beta release for testing"

# Skip tests (not recommended)
./scripts/release.sh -v v1.0.1 --skip-tests

# Force release (overwrites existing tag)
./scripts/release.sh -v v1.0.0 --force
```

#### Script Options

| Option | Description | Example |
|--------|-------------|---------|
| `-v, --version` | Release version (required) | `-v v1.0.0` |
| `-m, --message` | Release message | `-m "Bug fixes and improvements"` |
| `--skip-tests` | Skip running tests | `--skip-tests` |
| `--force` | Overwrite existing tag | `--force` |
| `-h, --help` | Show help | `-h` |

### Method 2: Manual Git Tags

For experienced users who prefer manual control:

```bash
# 1. Ensure repository is up to date
git fetch origin
git pull origin main

# 2. Run tests manually
go test ./...

# 3. Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### Method 3: GitHub Actions Manual Trigger

You can also trigger releases through GitHub Actions UI:

1. Go to **Actions** → **🚀 Build and Release Jumpstart CLI**
2. Click **Run workflow**
3. Enter version (e.g., `v1.0.0`)
4. Choose if it's a pre-release
5. Click **Run workflow**

## 🤖 Automated Workflow

When you push a tag or trigger the workflow, the following happens automatically:

### 1. 🧪 Validation and Testing (5-10 minutes)
- **Environment Setup**: Go 1.21, dependency caching
- **Code Quality**: `go vet`, `go fmt` checks
- **Comprehensive Testing**: All tests with race detection and coverage
- **Version Injection Testing**: Validates version embedding works
- **Upgrade System Testing**: Dry-run of upgrade functionality

### 2. 🏗️ Cross-Platform Building (10-15 minutes)
Builds binaries for all supported platforms:

| Platform | Architecture | Binary Name |
|----------|--------------|-------------|
| 🐧 Linux | AMD64 | `js-linux-amd64` |
| 🐧 Linux | ARM64 | `js-linux-arm64` |
| 🪟 Windows | AMD64 | `js-windows-amd64.exe` |
| 🍎 macOS | AMD64 | `js-darwin-amd64` |
| 🍎 macOS | ARM64 (M1/M2) | `js-darwin-arm64` |

### 3. 🧪 Integration Testing (5-10 minutes)
- **Cross-Platform Validation**: Tests binaries on Ubuntu, Windows, macOS
- **Basic Functionality**: Version, help, and config commands
- **Binary Integrity**: Ensures all builds work correctly

### 4. 🚀 Release Creation (2-5 minutes)
- **Asset Preparation**: Organizes binaries and generates checksums
- **Release Notes**: Auto-generated with commit history and download links
- **GitHub Release**: Creates release with all assets
- **Notification**: Updates GitHub release page and sends notifications

## 🎯 Upgrade System

### How It Works
The CLI includes a built-in upgrade system that:

1. **Checks for Updates**: Queries GitHub releases API
2. **Version Comparison**: Uses semantic versioning
3. **Platform Detection**: Automatically selects correct binary
4. **Binary Download**: Downloads and verifies new version
5. **Installation**: Replaces current binary with new version

### User Experience
```bash
# Check for updates
js upgrade check

# Upgrade to latest version
js upgrade

# Upgrade to specific version
js upgrade --version v1.2.0
```

### Configuration
The upgrade system is configured in `internal/upgrade/config/config.go`:

```go
const (
    DefaultOwner = "likamrat"          // Your GitHub username
    DefaultRepo  = "jumpstart-cli"     // Repository name
    
    // Binary naming patterns match release workflow
    LinuxBinaryPattern   = "js-linux-amd64"
    WindowsBinaryPattern = "js-windows-amd64.exe"
    DarwinBinaryPattern  = "js-darwin-amd64"
    // ... ARM64 variants
)
```

## 🔍 Troubleshooting

### Common Issues

#### ❌ "Tag already exists"
```bash
# Delete local tag
git tag -d v1.0.0

# Delete remote tag
git push origin :refs/tags/v1.0.0

# Or use --force flag
./scripts/release.sh -v v1.0.0 --force
```

#### ❌ "Tests failed"
```bash
# Run tests locally to debug
go test -v ./...

# Check specific test
go test -v ./cmd/version/

# Skip tests if urgent (not recommended)
./scripts/release.sh -v v1.0.0 --skip-tests
```

#### ❌ "Build failed"
```bash
# Test build locally
go build -ldflags "-X jumpstartcli/internal/utils.CliVersion=v1.0.0" .

# Check for platform-specific issues
GOOS=windows GOARCH=amd64 go build .
```

#### ❌ "Working directory not clean"
```bash
# Check what's uncommitted
git status

# Commit changes
git add .
git commit -m "Prepare for release"

# Or stash changes
git stash
```

#### ❌ "Not up to date with remote"
```bash
# Pull latest changes
git pull origin main

# Or rebase if you have local commits
git rebase origin/main
```

### GitHub Actions Issues

#### Workflow Not Triggering
1. Check if tag was pushed: `git ls-remote --tags origin`
2. Verify workflow file syntax in `.github/workflows/release.yml`
3. Check GitHub Actions tab for error messages

#### Build Failures
1. Check **Actions** tab for detailed logs
2. Look for specific error messages in build steps
3. Test builds locally using the same commands

#### Release Not Created
1. Verify GitHub token permissions
2. Check if release already exists
3. Look for upload failures in workflow logs

### Upgrade System Issues

#### Users Can't Find Updates
1. Verify repository URL in `config.go`
2. Check if releases are marked as drafts
3. Ensure binary naming matches patterns

#### Download Failures
1. Check GitHub release assets are public
2. Verify binary names match exactly
3. Test download URLs manually

## 📚 Advanced Usage

### Custom Release Notes
Edit the workflow to customize release notes generation in `.github/workflows/release.yml`:

```yaml
- name: 📝 Generate Release Notes
  id: release_notes
  run: |
    # Add your custom release notes logic here
```

### Additional Platforms
Add new platforms to the build matrix:

```yaml
- goos: linux
  goarch: arm
  name: js-linux-arm
  display: 🐧 Linux ARM32
```

### Webhook Integration
Set up webhooks for release notifications:

```yaml
- name: 📢 Notify Team
  run: |
    curl -X POST "${{ secrets.WEBHOOK_URL }}" \
      -H "Content-Type: application/json" \
      -d '{"version": "${{ needs.validate.outputs.version }}"}'
```

## 🔗 Useful Links

- **GitHub Releases**: https://github.com/likamrat/jumpstart-cli/releases
- **GitHub Actions**: https://github.com/likamrat/jumpstart-cli/actions
- **Issues**: https://github.com/likamrat/jumpstart-cli/issues
- **Original Project**: https://github.com/Azure/jumpstart-cli

## 📞 Support

If you encounter issues with the release process:

1. **Check this documentation** for common solutions
2. **Review GitHub Actions logs** for specific error details
3. **Test locally** using the same commands as the workflow
4. **Create an issue** with detailed error information

---

*This release system is designed to be robust, automated, and user-friendly. The comprehensive testing and validation ensure that releases are always high quality and ready for production use.*
