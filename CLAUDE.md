# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Essential Development Commands

### Build and Test

- `make build` - Build both CLI (`git-pr-cli`) and MCP server (`git-pr-mcp`) binaries
- `make test` - Run all tests with coverage
- `make test-coverage` - Generate detailed coverage report (creates coverage.html)
- `make ci` - Run all CI checks (fmt, vet, lint, deadcode, test)

### Code Quality

- `make fmt` - Format Go code with `go fmt`
- `make lint` - Run golangci-lint (install with `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)
- `make vet` - Run `go vet` static analysis
- `make deadcode` - Run deadcode analysis (install with `go install golang.org/x/tools/cmd/deadcode@latest`)

### Development Workflow

- `make run-cli` - Build and run the CLI application
- `make run-mcp` - Build and run the MCP server
- `make clean` - Remove build artifacts and temporary files
- `make deps` - Download and tidy Go dependencies

### Application Commands

- `make check` - Check pull request status across repositories
- `make merge` - Merge ready pull requests (use `make merge ARGS="--dry-run"` for testing)
- `make setup` - Run interactive setup wizard
- `make validate` - Validate configuration and connectivity
- `make stats` - Show repository and PR statistics

## Project Architecture

This is a Go-based Git PR automation tool with two main binaries:

### Core Applications

- **CLI Tool** (`cmd/git-pr-cli/`) - Command-line interface for managing pull requests
- **MCP Server** (`cmd/git-pr-mcp/`) - Model Context Protocol server for AI assistant integration

### Package Structure

- **`pkg/`** - Shared libraries accessible to external packages:
  - `config/` - Configuration management with YAML and environment variable support
  - `providers/` - Git provider implementations (GitHub, GitLab, Bitbucket)
  - `pr/` - Pull request processing logic and filtering
  - `merge/` - Merge strategies (squash, merge, rebase)
  - `notifications/` - Notification systems (Slack, email SMTP)
  - `utils/` - Shared utilities and helpers
  - `validation/` - Configuration and system validation
  - `wizard/` - Interactive setup wizard

- **`internal/`** - Application-specific code not meant for external use:
  - `cli/` - CLI command implementations using Cobra
  - `executor/` - Core execution logic
  - `mcp/` - MCP server implementation

### Key Dependencies

- **CLI Framework**: Cobra + Viper for commands and configuration
- **Git Providers**: Official Go clients (go-github, go-gitlab)
- **HTTP**: Resty with retry logic and rate limiting
- **Logging**: Structured logging with Logrus
- **Testing**: Testify + GoMock
- **MCP**: Model Context Protocol for AI assistant integration

### Configuration System

The application uses a hierarchical configuration system:

1. YAML configuration file (`config.yaml`)
2. Environment variables (e.g., `GITHUB_TOKEN`, `GITLAB_TOKEN`)
3. Command-line flags

Configuration includes:

- Repository definitions with auto-merge settings
- PR filtering rules (allowed actors, skip labels)
- Authentication tokens for providers
- Notification settings

### Safety and Security Features

- Only processes PRs from trusted actors (dependabot, renovate, etc.)
- Requires status checks and approval workflows
- Dry-run mode for testing
- Rate limiting and retry logic
- No sensitive data logging
- Environment variable-based token management

### Testing Strategy

- Unit tests for all packages using testify
- Coverage reports generated in `coverage.html`
- Integration tests for provider APIs
- Mock interfaces for external dependencies

### Build and Deployment

- Cross-platform builds via `make cross-compile`
- Version information embedded via build flags
- Native Go performance with concurrent processing
- Single binary distribution for each platform
