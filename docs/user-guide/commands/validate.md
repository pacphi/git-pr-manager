# validate - Validate Configuration and Connectivity

The `validate` command verifies your configuration file, authentication tokens, and connectivity to all configured providers and repositories.

## Synopsis

```bash
git-pr-cli validate [flags]
```

## Description

The validate command performs comprehensive validation of your Git PR CLI setup:

1. **Configuration Validation**: Checks YAML syntax and required fields
2. **Authentication Testing**: Verifies tokens and credentials for all providers
3. **Repository Access**: Tests connectivity and permissions for configured repositories
4. **Dependency Verification**: Ensures required tools and dependencies are available
5. **Network Connectivity**: Validates API endpoint accessibility

## Options

```bash
      --check-auth              Validate authentication tokens only
      --check-config            Validate configuration file syntax only
      --check-deps              Check required dependencies only
      --check-repos             Test repository access and permissions
  -h, --help                    help for validate
      --provider strings        Validate only specific providers (github,gitlab,bitbucket)
      --repos strings           Validate only specific repositories (comma-separated)
      --show-details            Show detailed validation results
      --timeout duration        Timeout for network checks (default 30s)
```

## Global Flags

```bash
  -c, --config string   Configuration file path (default "config.yaml")
      --debug           Enable debug logging
      --dry-run         Show what would be validated without making API calls
      --quiet           Suppress non-error output
```

## Examples

### Basic Configuration Validation

```bash
# Validate entire configuration
git-pr-cli validate

# Validate with detailed output
git-pr-cli validate --show-details
```

### Authentication Testing

```bash
# Test all authentication tokens
git-pr-cli validate --check-auth

# Test GitHub authentication only
git-pr-cli validate --check-auth --provider github

# Test authentication with timeout
git-pr-cli validate --check-auth --timeout 60s
```

### Repository Access Testing

```bash
# Test repository access and permissions
git-pr-cli validate --check-repos

# Test specific repositories
git-pr-cli validate --check-repos --repos "owner/repo1,owner/repo2"

# Test repositories for specific provider
git-pr-cli validate --check-repos --provider github
```

### Dependency Checking

```bash
# Check required dependencies
git-pr-cli validate --check-deps

# Show detailed dependency information
git-pr-cli validate --check-deps --show-details
```

### Targeted Validation

```bash
# Validate configuration syntax only
git-pr-cli validate --check-config

# Validate specific provider configuration
git-pr-cli validate --provider gitlab --show-details

# Validate with custom configuration file
git-pr-cli validate --config custom-config.yaml
```

## Output Formats

The validate command provides clear feedback on validation results:

### Success Output

```text
✅ Configuration validation passed
✅ GitHub authentication verified
✅ GitLab authentication verified
✅ Repository access validated (5 repositories)
✅ Dependencies satisfied
✅ All validations passed
```

### Error Output

```text
❌ Configuration validation failed
   - Missing required field: pr_filters.allowed_actors
   - Invalid merge strategy: "invalid" (must be: merge, squash, rebase)

❌ GitHub authentication failed
   - Invalid token: 401 Unauthorized
   - Check GITHUB_TOKEN environment variable

⚠️  Repository access issues
   - owner/repo1: 403 Forbidden (insufficient permissions)
   - owner/repo2: 404 Not Found (repository not found)

❌ 3 validations failed, 2 passed
```

## Validation Categories

### Configuration File Validation

- YAML syntax and structure
- Required field presence
- Valid enum values (merge strategies, providers)
- Repository name format
- Notification settings format

### Authentication Validation

**GitHub:**

- Token validity and permissions
- Rate limit status
- Organization access (if applicable)

**GitLab:**

- Token validity and scope
- API endpoint accessibility
- Project access permissions

**Bitbucket:**

- Username and app password validity
- Workspace access permissions
- API connectivity

### Repository Access Validation

- Repository existence and accessibility
- Required permissions for PR operations:
  - Read access to pull requests
  - Write access for merging (if auto_merge enabled)
  - Repository admin access (if required)
- Branch protection rules compatibility

### Dependency Validation

- Required system tools:
  - `git` command availability
  - Network connectivity tools
- Optional but recommended tools:
  - `jq` for JSON processing
  - `yq` for YAML processing

## Troubleshooting

### Common Validation Errors

#### Authentication Failures

```bash
# Error: GitHub token invalid
export GITHUB_TOKEN="your_new_token_here"
git-pr-cli validate --check-auth --provider github
```

#### Repository Access Issues

```bash
# Error: Repository not found
# Check repository name spelling and access permissions
git-pr-cli validate --check-repos --repos "correct/repo-name" --show-details
```

#### Configuration Syntax Errors

```bash
# Error: Invalid YAML syntax
# Use a YAML validator or editor with syntax highlighting
yq eval . config.yaml  # Validates YAML syntax
```

### Permissions Requirements

**GitHub Repositories:**

- `repo` scope for private repositories
- `public_repo` scope for public repositories
- `read:org` for organization repositories

**GitLab Projects:**

- `read_api` scope minimum
- `write_repository` for merge operations
- Project member access (Developer role or higher)

**Bitbucket Repositories:**

- Repository read access
- Pull request write access
- Workspace member permissions

## Exit Codes

- `0`: All validations passed
- `1`: One or more validations failed
- `2`: Configuration file not found or invalid
- `3`: Authentication failure
- `4`: Network connectivity issues

## Related Commands

- [`check`](check.md) - Check PR status (requires valid configuration)
- [`setup`](setup.md) - Interactive configuration setup
- [`merge`](merge.md) - Merge operations (requires valid authentication)

## Best Practices

1. **Run Before Operations**: Always validate before running check or merge commands
2. **Regular Validation**: Validate configuration after any changes
3. **CI/CD Integration**: Include validation in automated workflows
4. **Token Rotation**: Re-validate after rotating authentication tokens
5. **Troubleshooting**: Use `--show-details` for debugging configuration issues

## Examples of Common Issues

### Missing Environment Variables

```yaml
# config.yaml
auth:
  github:
    token: "${GITHUB_TOKEN}"  # Ensure this environment variable is set
```

```bash
# Solution
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxx"
git-pr-cli validate --check-auth
```

### Invalid Repository Configuration

```yaml
# Incorrect
repositories:
  github:
    - name: "invalid-repo-name"  # Missing owner/

# Correct
repositories:
  github:
    - name: "owner/repository-name"
```

### Network Issues

```bash
# Test with increased timeout for slow networks
git-pr-cli validate --timeout 60s

# Test specific connectivity
git-pr-cli validate --check-auth --provider github --show-details
```
