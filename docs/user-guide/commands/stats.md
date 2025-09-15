# stats - Repository and PR Statistics

The `stats` command provides detailed analytics and statistics about your configured repositories, pull requests, and automation activity.

## Synopsis

```bash
git-pr-cli stats [flags]
```

## Description

The stats command generates comprehensive statistics about:

1. **Repository Overview**: Total repositories, providers, and configuration status
2. **Pull Request Analytics**: PR counts, status distribution, and merge statistics
3. **Automation Metrics**: Success rates, processing times, and error analysis
4. **Actor Analysis**: Activity breakdown by PR authors and bots
5. **Time-based Trends**: Historical data and pattern analysis

## Options

```bash
      --detailed                Show detailed statistics for each repository
      --format string           Output format (text, json, yaml, csv) (default "text")
  -h, --help                    help for stats
      --include-closed          Include closed/merged PRs in statistics
      --period string           Time period for statistics (1d, 7d, 30d, all) (default "30d")
      --provider strings        Show stats for specific providers only
      --repos strings           Show stats for specific repositories only
      --sort string             Sort results by (name, prs, activity, last_update) (default "name")
      --top int                 Show top N repositories by activity (default 10)
```

## Global Flags

```bash
  -c, --config string   Configuration file path (default "config.yaml")
      --debug           Enable debug logging
      --quiet           Suppress non-error output
```

## Examples

### Basic Statistics

```bash
# Show overall statistics
git-pr-cli stats

# Show detailed statistics
git-pr-cli stats --detailed

# Show statistics for last 7 days
git-pr-cli stats --period 7d
```

### Provider-Specific Statistics

```bash
# GitHub repositories only
git-pr-cli stats --provider github

# Multiple providers
git-pr-cli stats --provider github,gitlab

# Compare providers with detailed output
git-pr-cli stats --detailed --format json
```

### Repository Analysis

```bash
# Specific repositories
git-pr-cli stats --repos "owner/repo1,owner/repo2"

# Top 5 most active repositories
git-pr-cli stats --top 5 --sort activity

# All repositories sorted by PR count
git-pr-cli stats --sort prs --detailed
```

### Historical Analysis

```bash
# Include historical data (closed/merged PRs)
git-pr-cli stats --include-closed --period all

# Last 30 days with CSV export
git-pr-cli stats --period 30d --format csv > pr-stats.csv
```

## Output Formats

### Default Text Format

```text
Git PR CLI Statistics
====================

Overview (Last 30 days)
-----------------------
Total Repositories:     25
  ├─ GitHub:           20 (80.0%)
  ├─ GitLab:            3 (12.0%)
  └─ Bitbucket:         2 (8.0%)

Active Repositories:    18 (72.0%)
Auto-merge Enabled:     23 (92.0%)

Pull Request Summary
-------------------
Total Open PRs:         47
  ├─ Ready to Merge:    12 (25.5%)
  ├─ Pending Checks:    23 (48.9%)
  ├─ Needs Approval:     8 (17.0%)
  └─ Blocked:            4 (8.5%)

Automation Activity
------------------
PRs Processed:          156
  ├─ Successfully Merged: 134 (85.9%)
  ├─ Failed Merges:        12 (7.7%)
  └─ Skipped:              10 (6.4%)

Average Processing Time: 2.3 minutes
Success Rate:           85.9%

Top Contributors
---------------
dependabot[bot]:        98 PRs (62.8%)
renovate[bot]:          45 PRs (28.8%)
github-actions[bot]:    13 PRs (8.3%)
```

### Detailed Repository View

```bash
git-pr-cli stats --detailed --top 3
```

```text
Repository Details
=================

1. owner/high-activity-repo (GitHub)
   ├─ Open PRs:           8 (3 ready, 4 pending, 1 blocked)
   ├─ Merged (30d):      23 PRs
   ├─ Success Rate:      91.3%
   ├─ Avg. Merge Time:   1.2 hours
   ├─ Auto-merge:        Enabled (squash)
   └─ Last Activity:     2 minutes ago

2. owner/web-application (GitHub)
   ├─ Open PRs:           5 (2 ready, 3 pending)
   ├─ Merged (30d):      18 PRs
   ├─ Success Rate:      94.7%
   ├─ Avg. Merge Time:   3.1 hours
   ├─ Auto-merge:        Enabled (merge)
   └─ Last Activity:     1 hour ago

3. group/api-service (GitLab)
   ├─ Open PRs:           3 (1 ready, 2 pending)
   ├─ Merged (30d):      12 PRs
   ├─ Success Rate:      83.3%
   ├─ Avg. Merge Time:   4.7 hours
   ├─ Auto-merge:        Disabled
   └─ Last Activity:     6 hours ago
```

### JSON Export

```bash
git-pr-cli stats --format json --period 7d
```

