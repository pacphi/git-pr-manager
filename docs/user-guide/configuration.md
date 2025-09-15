# Configuration Reference

Git PR CLI uses a YAML configuration file (`config.yaml`) to define repositories, filters, and settings.

## Configuration File Location

Git PR CLI looks for configuration in this order:

1. File specified by `--config` flag
2. `config.yaml` in current directory
3. `$HOME/.git-pr-cli/config.yaml`

## Complete Configuration Example

```yaml
# Authentication settings (use environment variables)
auth:
  github_token: ${GITHUB_TOKEN}
  gitlab_token: ${GITLAB_TOKEN}
  gitlab_url: ${GITLAB_URL:-https://gitlab.com}
  bitbucket_username: ${BITBUCKET_USERNAME}
  bitbucket_app_password: ${BITBUCKET_APP_PASSWORD}
  bitbucket_workspace: ${BITBUCKET_WORKSPACE}

# Pull request filtering rules
pr_filters:
  allowed_actors:
    - "dependabot[bot]"
    - "renovate[bot]"
    - "github-actions[bot]"
  skip_labels:
    - "do-not-merge"
    - "wip"
    - "hold"
    - "draft"
  require_status_checks: true
  require_up_to_date: false
  max_age_days: 30

# Repository configuration per provider
repositories:
  github:
    - name: "myorg/web-app"
      auto_merge: true
      merge_strategy: "squash"
      labels: ["dependencies"]
    - name: "myorg/api-service"
      auto_merge: true
      merge_strategy: "merge"
      require_reviews: 1
    - name: "myorg/*"  # Wildcard pattern
      auto_merge: false
      merge_strategy: "squash"

  gitlab:
    - name: "group/project"
      auto_merge: true
      merge_strategy: "merge"
      target_branch: "main"
    - name: "another-group/*"
      auto_merge: false

  bitbucket:
    - name: "workspace/repository"
      auto_merge: true
      merge_strategy: "squash"

# Notification settings
notifications:
  slack:
    webhook_url: ${SLACK_WEBHOOK_URL}
    channel: "#deployments"
    mention_on_failure: true
    template: |
      {{if .Success}}✅{{else}}❌{{end}} Merged {{len .MergedPRs}} PRs across {{len .Repositories}} repositories

  email:
    smtp_host: ${SMTP_HOST}
    smtp_port: ${SMTP_PORT:-587}
    smtp_username: ${SMTP_USERNAME}
    smtp_password: ${SMTP_PASSWORD}
    from: ${EMAIL_FROM}
    to: ["devops@company.com"]
    subject: "PR Automation Report - {{.Date}}"

# Concurrency and rate limiting
settings:
  max_concurrent_requests: 5
  request_timeout: "30s"
  retry_attempts: 3
  retry_delay: "1s"
  rate_limit_requests: 100
  rate_limit_duration: "1h"
```

## Section Details

### Authentication (`auth`)

All authentication uses environment variables for security:

```yaml
auth:
  github_token: ${GITHUB_TOKEN}          # Required for GitHub
  gitlab_token: ${GITLAB_TOKEN}          # Required for GitLab
  gitlab_url: ${GITLAB_URL}              # Optional, defaults to gitlab.com
  bitbucket_username: ${BITBUCKET_USERNAME}      # Required for Bitbucket
  bitbucket_app_password: ${BITBUCKET_APP_PASSWORD}  # Required for Bitbucket
  bitbucket_workspace: ${BITBUCKET_WORKSPACE}    # Optional
```

### PR Filters (`pr_filters`)

Control which pull requests are processed:

```yaml
pr_filters:
  allowed_actors:                    # Only PRs from these users/bots
    - "dependabot[bot]"
    - "renovate[bot]"
    - "github-actions[bot]"

  skip_labels:                       # Skip PRs with these labels
    - "do-not-merge"
    - "wip"
    - "draft"
    - "breaking-change"

  require_status_checks: true        # Wait for CI/CD to pass
  require_up_to_date: false         # Require branch to be up-to-date
  max_age_days: 30                  # Skip PRs older than this
  min_approvals: 0                  # Minimum required reviews
```

### Repositories

Configure repositories per provider:

#### Repository Options

```yaml
repositories:
  github:
    - name: "owner/repo"             # Repository identifier
      auto_merge: true               # Enable automatic merging
      merge_strategy: "squash"       # squash, merge, or rebase
      target_branch: "main"          # Target branch (default: main)
      labels: ["dependencies"]       # Required labels
      require_reviews: 1             # Minimum reviews needed
      skip_labels: ["hold"]          # Additional skip labels
      max_age_hours: 48             # Repository-specific age limit
```

