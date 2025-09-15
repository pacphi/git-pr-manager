# Zed MCP Integration

Zed is a high-performance code editor with built-in AI capabilities and MCP support for integrating with Git PR automation.

## Prerequisites

1. Complete the [MCP server setup](../README.md#quick-start)
2. **Zed Editor** with AI features enabled
3. **MCP support** (available in Zed 0.118.0+)

## Configuration

### Method 1: Zed Settings UI

1. Open Zed
2. Go to **Zed** → **Settings** (macOS) or **File** → **Settings** (Linux/Windows)
3. Navigate to **AI** → **MCP Servers**
4. Add new MCP server:
   - **Name**: `git-pr-automation`
   - **Command**: `/absolute/path/to/git-pr-mcp`
   - **Working Directory**: `/absolute/path/to/your/project`

### Method 2: Configuration File

Edit Zed's settings file:

**Location:**

- **macOS**: `~/Library/Application Support/Zed/settings.json`
- **Linux**: `~/.config/zed/settings.json`
- **Windows**: `%APPDATA%\Zed\settings.json`

**Configuration:**

```json
{
  "languages": {
    "mcp_servers": {
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
}
```

### Method 3: Project Settings

Add to your project's `.zed/settings.json`:

```json
{
  "mcp_servers": {
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

# Launch Zed from terminal
zed /path/to/your/project
```

### Option 2: Zed Environment Configuration

Include environment variables in the MCP server configuration (see examples above).

### Option 3: Shell Integration

Zed inherits environment from your shell, so setting variables in your shell profile works:

```bash
# Add to ~/.zshrc, ~/.bashrc, etc.
export GITHUB_TOKEN="ghp_your_token_here"
export GITLAB_TOKEN="glpat_your_token_here"
```

## Usage

Once configured and Zed is restarted:

1. **Open AI Assistant** (Ctrl/Cmd + Shift + A or click AI button in toolbar)
2. **Start a conversation** with Zed's AI
3. **Use natural language** to interact with Git PR automation

### Example Conversations

**Check PR Status:**
> "Check all pull requests across my repositories and show me which ones are ready to merge"

**Merge Operations:**
> "Show me dependabot PRs that are ready to merge, then merge them after I approve"

**Repository Management:**
> "What repositories do I have configured and what's their current status?"

**Configuration:**
> "Validate my Git PR automation configuration and show any issues"


## Verification

Test the MCP integration:

1. **Check Connection:**
   > "Are you connected to the Git PR automation MCP server?"

2. **List Available Tools:**
   > "What Git PR automation tools do you have access to?"

3. **Test Basic Command:**
   > "Show me help for checking pull requests"

## Troubleshooting

### MCP Server Not Connecting

1. **Check Binary Path:**

   ```bash
   ls -la /path/to/git-pr-mcp
   chmod +x /path/to/git-pr-mcp
   ```

2. **Verify Configuration:**

   Check Zed's settings file for syntax errors:

   ```bash
   cat ~/.config/zed/settings.json | jq .
   ```

3. **Check Zed Logs:**
   - Open **Command Palette** (Ctrl/Cmd + Shift + P)
   - Run "zed: open log"
   - Look for MCP-related errors

### Environment Issues

1. **Test Environment Variables:**

   ```bash
   echo $GITHUB_TOKEN
   ./git-pr-mcp --help
   ```

2. **Launch from Terminal:**
   Most reliable method for environment variable inheritance

### Commands Not Working

1. **Test CLI Directly:**

   ```bash
   cd /path/to/project
   ./git-pr-cli validate --check-auth
   ```

2. **Check Configuration File:**
   Ensure `config.yaml` exists and is valid

## Advanced Configuration

### Custom Arguments

```json
{
  "languages": {
    "mcp_servers": {
      "git-pr-automation": {
        "command": "/path/to/git-pr-mcp",
        "args": ["--log-level", "debug"],
        "env": {
          "LOG_LEVEL": "debug"
        }
      }
    }
  }
}
```

### Multiple Projects

```json
{
  "languages": {
    "mcp_servers": {
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
}
```

### Performance Tuning

```json
{
  "languages": {
    "mcp_servers": {
      "git-pr-automation": {
        "command": "/path/to/git-pr-mcp",
        "env": {
          "LOG_LEVEL": "warn",
          "MAX_CONCURRENT_REQUESTS": "3"
        }
      }
    }
  }
}
```

## Zed-Specific Features

### Multi-Buffer Editing
Use Zed's multi-buffer feature to work on multiple files while managing PRs.

### Collaborative Editing
Share your Git PR automation setup with collaborators using Zed's collaboration features.

### Extensions Integration
Zed's extension system can be enhanced with Git PR automation commands.

### Terminal Integration
Use Zed's integrated terminal for direct CLI commands:

```bash
git-pr-cli check --verbose
git-pr-cli merge --dry-run
```

## Best Practices

1. **Use absolute paths** for reliable execution
2. **Test CLI first** before configuring MCP
3. **Launch from terminal** for environment variables
4. **Keep Zed updated** for latest MCP features
5. **Use specific queries** for better AI responses
6. **Monitor performance** with large repository sets

## Common Issues

- **"MCP server not found"**: Check absolute paths and permissions
- **"Authentication failed"**: Verify environment variables are set
- **"Config not found"**: Ensure `cwd` points to correct directory
- **Slow responses**: Consider performance tuning for large repo sets
- **Connection timeout**: Check network and API rate limits

## Zed Version Compatibility

- **Zed 0.118.0+**: Native MCP support
- **Earlier versions**: Limited or no MCP support
- **Latest stable**: Recommended for best MCP experience

Update Zed regularly to get the latest MCP improvements and bug fixes.