```json
{
  "period": "7d",
  "generated_at": "2024-01-15T10:30:00Z",
  "overview": {
    "total_repositories": 25,
    "providers": {
      "github": {"count": 20, "percentage": 80.0},
      "gitlab": {"count": 3, "percentage": 12.0},
      "bitbucket": {"count": 2, "percentage": 8.0}
    },
    "active_repositories": 18,
    "auto_merge_enabled": 23
  },
  "pull_requests": {
    "total_open": 47,
    "ready_to_merge": 12,
    "pending_checks": 23,
    "needs_approval": 8,
    "blocked": 4
  },
  "automation": {
    "processed": 156,
    "successfully_merged": 134,
    "failed_merges": 12,
    "skipped": 10,
    "success_rate": 85.9,
    "average_processing_time_minutes": 2.3
  },
  "repositories": [
    {
      "name": "owner/repo1",
      "provider": "github",
      "open_prs": 8,
      "merged_prs_period": 23,
      "success_rate": 91.3,
      "auto_merge_enabled": true,
      "merge_strategy": "squash",
      "last_activity": "2024-01-15T10:28:00Z"
    }
  ]
}
```

## Statistics Categories

### Repository Metrics

- **Total Count**: Number of configured repositories per provider
- **Activity Status**: Repositories with recent PR activity
- **Configuration Status**: Auto-merge settings, merge strategies
- **Health Score**: Based on success rate and activity

### Pull Request Analytics

- **Status Distribution**: Ready, pending, blocked PRs
- **Age Analysis**: PR age distribution and stale PR identification
- **Size Metrics**: Lines changed, files modified
- **Review Statistics**: Approval rates and review times

### Automation Performance

- **Success Rates**: Percentage of successful merges
- **Processing Times**: Average time from ready to merged
- **Error Analysis**: Common failure reasons and patterns
- **Throughput**: PRs processed per time period

### Actor Analysis

- **Bot Activity**: Breakdown by dependabot, renovate, etc.
- **Human Contributors**: Manual PR activity
- **Update Patterns**: Frequency and types of updates

## Time Period Options

- `1d`: Last 24 hours
- `7d`: Last 7 days (default for most operations)
- `30d`: Last 30 days (default for overall stats)
- `90d`: Last 90 days
- `all`: All available historical data

## Sorting Options

- `name`: Alphabetical by repository name
- `prs`: By number of open PRs
- `activity`: By recent activity (default)
- `last_update`: By last PR update time
- `success_rate`: By merge success rate

## Export Formats

### CSV Export

```bash
git-pr-cli stats --format csv --period 30d > monthly-report.csv
```

Includes columns: repository, provider, open_prs, merged_prs, success_rate, avg_merge_time, auto_merge, last_activity

### JSON Export

Structured data suitable for integration with monitoring systems, dashboards, or further analysis.

### YAML Export

Human-readable structured format for documentation or configuration management.

## Use Cases

### Daily Monitoring

```bash
# Quick daily overview
git-pr-cli stats --period 1d

# Check for issues
git-pr-cli stats --sort success_rate --detailed | grep -E "(Failed|Error)"
```

### Weekly Reports

```bash
# Generate weekly summary
git-pr-cli stats --period 7d --format json > weekly-stats.json

# Most active repositories
git-pr-cli stats --period 7d --top 10 --sort activity
```

### Health Monitoring

```bash
# Identify problematic repositories
git-pr-cli stats --detailed --sort success_rate --format csv | \
  awk -F',' '$5 < 80 { print $1, $5 }'  # Repos with <80% success rate

# Find stale repositories
git-pr-cli stats --sort last_update --detailed
```

### Integration Examples

#### Slack Notifications

```bash
#!/bin/bash
STATS=$(git-pr-cli stats --format json --period 1d)
SUCCESS_RATE=$(echo "$STATS" | jq '.automation.success_rate')

if (( $(echo "$SUCCESS_RATE < 90" | bc -l) )); then
  curl -X POST -H 'Content-type: application/json' \
    --data "{\"text\":\"⚠️ PR automation success rate dropped to ${SUCCESS_RATE}%\"}" \
    "$SLACK_WEBHOOK_URL"
fi
```

#### Prometheus Metrics

```bash
#!/bin/bash
git-pr-cli stats --format json | jq -r '
  .repositories[] |
  "git_pr_open_prs{repo=\"\(.name)\",provider=\"\(.provider)\"} \(.open_prs)"
' > /var/lib/prometheus/git-pr-metrics.prom
```

## Troubleshooting

### Performance Considerations

```bash
# For large numbers of repositories, use specific filters
git-pr-cli stats --provider github --top 20  # Instead of all repos

# Use shorter time periods for faster results
git-pr-cli stats --period 7d  # Instead of --period all
```

### Data Accuracy

- Statistics are generated from current API data
- Historical data depends on API retention policies
- Rate limiting may affect data completeness for large configurations

## Related Commands

- [`check`](check.md) - Current PR status (real-time data)
- [`validate`](validate.md) - Verify configuration before generating stats
- [`watch`](watch.md) - Monitor live activity

## Best Practices

1. **Regular Monitoring**: Run daily stats to track trends
2. **Export Data**: Keep historical records with CSV/JSON exports
3. **Set Thresholds**: Monitor success rates and alert on degradation
4. **Filter Results**: Use provider/repo filters for focused analysis
5. **Time-based Analysis**: Compare different time periods for trend analysis
