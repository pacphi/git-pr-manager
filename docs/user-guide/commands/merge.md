# merge - Merge Ready Pull Requests

The `merge` command automatically merges pull requests that meet all configured criteria and requirements.

## Synopsis

```bash
git-pr-cli merge [flags]
```

## Description

The merge command performs the following operations:

1. **Discovery**: Finds ready-to-merge PRs using same logic as `check`
2. **Validation**: Verifies each PR still meets merge requirements
3. **Execution**: Merges PRs using configured strategy for each repository
4. **Notification**: Sends notifications if configured
5. **Reporting**: Provides summary of merge results

## Options

```bash
      --auto-approve            Skip confirmation prompts
      --dry-run                 Show what would be merged without actually merging
  -h, --help                    help for merge
      --max-merges int          Maximum number of PRs to merge (default unlimited)
      --output string           Output format (text, json, yaml) (default "text")
      --provider strings        Merge only from specific providers (github,gitlab,bitbucket)
      --repos strings           Merge only from specific repositories (comma-separated)
      --strategy string         Override merge strategy for all repos (merge,squash,rebase)
```

## Examples

### Basic Usage

```bash
# Dry run first (recommended)
git-pr-cli merge --dry-run

# Interactive merge (asks for confirmation)
git-pr-cli merge

# Auto-approve all merges
git-pr-cli merge --auto-approve

# Limit number of merges
git-pr-cli merge --max-merges=5 --auto-approve
```

### Provider and Repository Selection

```bash
# Merge only GitHub PRs
git-pr-cli merge --provider=github

# Merge from specific repositories
git-pr-cli merge --repos="owner/repo1,owner/repo2"

# Combine provider and repository filters
git-pr-cli merge --provider=github --repos="myorg/*"
```

### Merge Strategy Override

```bash
# Force squash merge for all PRs
git-pr-cli merge --strategy=squash

# Force merge commit for all PRs
git-pr-cli merge --strategy=merge

# Use rebase strategy for all PRs
git-pr-cli merge --strategy=rebase
```

## Sample Output

### Dry Run Output

```text
🔍 Dry Run Mode - No actual merges will be performed

Found 4 PRs ready to merge:

GitHub (myorg/web-app):
├── PR #123: dependabot[bot] - Bump lodash from 4.17.20 to 4.17.21
│   Strategy: squash │ Target: main │ Status: ✅ Ready
└── PR #124: renovate[bot] - Update Node.js to v18.17.1
    Strategy: squash │ Target: main │ Status: ✅ Ready

GitLab (group/project):
└── MR !45: renovate[bot] - Update Go modules
    Strategy: merge │ Target: main │ Status: ✅ Ready

Bitbucket (workspace/repo):
└── PR #67: dependabot[bot] - Bump axios from 0.21.1 to 0.27.2
    Strategy: squash │ Target: develop │ Status: ✅ Ready

Summary:
- Total PRs to merge: 4
- GitHub: 2 PRs
- GitLab: 1 MR
- Bitbucket: 1 PR

Use --auto-approve to skip confirmation prompts
```

### Interactive Merge

```text
Found 4 PRs ready to merge:

GitHub (myorg/web-app):
├── PR #123: dependabot[bot] - Bump lodash from 4.17.20 to 4.17.21
└── PR #124: renovate[bot] - Update Node.js to v18.17.1

Proceed with merging 4 PRs? [y/N]: y

Merging PRs...
✅ Merged GitHub PR #123 (myorg/web-app): Bump lodash from 4.17.20 to 4.17.21
✅ Merged GitHub PR #124 (myorg/web-app): Update Node.js to v18.17.1
✅ Merged GitLab MR !45 (group/project): Update Go modules
❌ Failed to merge Bitbucket PR #67 (workspace/repo): Merge conflict detected

Summary:
- Successfully merged: 3 PRs
- Failed to merge: 1 PR
- Total time: 12.3s

Notifications sent to:
- Slack: #deployments
- Email: devops@company.com
```

### JSON Output

```json
{
  "timestamp": "2025-01-15T10:30:00Z",
  "dry_run": false,
  "summary": {
    "total_ready": 4,
    "successfully_merged": 3,
    "failed_to_merge": 1,
    "skipped": 0,
    "execution_time": "12.3s"
  },
  "results": [
    {
      "repository": "myorg/web-app",
      "provider": "github",
      "number": 123,
      "title": "Bump lodash from 4.17.20 to 4.17.21",
      "author": "dependabot[bot]",
      "status": "merged",
      "merge_strategy": "squash",
      "merge_commit": "a1b2c3d",
      "url": "https://github.com/myorg/web-app/pull/123"
    },
    {
      "repository": "workspace/repo",
      "provider": "bitbucket",
      "number": 67,
      "title": "Bump axios from 0.21.1 to 0.27.2",
      "author": "dependabot[bot]",
      "status": "failed",
      "error": "Merge conflict detected",
      "url": "https://bitbucket.org/workspace/repo/pull-requests/67"
    }
  ],
  "notifications": {
    "slack": {
      "sent": true,
      "channel": "#deployments"
    },
    "email": {
      "sent": true,
      "recipients": ["devops@company.com"]
    }
  }
}
```

