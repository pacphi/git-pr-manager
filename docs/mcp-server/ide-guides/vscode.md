# VS Code MCP Integration

VS Code supports MCP through extensions and workspace configuration, allowing integration with Git PR automation.

## Prerequisites

1. Complete the [MCP server setup](../README.md#quick-start)
2. Install an MCP-compatible AI extension:
   - **Claude for VS Code** (recommended)
   - **GitHub Copilot** (with MCP extension)
   - **Continue** (open-source AI extension)

## Configuration

### Method 1: Workspace Settings

Create or edit `.vscode/settings.json` in your project:

```json
{
  "mcp.servers": {
    "git-pr-automation": {
      "command": "/absolute/path/to/git-pr-mcp",
      "args": [],
      "cwd": "${workspaceFolder}",
      "env": {
        "GITHUB_TOKEN": "${env:GITHUB_TOKEN}",
        "GITLAB_TOKEN": "${env:GITLAB_TOKEN}",
        "BITBUCKET_USERNAME": "${env:BITBUCKET_USERNAME}",
        "BITBUCKET_APP_PASSWORD": "${env:BITBUCKET_APP_PASSWORD}"
      }
    }
  }
}
```

### Method 2: User Settings

Add to your global VS Code settings (Ctrl/Cmd + ,):

```json
{
  "mcp.servers": {
    "git-pr-automation": {
      "command": "/absolute/path/to/git-pr-mcp",
      "args": [],
      "cwd": "/absolute/path/to/your/project"
    }
  }
}
```

## Environment Variables

### Option 1: Launch from Terminal (Recommended)

```bash
export GITHUB_TOKEN="ghp_your_token_here"
export GITLAB_TOKEN="glpat_your_token_here"
code /path/to/your/project
```

### Option 2: VS Code Terminal

Set in VS Code's integrated terminal:

```bash
export GITHUB_TOKEN="your_token"
export GITLAB_TOKEN="your_token"
```

### Option 3: Extension Configuration

Some AI extensions provide their own environment variable configuration.

## Usage

Once configured:

1. **Open Command Palette** (Ctrl/Cmd + Shift + P)
2. **Start AI Chat** (depends on your AI extension)
3. **Use natural language** to interact with Git PR automation

### Example Commands

**Check PRs:**
> "Check all pull requests and show me which ones are ready to merge"

**Merge PRs:**
> "Show me dependabot PRs that are ready and merge them after confirmation"

**Statistics:**
> "What are the statistics for my configured repositories?"

**Configuration:**
> "Validate my Git PR configuration and show any issues"

## Extension-Specific Setup

### Claude for VS Code

1. Install the Claude extension
2. Configure MCP server in workspace settings
3. Start a conversation with Claude
4. Claude will automatically discover and use MCP tools

### GitHub Copilot + MCP Extension

1. Install GitHub Copilot
2. Install MCP extension for Copilot
3. Configure MCP servers in settings
4. Use Copilot chat with MCP capabilities

### Continue Extension

1. Install Continue extension
2. Configure MCP in Continue settings
3. Use Continue chat interface

## Troubleshooting

### MCP Server Not Found

1. **Check Extension Installation:**
   Ensure you have an MCP-compatible AI extension installed

2. **Verify Binary Path:**

   ```bash
   ls -la /path/to/git-pr-mcp
   chmod +x /path/to/git-pr-mcp
   ```

3. **Check VS Code Output:**
   - Go to **View** → **Output**
   - Select your AI extension from dropdown
   - Look for MCP connection errors

### Environment Issues

1. **Test Environment Variables:**
   Open VS Code terminal and run:

   ```bash
   echo $GITHUB_TOKEN
   ./git-pr-mcp --help
   ```

2. **Use Absolute Paths:**
   Replace `${workspaceFolder}` with absolute path if needed

### Commands Not Working

1. **Test CLI Directly:**

   ```bash
   cd /path/to/project
   git-pr-cli validate
   ```

2. **Check Extension Logs:**
   Each AI extension has its own output channel in VS Code

## Advanced Configuration

### Multiple Projects

```json
{
  "mcp.servers": {
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

### Custom Arguments

```json
{
  "mcp.servers": {
    "git-pr-automation": {
      "command": "/path/to/git-pr-mcp",
      "args": ["--log-level", "debug"],
      "env": {
        "LOG_LEVEL": "debug"
      }
    }
  }
}
```

### Workspace-Specific Config

Create `.vscode/git-pr.json` for project-specific MCP settings:

```json
{
  "servers": {
    "git-pr-automation": {
      "command": "./git-pr-mcp",
      "cwd": "."
    }
  }
}
```

## VS Code Features

### Integrated Terminal

Run CLI commands directly:

```bash
git-pr-cli check --verbose
```

### Task Runner

Create `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Check PRs",
      "type": "shell",
      "command": "./git-pr-cli",
      "args": ["check", "--verbose"],
      "group": "build"
    }
  ]
}
```

### Debug Configuration

Create `.vscode/launch.json` for debugging the MCP server:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug MCP Server",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "./cmd/git-pr-mcp"
    }
  ]
}
```

## Best Practices

1. **Use workspace settings** for project-specific configuration
2. **Launch from terminal** for reliable environment variables
3. **Test CLI first** before configuring MCP
4. **Use VS Code variables** like `${workspaceFolder}` when possible
5. **Check extension compatibility** with MCP features

## Common Issues

- **"Extension not found"**: Install an MCP-compatible AI extension
- **"Server not responding"**: Check binary permissions and paths
- **"Environment variables missing"**: Launch VS Code from terminal
- **Slow responses**: Consider using specific queries vs. broad requests
