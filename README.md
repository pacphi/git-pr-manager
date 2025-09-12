# Multi-Gitter Pull-Request Automation

🤖 **Automate pull request management across multiple Git repositories**

Manage PR automation centrally through YAML configuration or natural language commands via AI assistants. Watch repositories, check PR status, and automatically merge ready PRs across GitHub, GitLab, and Bitbucket.

**Two ways to use:**

- 🖥️ **Command Line**: Traditional make commands and shell scripts
- 🤖 **AI Assistant**: Natural language via MCP (Model Context Protocol) server in your IDE

## ✨ Key Benefits

- ✅ **Central Configuration** - Manage all repositories from one YAML file
- 🔄 **Multi-Provider Support** - GitHub, GitLab, Bitbucket in one tool
- 🎯 **Smart Filtering** - Only merge PRs from trusted bots (dependabot, renovate)
- 🛡️ **Safe by Default** - Dry-run mode, status checks, approval requirements
- 📱 **Notifications** - Slack and email alerts for merged PRs
- ⚡ **Simple CLI** - Easy-to-use commands for all operations
- 🖥️ **Cross-Platform** - Works on macOS and Linux with automatic dependency management
- 🤖 **AI Integration** - Natural language interface via MCP server for IDEs

## 🚀 Quick Start (5 minutes)

**🤖 AI Assistant Setup**

Use natural language commands in your IDE:

```bash
# 1. Build the MCP server
cd mcp-server && go build -o multi-gitter-pr-a8n-mcp .

# 2. Follow IDE-specific setup guide
# See docs/mcp-server/MCP_SETUP.md
```

Then in your IDE: *"Check all pull requests and merge ready dependabot PRs"*

**📖 Full MCP Setup Guide**: [`docs/mcp-server/MCP_SETUP.md`](docs/mcp-server/MCP_SETUP.md)

**🎯 Command Line Setup (Traditional)**

```bash
# 1. Complete automated setup with wizard
make setup-full

# 2. Validate your configuration
make validate

# 3. Check what PRs are ready
make check-prs

# 4. See what would be merged (safe!)
make dry-run

# 5. Actually merge the PRs
make merge-prs
```

**⚙️ Manual Setup (Alternative)**

```bash
# 1. Install dependencies
make install

# 2. Create your configuration file
make setup-config

# 3. Edit config.yaml to add your repositories
# (This copies config.sample to config.yaml for you to customize)

# 4. Validate and test
make validate && make check-prs
```

**📚 New here?** Start with the **[Quick Start Guide](docs/QUICKSTART.md)** →

## 📋 Configuration Options

**🎯 Automatic Configuration (Recommended)**

The setup wizard discovers and configures repositories automatically:

```bash
make setup-wizard      # Interactive repository discovery
make wizard-preview    # Preview what would be configured
make wizard-additive   # Add to existing configuration
```

**⚙️ Manual Configuration**

Copy the sample and edit manually:

```bash
make setup-config  # Creates config.yaml from config.sample
```

Then edit **config.yaml** with your repositories:

```yaml
# Only merge PRs from trusted bots
config:
  pr_filters:
    allowed_actors:
      - "dependabot[bot]"
      - "renovate[bot]"

# Your repositories across providers
repositories:
  github:
    - name: "company/frontend"
      auto_merge: true
    - name: "company/backend"
      auto_merge: true

  gitlab:
    - name: "team/microservice"
      auto_merge: true

# Get notified when PRs are merged
notifications:
  slack:
    webhook_url: "${SLACK_WEBHOOK_URL}"
    enabled: true
```

**Environment Variables:**

```bash
# Set your tokens (add to ~/.bashrc or ~/.zshrc)
export GITHUB_TOKEN="ghp_your_token_here"
export GITLAB_TOKEN="glpat_your_token_here"
export SLACK_WEBHOOK_URL="https://hooks.slack.com/..."
```

## 📖 Documentation

Choose your path:

| 📚 Guide | 🎯 Purpose | ⏱️ Time |
|----------|------------|----------|
| **[📋 Quick Start](docs/QUICKSTART.md)** | Get running in 5 minutes | 5 min |
| **[🤖 MCP Server Setup](docs/mcp-server/MCP_SETUP.md)** | AI assistant integration | 10 min |
| **[📖 Reference](docs/REFERENCE.md)** | Complete configuration guide | 15 min |
| **[🔧 Troubleshooting](docs/TROUBLESHOOTING.md)** | Fix common issues | As needed |
| **[📧 Notifications](docs/NOTIFICATIONS.md)** | Setup Slack & email alerts | 10 min |

## 🛠️ Available Commands

**Setup Commands**

```bash
make setup-full             # Complete automated setup with wizard
make setup-wizard           # Interactive repository discovery wizard
make wizard-preview         # Preview what wizard would configure
make wizard-additive        # Add repositories to existing config
make setup-config           # Copy config.sample to config.yaml (manual)
```

**Core Commands**

```bash
make check-prs              # Check PR status across repositories
make dry-run                # Preview what would be merged (safe!)
make merge-prs              # Actually merge ready PRs
make watch                  # Monitor continuously (30s refresh)
```

**Maintenance Commands**

```bash
make validate               # Check configuration
make test                   # Test functionality
make backup-config          # Backup current configuration
make restore-config         # Restore from backup
make help                   # Show all commands
```

