# watch - Continuously Monitor Pull Requests

The `watch` command provides continuous monitoring of pull requests across all configured repositories, automatically processing and merging ready PRs based on your configuration.

## Synopsis

```bash
git-pr-cli watch [flags]
```

## Description

The watch command runs continuously and performs automated PR management:

1. **Periodic Scanning**: Regularly checks all configured repositories for new or updated PRs
2. **Real-time Processing**: Evaluates PR status, approvals, and CI checks
3. **Automatic Merging**: Merges qualifying PRs based on your auto-merge settings
4. **Live Reporting**: Provides real-time status updates and activity logs
5. **Error Handling**: Automatically retries failed operations and reports issues

## Options

```bash
      --check-interval duration   Time between PR checks (default 5m0s)
      --dry-run                   Show what would be processed without making changes
  -h, --help                      help for watch
      --max-concurrent int        Maximum concurrent operations (default 5)
      --max-retries int           Maximum retry attempts for failed operations (default 3)
      --notify-on-error           Send notifications for errors (requires notification config)
      --notify-on-merge           Send notifications for successful merges
      --once                      Run once instead of continuously
      --provider strings          Watch only specific providers (github,gitlab,bitbucket)
      --repos strings             Watch only specific repositories (comma-separated)
      --timeout duration          Timeout for individual operations (default 30s)
```

## Global Flags

```bash
  -c, --config string   Configuration file path (default "config.yaml")
      --debug           Enable debug logging
      --quiet           Suppress non-error output
```

## Examples

### Basic Watching

```bash
# Start continuous monitoring with default 5-minute intervals
git-pr-cli watch

# Watch with custom interval
git-pr-cli watch --check-interval 2m

# Run once and exit (useful for cron jobs)
git-pr-cli watch --once
```

### Dry Run Mode

```bash
# See what would happen without making changes
git-pr-cli watch --dry-run --once

# Watch in dry-run mode with detailed logging
git-pr-cli watch --dry-run --debug --check-interval 1m
```

### Targeted Watching

```bash
# Watch specific repositories
git-pr-cli watch --repos "owner/repo1,owner/repo2"

# Watch GitHub repositories only
git-pr-cli watch --provider github

# Watch multiple providers
git-pr-cli watch --provider github,gitlab
```

### Production Configuration

```bash
# Production setup with notifications
git-pr-cli watch \
  --check-interval 10m \
  --max-concurrent 3 \
  --notify-on-merge \
  --notify-on-error

# High-frequency monitoring for critical repos
git-pr-cli watch \
  --repos "critical/app1,critical/app2" \
  --check-interval 1m \
  --timeout 60s
```

## Watch Process Flow

### 1. Repository Discovery

- Loads configured repositories from config file
- Applies provider and repository filters
- Validates authentication for each provider

### 2. PR Scanning

- Fetches open pull requests from each repository
- Applies PR filters (allowed actors, skip labels, age limits)
- Evaluates CI status and approval requirements

### 3. Processing Decision

- Checks auto-merge configuration for each repository
- Verifies PR meets merge criteria
- Determines appropriate merge strategy

### 4. Action Execution

- Merges qualifying PRs using configured strategy
- Updates PR status and adds merge commit information
- Logs all actions and results

### 5. Notification Dispatch

- Sends success notifications for merged PRs
- Reports errors and failures
- Updates monitoring systems (if configured)

## Output and Logging

### Console Output

```text
Git PR CLI - Watch Mode
======================
Started: 2024-01-15 10:30:00 UTC
Interval: 5m0s
Repositories: 25 (20 GitHub, 3 GitLab, 2 Bitbucket)

[10:30:00] Starting scan cycle #1
[10:30:02] ✅ owner/repo1: Found 2 PRs, 1 ready to merge
[10:30:03] 🔄 owner/repo1: Merging PR #123 "Bump dependency version"
[10:30:05] ✅ owner/repo1: PR #123 merged successfully (squash)
[10:30:07] ℹ️  owner/repo2: Found 1 PR, waiting for CI checks
[10:30:15] ✅ Scan cycle completed: 1 merged, 0 errors
[10:35:00] Starting scan cycle #2
...
```