## Merge Strategies

### Squash Merge (Default)

Combines all commits from the PR into a single commit:

```yaml
repositories:
  github:
    - name: "owner/repo"
      merge_strategy: "squash"
```

**Advantages:**

- Clean, linear history
- Single commit per feature/fix
- Removes intermediate commits

### Merge Commit

Creates a merge commit preserving the PR branch history:

```yaml
repositories:
  github:
    - name: "owner/repo"
      merge_strategy: "merge"
```

**Advantages:**

- Preserves complete commit history
- Shows branch structure
- Maintains author information

### Rebase Merge

Rebases PR commits onto target branch without merge commit:

```yaml
repositories:
  github:
    - name: "owner/repo"
      merge_strategy: "rebase"
```

**Advantages:**

- Linear history
- Preserves individual commits
- No merge commits

## Safety Features

### Pre-merge Validation

Before merging, the command verifies:

- ✅ PR still exists and is open
- ✅ CI/CD checks are still passing
- ✅ Required approvals are still present
- ✅ Branch is up-to-date (if required)
- ✅ No merge conflicts exist

### Rate Limiting

Respects API rate limits:

```yaml
behavior:
  rate_limit:
    requests_per_second: 2.0
    burst: 5
    timeout: "30s"
```

### Retry Logic

Handles transient failures:

```yaml
behavior:
  retry:
    max_attempts: 3
    backoff: "1s"
    max_backoff: "30s"
```

## Notifications

### Slack Notifications

```yaml
notifications:
  slack:
    webhook_url: "${SLACK_WEBHOOK_URL}"
    channel: "#deployments"
    enabled: true
```

Sample notification:

```text
🚀 Git PR CLI Merge Report

✅ Successfully merged 3 PRs:
• myorg/web-app #123: Bump lodash
• myorg/web-app #124: Update Node.js
• group/project !45: Update Go modules

❌ Failed to merge 1 PR:
• workspace/repo #67: Merge conflict

Total execution time: 12.3s
```

### Email Notifications

```yaml
notifications:
  email:
    smtp_host: "${SMTP_HOST}"
    smtp_port: 587
    from: "${EMAIL_FROM}"
    to: ["devops@company.com"]
    enabled: true
```

## Error Handling

### Common Merge Failures

**Merge Conflicts**:

- PR has conflicts with target branch
- Requires manual resolution
- PR will be skipped

**Status Check Failures**:

- CI/CD checks failed since discovery
- PR will be skipped
- Retry on next run

**Permission Errors**:

- Token lacks merge permissions
- Repository settings prevent merge
- Check token scopes and branch protection

**Rate Limit Exceeded**:

- Too many API requests
- Command will retry with backoff
- Adjust rate limiting settings

### Partial Failures

If some PRs fail to merge:

- Successfully merged PRs remain merged
- Failed PRs are reported with error details
- Command continues with remaining PRs
- Exit code indicates partial failure

## Exit Codes

- `0`: All PRs merged successfully (or none found)
- `1`: Error occurred (auth, config, network)
- `2`: Some PRs failed to merge (partial success)
- `3`: No PRs ready to merge

## Best Practices

### Always Dry Run First

```bash
# Check what would be merged
git-pr-cli merge --dry-run

# Then execute if satisfied
git-pr-cli merge --auto-approve
```

### Limit Concurrent Merges

For large numbers of PRs:

```bash
# Merge in small batches
git-pr-cli merge --max-merges=5 --auto-approve
```

### Monitor and Validate

```bash
# Check results after merging
git-pr-cli stats --detailed

# Validate configuration regularly
git-pr-cli validate --check-repos
```

## Troubleshooting

### No PRs to Merge

- Run `git-pr-cli check` first to see available PRs
- Verify PR filters and requirements
- Check repository access permissions

### Merge Failures

- Review error messages in output
- Check repository branch protection rules
- Verify token has merge permissions
- Ensure CI/CD checks are passing

### Notification Failures

- Test notifications with `git-pr-cli test --notifications`
- Verify webhook URLs and email settings
- Check network connectivity

## Related Commands

- [`check`](check.md) - Check PR status before merging
- [`validate`](validate.md) - Verify configuration and permissions
- [`watch`](watch.md) - Continuously monitor and merge
- [`test`](test.md) - Test merge functionality