**MCP Server Commands**

```bash
# Build the MCP server
cd mcp-server && go build -o multi-gitter-pr-a8n-mcp .

# Then use natural language in your IDE:
# "Check all pull requests across my repositories"
# "Merge ready dependabot PRs after showing me a dry run"
# "Validate my configuration and test notifications"
```

## 🏗️ Architecture

**Traditional Command Line:**

```text
┌─────────────────┐    ┌──────────────┐    ┌─────────────────┐
│   config.yaml   │───►│ check-prs.sh │───►│   PR Status     │
│ (repositories)  │    │              │    │   Report        │
└─────────────────┘    └──────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌──────────────┐    ┌─────────────────┐
                       │ merge-prs.sh │───►│  Notifications  │
                       │              │    │ (Slack/Email)   │
                       └──────────────┘    └─────────────────┘
```

**AI Assistant Integration (MCP):**

```text
┌─────────────────┐    ┌──────────────┐    ┌─────────────────┐
│   IDE/AI Chat   │───►│  MCP Server  │───►│   Make Commands │
│ "Check all PRs" │    │   (Go)       │    │  (same scripts) │
└─────────────────┘    └──────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌──────────────┐    ┌─────────────────┐
                       │ Natural Lang │───►│  Tool Results   │
                       │  Response    │    │   & Status      │
                       └──────────────┘    └─────────────────┘
```

**Core Components:**

- **`check-prs.sh`** - Scans repositories, reports PR status
- **`merge-prs.sh`** - Merges ready PRs, sends notifications
- **`test-notifications.sh`** - Tests Slack and email setup
- **`setup-wizard.sh`** - Interactive repository discovery and configuration
- **`mcp-server/`** - Go-based MCP server for AI assistant integration
- **`Makefile`** - Convenient command interface

## 🔐 Authentication Quick Setup

### GitHub

1. **Create Token**: https://github.com/settings/tokens → `repo` + `workflow` scopes
2. **Set Variable**: `export GITHUB_TOKEN="ghp_your_token"`

### GitLab

1. **Create Token**: https://gitlab.com/-/profile/personal_access_tokens → `api` scope
2. **Set Variable**: `export GITLAB_TOKEN="glpat_your_token"`

### Bitbucket

1. **Create App Password**: https://bitbucket.org/account/settings/app-passwords/ → `Repositories: Write` + `Pull requests: Write`
2. **Set Variables**:

   ```bash
   export BITBUCKET_USERNAME="your_username"
   export BITBUCKET_APP_PASSWORD="your_app_password"
   ```

> 💡 **Tip**: Add exports to `~/.bashrc` or `~/.zshrc` for persistence

## ✨ Key Features

**🎯 Smart PR Filtering**

- Only processes PRs from trusted bots (dependabot, renovate, etc.)
- Skips PRs with labels like `do-not-merge`, `wip`
- Waits for status checks to pass

**🔀 Flexible Merge Strategies**

- `squash` - Clean commit history
- `merge` - Preserve branch structure
- `rebase` - Linear history

**📱 Notifications**

- Slack webhook integration
- Email notifications via SMTP
- Customizable message formats

**🛡️ Safe Operations**

- Dry-run mode shows what would happen
- Requires status checks to pass
- Optional approval requirements
- Rate limiting protection

## 🎯 Perfect For

- **Dependency Updates** - Auto-merge dependabot/renovate PRs
- **Multi-Repository Management** - Manage dozens of repos from one place
- **Team Automation** - Reduce manual PR review overhead
- **AI-Assisted Workflows** - Natural language PR management in your IDE
- **CI/CD Integration** - Incorporate into deployment pipelines

## 🤝 Contributing & Support

- 📖 **Documentation**: All guides in [`docs/`](docs/) directory
- 🐛 **Issues**: Found a bug? Open an issue
- 💡 **Ideas**: Suggestions welcome
- 🔧 **Development**: Run `make test` and `make lint`

**Need Help?** Start with [📋 Quick Start](docs/QUICKSTART.md) or check [🔧 Troubleshooting](docs/TROUBLESHOOTING.md)

## 🙏 Acknowledgments

This project stands on the shoulders of amazing open-source tools:

- **[multi-gitter](https://github.com/lindell/multi-gitter)** - The inspiration for this project's name and approach to bulk repository operations
- **[yq](https://github.com/mikefarah/yq)** - YAML processor that makes configuration parsing effortless
- **[jq](https://github.com/jqlang/jq)** - JSON processor for handling API responses with precision
- **[GitHub CLI (gh)](https://cli.github.com/)** - Enhanced GitHub operations and authentication
- **[shellcheck](https://github.com/koalaman/shellcheck)** - Shell script linting to keep our bash code clean
- **[Homebrew](https://brew.sh/)** - Package management for macOS dependencies
- **curl** - The reliable HTTP client powering all our API interactions

Special thanks to the **multi-gitter** project for pioneering the approach of managing operations across multiple repositories. While this tool focuses specifically on PR automation, multi-gitter's broader vision of bulk repository operations inspired our approach.

---

**Ready to get started?** 🚀 **[📋 Quick Start Guide](docs/QUICKSTART.md)**

<sub>Built for teams who want to automate PR management without cluttering every repository with GitHub Actions. Focus on code, let multi-gitter handle the merges.</sub>