### Dry Run Output

```text
[DRY RUN] Git PR CLI - Watch Mode
================================
[10:30:00] 🔍 owner/repo1: Would merge PR #123 "Bump dependency version" (squash)
[10:30:01] ⏸️  owner/repo2: Would skip PR #456 "WIP: Feature development" (has skip label)
[10:30:02] ⏳ owner/repo3: Would wait for PR #789 "Security update" (CI pending)
[10:30:03] 📊 Summary: Would merge 1 PR, skip 1 PR, wait for 1 PR
```

### Debug Logging

```bash
git-pr-cli watch --debug --once
```

```text
[DEBUG] Loading configuration from config.yaml
[DEBUG] Authenticating with GitHub (20 repositories)
[DEBUG] Authenticating with GitLab (3 repositories)
[DEBUG] Repository owner/repo1: Fetching PRs...
[DEBUG] Repository owner/repo1: Found PR #123 by dependabot[bot]
[DEBUG] Repository owner/repo1: PR #123 status checks: ✅ CI, ✅ Tests
[DEBUG] Repository owner/repo1: PR #123 approvals: 2 required, 2 received
[DEBUG] Repository owner/repo1: PR #123 eligible for merge (squash strategy)
[DEBUG] Repository owner/repo1: Merging PR #123...
[DEBUG] Repository owner/repo1: PR #123 merge API response: 200 OK
[INFO]  Repository owner/repo1: PR #123 merged successfully
```

## Configuration Integration

### Auto-merge Settings

```yaml
repositories:
  github:
    - name: "owner/repo1"
      auto_merge: true          # Required for watch mode
      merge_strategy: "squash"
      require_checks: true

    - name: "owner/repo2"
      auto_merge: false         # Watch will monitor but not merge
      merge_strategy: "merge"
```

### Behavioral Configuration

```yaml
behavior:
  # Watch mode respects these settings
  concurrency: 5              # Max concurrent operations
  dry_run: false              # Global dry-run mode
  watch_interval: "5m"        # Default check interval
  require_approval: false     # Additional approval requirement

  rate_limit:
    requests_per_second: 5.0
    burst: 10
    timeout: "30s"

  retry:
    max_attempts: 3
    backoff: "1s"
    max_backoff: "30s"
```

## Notification Integration

### Slack Notifications

```yaml
notifications:
  slack:
    webhook_url: "${SLACK_WEBHOOK_URL}"
    channel: "#deployments"
    enabled: true
```

Watch mode messages:

- ✅ **Merged PR**: `PR #123 in owner/repo1 merged successfully`
- ❌ **Merge Failed**: `Failed to merge PR #456 in owner/repo2: CI checks failing`
- 🔄 **Watch Started**: `PR monitoring started for 25 repositories`

### Email Notifications

```yaml
notifications:
  email:
    smtp_host: "${SMTP_HOST}"
    smtp_port: 587
    from: "${EMAIL_FROM}"
    to: ["team@company.com"]
    enabled: true
```

## Error Handling

### Automatic Retry Logic

Watch mode includes sophisticated error handling:

- **Temporary failures**: Network issues, rate limiting
- **Authentication problems**: Token expiration, permission changes
- **Repository issues**: Branch protection, merge conflicts
- **CI/CD failures**: Failed status checks, pending approvals

### Error Categories

#### Recoverable Errors

```text
[WARN] owner/repo1: Rate limit exceeded, waiting 60s before retry
[INFO] owner/repo1: Retrying operation (attempt 2/3)
[INFO] owner/repo1: Operation succeeded on retry
```

#### Non-recoverable Errors

```text
[ERROR] owner/repo1: PR #123 has merge conflicts - manual intervention required
[ERROR] owner/repo2: Authentication failed - check GITHUB_TOKEN
[ERROR] owner/repo3: Repository not found or access denied
```

