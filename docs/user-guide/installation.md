# Installation Guide

Git PR CLI is a cross-platform Go-based tool for automating pull request management across multiple Git repositories.

## Quick Install

### Pre-built Binaries (Recommended)

Download the latest release for your platform from [GitHub Releases](https://github.com/pacphi/git-pr-manager/releases):

#### macOS

**Apple Silicon:**

```bash
curl -L -o git-pr-cli https://github.com/pacphi/git-pr-manager/releases/latest/download/git-pr-cli-darwin-arm64
chmod +x git-pr-cli
sudo mv git-pr-cli /usr/local/bin/
```

**Intel:**

```bash
curl -L -o git-pr-cli https://github.com/pacphi/git-pr-manager/releases/latest/download/git-pr-cli-darwin-amd64
chmod +x git-pr-cli
sudo mv git-pr-cli /usr/local/bin/
```

#### Linux

**x86_64 (Intel/AMD):**

```bash
curl -L -o git-pr-cli https://github.com/pacphi/git-pr-manager/releases/latest/download/git-pr-cli-linux-amd64
chmod +x git-pr-cli
sudo mv git-pr-cli /usr/local/bin/
```

**ARM64:**

```bash
curl -L -o git-pr-cli https://github.com/pacphi/git-pr-manager/releases/latest/download/git-pr-cli-linux-arm64
chmod +x git-pr-cli
sudo mv git-pr-cli /usr/local/bin/
```

#### Windows

**PowerShell:**

```powershell
# Download to current directory
Invoke-WebRequest -Uri "https://github.com/pacphi/git-pr-manager/releases/latest/download/git-pr-cli-windows-amd64.exe" -OutFile "git-pr-cli.exe"

# Move to PATH (optional)
Move-Item -Path "git-pr-cli.exe" -Destination "$env:USERPROFILE\AppData\Local\Microsoft\WindowsApps\git-pr-cli.exe"
```

**Command Prompt:**

```cmd
curl -L -o git-pr-cli.exe https://github.com/pacphi/git-pr-manager/releases/latest/download/git-pr-cli-windows-amd64.exe
```

### Build from Source

If pre-built binaries aren't available for your platform:

**Requirements:**

- Go 1.24+ ([Download Go](https://golang.org/dl/))
- Git
- Make (optional, for convenience)

```bash
# Clone repository
git clone https://github.com/pacphi/git-pr-manager.git
cd git-pr-manager

# Build with Make (recommended)
make build

# Or build manually
go build -o git-pr-cli ./cmd/git-pr-cli
go build -o git-pr-mcp ./cmd/git-pr-mcp

# Install globally
sudo cp git-pr-cli /usr/local/bin/
sudo cp git-pr-mcp /usr/local/bin/
```

**Cross-compilation for other platforms:**

```bash
# Build for all platforms
make cross-compile

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o git-pr-cli-linux-amd64 ./cmd/git-pr-cli
GOOS=windows GOARCH=amd64 go build -o git-pr-cli-windows-amd64.exe ./cmd/git-pr-cli
```

## Dependencies

Git PR CLI requires these external tools for full functionality:

### Required Dependencies

- **git** - Version control (usually pre-installed)
- **curl** - HTTP client (usually pre-installed)

### Optional Dependencies

- **jq** - JSON processor (for advanced scripting and debugging)
- **yq** - YAML processor (for config manipulation)
- **gh** - GitHub CLI (for enhanced GitHub operations)

### Install Optional Dependencies

**macOS (Homebrew):**

```bash
# Install via Makefile (recommended)
make install-macos

# Or manually
brew install jq yq gh
```

**Linux (Ubuntu/Debian):**

```bash
# Install via Makefile (recommended)
make install-linux-apt

# Or manually
sudo apt-get update
sudo apt-get install -y jq curl wget git

# Install yq
sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
sudo chmod +x /usr/local/bin/yq

# Install GitHub CLI
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt update
sudo apt install gh
```

**Linux (CentOS/RHEL/Rocky):**

```bash
# Install via Makefile (recommended)
make install-linux-yum

# Or manually
sudo yum install -y jq curl wget git

# Install yq
sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
sudo chmod +x /usr/local/bin/yq

# Install GitHub CLI
sudo dnf install 'dnf-command(config-manager)'
sudo dnf config-manager --add-repo https://cli.github.com/packages/rpm/gh-cli.repo
sudo dnf install gh
```

**Linux (Fedora):**

```bash
# Install via Makefile (recommended)
make install-linux-dnf

# Or manually
sudo dnf install -y jq curl wget git gh

# Install yq
sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
sudo chmod +x /usr/local/bin/yq
```

**Linux (Arch):**

```bash
# Install via Makefile (recommended)
make install-linux-pacman

# Or manually
sudo pacman -S jq curl wget git github-cli

# Install yq
sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
sudo chmod +x /usr/local/bin/yq
```

**Windows:**

The easiest way is to use a package manager if you have one:

```powershell
# Using Chocolatey (if installed)
choco install jq gh

# Using Scoop (if installed)
scoop install jq gh

# Using Winget (if available)
winget install jqlang.jq GitHub.cli
```

Otherwise, download binaries manually from their respective GitHub releases pages.

## Authentication Setup

Git PR CLI requires authentication tokens for each Git provider you plan to use.

### Environment Variables

Set these in your shell profile (`.bashrc`, `.zshrc`, `.profile`, etc.):

```bash
# GitHub (required for GitHub repositories)
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# GitLab (required for GitLab repositories)
export GITLAB_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export GITLAB_URL="https://gitlab.com"  # Optional: defaults to gitlab.com

# Bitbucket (required for Bitbucket repositories)
export BITBUCKET_USERNAME="your-username"
export BITBUCKET_APP_PASSWORD="xxxxxxxxxxxxxxxxxxxx"
export BITBUCKET_WORKSPACE="your-workspace"  # Optional

# Notifications (optional)
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/..."
export SMTP_HOST="smtp.gmail.com"
export SMTP_USERNAME="notifications@company.com"
export SMTP_PASSWORD="app-password"
export EMAIL_FROM="Git PR Bot <notifications@company.com>"
```

### Platform-Specific Configuration

**macOS/Linux:**

```bash
# Add to ~/.bashrc, ~/.zshrc, or ~/.profile
echo 'export GITHUB_TOKEN="your-token"' >> ~/.bashrc
echo 'export GITLAB_TOKEN="your-token"' >> ~/.bashrc
source ~/.bashrc
```

**Windows (PowerShell):**

```powershell
# Set user environment variables (persistent)
[Environment]::SetEnvironmentVariable("GITHUB_TOKEN", "your-token", "User")
[Environment]::SetEnvironmentVariable("GITLAB_TOKEN", "your-token", "User")

# Or add to PowerShell profile
Add-Content $PROFILE '$env:GITHUB_TOKEN="your-token"'
Add-Content $PROFILE '$env:GITLAB_TOKEN="your-token"'
```

**Windows (Command Prompt):**

```cmd
setx GITHUB_TOKEN "your-token"
setx GITLAB_TOKEN "your-token"
```

### Token Creation

#### GitHub Personal Access Token

1. Go to [GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Select scopes:
   - `repo` (for private repositories)
   - `public_repo` (for public repositories)
   - `read:org` (to discover organization repositories)
4. Copy token and set as `GITHUB_TOKEN`

#### GitLab Personal Access Token

1. Go to [GitLab User Settings → Access Tokens](https://gitlab.com/-/profile/personal_access_tokens)
2. Create token with scopes:
   - `api` (full API access)
   - `read_repository` (read repository data)
   - `write_repository` (merge pull requests)
3. Copy token and set as `GITLAB_TOKEN`

#### Bitbucket App Password

1. Go to [Bitbucket Account settings → App passwords](https://bitbucket.org/account/settings/app-passwords/)
2. Create app password with permissions:
   - `Repositories: Read`
   - `Pull requests: Write`
3. Copy password and set as `BITBUCKET_APP_PASSWORD`

## Verification

### Test Installation

```bash
# Check version and build info
git-pr-cli --version

# Test basic functionality
git-pr-cli --help
```

### Validate Configuration

```bash
# Basic validation (checks config syntax)
git-pr-cli validate

# Test authentication with all providers
git-pr-cli validate --check-auth

# Test repository access
git-pr-cli validate --check-repos

# Show current configuration
git-pr-cli validate --show-config
```

### Quick Setup

```bash
# Run interactive setup wizard
git-pr-cli setup wizard

# Or copy sample configuration
cp config.sample config.yaml
# Edit config.yaml with your repositories

# Test configuration
git-pr-cli check --dry-run
```

## Troubleshooting Installation

### Common Issues

**Binary not found:**

```bash
# Check if binary is in PATH
which git-pr-cli
echo $PATH

# Add to PATH if needed (macOS/Linux)
export PATH="/usr/local/bin:$PATH"
```

**Permission denied:**

```bash
# Make binary executable
chmod +x git-pr-cli

# Check file permissions
ls -la git-pr-cli
```

**Go build errors:**

```bash
# Update Go to latest version
go version  # Should be 1.24+

# Clean Go modules
go clean -modcache
go mod download
go mod tidy
```

**Dependency installation fails:**

```bash
# Check package manager availability
which brew apt yum dnf pacman

# Update package lists
sudo apt update  # Ubuntu/Debian
brew update      # macOS
```

### Platform-Specific Issues

**macOS Gatekeeper:**

```bash
# If binary is blocked by Gatekeeper
sudo xattr -rd com.apple.quarantine git-pr-cli

# Or allow in System Preferences → Security & Privacy
```

**Windows Defender:**

```powershell
# If binary is flagged by Windows Defender
Add-MpPreference -ExclusionPath "C:\path\to\git-pr-cli.exe"
```

**Linux SELinux:**

```bash
# If blocked by SELinux
sudo setsebool -P allow_execmem 1

# Or set proper context
sudo semanage fcontext -a -t bin_t "/usr/local/bin/git-pr-cli"
sudo restorecon /usr/local/bin/git-pr-cli
```

## Next Steps

After successful installation:

1. **[Getting Started Guide](getting-started.md)** - Step-by-step tutorial
2. **[Configuration Reference](configuration.md)** - Detailed configuration options
3. **[Command Reference](commands/)** - Complete command documentation
4. **[Troubleshooting Guide](troubleshooting.md)** - Common issues and solutions

## Uninstallation

### Remove Binary

```bash
# Remove main binaries
sudo rm /usr/local/bin/git-pr-cli
sudo rm /usr/local/bin/git-pr-mcp

# Remove configuration (optional)
rm -rf ~/.config/git-pr
rm config.yaml
```

### Remove Optional Dependencies

```bash
# macOS
brew uninstall jq yq gh

# Ubuntu/Debian
sudo apt remove jq gh
sudo rm /usr/local/bin/yq

# CentOS/RHEL/Fedora
sudo yum remove jq gh  # or dnf
sudo rm /usr/local/bin/yq
```

### Remove Environment Variables

```bash
# Remove from shell profile
nano ~/.bashrc  # Remove export statements

# Windows
[Environment]::SetEnvironmentVariable("GITHUB_TOKEN", $null, "User")
```
