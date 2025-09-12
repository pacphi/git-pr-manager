# Claude Desktop MCP Integration

This guide covers setting up the Multi-Gitter PR Automation MCP server with Claude Desktop.

## Prerequisites

1. **Build the MCP server** (from project root):

   ```bash
   cd mcp-server
   go build -o multi-gitter-pr-a8n-mcp .
   ```

2. **Set up environment variables** as described in the main MCP setup guide

3. **Configure your repositories** using the setup wizard

## Configuration

### macOS

Claude Desktop stores its configuration at:

```bash
~/Library/Application Support/Claude/claude_desktop_config.json
```

Add the MCP server configuration:

```json
{
  "mcpServers": {
    "multi-gitter-pr-automation": {
      "command": "/absolute/path/to/your/project/mcp-server/multi-gitter-pr-a8n-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

### Linux

Configuration location:

```bash
~/.config/Claude/claude_desktop_config.json
```

Same format as macOS with absolute paths.

## Usage

1. **Restart Claude Desktop** after adding the configuration
2. **Verify connection** - The MCP server should appear in Claude's available tools
3. **Start using natural language commands**:
   - "Check all my pull requests"
   - "Merge ready dependabot PRs"
   - "Show repository statistics"
   - "Validate my configuration"

## Troubleshooting

### Server Not Starting

1. **Check absolute paths**: Ensure the command path is absolute, not relative
2. **Test manually**: Run the binary directly to verify it works

   ```bash
   cd /path/to/your/project
   ./mcp-server/multi-gitter-pr-a8n-mcp
   ```

3. **Check environment**: Make sure required tokens are available in your shell environment

### No Tools Available

1. **Check Claude Desktop logs** (macOS):

   ```bash
   tail -f ~/Library/Logs/Claude/claude_desktop.log
   ```

2. **Verify MCP server output**: Look for initialization messages and errors

### Permission Issues

Ensure the binary is executable:

```bash
chmod +x /path/to/project/mcp-server/multi-gitter-pr-a8n-mcp
```

## Advanced Configuration

### Environment Variables

You can pass environment variables directly in the configuration:

```json
{
  "mcpServers": {
    "multi-gitter-pr-automation": {
      "command": "/path/to/project/mcp-server/multi-gitter-pr-a8n-mcp",
      "args": [],
      "env": {
        "GITHUB_TOKEN": "ghp_your_token_here"
      }
    }
  }
}
```

**Security Note**: Avoid putting tokens directly in configuration files. Use environment variables instead.

### Working Directory

The MCP server automatically uses the directory containing the binary as its working directory, which should be your project root.

## Example Commands

Once configured, you can use these natural language commands in Claude Desktop:

- "Check the status of all pull requests across my repositories"
- "Show me what PRs would be merged in a dry run"
- "Merge all ready dependency update PRs"
- "Get statistics about my configured repositories"
- "Test my Slack notification setup"
- "Validate my current configuration file"

## Next Steps

- See the main MCP setup guide for general usage patterns
- Check the troubleshooting section for common issues
- Review your `config.yaml` to ensure proper repository configuration
