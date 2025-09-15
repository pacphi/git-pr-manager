# Claude Code MCP Integration

Claude Code has native MCP (Model Context Protocol) support, making integration with Git PR automation seamless.

## Prerequisites

Complete the [MCP server setup](../README.md#quick-start) first.

## Configuration

### Method 1: Using Claude Code Settings

1. Open Claude Code
2. Go to Settings/Preferences
3. Navigate to "MCP Servers" or "Model Context Protocol" section
4. Add a new server configuration:

```json
{
  "mcpServers": {
    "git-pr-automation": {
      "command": "/absolute/path/to/git-pr-mcp",
      "args": [],
      "cwd": "/absolute/path/to/your/project"
    }
  }
}
```

### Method 2: Configuration File

Claude Code typically stores MCP configuration in:

- **macOS**: `~/Library/Application Support/Claude Code/mcp-servers.json`
- **Linux**: `~/.config/claude-code/mcp-servers.json`
- **Windows**: `%APPDATA%\Claude Code\mcp-servers.json`

Create or edit the file with:

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
export GITHUB_TOKEN="ghp_your_token_here"
export GITLAB_TOKEN="glpat_your_token_here"
# Launch Claude Code from terminal to inherit environment
claude-code /path/to/your/project
```

### Option 2: Set in Configuration

Include environment variables directly in the MCP configuration (see example above).

## Usage

Once configured and Claude Code is restarted:

1. **Start a conversation** with Claude in your project
2. **Use natural language** to interact with Git PR automation:

### Example Conversations

**Check PR Status:**
> "Can you check the status of all pull requests across my repositories?"

**Merge Ready PRs:**
> "Show me what PRs are ready to merge, then merge the dependabot ones"

**Repository Statistics:**
> "What are the statistics for all my configured repositories?"

**Configuration Management:**
> "Validate my current configuration and show me any issues"

**Setup and Maintenance:**
> "Test my notification setup and show me the results"

## Verification

To verify the MCP server is working:

1. **Check MCP Connection:**
   > "Are you connected to the Git PR automation MCP server?"

2. **List Available Tools:**
   > "What Git PR automation tools do you have access to?"

3. **Test Basic Functionality:**
   > "Can you show me the help for checking pull requests?"

## Troubleshooting

### Server Not Found

- Ensure the binary path is absolute and correct
- Verify the binary is executable: `chmod +x git-pr-mcp`
- Check Claude Code logs for MCP connection errors

### Tools Not Working

- Verify environment variables are set correctly
- Test CLI directly: `./git-pr-mcp --help`
- Check that `config.yaml` exists and is valid

### Performance Issues

- Large repositories may take time to process
- Use specific commands rather than broad queries
- Consider using `--dry-run` for preview operations

## Advanced Configuration

### Custom Environment

```json
{
  "mcpServers": {
    "git-pr-automation": {
      "command": "/absolute/path/to/git-pr-mcp",
      "args": ["--log-level", "debug"],
      "cwd": "/absolute/path/to/your/project",
      "env": {
        "CONFIG_FILE": "/custom/path/to/config.yaml",
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
    "git-pr-project-1": {
      "command": "/absolute/path/to/git-pr-mcp",
      "cwd": "/path/to/project1"
    },
    "git-pr-project-2": {
      "command": "/absolute/path/to/git-pr-mcp",
      "cwd": "/path/to/project2"
    }
  }
}
```

## Best Practices

1. **Use absolute paths** for both command and cwd
2. **Test CLI first** before configuring MCP
3. **Launch from terminal** to ensure environment variables
4. **Restart Claude Code** after configuration changes
5. **Use specific commands** for better performance