#### Merge Strategies

- **`squash`**: Squash commits into single commit (recommended for dependency updates)
- **`merge`**: Standard merge commit (preserves commit history)
- **`rebase`**: Rebase and merge (clean linear history)

#### Wildcard Patterns

Use wildcards to match multiple repositories:

```yaml
repositories:
  github:
    - name: "myorg/*"               # All repositories in organization
      auto_merge: false
    - name: "myorg/web-*"          # All repositories starting with 'web-'
      auto_merge: true
      merge_strategy: "squash"
```

### Notifications

#### Slack Notifications

```yaml
notifications:
  slack:
    webhook_url: ${SLACK_WEBHOOK_URL}
    channel: "#deployments"           # Optional: override webhook channel
    username: "Git PR Bot"           # Optional: custom username
    icon_emoji: ":robot_face:"       # Optional: custom emoji
    mention_on_failure: true         # Mention @channel on failures
    mention_users: ["@devops"]       # Users to mention
    template: |                      # Optional: custom message template
      {{if .Success}}✅{{else}}❌{{end}} {{.Summary}}
```

#### Email Notifications

```yaml
notifications:
  email:
    smtp_host: ${SMTP_HOST}          # SMTP server
    smtp_port: ${SMTP_PORT:-587}     # Port (default: 587)
    smtp_username: ${SMTP_USERNAME}  # SMTP username
    smtp_password: ${SMTP_PASSWORD}  # SMTP password
    from: ${EMAIL_FROM}              # From address
    to: ["team@company.com"]         # Recipients
    subject: "PR Report - {{.Date}}" # Subject template
    html_template: "custom.html"     # Optional: custom HTML template
```

### Settings

Performance and reliability settings:

```yaml
settings:
  max_concurrent_requests: 5         # Parallel API requests
  request_timeout: "30s"            # Individual request timeout
  retry_attempts: 3                 # Failed request retries
  retry_delay: "1s"                 # Delay between retries
  rate_limit_requests: 100          # Requests per duration
  rate_limit_duration: "1h"         # Rate limit window
  dry_run: false                    # Global dry-run mode
  log_level: "info"                 # debug, info, warn, error
```

## Configuration Validation

Validate your configuration:

```bash
# Basic validation
git-pr-cli validate

# Validate with authentication test
git-pr-cli validate --check-auth

# Validate with repository access test
git-pr-cli validate --check-repos

# Validate specific provider
git-pr-cli validate --provider=github
```

## Environment Variables

Required environment variables:

```bash
# GitHub
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxx"

# GitLab
export GITLAB_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
export GITLAB_URL="https://gitlab.company.com"  # Optional

# Bitbucket
export BITBUCKET_USERNAME="username"
export BITBUCKET_APP_PASSWORD="xxxxxxxxxxxx"
export BITBUCKET_WORKSPACE="workspace"  # Optional

# Notifications (optional)
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/..."
export SMTP_HOST="smtp.gmail.com"
export SMTP_USERNAME="notifications@company.com"
export SMTP_PASSWORD="app-password"
export EMAIL_FROM="Git PR Bot <notifications@company.com>"
```

## Configuration Management

### Generate Configuration

```bash
# Interactive wizard
git-pr-cli setup wizard

# Add to existing config
git-pr-cli setup wizard --additive

# Preview without saving
git-pr-cli setup preview
```

### Backup and Restore

```bash
# Backup current config
git-pr-cli setup backup

# Restore from backup
git-pr-cli setup restore
```

### Template Management

```bash
# Copy sample config
cp config.sample config.yaml

# Validate and migrate
git-pr-cli setup migrate
```

## Advanced Configuration

### Custom Templates

Create custom notification templates:

```yaml
notifications:
  slack:
    template: |
      🚀 *PR Automation Report*
      {{range .MergedPRs}}
      • {{.Repository}}: {{.Title}} by {{.Author}}
      {{end}}

      📊 *Summary*
      • Total PRs merged: {{len .MergedPRs}}
      • Repositories updated: {{len .Repositories}}
      • Execution time: {{.Duration}}
```

### Conditional Configuration

Use environment-specific configs:

```bash
# Development
git-pr-cli --config config.dev.yaml check

# Production
git-pr-cli --config config.prod.yaml merge
```

### Provider-Specific Settings

Override settings per provider:

```yaml
providers:
  github:
    rate_limit_requests: 5000
    request_timeout: "60s"
  gitlab:
    rate_limit_requests: 300
    request_timeout: "30s"
  bitbucket:
    rate_limit_requests: 1000
    request_timeout: "45s"
```
