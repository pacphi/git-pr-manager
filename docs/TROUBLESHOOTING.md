# Troubleshooting Guide

Common issues and solutions for multi-gitter automation.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Authentication Problems](#authentication-problems)
- [Configuration Errors](#configuration-errors)
- [Script Execution Problems](#script-execution-problems)
- [Network and API Issues](#network-and-api-issues)
- [Notification Problems](#notification-problems)
- [Debug Mode](#debug-mode)

## Installation Issues

### Missing Dependencies

**Problem**: Error messages about missing `yq`, `jq`, or `gh` commands.

```text
[ERROR] Missing dependencies: yq jq
```

**Solution**: Install dependencies (auto-detects your platform):

```bash
# Install all dependencies (works on both macOS and Linux)
make install
```

**Platform-Specific Installation**:

**macOS (using Homebrew)**:

```bash
make install-macos
# Or manually: brew install yq jq gh curl
```

**Linux (auto-detects package manager)**:

```bash
make install-linux
```

**Manual Linux Installation by Distribution**:

```bash
# Ubuntu/Debian (apt)
sudo apt-get update && sudo apt-get install -y jq curl
sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
sudo chmod +x /usr/local/bin/yq

# RHEL/CentOS (yum)
sudo yum install -y jq curl
sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
sudo chmod +x /usr/local/bin/yq

# Fedora (dnf)
sudo dnf install -y jq curl
sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
sudo chmod +x /usr/local/bin/yq

# Arch Linux (pacman)
sudo pacman -S jq yq curl

# GitHub CLI (optional, for all Linux distributions)
# See: https://cli.github.com/manual/installation
```

### Permission Errors

**Problem**: Scripts are not executable.

```bash
bash: ./check-prs.sh: Permission denied
```

**Solution**: Make scripts executable:

```bash
chmod +x *.sh
```

## Authentication Problems

### GitHub Authentication

**Problem**: GitHub API returns 401 Unauthorized.

```text
[ERROR] GITHUB_TOKEN environment variable not set
```

**Solutions**:

1. **Set the token**:

   ```bash
   export GITHUB_TOKEN="ghp_your_token_here"
   ```

2. **Check token validity**:

   ```bash
   curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user
   ```

3. **Verify token scopes**:
   - Go to https://github.com/settings/tokens
   - Ensure scopes include: `repo`, `workflow`

4. **Organization access**:
   - For organization repos, ensure token has org access
   - Enable SSO if organization requires it

**Problem**: Rate limiting errors.

```text
API rate limit exceeded for user
```

**Solution**: Wait for rate limit reset or use a different token:

```bash
# Check rate limit status
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/rate_limit
```

### GitLab Authentication

**Problem**: GitLab API returns 401 Unauthorized.

```text
[ERROR] GITLAB_TOKEN environment variable not set
```

**Solutions**:

1. **Set the token**:

   ```bash
   export GITLAB_TOKEN="glpat-your_token_here"
   ```

2. **Test token**:

   ```bash
   curl --header "PRIVATE-TOKEN: $GITLAB_TOKEN" https://gitlab.com/api/v4/user
   ```

3. **For self-hosted GitLab**:

   - Update `auth.gitlab.url` in config.yaml
   - Use your GitLab instance URL

### Bitbucket Authentication

**Problem**: Bitbucket API returns 401 Unauthorized.

```text
[ERROR] BITBUCKET_USERNAME or BITBUCKET_APP_PASSWORD environment variable not set
```

**Solutions**:

1. **Set credentials**:

   ```bash
   export BITBUCKET_USERNAME="your_username"
   export BITBUCKET_APP_PASSWORD="your_app_password"
   ```

2. **Test credentials**:

   ```bash
   curl -u "$BITBUCKET_USERNAME:$BITBUCKET_APP_PASSWORD" https://api.bitbucket.org/2.0/user
   ```

3. **Check app password permissions**:

   - Go to https://bitbucket.org/account/settings/app-passwords/
   - Ensure permissions include: `Repositories: Read, Write`, `Pull requests: Read, Write`

## Configuration Errors

### Invalid YAML Syntax

**Problem**: Configuration file has syntax errors.

```text
[ERROR] Invalid YAML syntax in config.yaml
```

**Solution**: Validate and fix YAML:

```bash
# Check syntax
make validate

# Use online YAML validator or
yq eval '.' config.yaml
```

Common YAML issues:

- Missing quotes around strings with special characters
- Incorrect indentation (use spaces, not tabs)
- Missing colons after keys
- Unmatched brackets or quotes

### Missing Configuration Sections

**Problem**: Required configuration sections are missing.

**Solution**: Ensure config.yaml has all required sections:

```yaml
config: {}
repositories: {}
auth: {}
notifications: {}
```

Use `make config-template` to generate a template.

### Repository Name Format

**Problem**: Repository names don't match expected format.

**Solution**: Use correct formats:

- GitHub: `"owner/repository"`
- GitLab: `"group/project"` or `"user/project"`
- Bitbucket: `"workspace/repository"`

## Script Execution Problems

### Script Not Found

**Problem**: Script file not found or not in PATH.

```bash
./check-prs.sh: No such file or directory
```

**Solutions**:

1. Ensure you're in the correct directory
2. Check if script exists: `ls -la *.sh`
3. Use absolute path: `/path/to/multi-gitter-automation/check-prs.sh`

### No PRs Found

**Problem**: Scripts report no PRs found when PRs exist.

**Possible causes and solutions**:

1. **Wrong repository names**: Verify repository names in config.yaml match exactly
2. **Authentication issues**: Check tokens have correct permissions
3. **PR filters too restrictive**: Review `pr_filters` in config
4. **Network issues**: Test API connectivity manually

### Merge Failures

**Problem**: PRs fail to merge even when marked as ready.

**Solutions**:

1. **Check PR status manually**:

   ```bash
   gh pr view <PR_NUMBER> --repo <REPO_NAME>
   ```

2. **Review merge requirements**:
   - Status checks must pass
   - Branch must be up to date
   - No merge conflicts
   - Required reviews completed

3. **Use force mode** (carefully):

   ```bash
   make merge-prs FORCE=true
   ```

## Network and API Issues

### Connection Timeouts

**Problem**: Scripts timeout when connecting to APIs.

**Solutions**:

1. **Check network connectivity**:

   ```bash
   curl -I https://api.github.com
   curl -I https://gitlab.com/api/v4/projects
   curl -I https://api.bitbucket.org/2.0
   ```

2. **Check for proxy/firewall issues**
3. **Retry with longer timeout** (modify scripts if needed)

### SSL/TLS Issues

**Problem**: SSL certificate errors.

**Solutions**:

1. **Update certificates**:

   ```bash
   # macOS
   brew update && brew upgrade ca-certificates

   # Linux
   sudo apt-get update && sudo apt-get upgrade ca-certificates
   ```

2. **For self-hosted instances**, verify SSL configuration

### Rate Limiting

**Problem**: API rate limits exceeded.

**Solutions**:

1. **Check current limits**:

   ```bash
   # GitHub
   curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/rate_limit

   # GitLab (headers in response)
   curl -I --header "PRIVATE-TOKEN: $GITLAB_TOKEN" https://gitlab.com/api/v4/projects
   ```

2. **Wait for reset** or **use different tokens**
3. **Reduce frequency** of script execution

## Notification Problems

### Slack Notifications Not Working

**Problem**: Slack notifications fail to send.

**Solutions**:

1. **Test webhook manually**:

   ```bash
   curl -X POST -H 'Content-type: application/json' \
     --data '{"text":"Test message"}' \
     "$SLACK_WEBHOOK_URL"
   ```

2. **Verify webhook URL** is correct and active
3. **Check workspace permissions** for the webhook

### Email Notifications Not Working

**Problem**: Email notifications fail to send.

**Solutions**:

1. **Test SMTP settings**:

   ```bash
   ./test-notifications.sh --email
   ```

2. **Common SMTP issues**:
   - **Gmail**: Use app-specific password, not regular password
   - **Wrong port**: Try 587 (STARTTLS) or 465 (SSL)
   - **Firewall**: Ensure SMTP ports are open

3. **Enable "Less secure apps"** for Gmail (not recommended)
4. **Use app-specific passwords** for 2FA-enabled accounts

## Debug Mode

### Enable Debug Mode

Run scripts with debug information:

```bash
# Using make
make debug

# Direct script execution
set -x
./check-prs.sh
./merge-prs.sh --dry-run
set +x
```

### Verbose Output

Get more detailed output:

```bash
# Enable verbose mode
export VERBOSE=true

# Check specific provider only
make check-github
```

### Log Analysis

Examine logs for issues:

```bash
# Redirect output to log file
make check-prs > pr-check.log 2>&1

# Check for specific errors
grep -i error pr-check.log
grep -i fail pr-check.log
```

## Common Error Messages

### "yq: command not found"

**Solution**: Install yq: `brew install yq`

### "gh: command not found"

**Solution**: Install GitHub CLI: `brew install gh`

### "curl: command not found"

**Solution**: Install curl (usually pre-installed on most systems)

### "No open PRs found"

**Possible causes**:

- Wrong repository name format
- Authentication issues
- No PRs actually open
- PR filters excluding all PRs

### "Failed to merge PR"

**Possible causes**:

- PR not in mergeable state
- Missing required reviews
- Failing status checks
- Merge conflicts
- Branch protection rules

## Getting Additional Help

### Diagnostic Commands

Run these to gather diagnostic information:

```bash
# Check system info
uname -a
which bash yq jq gh curl

# Test configuration
make validate
make test

# Check authentication
./test-notifications.sh --config-test

# Show repository stats
make stats
```

### Enable Detailed Logging

For complex issues, enable detailed logging:

```bash
# Create log directory
mkdir -p logs

# Run with full logging
make check-prs > logs/check-$(date +%Y%m%d_%H%M%S).log 2>&1
make merge-prs > logs/merge-$(date +%Y%m%d_%H%M%S).log 2>&1
```

### Report Issues

When reporting issues, include:

1. **System information**: `uname -a`
2. **Dependency versions**: `yq --version`, `jq --version`, `gh --version`
3. **Configuration** (without sensitive data)
4. **Full error messages**
5. **Steps to reproduce**
6. **Expected vs actual behavior**

### Quick Diagnostic Script

Create a diagnostic script:

```bash
#!/bin/bash
echo "=== Multi-Gitter Diagnostics ==="
echo "System: $(uname -a)"
echo "Dependencies:"
echo "  yq: $(yq --version 2>/dev/null || echo 'NOT INSTALLED')"
echo "  jq: $(jq --version 2>/dev/null || echo 'NOT INSTALLED')"
echo "  gh: $(gh --version 2>/dev/null | head -1 || echo 'NOT INSTALLED')"
echo "  curl: $(curl --version 2>/dev/null | head -1 || echo 'NOT INSTALLED')"

echo "Environment Variables:"
echo "  GITHUB_TOKEN: ${GITHUB_TOKEN:+SET}"
echo "  GITLAB_TOKEN: ${GITLAB_TOKEN:+SET}"
echo "  BITBUCKET_USERNAME: ${BITBUCKET_USERNAME:+SET}"
echo "  SLACK_WEBHOOK_URL: ${SLACK_WEBHOOK_URL:+SET}"

echo "Configuration:"
if [ -f config.yaml ]; then
  echo "  Config file: EXISTS"
  echo "  GitHub repos: $(yq '.repositories.github | length' config.yaml 2>/dev/null || echo 0)"
  echo "  GitLab repos: $(yq '.repositories.gitlab | length' config.yaml 2>/dev/null || echo 0)"
  echo "  Bitbucket repos: $(yq '.repositories.bitbucket | length' config.yaml 2>/dev/null || echo 0)"
else
  echo "  Config file: NOT FOUND"
fi

echo "Scripts:"
echo "  check-prs.sh: $([ -x ./check-prs.sh ] && echo 'EXECUTABLE' || echo 'NOT EXECUTABLE')"
echo "  merge-prs.sh: $([ -x ./merge-prs.sh ] && echo 'EXECUTABLE' || echo 'NOT EXECUTABLE')"
```
