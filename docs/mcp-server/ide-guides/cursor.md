# Cursor MCP Integration

Cursor supports MCP through its built-in configuration system, allowing seamless integration with Git PR automation.

## Prerequisites

Complete the [MCP server setup](../README.md#quick-start) first.

## Configuration

### Method 1: Cursor Settings UI

1. Open Cursor
2. Go to **Settings** → **Extensions** → **MCP**
3. Add a new MCP server:
   - **Name**: `git-pr-automation`
   - **Command**: `/absolute/path/to/git-pr-mcp`
   - **Working Directory**: `/absolute/path/to/your/project`

### Method 2: Configuration File

Create or edit Cursor's MCP configuration file:

**Location:**

- **macOS**: `~/Library/Application Support/Cursor/mcp-servers.json`
- **Linux**: `~/.config/cursor/mcp-servers.json`
- **Windows**: `%APPDATA%\Cursor\mcp-servers.json`

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

## Environment Variables

### Option 1: Launch from Terminal (Recommended)

```bash
# Set environment variables
export GITHUB_TOKEN="ghp_your_token_here"
export GITLAB_TOKEN="glpat_your_token_here"

# Launch Cursor from terminal
cursor /path/to/your/project
```

### Option 2: Cursor Environment Configuration

Add environment variables in the MCP configuration (see example above).

### Option 3: Cursor Workspace Settings

Add to your workspace `.cursor/settings.json`:

```json
{
  "mcp.servers": {
    "git-pr-automation": {
      "command": "./git-pr-mcp",
      "env": {
        "GITHUB_TOKEN": "your_token_here"
      }
    }
  }
}
```

## Usage

Once configured and Cursor is restarted:

1. **Open AI Chat** (Ctrl/Cmd + K or Chat panel)
2. **Use natural language** to interact with Git PR automation

### Example Conversations

**Check PR Status:**
> "Check all my pull requests and show me which ones are ready to merge"

**Merge PRs:**
> "Show me the dependabot PRs that are ready, then merge them"

**Repository Management:**
> "What repositories do I have configured and what's their status?"

**Configuration:**
> "Validate my Git PR automation configuration"

## Verification

Test the MCP integration:

1. **Check Connection:**
   > "Are you connected to the Git PR automation server?"

2. **List Tools:**
   > "What Git PR tools can you access?"

3. **Test Basic Command:**
   > "Can you check the help for the merge command?"

## Troubleshooting

### MCP Server Not Connecting

1. **Check Binary Path:**

   ```bash
   ls -la /path/to/git-pr-mcp
   chmod +x /path/to/git-pr-mcp  # Ensure executable
   ```

2. **Verify Configuration:**

   ```bash
   # Validate JSON syntax
   cat ~/.config/cursor/mcp-servers.json | jq .
   ```

3. **Check Cursor Logs:**
   - Go to **Help** → **Developer Tools** → **Console**
   - Look for MCP-related errors

### Environment Issues

1. **Test Environment Variables:**

   ```bash
   echo $GITHUB_TOKEN
   ./git-pr-mcp --help  # Should work if env is correct
   ```

2. **Launch from Terminal:**
   Most reliable way to ensure environment variables are available

### Commands Not Working

1. **Test CLI Directly:**

   ```bash
   cd /path/to/your/project
   ./git-pr-cli validate
   ```

2. **Check Configuration:**
   Ensure `config.yaml` exists and is valid

## Advanced Configuration

### With Custom Arguments

```json
{
  "mcpServers": {
    "git-pr-automation": {
      "command": "/absolute/path/to/git-pr-mcp",
      "args": ["--log-level", "debug", "--config", "custom-config.yaml"],
      "cwd": "/absolute/path/to/your/project"
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
      "cwd": "/path/to/work-project"
    },
    "git-pr-personal": {
      "command": "/path/to/git-pr-mcp",
      "cwd": "/path/to/personal-project"
    }
  }
}
```

### Performance Tuning

```json
{
  "mcpServers": {
    "git-pr-automation": {
      "command": "/path/to/git-pr-mcp",
      "env": {
        "LOG_LEVEL": "warn",
        "MAX_CONCURRENT": "3"
      }
    }
  }
}
```

## Cursor-Specific Features

### AI Code Generation
> "Generate a GitHub Actions workflow that uses git-pr-cli to merge dependabot PRs"

### Code Analysis
> "Analyze my config.yaml and suggest optimizations for my repository setup"

### Debugging
> "Help me debug why my GitLab authentication isn't working"

## Best Practices

1. **Use absolute paths** for reliable execution
2. **Test CLI independently** before configuring MCP
3. **Launch from terminal** for environment variables
4. **Restart Cursor** after configuration changes
5. **Monitor performance** with large repository sets
6. **Use specific queries** for better AI responses

## Common Issues

- **"MCP server not found"**: Check absolute paths and permissions
- **"Authentication failed"**: Verify environment variables are set
- **"Config not found"**: Ensure `cwd` points to correct directory
- **Slow responses**: Large repo sets may need performance tuning
