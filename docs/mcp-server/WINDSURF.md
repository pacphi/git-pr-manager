# Windsurf MCP Integration

This guide covers setting up the Multi-Gitter PR Automation MCP server with Windsurf IDE.

## Prerequisites

Complete the [common prerequisites](MCP_SETUP.md#common-prerequisites) in the main MCP setup guide first.

## Configuration

### Using Pre-configured File

1. **Copy the configuration** from the project:

   ```bash
   cp samples/windsurf-mcp.json .windsurf/mcp.json
   ```

2. **Update paths** to be absolute:

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

1. **Create MCP configuration directory**:

   ```bash
   mkdir -p .windsurf
   ```

2. **Create MCP configuration file** (`.windsurf/mcp.json`):

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

### Global Configuration

For system-wide configuration, place the MCP configuration in Windsurf's global settings directory (location varies by platform).

## Platform-Specific Setup

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

### AI Chat Integration

Windsurf's AI assistant can access MCP tools through natural language:

1. **Open AI chat** in Windsurf
2. **Use natural language commands**:
   - "Check all pull requests across my repositories"
   - "Show me repository statistics"
   - "Merge ready dependabot PRs"
   - "Validate my PR automation configuration"

### Available Tools

The MCP server provides these capabilities to Windsurf's AI:

#### Setup & Configuration

- `setup_repositories` - Interactive repository setup wizard
- `validate_config` - Validate configuration files
- `backup_restore_config` - Backup or restore configurations

#### PR Management

- `check_pull_requests` - Check PR status across repositories
- `merge_pull_requests` - Merge ready pull requests
- `watch_repositories` - Monitor PR status continuously

#### Repository Tools

- `get_repository_stats` - Get repository statistics
- `test_notifications` - Test Slack and email notifications
- `lint_scripts` - Lint shell scripts with shellcheck

#### Utilities

- `check_dependencies` - Check required dependencies
- `install_dependencies` - Install missing dependencies

## Troubleshooting

### MCP Server Not Loading

1. **Check Windsurf version**: Ensure you have MCP support
2. **Verify configuration file**: Check `.windsurf/mcp.json` syntax
3. **Test binary manually**:

   ```bash
   cd /path/to/your/project
   ./mcp-server/multi-gitter-pr-a8n-mcp
   ```

### Configuration File Issues

1. **Validate JSON syntax**:

   ```bash
   jq . .windsurf/mcp.json
   ```

2. **Check file permissions**:

   ```bash
   ls -la .windsurf/mcp.json
   ```

3. **Verify paths are absolute** and point to correct locations

### Binary Permission Issues

**Unix-like systems**:

```bash
chmod +x /path/to/project/mcp-server/multi-gitter-pr-a8n-mcp
```

### AI Assistant Not Seeing Tools

1. **Restart Windsurf** after configuration changes
2. **Check Windsurf logs** for MCP-related errors
3. **Verify environment variables** are available to the process

## Advanced Configuration

### Environment Variables

You can specify environment variables in the MCP configuration:

```json
{
  "servers": {
    "multi-gitter-pr-automation": {
      "command": "/path/to/mcp-server/multi-gitter-pr-a8n-mcp",
      "args": [],
      "cwd": "/path/to/project",
      "env": {
        "LOG_LEVEL": "debug",
        "CONFIG_FILE": "config.yaml"
      }
    }
  }
}
```

**Security Note**: Avoid putting sensitive tokens directly in configuration files.

### Multiple Projects

You can configure multiple MCP servers for different projects:

```json
{
  "servers": {
    "multi-gitter-pr-automation": {
      "command": "/path/to/project1/mcp-server/multi-gitter-pr-a8n-mcp",
      "args": [],
      "cwd": "/path/to/project1"
    },
    "pr-automation-project2": {
      "command": "/path/to/project2/mcp-server/multi-gitter-pr-a8n-mcp",
      "args": [],
      "cwd": "/path/to/project2"
    }
  }
}
```

### Custom Arguments

Pass custom arguments to the MCP server:

```json
{
  "servers": {
    "multi-gitter-pr-automation": {
      "command": "/path/to/mcp-server/multi-gitter-pr-a8n-mcp",
      "args": ["--config", "custom-config.yaml", "--verbose"],
      "cwd": "/path/to/project"
    }
  }
}
```

## Integration with Windsurf Features

### Code Context Awareness

Windsurf's AI can use MCP tools while maintaining context about your code:

- Ask about PR status while viewing specific files
- Merge PRs related to the code you're currently editing
- Get repository statistics relevant to your current project

### Workflow Integration

Combine MCP tools with Windsurf's development workflow:

1. **Code review**: Use PR checking tools while reviewing code
2. **Dependency updates**: Merge dependency PRs while working on features
3. **Repository management**: Monitor multiple repositories from a single workspace

## Example Usage Scenarios

### Daily Development Workflow

```text
AI Chat: "Check all my pull requests and show me which ones are ready to merge"

AI Response: [Uses check_pull_requests tool to scan all configured repositories]

Follow-up: "Merge the ready dependabot PRs but show me a dry run first"

AI Response: [Uses merge_pull_requests with dry-run option, then actual merge]
```

### Repository Setup

```text
AI Chat: "Help me set up PR automation for my repositories"

AI Response: [Uses setup_repositories tool to run interactive wizard]

Follow-up: "Validate my configuration and test notifications"

AI Response: [Uses validate_config and test_notifications tools]
```

### Monitoring and Statistics

```text
AI Chat: "Show me statistics for all my repositories and check for any issues"

AI Response: [Uses get_repository_stats and check_dependencies tools]
```

## Custom Windsurf Extensions

If you want to create custom Windsurf extensions that work with MCP:

### Basic Extension Structure

```typescript
// Windsurf extension that integrates with MCP
export class PRAutomationExtension {
  private mcpClient: MCPClient;

  constructor() {
    this.mcpClient = new MCPClient('multi-gitter-pr-automation');
  }

  async checkPRs() {
    const result = await this.mcpClient.callTool('check_pull_requests');
    return result;
  }
}
```

### Integration Points

- Custom commands for common PR operations
- Status bar indicators for PR counts
- Notification integration for PR updates
- Custom views for repository statistics

## Next Steps

- Configure your `.windsurf/mcp.json` file with correct paths
- Test the integration with basic commands through the AI chat
- Explore advanced features and workflow integration
- Set up your repositories using the setup wizard
- Configure notifications for Slack or email updates

## Support and Resources

- Check Windsurf documentation for MCP-specific features
- Review the main MCP setup guide for general configuration
- See troubleshooting documentation for common issues
- Join Windsurf community forums for MCP-related discussions
