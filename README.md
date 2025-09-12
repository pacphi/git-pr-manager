# Multi-Gitter Pull-Request Automation

🤖 **Automate pull request management across multiple Git repositories**

Manage PR automation centrally through YAML configuration. Watch repositories, check PR status, and automatically merge ready PRs across GitHub, GitLab, and Bitbucket.

## ✨ Why Multi-Gitter?

- ✅ **Central Configuration** - Manage all repositories from one YAML file
- 🔄 **Multi-Provider Support** - GitHub, GitLab, Bitbucket in one tool
- 🎯 **Smart Filtering** - Only merge PRs from trusted bots (dependabot, renovate)
- 🛡️ **Safe by Default** - Dry-run mode, status checks, approval requirements
- 📱 **Notifications** - Slack and email alerts for merged PRs
- ⚡ **Simple CLI** - Easy-to-use commands for all operations
- 🖥️ **Cross-Platform** - Works on macOS and Linux with automatic dependency management

## 🚀 Quick Start (5 minutes)

```bash
# 1. Install dependencies
make install

# 2. Create your configuration file
make setup-config

# 3. Edit config.yaml to add your repositories
# (This copies config.sample to config.yaml for you to customize)

# 4. Setup authentication
make setup

# 5. Check what PRs are ready
make check-prs

# 6. See what would be merged (safe!)
make dry-run

# 7. Actually merge the PRs
make merge-prs
```

**📚 New here?** Start with the **[Quick Start Guide](docs/QUICKSTART.md)** →

## 📋 Example Configuration

**config.sample** → **config.yaml** - Copy the sample to create your configuration:

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
| **[📖 Reference](docs/REFERENCE.md)** | Complete configuration guide | 15 min |
| **[🔧 Troubleshooting](docs/TROUBLESHOOTING.md)** | Fix common issues | As needed |
| **[📧 Notifications](docs/NOTIFICATIONS.md)** | Setup Slack & email alerts | 10 min |

## 🛠️ Available Commands

```bash
# First-time setup
make setup-config           # Copy config.sample to config.yaml
make setup                  # Install deps, config, and auth

# Core commands
make check-prs              # Check PR status across repositories
make dry-run                # Preview what would be merged (safe!)
make merge-prs              # Actually merge ready PRs
make watch                  # Monitor continuously (30s refresh)

# Maintenance commands
make validate               # Check configuration
make test                   # Test functionality

# Get help anytime
make help                   # Show all commands
```

## 🏗️ Architecture

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

**Core Scripts:**

- **`check-prs.sh`** - Scans repositories, reports PR status
- **`merge-prs.sh`** - Merges ready PRs, sends notifications
- **`test-notifications.sh`** - Tests Slack and email setup
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
- **CI/CD Integration** - Incorporate into deployment pipelines

## 🤝 Contributing & Support

- 📖 **Documentation**: All guides in [`docs/`](docs/) directory
- 🐛 **Issues**: Found a bug? Open an issue
- 💡 **Ideas**: Suggestions welcome
- 🔧 **Development**: Run `make test` and `make lint`

**Need Help?** Start with [📋 Quick Start](docs/QUICKSTART.md) or check [🔧 Troubleshooting](docs/TROUBLESHOOTING.md)

---

**Ready to get started?** 🚀 **[📋 Quick Start Guide](docs/QUICKSTART.md)**

<sub>Built for teams who want to automate PR management without cluttering every repository with GitHub Actions. Focus on code, let multi-gitter handle the merges.</sub>
