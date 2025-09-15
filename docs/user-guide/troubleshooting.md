# Troubleshooting Guide

This guide covers common issues and solutions when using Git PR CLI.

## Quick Diagnostics

Start with these basic validation commands:

```bash
# Check configuration syntax
git-pr-cli validate

# Test authentication with providers
git-pr-cli validate --check-auth

# Verify repository access
git-pr-cli validate --check-repos

# Show current configuration
git-pr-cli validate --show-config
```

## Authentication Issues

### GitHub Authentication

**Problem**: `Error: GitHub authentication failed`

**Solutions**:

1. **Verify token exists and is set**:

   ```bash
   echo $GITHUB_TOKEN
   ```

2. **Check token permissions** - Token needs these scopes:
   - `repo` (for private repositories)
   - `public_repo` (for public repositories)
   - `write:repo_hook` (if using webhooks)

3. **Create new token**:
   - Go to GitHub → Settings → Developer settings → Personal access tokens
   - Generate new token with required scopes
   - Set in environment: `export GITHUB_TOKEN="ghp_..."`

4. **Test token manually**:

   ```bash
   curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user
   ```

### GitLab Authentication

**Problem**: `Error: GitLab authentication failed`

**Solutions**:

1. **Verify token and URL**:

   ```bash
   echo $GITLAB_TOKEN
   echo $GITLAB_URL
   ```

2. **For GitLab.com** (default):

   ```bash
   export GITLAB_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
   # GITLAB_URL is optional for gitlab.com
   ```

3. **For self-hosted GitLab**:

   ```bash
   export GITLAB_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
   export GITLAB_URL="https://gitlab.company.com"
   ```

4. **Check token permissions** - Token needs:
   - `api` scope
   - `read_repository` scope
   - `write_repository` scope (for merging)

5. **Test token manually**:

   ```bash
   curl -H "Authorization: Bearer $GITLAB_TOKEN" $GITLAB_URL/api/v4/user
   ```

### Bitbucket Authentication

**Problem**: `Error: Bitbucket authentication failed`

**Solutions**:

1. **Verify credentials**:

   ```bash
   echo $BITBUCKET_USERNAME
   echo $BITBUCKET_APP_PASSWORD
   echo $BITBUCKET_WORKSPACE
   ```

2. **Create app password**:
   - Go to Bitbucket → Personal settings → App passwords
   - Create password with `Repositories: Read` and `Pull requests: Write`

3. **Set environment variables**:

   ```bash
   export BITBUCKET_USERNAME="your-username"
   export BITBUCKET_APP_PASSWORD="app-password-here"
   export BITBUCKET_WORKSPACE="workspace-name"  # Optional
   ```

4. **Test credentials manually**:

   ```bash
   curl -u $BITBUCKET_USERNAME:$BITBUCKET_APP_PASSWORD \
        https://api.bitbucket.org/2.0/user
   ```

## Configuration Issues

### Invalid Configuration File

**Problem**: `Error: failed to parse config file`

**Solutions**:

1. **Validate YAML syntax**:

   ```bash
   # Using yq (if installed)
   yq eval . config.yaml

   # Using Python
   python -c "import yaml; yaml.safe_load(open('config.yaml'))"
   ```

2. **Check indentation** - YAML is sensitive to spaces vs tabs
3. **Verify required fields** are present:
   - `pr_filters.allowed_actors`
   - `repositories`
   - `auth`

4. **Use sample config as reference**:

   ```bash
   cp config.sample config.yaml
   # Edit as needed
   ```

### Repository Not Found

**Problem**: `Error: repository not found: owner/repo`

**Solutions**:

1. **Verify repository name** format:
   - GitHub: `owner/repository`
   - GitLab: `group/project`
   - Bitbucket: `workspace/repository`

2. **Check repository access**:

   ```bash
   git-pr-cli validate --check-repos --repos="owner/repo"
   ```

3. **Verify token has access** to the repository
4. **For private repositories** - ensure token has appropriate scope

### Missing Environment Variables

**Problem**: `Error: required environment variable not set`

**Solutions**:

1. **Check which variables are missing**:

   ```bash
   git-pr-cli validate --check-auth
   ```

2. **Set missing variables**:

   ```bash
   export GITHUB_TOKEN="your-token"
   export GITLAB_TOKEN="your-token"
   # etc.
   ```

3. **Make variables persistent** by adding to shell profile:

   ```bash
   # Add to ~/.bashrc, ~/.zshrc, etc.
   echo 'export GITHUB_TOKEN="your-token"' >> ~/.bashrc
   ```

4. **Use .env file** (if supported):

   ```bash
   cat > .env << EOF
   GITHUB_TOKEN=your-token
   GITLAB_TOKEN=your-token
   EOF
   ```

## Runtime Issues

### No Pull Requests Found

**Problem**: `No pull requests found` even though PRs exist

**Solutions**:

1. **Check PR filters**:

   ```yaml
   pr_filters:
     allowed_actors:
       - "dependabot[bot]"
       - "renovate[bot]"
   ```

2. **Verify PR author** matches allowed actors
3. **Check for skip labels** on PRs:

   ```yaml
   pr_filters:
     skip_labels:
       - "do-not-merge"
       - "wip"
   ```

4. **Check PR age** if `max_age` is set
5. **Run with verbose output**:

   ```bash
   git-pr-cli check --verbose
   ```

### Pull Request Won't Merge

**Problem**: PR is found but won't merge

**Solutions**:

1. **Check required status checks**:

   ```bash
   git-pr-cli check --show-status
   ```

2. **Verify merge strategy** is allowed by repository:

   ```yaml
   repositories:
     github:
       - name: "owner/repo"
         merge_strategy: "squash"  # or "merge", "rebase"
   ```

