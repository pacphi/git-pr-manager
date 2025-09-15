# Git PR Manager

🚀 **Modern Go-based tool for automating pull request management across multiple Git repositories**

Git PR CLI provides a fast, reliable, and feature-rich solution for managing dependency updates and pull requests across GitHub, GitLab, and Bitbucket repositories from a single configuration.

## ✨ Features

- 🔄 **Multi-Provider Support** - GitHub, GitLab, and Bitbucket in one tool
- 🎯 **Smart Filtering** - Only merge PRs from trusted bots (dependabot, renovate, etc.)
- 🛡️ **Safety First** - Dry-run mode, status checks, and approval requirements
- ⚡ **High Performance** - Native Go performance with concurrent processing
- 📱 **Notifications** - Slack webhook and email SMTP notifications
- 🔧 **Interactive Setup** - Automated repository discovery and configuration wizard
- 📊 **Rich Statistics** - Detailed repository and PR analytics
- 🖥️ **Cross-Platform** - Native binaries for macOS, Linux, and Windows

## 🚀 Quick Start

### Install

Download the latest release for your platform:

```bash
# macOS (Apple Silicon)
curl -L -o git-pr-cli https://github.com/pacphi/git-pr-manager/releases/latest/download/git-pr-cli-darwin-arm64
chmod +x git-pr-cli
sudo mv git-pr-cli /usr/local/bin/

# Or build from source
git clone https://github.com/pacphi/git-pr-manager.git
cd git-pr-manager
make build
```

### Configure Authentication

```bash
# Set required environment variables
export GITHUB_TOKEN="ghp_..."
export GITLAB_TOKEN="glpat_..."
export BITBUCKET_USERNAME="username"
export BITBUCKET_APP_PASSWORD="password"
```

### Setup Repositories

```bash
# Interactive setup wizard
git-pr-cli setup wizard

# Validate configuration
git-pr-cli validate --check-repos
```

### Start Automating

```bash
# Check what's ready to merge
git-pr-cli check

# Dry run to see what would be merged
git-pr-cli merge --dry-run

# Merge ready PRs
git-pr-cli merge

# Watch continuously
git-pr-cli watch --interval=5m
```

## 📋 Commands

| Command | Description |
|---------|-------------|
| `check` | Check pull request status across repositories |
| `completion` | Generate the autocompletion script for the specified shell |
| `help` | Help about any command |
| `info` | Show configuration and provider information |
| `merge` | Merge ready pull requests |
| `setup wizard` | Interactive configuration setup |
| `stats` | Show repository and PR statistics |
| `validate` | Validate configuration and connectivity |
| `watch` | Continuously monitor pull requests |

## 📖 Documentation

- **[Installation Guide](docs/user-guide/installation.md)** - Complete installation and setup
- **[Getting Started](docs/user-guide/getting-started.md)** - Step-by-step tutorial
- **[Configuration Reference](docs/user-guide/configuration.md)** - All configuration options
- **[Command Reference](docs/user-guide/commands/)** - Detailed command documentation
- **[MCP Server Guide](docs/mcp-server/)** - AI assistant integration setup
- **[Architecture](docs/developer-guide/architecture.md)** - Technical architecture overview

## 🔧 Configuration Example

```yaml
# config.yaml - Updated for Go-based schema
pr_filters:
  allowed_actors:
    - "dependabot[bot]"
    - "renovate[bot]"
  skip_labels:
    - "do-not-merge"
    - "wip"

repositories:
  github:
    - name: "myorg/web-app"
      auto_merge: true
      merge_strategy: "squash"
      require_checks: true
    - name: "myorg/*"
      auto_merge: false
      require_checks: true

auth:
  github:
    token: "${GITHUB_TOKEN}"
  gitlab:
    token: "${GITLAB_TOKEN}"
    url: "${GITLAB_URL}"
  bitbucket:
    username: "${BITBUCKET_USERNAME}"
    app_password: "${BITBUCKET_APP_PASSWORD}"

notifications:
  slack:
    webhook_url: "${SLACK_WEBHOOK_URL}"
    channel: "#deployments"
    enabled: false
```

## 🏗️ Development

### Build

```bash
# Build both CLI and MCP server
make build

# Cross-platform builds
make cross-compile

# Run tests
make test

# Run linting
make lint
```

### Project Structure

```text
├── cmd/
│   ├── git-pr-cli/     # CLI application
│   └── git-pr-mcp/     # MCP server for AI assistants
├── pkg/                # Shared libraries
│   ├── config/         # Configuration management
│   ├── providers/      # Git provider implementations
│   ├── pr/            # PR processing logic
│   ├── merge/         # Merge strategies
│   └── notifications/ # Notification systems
├── internal/          # Application-specific code
└── docs/             # Documentation
```

### Technology Stack

- **CLI**: Cobra + Viper
- **MCP Server**: Model Context Protocol for AI assistant integration
- **Providers**: Official Go clients (go-github, go-gitlab, go-bitbucket)
- **HTTP**: Resty with retry logic and rate limiting
- **Configuration**: YAML with environment variable support
- **Logging**: Structured logging with Logrus
- **Testing**: Testify + GoMock

## 🔐 Security

- Tokens stored only in environment variables
- HTTPS/TLS enforcement for all API calls
- Input validation and sanitization
- Rate limiting and retry logic
- No sensitive data logging

## 📊 Performance

- Native Go performance
- Concurrent API processing with controlled limits
- Built-in rate limiting per provider
- HTTP connection pooling and keepalive
- Memory-efficient streaming for large datasets

## 🚀 Why Git PR CLI

This modern Go-based solution provides:

- **High Performance** - Native Go performance with concurrent processing
- **Robust Error Handling** - Structured error messages and recovery
- **Enterprise Reliability** - Retry logic, timeout handling, and rate limiting
- **Rich Features** - Interactive wizard, detailed statistics, flexible notifications
- **Cross-Platform** - Single binary works on macOS, Linux, and Windows

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Run `make ci` to verify
5. Submit a pull request

## 📄 License

[MIT License](LICENSE) - see LICENSE file for details.

## 🙏 Acknowledgments

Built with these excellent Go libraries:

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [mcp-go](https://github.com/mark3labs/mcp-go) - Model Context Protocol implementation
- [go-github](https://github.com/google/go-github) - GitHub API client
- [go-gitlab](https://github.com/xanzy/go-gitlab) - GitLab API client
- [go-resty](https://github.com/go-resty/resty) - HTTP client
- [Logrus](https://github.com/sirupsen/logrus) - Structured logging
