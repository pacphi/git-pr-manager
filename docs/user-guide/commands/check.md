# check - Check Pull Request Status

The `check` command scans all configured repositories to find pull requests from trusted actors and displays their status.

## Synopsis

```bash
git-pr-cli check [flags]
```

## Description

The check command performs the following operations:

1. **Discovery**: Scans all configured repositories across GitHub, GitLab, and Bitbucket
2. **Filtering**: Applies PR filters (allowed actors, skip labels, age limits)
3. **Status Evaluation**: Checks CI status, approvals, and merge requirements
4. **Reporting**: Displays summary and detailed status information

## Options

```bash
      --dry-run                 Show what would be processed without making API calls
  -h, --help                    help for check
      --output string           Output format (text, json, yaml) (default "text")
      --provider strings        Check only specific providers (github,gitlab,bitbucket)
      --repos strings           Check only specific repositories (comma-separated)
      --show-details            Show detailed information for each PR
      --show-status             Show CI/CD status for each PR
```

## Examples

### Basic Usage

```bash
# Check all configured repositories
git-pr-cli check

# Check with verbose output
git-pr-cli check --verbose

# Check specific provider only
git-pr-cli check --provider=github

# Check specific repositories
git-pr-cli check --repos="owner/repo1,owner/repo2"
```

### Output Formats

```bash
# Default text output
git-pr-cli check

# JSON output for scripting
git-pr-cli check --output=json

# YAML output
git-pr-cli check --output=yaml
```

### Detailed Information

```bash
# Show detailed PR information
git-pr-cli check --show-details

# Show CI/CD status for each PR
git-pr-cli check --show-status

# Combine options
git-pr-cli check --show-details --show-status --verbose
```

## Sample Output

### Text Format (Default)

```text
✅ GitHub: 15 repositories, 8 PRs found, 3 ready to merge
✅ GitLab: 5 repositories, 2 PRs found, 1 ready to merge
✅ Bitbucket: 3 repositories, 1 PR found, 0 ready to merge

Ready to merge:
├── myorg/web-app: dependabot[bot] - Bump lodash from 4.17.20 to 4.17.21
│   ✅ CI: passing │ ✅ Reviews: 1/1 │ ✅ Checks: 3/3 passed
├── myorg/api-service: renovate[bot] - Update Node.js to v18.17.1
│   ✅ CI: passing │ ✅ Reviews: 2/1 │ ✅ Checks: 5/5 passed
└── gitlab-group/my-project: renovate[bot] - Update Go modules
    ✅ CI: passing │ ✅ Reviews: 1/0 │ ✅ Checks: 2/2 passed

Summary:
- Total repositories: 23
- Total PRs found: 11
- Ready to merge: 4
- Waiting for CI: 2
- Needs approval: 3
- Has conflicts: 1
- Skipped (labels): 1
```

### JSON Format

```json
{
  "timestamp": "2025-01-15T10:30:00Z",
  "summary": {
    "total_repositories": 23,
    "total_prs": 11,
    "ready_to_merge": 4,
    "waiting_for_ci": 2,
    "needs_approval": 3,
    "has_conflicts": 1,
    "skipped": 1
  },
  "providers": {
    "github": {
      "repositories_checked": 15,
      "prs_found": 8,
      "ready_to_merge": 3
    },
    "gitlab": {
      "repositories_checked": 5,
      "prs_found": 2,
      "ready_to_merge": 1
    },
    "bitbucket": {
      "repositories_checked": 3,
      "prs_found": 1,
      "ready_to_merge": 0
    }
  },
  "ready_prs": [
    {
      "repository": "myorg/web-app",
      "provider": "github",
      "number": 123,
      "title": "Bump lodash from 4.17.20 to 4.17.21",
      "author": "dependabot[bot]",
      "url": "https://github.com/myorg/web-app/pull/123",
      "status": "ready",
      "ci_status": "success",
      "reviews": {
        "required": 1,
        "approved": 1
      },
      "checks": {
        "total": 3,
        "passed": 3,
        "failed": 0
      }
    }
  ]
}
```

## Filtering Behavior

The check command respects all configured filters:

### Allowed Actors

Only processes PRs from configured actors:

```yaml
pr_filters:
  allowed_actors:
    - "dependabot[bot]"
    - "renovate[bot]"
    - "github-actions[bot]"
```

### Skip Labels

Skips PRs with configured labels:

```yaml
pr_filters:
  skip_labels:
    - "do-not-merge"
    - "wip"
    - "draft"
```

### Age Limits

Skips PRs older than configured age:

```yaml
pr_filters:
  max_age: "30d"  # Skip PRs older than 30 days
```

## Status Evaluation

For each PR, the command evaluates:

### CI/CD Status

- ✅ **Passing**: All checks successful
- ⏳ **Pending**: Checks still running
- ❌ **Failed**: One or more checks failed
- ⚠️ **Missing**: No status checks configured

### Review Requirements

- ✅ **Approved**: Has required approvals
- ⏳ **Pending**: Awaiting reviews
- ❌ **Changes Requested**: Review requested changes

### Merge Requirements

- ✅ **Ready**: Can be merged
- ⚠️ **Conflicts**: Has merge conflicts
- ⏳ **Branch Protection**: Waiting for requirements

## Exit Codes

- `0`: Success, PRs found and processed
- `1`: Error occurred (auth, config, network)
- `2`: No PRs found matching criteria
- `3`: Some repositories inaccessible

## Performance Notes

- Uses concurrent API requests (configurable via `behavior.concurrency`)
- Respects rate limits for each provider
- Caches repository metadata when possible
- Large numbers of repositories may take time to process

## Troubleshooting

### No PRs Found

- Verify PR authors match `allowed_actors`
- Check for skip labels on PRs
- Ensure PRs are not older than `max_age`
- Confirm repository access with `--show-details`

### Slow Performance

- Reduce concurrent requests: `behavior.concurrency: 2`
- Use `--repos` to check specific repositories
- Enable caching if available

### Authentication Errors

- Verify tokens with `git-pr-cli validate --check-auth`
- Check token permissions for repository access
- Ensure network connectivity to provider APIs

## Related Commands

- [`merge`](merge.md) - Execute merges for ready PRs
- [`validate`](validate.md) - Check configuration and connectivity
- [`stats`](stats.md) - Get detailed statistics
- [`watch`](watch.md) - Continuously monitor PR status
