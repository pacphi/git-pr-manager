# Cursor MCP Integration

## Overview

Cursor supports MCP (Model Context Protocol) servers through configuration files, enabling natural language interaction with the Multi-Gitter PR Automation tool.

## Prerequisites

Complete the [common prerequisites](MCP_SETUP.md#common-prerequisites) in the main MCP setup guide first.

## Configuration

Create the MCP configuration directory and file:

```bash
mkdir -p .cursor
```

Copy the configuration from `samples/cursor-mcp.json` to `.cursor/mcp.json`:

```json
{
  "servers": {
    "multi-gitter-pr-automation": {
      "command": "./mcp-server/multi-gitter-mcp",
      "args": [],
      "cwd": "."
    }
  }
}
```

### 3. Restart Cursor

After adding the configuration file, restart Cursor to load the MCP server.

## Configuration Options

The MCP server can be customized with additional options:

```json
{
  "servers": {
    "multi-gitter-pr-automation": {
      "command": "./mcp-server/multi-gitter-mcp",
      "args": [],
      "cwd": ".",
      "env": {
        "CONFIG_FILE": "config.yaml"
      }
    }
  }
}
```

## Usage

### Basic Commands

Once configured, interact with the tool through Cursor's AI assistant:

- "Check the status of all pull requests"
- "Merge dependabot PRs that are ready"
- "Show repository statistics"
- "Validate my configuration file"

### Advanced Usage

- "Check GitHub PRs only with JSON output"
- "Run a dry run to see what PRs would be merged"
- "Test my Slack notifications"
- "Install missing dependencies for macOS"

## Available Functionality

### Tools

- **setup_repositories** - Configure repositories with wizard
- **check_pull_requests** - Check PR status across repos
- **merge_pull_requests** - Merge ready PRs safely
- **validate_config** - Validate configuration
- **get_repository_stats** - Get repo statistics
- **test_notifications** - Test Slack/email setup
- **check_dependencies** - Verify required tools
- **install_dependencies** - Install missing tools

### Resources

- **Current Configuration** - Live config.yaml data
- **Repository Stats** - Statistics by provider
- **Makefile Targets** - Available commands
- **Environment Status** - Dependencies and env vars

## Troubleshooting

For common issues, see the [Common Troubleshooting](MCP_SETUP.md#common-troubleshooting) section in the main setup guide.

### Cursor-Specific Issues

1. **Check Cursor logs:** Look for MCP-related errors in Cursor's developer console
2. **Verify file paths:** Ensure `.cursor/mcp.json` exists and has correct paths
3. **Test underlying functionality:**

   ```bash
   make check-prs
   make validate
   ```

## Project Structure

Your project should look like this:

```text
project-root/
├── .cursor/
│   └── mcp.json              # Cursor MCP configuration
├── mcp-server/
│   ├── multi-gitter-mcp      # Built binary
│   └── ...                   # Go source files
├── config.yaml               # PR automation configuration
├── Makefile                  # Original tool interface
└── ...
```

## Benefits

- **AI-Powered Interface**: Natural language commands through Cursor
- **Full Tool Access**: All Makefile functionality available
- **Safe Operations**: Dry-run and validation capabilities
- **Real-time Data**: Access to live configuration and status
- **IDE Integration**: Works within your development workflow

## Next Steps

1. **Test the integration:**
   Ask Cursor: "What PR automation tools are available?"

2. **Check your setup:**
   Ask: "What's the status of my environment and dependencies?"

3. **Start managing PRs:**
   Ask: "Show me all PRs that are ready to merge"

4. **Explore capabilities:**
   Ask: "Help me set up repository automation"