## Performance Considerations

### Resource Usage

- **Memory**: Scales with number of configured repositories
- **CPU**: Minimal during idle intervals, increases during PR processing
- **Network**: API calls proportional to repository count and check frequency

### Optimization Tips

```bash
# Reduce load for large configurations
git-pr-cli watch --check-interval 15m --max-concurrent 3

# Focus on high-priority repositories
git-pr-cli watch --repos "critical/app1,critical/app2" --check-interval 1m

# Balance thoroughness with performance
git-pr-cli watch --provider github --check-interval 5m
```

## Production Deployment

### Systemd Service

```ini
[Unit]
Description=Git PR CLI Watch Service
After=network.target

[Service]
Type=simple
User=git-pr-cli
WorkingDirectory=/opt/git-pr-cli
ExecStart=/usr/local/bin/git-pr-cli watch --check-interval 10m
Restart=always
RestartSec=30
Environment=GITHUB_TOKEN=your_token_here

[Install]
WantedBy=multi-user.target
```

### Docker Container

```dockerfile
FROM alpine:latest
RUN apk add --no-cache ca-certificates git
COPY git-pr-cli /usr/local/bin/
COPY config.yaml /etc/git-pr-cli/
WORKDIR /app
CMD ["git-pr-cli", "watch", "--config", "/etc/git-pr-cli/config.yaml"]
```

### Kubernetes CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: git-pr-watch
spec:
  schedule: "*/10 * * * *"  # Every 10 minutes
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: git-pr-cli
            image: git-pr-cli:latest
            command: ["git-pr-cli", "watch", "--once"]
            env:
            - name: GITHUB_TOKEN
              valueFrom:
                secretKeyRef:
                  name: git-tokens
                  key: github-token
          restartPolicy: OnFailure
```

## Monitoring and Observability

### Health Checks

```bash
# Check if watch process is healthy
git-pr-cli validate --check-auth --quiet
echo "Watch health: $?"

# Monitor watch activity through logs
tail -f /var/log/git-pr-cli/watch.log | grep -E "(ERROR|merged|started)"
```

### Metrics Collection

```bash
#!/bin/bash
# Simple metrics script for monitoring
while true; do
    STATS=$(git-pr-cli stats --format json --period 1h)
    MERGED=$(echo "$STATS" | jq '.automation.successfully_merged')
    FAILED=$(echo "$STATS" | jq '.automation.failed_merges')

    echo "git_pr_merged_total $MERGED" | curl -X POST --data-binary @- \
        http://pushgateway:9091/metrics/job/git-pr-cli

    sleep 300
done
```

## Troubleshooting

### Common Issues

#### High Resource Usage

```bash
# Reduce concurrent operations and increase interval
git-pr-cli watch --max-concurrent 2 --check-interval 15m
```

#### Frequent Authentication Errors

```bash
# Validate tokens before starting watch
git-pr-cli validate --check-auth
```

#### Missing PRs

```bash
# Check PR filters and repository configuration
git-pr-cli check --repos "owner/repo" --show-details
```

### Debugging Tips

1. **Start with dry-run**: Always test with `--dry-run --once` first
2. **Use debug logging**: Add `--debug` for detailed operation logs
3. **Test individual repos**: Use `--repos` to isolate problematic repositories
4. **Monitor notifications**: Ensure notification systems are receiving messages

## Best Practices

1. **Start Conservatively**: Begin with longer intervals and fewer repositories
2. **Monitor Initially**: Watch logs and notifications for the first few cycles
3. **Use Dry-Run**: Test configuration changes with dry-run mode first
4. **Set Up Notifications**: Configure alerts for errors and successful merges
5. **Regular Validation**: Periodically validate configuration and authentication
6. **Resource Planning**: Monitor system resources with large repository counts

## Related Commands

- [`check`](check.md) - One-time PR status check
- [`merge`](merge.md) - Manual PR merging
- [`validate`](validate.md) - Validate configuration before watching
- [`stats`](stats.md) - Monitor automation performance