3. **Check branch protection rules** in repository settings
4. **Ensure PR is not in draft state**
5. **Verify minimum approvals** are met if required

### Rate Limiting

**Problem**: `Error: rate limit exceeded`

**Solutions**:

1. **Reduce concurrency**:

   ```yaml
   behavior:
     concurrency: 2  # Reduce from default
   ```

2. **Adjust rate limiting**:

   ```yaml
   behavior:
     rate_limit:
       requests_per_second: 2.0  # Reduce from default
       burst: 5
   ```

3. **Add delays between requests**:

   ```yaml
   behavior:
     retry:
       backoff: "5s"  # Increase from default
   ```

4. **Check provider-specific limits**:
   - GitHub: 5,000 requests/hour for authenticated requests
   - GitLab: 2,000 requests/hour for authenticated requests
   - Bitbucket: 1,000 requests/hour

## Network Issues

### Connection Timeouts

**Problem**: `Error: request timeout`

**Solutions**:

1. **Increase timeout**:

   ```yaml
   behavior:
     rate_limit:
       timeout: "60s"  # Increase from default 30s
   ```

2. **Check network connectivity**:

   ```bash
   curl -I https://api.github.com
   curl -I https://gitlab.com/api/v4
   curl -I https://api.bitbucket.org
   ```

3. **Configure proxy if needed**:

   ```bash
   export HTTP_PROXY=http://proxy.company.com:8080
   export HTTPS_PROXY=http://proxy.company.com:8080
   ```

### SSL/TLS Issues

**Problem**: `Error: certificate verify failed`

**Solutions**:

1. **Update CA certificates**:

   ```bash
   # macOS
   brew install ca-certificates

   # Ubuntu/Debian
   sudo apt-get update && sudo apt-get install ca-certificates
   ```

2. **For self-hosted GitLab with custom CA**:

   ```bash
   export SSL_CERT_DIR=/path/to/custom/certs
   export SSL_CERT_FILE=/path/to/custom/ca.pem
   ```

## MCP Server Issues

### MCP Server Won't Start

**Problem**: IDE can't connect to MCP server

**Solutions**:

1. **Check binary exists and is executable**:

   ```bash
   ls -la git-pr-mcp
   chmod +x git-pr-mcp  # If needed
   ```

2. **Test server manually**:

   ```bash
   ./git-pr-mcp --help
   ```

3. **Use absolute paths** in IDE configuration:

   ```json
   {
     "command": "/full/path/to/git-pr-mcp",
     "cwd": "/full/path/to/project"
   }
   ```

4. **Check IDE logs** for MCP-related errors
5. **Restart IDE** after configuration changes

### MCP Tools Not Working

**Problem**: AI assistant can't execute tools

**Solutions**:

1. **Verify environment variables** in MCP config:

   ```json
   {
     "env": {
       "GITHUB_TOKEN": "your-token"
     }
   }
   ```

2. **Launch IDE from terminal** to inherit environment:

   ```bash
   export GITHUB_TOKEN="your-token"
   code .  # or your IDE command
   ```

3. **Test CLI directly**:

   ```bash
   git-pr-cli check --help
   ```

## Performance Issues

### Slow Execution

**Problem**: Commands take too long to execute

**Solutions**:

1. **Increase concurrency**:

   ```yaml
   behavior:
     concurrency: 10  # Increase from default
   ```

2. **Reduce repository scope**:

   ```bash
   git-pr-cli check --repos="critical-repo1,critical-repo2"
   ```

3. **Use filters to reduce processing**:

   ```yaml
   pr_filters:
     max_age: "7d"  # Only process recent PRs
   ```

4. **Enable caching** if available
5. **Profile execution**:

   ```bash
   time git-pr-cli check --verbose
   ```

### High Memory Usage

**Problem**: Tool uses excessive memory

**Solutions**:

1. **Reduce concurrency**:

   ```yaml
   behavior:
     concurrency: 2
   ```

2. **Process repositories in batches**
3. **Increase system memory** if possible
4. **Monitor with system tools**:

   ```bash
   top -p $(pgrep git-pr-cli)
   ```

## Debugging Tips

### Enable Debug Logging

```bash
# Set log level to debug
git-pr-cli --log-level=debug check

# Output to file for analysis
git-pr-cli --log-level=debug check > debug.log 2>&1
```

### Dry Run Mode

```bash
# See what would happen without actually doing it
git-pr-cli merge --dry-run
```

### Configuration Validation

```bash
# Show current configuration
git-pr-cli validate --show-config

# Test specific provider
git-pr-cli validate --provider=github

# Check specific repository
git-pr-cli validate --repos="owner/repo"
```

### Trace Network Requests

```bash
# Enable HTTP request tracing (if supported)
export DEBUG_HTTP=1
git-pr-cli check
```

## Getting Help

### Check Version and Build Info

```bash
git-pr-cli version
```

### System Information

```bash
# Check system requirements
go version
git --version

# Check available tools
which yq jq curl
```

### Community Resources

- **GitHub Issues**: Report bugs and feature requests
- **Documentation**: Check latest docs for updates
- **Configuration Examples**: Review sample configurations

### Creating Support Issues

When reporting issues, include:

1. **Version information**:

   ```bash
   git-pr-cli version
   ```

2. **Configuration** (sanitized):

   ```bash
   git-pr-cli validate --show-config | grep -v token
   ```

3. **Error messages** with debug logging:

   ```bash
   git-pr-cli --log-level=debug command > issue.log 2>&1
   ```

4. **Environment details**:
   - Operating system
   - Go version
   - Network setup (proxy, etc.)

5. **Steps to reproduce** the issue

This systematic approach will help identify and resolve most common issues with Git PR CLI.
