# Claude Code MCP Integration

## Overview

Claude Code has native MCP (Model Context Protocol) support, making integration with the Multi-Gitter PR Automation server seamless.

## Prerequisites

Complete the [common prerequisites](MCP_SETUP.md#common-prerequisites) in the main MCP setup guide first.

## Configuration

Claude Code settings can be accessed through the IDE interface:

1. Open Claude Code
2. Go to Settings/Preferences
3. Look for "MCP Servers" or "Model Context Protocol" section
4. Add a new server configuration

Use the configuration from `samples/claude-desktop.json`:

```json
{
  "mcpServers": {
    "multi-gitter-pr-automation": {
      "command": "./mcp-server/multi-gitter-mcp",
      "args": [],
      "cwd": "."
    }
  }
}
```

### 3. Restart Claude Code

After adding the configuration, restart Claude Code to initialize the MCP server.

## Usage

Once configured, you can interact with the PR automation tool through natural language:

### Example Commands

- "Check all pull requests across my repositories"
- "Merge ready dependabot updates"
- "Show me repository statistics"
- "Validate my configuration"
- "What environment variables are missing?"

### Advanced Usage

- "Check PRs for GitHub only and output in JSON format"
- "Do a dry run of merging to see what would be affected"
- "Run the setup wizard to configure new repositories"

## Features Available

All MCP tools and resources are available through Claude Code:

### Tools

- Repository setup and configuration
- PR checking and merging
- Dependency management
- Notification testing
- Script linting

### Resources

- Live configuration data
- Repository statistics
- Environment status
- Available Makefile targets

## Troubleshooting

For common issues, see the [Common Troubleshooting](MCP_SETUP.md#common-troubleshooting) section in the main setup guide.

### Claude Code-Specific Issues

1. **Check Claude Code logs** for MCP-related errors
2. **Verify working directory** in settings matches your project root
3. **Test setup:**
   ```bash
   make setup-full
   make check-prs
   ```

## Benefits

- **Seamless Integration**: Native MCP support in Claude Code
- **Natural Language**: Interact using conversational commands
- **Full Feature Access**: All Makefile functionality available
- **Real-time Data**: Access to live configuration and status information
- **Safety Preserved**: All original safety features maintained

## Next Steps

- Explore the available tools by asking "What can you help me with for PR automation?"
- Check your setup by asking "What's the status of my environment variables?"
- Start managing PRs by asking "Show me all pull requests that are ready to merge"
