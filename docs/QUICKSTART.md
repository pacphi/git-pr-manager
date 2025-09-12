# Quick Start Guide

Get up and running with multi-gitter automation in 5 minutes.

## Prerequisites

- macOS or Linux system
- Bash shell
- Internet connection

## 1. Install Dependencies

```bash
make install
```

This installs `yq`, `jq`, and GitHub CLI (`gh`).

## 2. Setup Authentication

### Automatic Setup

```bash
make setup
```

### Manual Setup

Set these environment variables:

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

## 3. Configure Repositories

Edit `config.yaml` to add your repositories:

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

## Common First-Time Tasks

### Add More Repositories

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
