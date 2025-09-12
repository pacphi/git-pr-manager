# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a bash-based tool for automating pull request management across multiple Git repositories. It supports GitHub, GitLab, and Bitbucket, focusing on safely merging dependency updates from trusted bots like dependabot and renovate.

## Architecture

The project consists of three core bash scripts orchestrated by a comprehensive Makefile:

- **`check-prs.sh`** - Scans configured repositories and reports PR status
- **`merge-prs.sh`** - Merges ready PRs and sends notifications
- **`test-notifications.sh`** - Tests Slack and email notification setup
- **`Makefile`** - Provides convenient command interface with dependency management
- **`config.yaml`** - YAML configuration file (copied from `config.sample`)

The tool reads repository configurations from YAML and uses REST APIs to interact with Git providers.

## Essential Commands

### Setup and Installation
```bash
make install           # Install dependencies (yq, jq, gh, curl)
make setup-config      # Copy config.sample to config.yaml
make setup             # Full setup including auth
make validate          # Validate configuration file
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

## Dependencies

Required tools (auto-installed by `make install`):
- **yq** - YAML processor for parsing config files
- **jq** - JSON processor for API responses
- **curl** - HTTP client for API calls
- **gh** - GitHub CLI (optional, for enhanced GitHub operations)

## Configuration

The main configuration is in `config.yaml` (created from `config.sample`). Key sections:

- `config.pr_filters.allowed_actors` - Trusted bots (e.g., "dependabot[bot]", "renovate[bot]")
- `repositories` - Per-provider repository lists with auto_merge settings
- `auth` - Environment variable references for tokens
- `notifications` - Slack webhook and email SMTP configuration

## Environment Variables

Authentication tokens (set in shell profile):
```bash
export GITHUB_TOKEN="ghp_..."
export GITLAB_TOKEN="glpat_..."
export BITBUCKET_USERNAME="username"
export BITBUCKET_APP_PASSWORD="app_password"
export SLACK_WEBHOOK_URL="https://hooks.slack.com/..."
```

## Safety Features

- Dry-run mode shows planned actions without executing them
- Only processes PRs from configured trusted actors
- Waits for status checks to pass before merging
- Skips PRs with labels like "do-not-merge", "wip"
- Rate limiting protection built into API calls

## Development Notes

- Scripts use `set -euo pipefail` for strict error handling
- Platform detection supports macOS and Linux with different package managers
- Color-coded output with consistent logging functions
- JSON output mode available for programmatic integration
- Configuration validation prevents runtime errors