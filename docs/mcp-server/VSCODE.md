# VSCode MCP Integration

This guide covers setting up the Multi-Gitter PR Automation MCP server with Visual Studio Code.

## Prerequisites

Complete the [common prerequisites](MCP_SETUP.md#common-prerequisites) in the main MCP setup guide first.

## Configuration

### Method 1: Workspace Settings (Recommended)

1. **Copy pre-configured settings**:

   ```bash
   mkdir -p .vscode
   cp samples/vscode-settings.json .vscode/settings.json
   ```

2. **Update paths** to be absolute:

   ```json
   {
     "mcp.servers": {
       "multi-gitter-pr-automation": {
         "command": "/absolute/path/to/your/project/mcp-server/multi-gitter-pr-a8n-mcp",
         "args": [],
         "cwd": "/absolute/path/to/your/project"
       }
     }
   }
   ```

### Method 2: Manual Configuration

1. **Open workspace settings**: `Ctrl/Cmd + Shift + P` → "Preferences: Open Workspace Settings (JSON)"

2. **Add MCP configuration**:

   ```json
   {
     "mcp.servers": {
       "multi-gitter-pr-automation": {
         "command": "./mcp-server/multi-gitter-pr-a8n-mcp",
         "args": [],
         "cwd": "${workspaceFolder}"
       }
     }
   }
   ```

### Method 3: User Settings

For global configuration across all workspaces:

1. **Open user settings**: `Ctrl/Cmd + Shift + P` → "Preferences: Open User Settings (JSON)"
2. **Add the same MCP configuration** with absolute paths

## MCP Extensions

### Available Extensions

Check the VSCode Marketplace for MCP extensions:

- Search for "MCP" or "Model Context Protocol"
- Popular extensions may include MCP client implementations

### Installing MCP Support

1. **Install an MCP extension** from the marketplace
2. **Reload VSCode** after installation
3. **Configure server settings** as shown above

## Platform-Specific Configuration

### macOS

```json
{
  "mcp.servers": {
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
  "mcp.servers": {
    "multi-gitter-pr-automation": {
      "command": "/home/yourusername/path/to/project/mcp-server/multi-gitter-pr-a8n-mcp",
      "args": [],
      "cwd": "/home/yourusername/path/to/project"
    }
  }
}
```

## Usage

### Command Palette Integration

With MCP extensions installed:

1. **Open Command Palette**: `Ctrl/Cmd + Shift + P`
2. **Look for MCP commands**: Type "MCP" to see available commands
3. **Use natural language**: Many extensions provide chat interfaces

### AI Assistant Integration

If using VSCode with AI assistants (like GitHub Copilot Chat):

- "Check all pull requests across my repositories"
- "Show me statistics for my configured repos"
- "Merge ready dependabot PRs after a dry run"
- "Validate my PR automation configuration"

## Troubleshooting

For common issues, see the [Common Troubleshooting](MCP_SETUP.md#common-troubleshooting) section in the main setup guide.

### VSCode-Specific Issues

1. **Check extension installation**:

   ```bash
   code --list-extensions | grep -i mcp
   ```

2. **Check VSCode output**: Look in the Output panel for MCP-related logs

3. **Environment variables in VSCode**:

   ```json
   {
     "mcp.servers": {
       "multi-gitter-pr-automation": {
         "command": "/path/to/mcp-server/multi-gitter-pr-a8n-mcp",
         "args": [],
         "cwd": "/path/to/project",
         "env": {
           "GITHUB_TOKEN": "${env:GITHUB_TOKEN}"
         }
       }
     }
   }
   ```

## Development Workflow

### Integrated Development

With MCP integration, you can:

1. **Code and manage PRs** in the same environment
2. **Use AI assistance** for both coding and PR management
3. **Automate workflows** with custom tasks and extensions

### Custom Tasks

Add to `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Check PRs",
      "type": "shell",
      "command": "make",
      "args": ["check-prs"],
      "group": "build",
      "presentation": {
        "echo": true,
        "reveal": "always",
        "panel": "new"
      }
    },
    {
      "label": "Dry Run Merge",
      "type": "shell",
      "command": "make",
      "args": ["dry-run"],
      "group": "build"
    }
  ]
}
```

## Extension Development

If you want to create a custom MCP extension for VSCode:

### Basic Structure

```typescript
import * as vscode from 'vscode';

export function activate(context: vscode.ExtensionContext) {
    // Register MCP client
    const mcpClient = new MCPClient();

    // Register commands
    const disposable = vscode.commands.registerCommand('mcp.checkPRs', () => {
        mcpClient.callTool('check_pull_requests');
    });

    context.subscriptions.push(disposable);
}
```

### Package.json Contribution

```json
{
  "contributes": {
    "commands": [
      {
        "command": "mcp.checkPRs",
        "title": "Check Pull Requests",
        "category": "MCP"
      }
    ],
    "configuration": {
      "title": "MCP Servers",
      "properties": {
        "mcp.servers": {
          "type": "object",
          "description": "MCP server configurations"
        }
      }
    }
  }
}
```

## Example Commands

Natural language commands available through MCP:

### Repository Management

- "Show me all repositories with pending PRs"
- "Get statistics for GitHub repositories only"
- "Check which repos need configuration updates"

### PR Operations

- "Check all pull requests and show status"
- "Merge all ready dependency updates"
- "Show what would be merged in a dry run"

### Configuration

- "Validate my current configuration"
- "Test my Slack notification setup"
- "Show me missing environment variables"

## Next Steps

- Install a suitable MCP extension from the VSCode marketplace
- Configure your workspace settings with the MCP server
- Test the integration with basic commands
- Explore advanced features and custom workflows
- Check the main documentation for detailed configuration options
