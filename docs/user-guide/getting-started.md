# Getting Started

This guide will help you set up and start using Git PR CLI to automate pull request management.

## Quick Start

### 1. Install Git PR CLI

Follow the [Installation Guide](installation.md) to install the CLI and set up authentication tokens.

### 2. Initial Setup

Run the interactive setup wizard to discover and configure your repositories:

```bash
git-pr-cli setup wizard
```

The wizard will:

- Test authentication with your Git providers
- Discover repositories from GitHub, GitLab, and Bitbucket
- Allow you to filter and select repositories
- Generate a `config.yaml` file

### 3. Validate Configuration

Verify your setup works correctly:

```bash
git-pr-cli validate --check-repos
```

This command:

- ✅ Validates configuration syntax
- ✅ Tests provider authentication
- ✅ Verifies access to configured repositories

### 4. Check Pull Requests

See what pull requests are available for merging:

```bash
git-pr-cli check --verbose
```

Example output:

```
✅ GitHub: 15 repositories, 8 PRs found, 3 ready to merge
✅ GitLab: 5 repositories, 2 PRs found, 1 ready to merge
✅ Bitbucket: 3 repositories, 1 PR found, 0 ready to merge

Ready to merge:
├── myorg/web-app: dependabot[bot] - Bump lodash from 4.17.20 to 4.17.21
├── myorg/api-service: renovate[bot] - Update Node.js to v18.17.1
├── myorg/mobile-app: dependabot[bot] - Bump react-native from 0.71.8 to 0.72.3
└── gitlab-org/my-project: renovate[bot] - Update Go modules
```

### 5. Dry Run Merge

See what would be merged without actually doing it:

```bash
git-pr-cli merge --dry-run
```

This shows you exactly what actions will be taken.

### 6. Merge Pull Requests

When you're ready, merge the PRs:

```bash
git-pr-cli merge
```

## Core Concepts

### Trusted Actors

Git PR CLI only processes PRs from trusted actors configured in your `config.yaml`:

```yaml
pr_filters:
  allowed_actors:
    - "dependabot[bot]"
    - "renovate[bot]"
    - "github-actions[bot]"
```

### Safety Features

- **Status Checks**: Waits for CI/CD to pass before merging
- **Skip Labels**: Automatically skips PRs with labels like `do-not-merge`, `wip`
- **Rate Limiting**: Respects API rate limits
- **Dry Run**: Always test with `--dry-run` first

### Merge Strategies

Configure per-repository merge strategies:

```yaml
repositories:
  github:
    - name: "myorg/web-app"
      auto_merge: true
      merge_strategy: "squash"  # squash, merge, rebase
```

## Common Workflows

### Daily Automation

Set up a cron job or GitHub Action to run daily:

```bash
# Check and merge dependency updates
git-pr-cli check --quiet && git-pr-cli merge --auto-approve
```

### Continuous Monitoring

Watch for new PRs continuously:

```bash
git-pr-cli watch --interval=5m
```

### Selective Merging

Merge only specific repositories:

```bash
git-pr-cli merge --repos="myorg/critical-app,myorg/web-service"
```

### Statistics and Reporting

Get insights into your repository activity:

```bash
git-pr-cli stats --detailed
```

## Configuration Management

### Backup Configuration

```bash
git-pr-cli setup backup
```

### Add More Repositories

```bash
git-pr-cli setup wizard --additive
```

### Preview Configuration Changes

```bash
git-pr-cli setup preview
```

## Testing and Validation

### Validate Specific Provider

```bash
git-pr-cli validate --provider=github
```

### Check Single Repository

```bash
git-pr-cli check --repos="myorg/specific-repo"
```

## Next Steps

- [Configuration Reference](configuration.md) - Detailed configuration options
- [Command Reference](commands/) - Complete command documentation
- [MCP Server Guide](../mcp-server/) - AI assistant integration setup

## Tips

1. **Start Small**: Begin with a few non-critical repositories
2. **Use Dry Run**: Always test with `--dry-run` before real merges
3. **Monitor First**: Use `watch` mode to understand PR patterns
4. **Configure Notifications**: Set up Slack/email for merge notifications
5. **Regular Validation**: Run `validate` periodically to catch issues early
