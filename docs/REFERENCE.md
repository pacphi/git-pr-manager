# Reference Documentation

Complete configuration reference and advanced usage patterns.

## Table of Contents

- [Setup Wizard Reference](#setup-wizard-reference)
- [Configuration File Reference](#configuration-file-reference)
- [Environment Variables](#environment-variables)
- [Command Reference](#command-reference)
- [Script Reference](#script-reference)
- [Authentication Guide](#authentication-guide)
- [Advanced Usage Patterns](#advanced-usage-patterns)

## Setup Wizard Reference

### Wizard Commands

| Command | Description | Options |
|---------|-------------|---------|
| `make setup-full` | Complete automated setup | Installs dependencies + runs wizard |
| `make setup-wizard` | Interactive repository discovery | Default mode |
| `make wizard-preview` | Preview configuration | `--preview` mode, no changes made |
| `make wizard-additive` | Add to existing configuration | `--additive` mode |

### Direct Script Usage

```bash
./setup-wizard.sh [OPTIONS]

Options:
  -h, --help              Show help message
  -c, --config FILE       Configuration file (default: config.yaml)
  -p, --preview           Preview mode - show what would be configured
  -a, --additive          Add to existing configuration instead of replacing
  --backup-dir DIR        Directory for configuration backups (default: backups)

Environment Variables:
  GITHUB_TOKEN            GitHub personal access token
  GITLAB_TOKEN            GitLab personal access token
  GITLAB_URL              GitLab instance URL (default: https://gitlab.com)
  BITBUCKET_USERNAME      Bitbucket username
  BITBUCKET_APP_PASSWORD  Bitbucket app password

Examples:
  ./setup-wizard.sh                      # Run interactive wizard
  ./setup-wizard.sh --preview            # Preview what would be configured
  ./setup-wizard.sh --additive           # Add repositories to existing config
  ./setup-wizard.sh -c custom.yaml       # Use custom configuration file
```

### Wizard Capabilities

**Repository Discovery**:
- Automatically discovers repositories from authenticated Git providers
- Supports GitHub, GitLab, and Bitbucket simultaneously
- Fetches personal repositories, organization/group repositories, and workspace repositories

**Smart Filtering**:
- Filter by visibility (public, private, internal)
- Filter by owner type (personal, organization/group/workspace)
- Filter by activity (last updated: 30/90/365 days)
- Filter by name patterns (wildcards supported)
- Custom filters for forks, archived projects, language, etc.

**Interactive Selection**:
- Choose all repositories for auto-merge
- Select repositories by provider
- Interactive repository selection (planned feature)
- Configure merge strategies per selection

**Safety Features**:
- Preview mode shows what would be configured without making changes
- Automatic backup of existing configurations
- Comprehensive validation of authentication tokens
- Error handling and progress tracking

## Configuration File Reference

### Complete Example

```yaml
# Global configuration
config:
  default_merge_strategy: "squash"  # squash, merge, rebase
  auto_merge:
    enabled: true
    wait_for_checks: true
    require_approval: true
  pr_filters:
    allowed_actors:
      - "dependabot[bot]"
      - "renovate[bot]"
      - "github-actions[bot]"
    skip_labels:
      - "do-not-merge"
      - "wip"
      - "draft"
      - "breaking-change"

# Repository configurations
repositories:
  github:
    - name: "owner/repo1"
      url: "https://github.com/owner/repo1"
      provider: "github"
      auth_type: "token"
      merge_strategy: "squash"
      auto_merge: true

    - name: "owner/repo2"
      url: "https://github.com/owner/repo2"
      provider: "github"
      auth_type: "token"
      merge_strategy: "merge"
      auto_merge: false  # Disabled for this repo

  gitlab:
    - name: "group/project1"
      url: "https://gitlab.com/group/project1"
      provider: "gitlab"
      auth_type: "token"
      merge_strategy: "squash"
      auto_merge: true

    - name: "group/project2"
      url: "https://my-gitlab.company.com/group/project2"
      provider: "gitlab"
      auth_type: "token"
      merge_strategy: "merge"
      auto_merge: true

  bitbucket:
    - name: "workspace/repo1"
      url: "https://bitbucket.org/workspace/repo1"
      provider: "bitbucket"
      auth_type: "app-password"
      merge_strategy: "squash"
      auto_merge: true

# Authentication configuration
auth:
  github:
    token: "${GITHUB_TOKEN}"

  gitlab:
    token: "${GITLAB_TOKEN}"
    url: "https://gitlab.com"  # For self-hosted GitLab

  bitbucket:
    username: "${BITBUCKET_USERNAME}"
    app_password: "${BITBUCKET_APP_PASSWORD}"
    workspace: "${BITBUCKET_WORKSPACE}"

# Notification settings
notifications:
  slack:
    webhook_url: "${SLACK_WEBHOOK_URL}"
    channel: "#deployments"
    enabled: false

  email:
    smtp_server: "smtp.gmail.com"
    smtp_port: 587
    username: "${EMAIL_USERNAME}"
    password: "${EMAIL_PASSWORD}"
    recipient: "${EMAIL_RECIPIENT}"
    enabled: false
```

### Configuration Sections

#### Global Config (`config`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `default_merge_strategy` | string | `"squash"` | Default merge strategy: `squash`, `merge`, `rebase` |
| `auto_merge.enabled` | boolean | `true` | Enable auto-merge globally |
| `auto_merge.wait_for_checks` | boolean | `true` | Wait for CI/CD checks to pass |
| `auto_merge.require_approval` | boolean | `true` | Require PR approval before merge |
| `pr_filters.allowed_actors` | array | `[]` | Only process PRs from these users/bots |
| `pr_filters.skip_labels` | array | `[]` | Skip PRs with these labels |

#### Repository Config

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | ✅ | Repository identifier (owner/repo format) |
| `url` | string | ✅ | Full repository URL |
| `provider` | string | ✅ | Git provider: `github`, `gitlab`, `bitbucket` |
| `auth_type` | string | ✅ | Authentication type: `token`, `app-password` |
| `merge_strategy` | string | ❌ | Override global merge strategy |
| `auto_merge` | boolean | ❌ | Override global auto-merge setting |

#### Notification Config

| Field | Type | Description |
|-------|------|-------------|
| `slack.webhook_url` | string | Slack webhook URL |
| `slack.channel` | string | Target Slack channel |
| `slack.enabled` | boolean | Enable Slack notifications |
| `email.smtp_server` | string | SMTP server hostname |
| `email.smtp_port` | number | SMTP port (usually 587 or 465) |
| `email.username` | string | SMTP username |
| `email.password` | string | SMTP password/app password |
| `email.recipient` | string | Email recipient (defaults to username) |
| `email.enabled` | boolean | Enable email notifications |

## Environment Variables

### Required Variables

#### GitHub

```bash
export GITHUB_TOKEN="ghp_your_token_here"
```

#### GitLab

```bash
export GITLAB_TOKEN="glpat-your_token_here"
```

#### Bitbucket

```bash
export BITBUCKET_USERNAME="your_username"
export BITBUCKET_APP_PASSWORD="your_app_password"
export BITBUCKET_WORKSPACE="your_workspace"  # Optional
```

### Optional Variables

#### Notifications

```bash
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/..."
export EMAIL_USERNAME="your_email@example.com"
export EMAIL_PASSWORD="your_app_password"
export EMAIL_RECIPIENT="recipient@example.com"  # Optional
```

#### Behavior Control

```bash
export CONFIG_FILE="custom-config.yaml"  # Default: config.yaml
export DRY_RUN="true"                    # Enable dry-run mode
export FORCE="true"                      # Force merge non-mergeable PRs
export OUTPUT_FORMAT="json"             # Output format: table, json
```

## Command Reference

### Core Commands

| Command | Description | Example |
|---------|-------------|---------|
| `make help` | Show help message | `make help` |
| `make check-prs` | Check PR status | `make check-prs` |
| `make merge-prs` | Merge ready PRs | `make merge-prs` |
| `make dry-run` | Show what would be merged | `make dry-run` |
| `make status` | Alias for check-prs | `make status` |

### Setup Commands

| Command | Description | Example |
|---------|-------------|---------|
| `make setup-full` | Complete automated setup with wizard | `make setup-full` |
| `make setup-wizard` | Interactive repository discovery wizard | `make setup-wizard` |
| `make wizard-preview` | Preview what wizard would configure | `make wizard-preview` |
| `make wizard-additive` | Add repositories to existing config | `make wizard-additive` |
| `make install` | Install dependencies | `make install` |
| `make setup-config` | Copy config.sample to config.yaml | `make setup-config` |
| `make validate` | Validate config file | `make validate` |
| `make test` | Run functionality tests | `make test` |

### Provider-Specific Commands

| Command | Description | Example |
|---------|-------------|---------|
| `make check-github` | Check only GitHub repos | `make check-github` |
| `make check-gitlab` | Check only GitLab repos | `make check-gitlab` |
| `make check-bitbucket` | Check only Bitbucket repos | `make check-bitbucket` |

### Utility Commands

| Command | Description | Example |
|---------|-------------|---------|
| `make stats` | Show repository statistics | `make stats` |
| `make watch` | Continuously monitor PRs | `make watch` |
| `make clean` | Clean temporary files | `make clean` |
| `make debug` | Run in debug mode | `make debug` |
| `make lint` | Lint shell scripts | `make lint` |

### Configuration Commands

| Command | Description | Example |
|---------|-------------|---------|
| `make config-template` | Create config template | `make config-template` |
| `make backup-config` | Backup configuration | `make backup-config` |
| `make restore-config` | Restore from backup | `make restore-config` |

## Script Reference

### check-prs.sh

Check PR status across repositories.

```bash
./check-prs.sh [OPTIONS]

Options:
  -h, --help          Show help
  -c, --config FILE   Config file (default: config.yaml)
  -f, --format FORMAT Output format: table, json (default: table)

Examples:
  ./check-prs.sh                    # Default table output
  ./check-prs.sh -f json            # JSON output
  ./check-prs.sh -c custom.yaml     # Custom config
```

### merge-prs.sh

Merge ready PRs across repositories.

```bash
./merge-prs.sh [OPTIONS]

Options:
  -h, --help          Show help
  -c, --config FILE   Config file (default: config.yaml)
  -n, --dry-run       Show what would be done
  -f, --force         Force merge non-mergeable PRs

Examples:
  ./merge-prs.sh                    # Merge ready PRs
  ./merge-prs.sh --dry-run          # Preview merges
  ./merge-prs.sh -c custom.yaml     # Custom config
```

### test-notifications.sh

Test notification functionality.

```bash
./test-notifications.sh [OPTIONS]

Options:
  -h, --help          Show help
  -c, --config FILE   Config file (default: config.yaml)
  -s, --slack         Test only Slack
  -e, --email         Test only email
  --config-test       Test config only

Examples:
  ./test-notifications.sh           # Test all notifications
  ./test-notifications.sh --slack   # Test only Slack
  ./test-notifications.sh --email   # Test only email
```

## Authentication Guide

### GitHub Authentication

1. **Create Personal Access Token**:
   - Go to https://github.com/settings/tokens
   - Click "Generate new token (classic)"
   - Select scopes: `repo`, `workflow`
   - Copy the token

2. **Set Environment Variable**:

   ```bash
   export GITHUB_TOKEN="ghp_your_token_here"
   ```

3. **For Organizations**:
   - Ensure token has access to organization repositories
   - May need to enable SSO if organization uses SAML

### GitLab Authentication

1. **Create Personal Access Token**:
   - Go to https://gitlab.com/-/profile/personal_access_tokens
   - Set name and expiration
   - Select scopes: `api`, `read_repository`, `write_repository`
   - Copy the token

2. **Set Environment Variable**:

   ```bash
   export GITLAB_TOKEN="glpat-your_token_here"
   ```

3. **For Self-Hosted GitLab**:
   - Update `auth.gitlab.url` in config.yaml
   - Use your GitLab instance URL

### Bitbucket Authentication

1. **Create App Password**:

   - Go to https://bitbucket.org/account/settings/app-passwords/
   - Click "Create app password"
   - Select permissions: `Repositories: Read, Write`, `Pull requests: Read, Write`
   - Copy the password

2. **Set Environment Variables**:

   ```bash
   export BITBUCKET_USERNAME="your_username"
   export BITBUCKET_APP_PASSWORD="your_app_password"
   ```

## Advanced Usage Patterns

### Multiple Configuration Files

Use different configs for different environments:

```bash
# Development repos
make CONFIG_FILE=config-dev.yaml check-prs

# Production repos
make CONFIG_FILE=config-prod.yaml merge-prs

# Staging repos
make CONFIG_FILE=config-staging.yaml dry-run
```

### Selective Repository Processing

Process only specific providers:

```bash
# Only GitHub repos
make check-github

# Only GitLab repos
make check-gitlab | grep "ready_to_merge.*yes"

# Chain commands
make check-bitbucket && make merge-prs
```

### Automated Workflows

Create automated workflows with cron:

```bash
# Check PRs every hour
0 * * * * cd /path/to/multi-gitter && make check-prs

# Merge PRs twice daily
0 9,17 * * * cd /path/to/multi-gitter && make merge-prs

# Send daily summary
0 8 * * * cd /path/to/multi-gitter && make stats | mail -s "PR Summary" admin@company.com
```

### Integration with CI/CD

Use in CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Check and merge PRs
  run: |
    make validate
    make check-prs
    make dry-run
    make merge-prs
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    GITLAB_TOKEN: ${{ secrets.GITLAB_TOKEN }}
```

### Custom Filtering

Filter PRs programmatically:

```bash
# Only dependabot PRs ready to merge
make check-prs-json | jq '.[] | select(.author == "dependabot[bot]" and .ready_to_merge == true)'

# PRs by specific author
make check-prs-json | jq '.[] | select(.author == "renovate[bot]")'

# PRs with passing checks
make check-prs-json | jq '.[] | select(.checks_status == "passing")'
```

### Batch Operations

Process repositories in batches:

```bash
# Process first 5 repos only
head -5 repo-list.txt | while read repo; do
  echo "Processing $repo"
  # Custom processing logic
done

# Process repos by provider
yq '.repositories.github[].name' config.yaml | while read repo; do
  echo "GitHub repo: $repo"
done
```

### Monitoring and Alerting

Set up monitoring:

```bash
# Check for failed merges
if ! make merge-prs; then
  echo "Some merges failed!" | mail -s "Merge Alert" admin@company.com
fi

# Monitor PR counts
PR_COUNT=$(make check-prs-json | jq length)
if [ "$PR_COUNT" -gt 10 ]; then
  echo "$PR_COUNT PRs pending review" | slack-notify
fi
```

### Error Handling

Robust error handling patterns:

```bash
# Retry on failure
for i in {1..3}; do
  if make merge-prs; then
    break
  else
    echo "Attempt $i failed, retrying in 30 seconds..."
    sleep 30
  fi
done

# Conditional execution
make validate && make check-prs && make merge-prs || {
  echo "Pipeline failed at step: $?"
  exit 1
}
```
