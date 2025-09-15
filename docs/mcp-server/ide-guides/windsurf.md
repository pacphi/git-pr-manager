# Windsurf MCP Integration

Windsurf is an AI-powered development environment with built-in MCP support for seamless integration with Git PR automation.

## Prerequisites

Complete the [MCP server setup](../README.md#quick-start) first.

## Configuration

### Method 1: Windsurf Settings UI

1. Open Windsurf
2. Go to **Settings** → **AI & MCP** → **MCP Servers**
3. Add a new MCP server:
   - **Name**: `git-pr-automation`
   - **Command**: `/absolute/path/to/git-pr-mcp`
   - **Working Directory**: `/absolute/path/to/your/project`
   - **Arguments**: `[]` (leave empty)

### Method 2: Configuration File

Create or edit Windsurf's MCP configuration file:

**Location:**

- **macOS**: `~/Library/Application Support/Windsurf/mcp-servers.json`
- **Linux**: `~/.config/windsurf/mcp-servers.json`
- **Windows**: `%APPDATA%\Windsurf\mcp-servers.json`

**Configuration:**

```json
{
  "mcpServers": {
    "git-pr-automation": {
      "command": "/absolute/path/to/git-pr-mcp",
      "args": [],
      "cwd": "/absolute/path/to/your/project",
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}",
        "GITLAB_TOKEN": "${GITLAB_TOKEN}",
        "BITBUCKET_USERNAME": "${BITBUCKET_USERNAME}",
        "BITBUCKET_APP_PASSWORD": "${BITBUCKET_APP_PASSWORD}"
      }
    }
  }
}
```

### Method 3: Project Configuration

Add to your project's `.windsurf/mcp.json`:

```json
{
  "servers": {
    "git-pr-automation": {
      "command": "./git-pr-mcp",
      "args": [],
      "cwd": "."
    }
  }
}
```

## Environment Variables

### Option 1: Launch from Terminal (Recommended)

```bash
# Set environment variables
export GITHUB_TOKEN="ghp_your_token_here"
export GITLAB_TOKEN="glpat_your_token_here"
export BITBUCKET_USERNAME="your_username"
export BITBUCKET_APP_PASSWORD="your_app_password"

# Launch Windsurf from terminal
windsurf /path/to/your/project
```

### Option 2: Windsurf Environment Settings

Configure in Windsurf's environment settings:

1. Go to **Settings** → **Environment**
2. Add environment variables:
   - `GITHUB_TOKEN`
   - `GITLAB_TOKEN`
   - `BITBUCKET_USERNAME`
   - `BITBUCKET_APP_PASSWORD`

### Option 3: Project Environment File

Create `.windsurf/.env` in your project:

```bash
GITHUB_TOKEN=ghp_your_token_here
GITLAB_TOKEN=glpat_your_token_here
BITBUCKET_USERNAME=your_username
BITBUCKET_APP_PASSWORD=your_app_password
```

## Usage

Once configured and Windsurf is restarted:

1. **Open AI Chat Panel** (usually on the right side)
2. **Start a conversation** with Windsurf's AI
3. **Use natural language** to interact with Git PR automation

### Example Conversations

**Check PR Status:**
> "Can you check all my pull requests and show me which repositories have PRs ready to merge?"

**Merge PRs:**
> "Show me the dependabot PRs that are ready to merge, then merge them after I confirm"

**Repository Management:**
> "What's the current status of all my configured repositories? Show me statistics."

**Configuration Tasks:**
> "Validate my Git PR automation setup and tell me if there are any issues"


## Verification

Test the MCP integration:

1. **Check MCP Connection:**
   > "Are you connected to the Git PR automation MCP server? What tools do you have access to?"

2. **List Available Commands:**
   > "What Git PR automation commands can you run for me?"

3. **Test Basic Functionality:**
   > "Can you show me the help documentation for checking pull requests?"

## Troubleshooting

### MCP Server Not Connecting

1. **Check Binary Path and Permissions:**

   ```bash
   ls -la /path/to/git-pr-mcp
   chmod +x /path/to/git-pr-mcp  # Ensure executable
   ```

2. **Verify Configuration Syntax:**

   ```bash
   # Validate JSON configuration
   cat ~/.config/windsurf/mcp-servers.json | jq .
   ```

3. **Check Windsurf Logs:**
   - Go to **Help** → **Show Logs** → **MCP Server Logs**
   - Look for connection errors or startup issues

### Environment Variables Not Working

1. **Test Environment Variables:**

   ```bash
   echo $GITHUB_TOKEN
   echo $GITLAB_TOKEN
   ./git-pr-mcp --help  # Should work if environment is correct
   ```

2. **Launch from Terminal:**
   Most reliable way to ensure environment variables are inherited

3. **Check Project Environment:**
   Verify `.windsurf/.env` file exists and contains correct values

### Commands Not Responding

1. **Test CLI Directly:**

   ```bash
   cd /path/to/your/project
   ./git-pr-cli validate --check-auth
   ```

2. **Check Configuration File:**
   Ensure `config.yaml` exists and is properly formatted

3. **Verify Working Directory:**
   Make sure `cwd` in MCP config points to the correct project directory

## Advanced Configuration

### Custom Arguments

```json
{
  "mcpServers": {
    "git-pr-automation": {
      "command": "/absolute/path/to/git-pr-mcp",
      "args": ["--log-level", "debug", "--config", "custom-config.yaml"],
      "cwd": "/absolute/path/to/your/project",
      "env": {
        "LOG_LEVEL": "debug"
      }
    }
  }
}
```

### Multiple Projects

```json
{
  "mcpServers": {
    "git-pr-work": {
      "command": "/path/to/git-pr-mcp",
      "cwd": "/path/to/work-projects"
    },
    "git-pr-personal": {
      "command": "/path/to/git-pr-mcp",
      "cwd": "/path/to/personal-projects"
    }
  }
}
```

### Performance Optimization

```json
{
  "mcpServers": {
    "git-pr-automation": {
      "command": "/path/to/git-pr-mcp",
      "env": {
        "LOG_LEVEL": "warn",
        "MAX_CONCURRENT_REQUESTS": "3",
        "REQUEST_TIMEOUT": "30s"
      }
    }
  }
}
```

## Windsurf-Specific Features

### AI Code Generation
> "Generate a GitHub Actions workflow that uses git-pr-cli to automatically merge dependabot PRs"

### Project Analysis
> "Analyze my repository configuration and suggest optimizations for better PR automation"

### Debugging Assistance
> "Help me debug why my GitLab repositories aren't being discovered during setup"

### Documentation Generation
> "Generate documentation for my PR automation setup based on my current configuration"

## Integration with Windsurf Features

### Terminal Integration

Use Windsurf's built-in terminal for direct CLI commands:

```bash
git-pr-cli check --verbose
git-pr-cli merge --dry-run
```

### File Explorer

- Right-click on `config.yaml` → "Analyze with AI" for configuration insights
- View PR automation logs directly in the file explorer

### Project Templates

Create project templates that include Git PR automation setup.

## Best Practices

1. **Use absolute paths** for reliable execution across sessions
2. **Test CLI independently** before setting up MCP integration
3. **Launch from terminal** for proper environment variable inheritance
4. **Monitor performance** with large repository sets
5. **Use specific queries** for better AI understanding and responses
6. **Restart Windsurf** after making configuration changes

## Common Issues

- **"MCP server not found"**: Check absolute paths and binary permissions
- **"Authentication failed"**: Verify environment variables are properly set
- **"Config not found"**: Ensure `cwd` points to directory containing `config.yaml`
- **Slow responses**: Consider performance tuning for large repository sets
- **Connection timeout**: Check network connectivity and API rate limits

## Security Notes

- Environment variables are processed securely by the MCP server
- All operations use the same safety features as the CLI tool
- Consider using project-specific environment files for better security isolation
- Always use `--dry-run` mode when testing new configurations
