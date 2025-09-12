# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a tool for automating pull request management across multiple Git repositories with two interfaces:

1. **Traditional CLI**: Bash-based scripts with Makefile commands
2. **AI Assistant Integration**: Go-based MCP (Model Context Protocol) server for natural language interaction

The tool supports GitHub, GitLab, and Bitbucket, focusing on safely merging dependency updates from trusted bots like dependabot and renovate.

## Architecture

The project has two main components:

### 1. Core CLI Scripts (Traditional Interface)

Bash-based scripts orchestrated by a comprehensive Makefile:

- **`check-prs.sh`** - Scans configured repositories and reports PR status
- **`merge-prs.sh`** - Merges ready PRs and sends notifications
- **`test-notifications.sh`** - Tests Slack and email notification setup
- **`setup-wizard.sh`** - Interactive repository discovery and configuration wizard
- **`Makefile`** - Provides convenient command interface with dependency management
- **`config.yaml`** - YAML configuration file (created from `config.sample` or wizard)

### 2. MCP Server (AI Assistant Interface)

Go-based MCP server for natural language interaction:

- **`mcp-server/main.go`** - MCP server entry point
- **`mcp-server/server/`** - Tool and resource implementations
- **`mcp-server/internal/`** - Core business logic (executor, config, types)
- **`docs/mcp-server/`** - MCP server documentation and IDE setup guides

The MCP server exposes the same underlying functionality through natural language commands, calling the same bash scripts and Makefile targets.

Both interfaces read repository configurations from YAML and use REST APIs to interact with Git providers.

## Essential Commands

### Quick Start (Recommended)

```bash
make setup-full        # Complete automated setup: dependencies + config wizard
make validate          # Validate generated configuration
make check-prs         # Test your setup by checking PR status
```

### Setup and Installation

```bash
make install           # Install dependencies (yq, jq, gh, curl)
make setup-config      # Copy config.sample to config.yaml (manual setup)
make setup-wizard      # Interactive repository discovery wizard
make wizard-preview    # Preview what wizard would configure
make wizard-additive   # Add repositories to existing config
```

### Core Operations

```bash
make check-prs         # Check PR status across all repositories
make dry-run           # Show what would be merged (safe preview)
make merge-prs         # Actually merge ready PRs
make watch             # Monitor continuously (30s refresh)
```

### Testing and Validation

```bash
make test              # Run basic functionality tests
make test-notifications # Test Slack/email setup
make lint              # Lint shell scripts (requires shellcheck)
```

### Platform-Specific Installation

```bash
make install-macos     # Install via Homebrew
make install-linux     # Install via package managers (apt, yum, dnf, pacman)
```

### Monitoring and Utilities

```bash
make watch             # Continuously monitor PR status (30s refresh)
make stats             # Show repository statistics
make backup-config     # Backup current configuration
make restore-config    # Restore from latest backup
```

### MCP Server Commands

```bash
# Build the MCP server
cd mcp-server
go build -o multi-gitter-pr-a8n-mcp .

# Then use natural language in your IDE:
# "Check all pull requests and show me the status"
# "Merge ready dependabot PRs after showing a dry run"
# "Validate my configuration and test my notifications"
# "Show me repository statistics for all configured repos"
```

## Dependencies

### CLI Tools (auto-installed by `make install`)

- **yq** - YAML processor for parsing config files
- **jq** - JSON processor for API responses
- **curl** - HTTP client for API calls
- **gh** - GitHub CLI (optional, for enhanced operations)

### MCP Server Dependencies

- **Go 1.19+** - Required for building the MCP server
- **MCP Go SDK** - Model Context Protocol implementation (`github.com/mark3labs/mcp-go`)

The MCP server dependencies are automatically managed by Go modules (`go.mod` / `go.sum`).

## Configuration

### Automatic Configuration (Recommended)

The setup wizard automatically discovers repositories from your Git providers and generates `config.yaml`:

```bash
make setup-wizard      # Interactive discovery and filtering
make wizard-preview    # See what would be configured
make wizard-additive   # Add to existing configuration
```

The wizard supports:

- **Repository Discovery**: Automatically finds repositories from GitHub, GitLab, and Bitbucket
- **Smart Filtering**: Filter by visibility, owner, activity, name patterns, and custom criteria
- **Interactive Selection**: Choose all repositories, by provider, or selectively
- **Merge Strategy Configuration**: Configure default merge strategies (squash, merge, rebase)

### Manual Configuration

Create `config.yaml` from `config.sample` for manual setup. Key sections:

- `config.pr_filters.allowed_actors` - Trusted bots (e.g., "dependabot[bot]", "renovate[bot]")
- `repositories` - Per-provider repository lists with auto_merge settings
- `auth` - Environment variable references for tokens
- `notifications` - Slack webhook and email SMTP configuration

## Environment Variables

Required authentication tokens (set in shell profile):

```bash
export GITHUB_TOKEN="ghp_..."              # GitHub personal access token
export GITLAB_TOKEN="glpat_..."            # GitLab personal access token
export GITLAB_URL="https://gitlab.com"     # GitLab instance URL (optional)
export BITBUCKET_USERNAME="username"       # Bitbucket username
export BITBUCKET_APP_PASSWORD="app_pass"   # Bitbucket app password
export BITBUCKET_WORKSPACE="workspace"     # Bitbucket workspace (optional)
export SLACK_WEBHOOK_URL="https://hooks.slack.com/..."  # Slack notifications (optional)
```

The setup wizard validates these tokens during repository discovery and shows authentication status for each provider.

## Safety Features

- Dry-run mode shows planned actions without executing them
- Only processes PRs from configured trusted actors
- Waits for status checks to pass before merging
- Skips PRs with labels like "do-not-merge", "wip"
- Rate limiting protection built into API calls

## Development Notes

### CLI Scripts

- Scripts use `set -euo pipefail` for strict error handling
- Platform detection supports macOS and Linux with different package managers
- Color-coded output with consistent logging functions
- JSON output mode available for programmatic integration (`make check-prs-json`)
- Configuration validation prevents runtime errors
- Setup wizard includes progress tracking, error handling, and comprehensive validation
- Backup system automatically preserves existing configurations

### MCP Server

- Built with Go 1.19+ using the `mcp-go` SDK
- Exposes tools that call underlying Makefile targets and bash scripts
- Inherits environment variables from parent process (IDE)
- Provides both tools (actions) and resources (data) to AI assistants
- Uses structured error handling with proper MCP protocol responses
- Supports all major IDEs through MCP protocol standardization
- Documentation includes IDE-specific setup guides for major editors
