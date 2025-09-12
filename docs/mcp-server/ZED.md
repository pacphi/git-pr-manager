# Zed MCP Integration

This guide covers setting up the Multi-Gitter PR Automation MCP server with Zed editor.

## Prerequisites

Complete the [common prerequisites](MCP_SETUP.md#common-prerequisites) in the main MCP setup guide first.

## Configuration

### Using Pre-configured File

1. **Locate Zed configuration directory**:
   - macOS: `~/.config/zed/`
   - Linux: `~/.config/zed/`

2. **Copy the configuration**:

   ```bash
   cp samples/zed-mcp.json ~/.config/zed/mcp.json
   ```

3. **Update paths** to be absolute in the copied file:

   ```json
   {
     "servers": {
       "multi-gitter-pr-automation": {
         "command": "/absolute/path/to/your/project/mcp-server/multi-gitter-pr-a8n-mcp",
         "args": [],
         "cwd": "/absolute/path/to/your/project"
       }
     }
   }
   ```

### Manual Configuration

1. **Create MCP configuration file** (`~/.config/zed/mcp.json`):

   ```json
   {
     "servers": {
       "multi-gitter-pr-automation": {
         "command": "/path/to/your/project/mcp-server/multi-gitter-pr-a8n-mcp",
         "args": [],
         "cwd": "/path/to/your/project"
       }
     }
   }
   ```

### Project-Specific Configuration

For project-specific MCP servers, create `.zed/mcp.json` in your project root:

```json
{
  "servers": {
    "multi-gitter-pr-automation": {
      "command": "./mcp-server/multi-gitter-pr-a8n-mcp",
      "args": [],
      "cwd": "."
    }
  }
}
```

## Platform-Specific Configuration

### macOS

```json
{
  "servers": {
    "multi-gitter-pr-automation": {
      "command": "/Users/yourusername/path/to/project/mcp-server/multi-gitter-pr-a8n-mcp",
      "args": [],
      "cwd": "/Users/yourusername/path/to/project"
    }
  }
}
```


### Linux

```json
{
  "servers": {
    "multi-gitter-pr-automation": {
      "command": "/home/yourusername/path/to/project/mcp-server/multi-gitter-pr-a8n-mcp",
      "args": [],
      "cwd": "/home/yourusername/path/to/project"
    }
  }
}
```

## Usage

### AI Assistant Integration

Zed's built-in AI assistant can access MCP tools through natural language commands:

1. **Open AI assistant**: `Cmd/Ctrl + Shift + A` or click the AI icon
2. **Use natural language**:
   - "Check the status of all pull requests"
   - "Merge ready dependabot PRs after showing me what would be merged"
   - "Show repository statistics for all my configured repos"
   - "Validate my PR automation configuration"

### Command Palette

Access MCP functionality through Zed's command palette:

1. **Open command palette**: `Cmd/Ctrl + Shift + P`
2. **Type "MCP"** to see available MCP-related commands
3. **Execute tools** directly or through AI assistant

### Available MCP Tools

The MCP server exposes these tools to Zed's AI:

#### Repository Setup

- **setup_repositories** - Interactive repository discovery and configuration
- **validate_config** - Validate YAML configuration file
- **backup_restore_config** - Backup or restore configuration files

#### PR Management

- **check_pull_requests** - Check PR status across all repositories
- **merge_pull_requests** - Merge ready PRs with safety checks
- **watch_repositories** - Monitor repositories continuously

#### Utilities

- **get_repository_stats** - Get detailed repository statistics
- **test_notifications** - Test Slack and email notification setup
- **check_dependencies** - Verify required tools are installed
- **install_dependencies** - Install missing dependencies
- **lint_scripts** - Lint shell scripts using shellcheck

## Troubleshooting

### MCP Server Not Starting

1. **Check Zed version**: Ensure you have MCP support (latest versions)
2. **Verify binary exists and is executable**:

   ```bash
   ls -la /path/to/project/mcp-server/multi-gitter-pr-a8n-mcp
   chmod +x /path/to/project/mcp-server/multi-gitter-pr-a8n-mcp
   ```

3. **Test binary manually**:

   ```bash
   cd /path/to/your/project
   ./mcp-server/multi-gitter-pr-a8n-mcp
   ```

### Configuration Issues

1. **Validate JSON syntax**:

   ```bash
   jq . ~/.config/zed/mcp.json
   ```

2. **Check file location**: Different platforms store Zed config in different locations
3. **Verify absolute paths**: Use full paths for command and cwd

### AI Assistant Not Seeing Tools

1. **Restart Zed** after configuration changes
2. **Check Zed logs**: Look for MCP-related errors in the Zed log output
3. **Verify configuration**: Use command palette to check MCP server status

### Permission and Environment Issues

**Unix-like systems**:

```bash
# Ensure binary is executable
chmod +x /path/to/project/mcp-server/multi-gitter-pr-a8n-mcp

# Check environment variables are available
echo $GITHUB_TOKEN
echo $GITLAB_TOKEN
```

## Advanced Configuration

### Environment Variables

Specify environment variables in the MCP configuration:

```json
{
  "servers": {
    "multi-gitter-pr-automation": {
      "command": "/path/to/mcp-server/multi-gitter-pr-a8n-mcp",
      "args": [],
      "cwd": "/path/to/project",
      "env": {
        "LOG_LEVEL": "debug"
      }
    }
  }
}
```

### Custom Arguments

Pass arguments to the MCP server:

```json
{
  "servers": {
    "multi-gitter-pr-automation": {
      "command": "/path/to/mcp-server/multi-gitter-pr-a8n-mcp",
      "args": ["--config", "custom-config.yaml"],
      "cwd": "/path/to/project"
    }
  }
}
```

### Multiple Servers

Configure multiple MCP servers for different projects:

```json
{
  "servers": {
    "multi-gitter-pr-automation": {
      "command": "/path/to/project1/mcp-server/multi-gitter-pr-a8n-mcp",
      "args": [],
      "cwd": "/path/to/project1"
    },
    "pr-automation-other": {
      "command": "/path/to/project2/mcp-server/multi-gitter-pr-a8n-mcp",
      "args": [],
      "cwd": "/path/to/project2"
    }
  }
}
```

## Zed-Specific Features

### Collaborative Editing

When using Zed's collaborative features, MCP tools are available to all collaborators, enabling shared PR management workflows.

### Language Server Integration

MCP tools work alongside Zed's language server integration, providing contextual PR management while coding.

### Vim Mode

If using Zed's vim mode, you can create custom key bindings for common MCP operations.

## Example Workflows

### Daily Development

```text
# In Zed AI assistant:
"Check all my pull requests and show me which dependency updates are ready"

# AI uses check_pull_requests tool and filters results

"Merge the ready ones but show me a dry run first"

# AI uses merge_pull_requests with dry-run, then actual merge
```

### Repository Setup

```text
# Initial setup:
"Help me configure PR automation for my repositories"

# AI guides through setup_repositories tool

"Validate my configuration and test my Slack notifications"

# AI uses validate_config and test_notifications tools
```

### Code Review Context

```text
# While viewing code:
"Are there any PRs related to this file that I should review?"

# AI checks PRs and correlates with current file context
```

## Customization

### Key Bindings

Add custom key bindings to `~/.config/zed/keymap.json`:

```json
[
  {
    "context": "Editor",
    "bindings": {
      "cmd-shift-p": "ai_assistant::ToggleAssistant",
      "cmd-alt-p": "mcp::check_pull_requests"
    }
  }
]
```

### Custom Commands

Create Zed extensions that integrate with MCP for common workflows.

## Performance Considerations

### Resource Usage

The MCP server runs as a separate process and uses minimal resources:

- Memory: ~10MB typical usage
- CPU: Minimal when idle, brief spikes during API calls
- Network: Only when interacting with Git providers

### Optimization

- Use `watch_repositories` sparingly as it runs continuously
- Configure appropriate API rate limiting in your Git provider tokens
- Consider using dry-run mode for testing complex operations

## Example Natural Language Commands

### Basic Operations

- "Check all my pull requests across GitHub and GitLab"
- "Show me which repositories have pending dependency updates"
- "Merge all approved dependabot PRs"

### Advanced Queries

- "Get statistics for repositories that haven't been updated in 30 days"
- "Test my notification setup and show me the results"
- "Validate my configuration and fix any issues you find"

### Workflow Integration

- "Check PRs for the repository I'm currently working in"
- "Show me what would happen if I merged all ready PRs"
- "Set up monitoring for new repositories I've been added to"

## Next Steps

- Configure your MCP server in Zed's settings
- Test basic functionality with simple commands
- Set up your repositories using the interactive wizard
- Configure notifications for your preferred channels
- Explore advanced workflows and automation
- Check the main documentation for detailed configuration options

## Community and Support

- Check Zed's documentation for the latest MCP features
- Join Zed community Discord for MCP-related discussions
- Report issues with MCP integration to the Zed team
- Contribute to MCP server improvements via the project repository
