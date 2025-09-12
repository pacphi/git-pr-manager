# Quick Start Guide

Get up and running with multi-gitter automation in 5 minutes.

## Prerequisites

- macOS or Linux system
- Bash shell
- Internet connection
- Git provider tokens (GitHub, GitLab, or Bitbucket)

## 🎯 Automated Setup (Recommended)

### 1. Complete Setup with Wizard

```bash
# All-in-one command: dependencies + configuration wizard
make setup-full
```

This will:

- Install all dependencies (`yq`, `jq`, `gh`, `curl`)
- Copy `config.sample` to `config.yaml`
- Launch the interactive repository discovery wizard
- Validate your authentication tokens
- Discover repositories from GitHub, GitLab, and Bitbucket
- Generate your complete configuration

### 2. Set Authentication Tokens First

Before running the wizard, set your authentication tokens:

```bash
# GitHub (required for GitHub repos)
export GITHUB_TOKEN="ghp_your_token_here"

# GitLab (required for GitLab repos)
export GITLAB_TOKEN="glpat-your_token_here"

# Bitbucket (required for Bitbucket repos)
export BITBUCKET_USERNAME="your_username"
export BITBUCKET_APP_PASSWORD="your_app_password"
```

**💡 Pro Tip**: Add these to your `~/.bashrc` or `~/.zshrc` for persistence.

### 3. Wizard Features

The setup wizard offers:

- **Repository Discovery**: Finds all your repositories automatically
- **Smart Filtering**: Filter by visibility, owner, activity, name patterns
- **Interactive Selection**: Choose all repositories, by provider, or selectively
- **Merge Strategy Configuration**: Set default merge strategies (squash, merge, rebase)
- **Preview Mode**: See what would be configured without making changes
- **Additive Mode**: Add repositories to existing configuration

## ⚙️ Manual Setup (Alternative)

### 1. Install Dependencies

```bash
make install
```

This installs `yq`, `jq`, and GitHub CLI (`gh`).

### 2. Configure Repositories Manually

```bash
make setup-config  # Creates config.yaml from config.sample
```

Then edit `config.yaml` to add your repositories:

```yaml
repositories:
  github:
    - name: "owner/repo-name"
      url: "https://github.com/owner/repo-name"
      auto_merge: true

  gitlab:
    - name: "group/project-name"
      url: "https://gitlab.com/group/project-name"
      auto_merge: true
```

## 4. Test Your Setup

```bash
# Validate configuration
make validate

# Check what PRs are available
make check-prs
```

## 5. Merge PRs Safely

```bash
# See what would be merged (recommended first!)
make dry-run

# Actually merge the PRs
make merge-prs
```

## 🎉 That's it!

You're now ready to automate PR management across multiple repositories.

## Next Steps

- **[📖 Full Reference](REFERENCE.md)** - Complete configuration options
- **[🔧 Troubleshooting](TROUBLESHOOTING.md)** - Fix common issues
- **[📧 Notifications](NOTIFICATIONS.md)** - Setup Slack/email alerts

## Wizard Usage Examples

### Preview Configuration Without Changes

```bash
make wizard-preview  # See what repositories would be configured
```

### Add Repositories to Existing Configuration

```bash
make wizard-additive  # Merge new discoveries with existing config
```

### Run Wizard Again with Different Filters

```bash
make setup-wizard    # Discover repositories with different filtering options
```

## Common First-Time Tasks

### Add More Repositories

**With Wizard** (Recommended):

```bash
make wizard-additive  # Discover and add new repositories
```

**Manual Method**:

```bash
# Edit the config file
nano config.yaml

# Validate your changes
make validate
```

### Monitor Continuously

```bash
# Watch for new PRs (refreshes every 30 seconds)
make watch
```

### Get Repository Stats

```bash
make stats
```

## Need Help?

- Run `make help` for all available commands
- Check `make test` to verify everything is working
- See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues
