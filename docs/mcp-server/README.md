# Git PR MCP Server

The Git PR MCP (Model Context Protocol) server provides a natural language interface to the Git PR CLI automation tool. It allows you to interact with the tool through AI assistants in your IDE using conversational commands.

## Overview

The MCP server acts as a **tool provider** that exposes Git PR automation capabilities to your IDE's AI assistant. The server itself doesn't make LLM calls - instead, your IDE's AI (Claude, GPT, etc.) calls the MCP server's tools based on your natural language requests.

**Architecture:**

```text
Your IDE + AI Assistant ←→ MCP Server ←→ Git PR CLI + Providers
```

## Features

### Available Tools

The MCP server exposes the following tools:

#### Setup & Configuration

- `setup_repositories` - Run interactive setup wizard to configure repositories
- `validate_configuration` - Validate configuration file and connectivity

#### PR Management

- `check_pull_requests` - Check PR status across all repositories
- `merge_pull_requests` - Merge ready pull requests
- `watch_repositories` - Monitor PR status continuously
- `get_repository_statistics` - Get detailed repository statistics

### Available Resources

The server provides contextual information through resources:

- `config://current` - Current configuration file contents
- `stats://repositories` - Repository statistics by provider
- `env://status` - Environment variables and system status
- `help://commands` - Available CLI commands and usage

## Quick Start

> 📋 **Prerequisites**: Complete the main [Git PR CLI setup](../../README.md#quick-start) first, including authentication and repository configuration.

### Test the MCP Server

```bash
# Verify MCP server is built
make build  # Builds both git-pr-cli and git-pr-mcp

# Test the server
./git-pr-mcp --help
```

## IDE Integration

Choose your IDE for detailed setup instructions:

- **[Claude Code](ide-guides/claude-code.md)** - Native MCP support
- **[Cursor](ide-guides/cursor.md)** - AI-first code editor
- **[VS Code](ide-guides/vscode.md)** - Visual Studio Code with MCP extensions
- **[Zed](ide-guides/zed.md)** - High-performance editor
- **[IntelliJ/JetBrains](ide-guides/intellij.md)** - Built-in MCP support
- **[Windsurf](ide-guides/windsurf.md)** - AI-powered development environment

## Common Configuration Pattern

Most IDEs use this JSON structure for MCP configuration:

```json
{
  "mcpServers": {
    "git-pr-automation": {
      "command": "/absolute/path/to/git-pr-mcp",
      "args": [],
      "cwd": "/absolute/path/to/project",
      "env": {
        "GITHUB_TOKEN": "your_token_here",
        "GITLAB_TOKEN": "your_token_here"
      }
    }
  }
}
```

**Configuration Notes:**

- Use **absolute paths** for production setups
- Ensure the binary has execute permissions (`chmod +x git-pr-mcp`)
- Set `cwd` to your project root directory
- Environment variables can be set in config or inherited from shell
- Restart your IDE after configuration changes

## Usage Examples

Once configured, interact with the tool through natural language:

### Basic Commands

- "Check the status of all pull requests"
- "Merge all ready dependabot PRs"
- "Show me repository statistics"
- "Validate my configuration"

### Advanced Usage

- "Check PRs only for GitHub repositories"
- "Do a dry run of merging PRs to see what would happen"
- "Show me what environment variables are missing"

### With Parameters

- "Check PRs and show detailed output"
- "Merge PRs for repositories matching pattern 'web-*'"

## Configuration Examples

Pre-configured setup files are available in the [examples](examples/) directory:

- `claude-desktop.json` - Claude Desktop configuration
- `cursor-mcp.json` - Cursor MCP configuration
- `vscode-settings.json` - VS Code workspace settings
- `zed-settings.json` - Zed configuration
- `intellij-mcp.xml` - IntelliJ IDEA MCP configuration

## Environment Variables

> 📋 **Environment Setup**: See the main [README authentication section](../../README.md#configure-authentication) for complete environment variable setup.

The MCP server inherits environment variables from its parent process. You can configure them in your MCP configuration:

```json
{
  "mcpServers": {
    "git-pr-automation": {
      "command": "./git-pr-mcp",
      "env": {
        "GITHUB_TOKEN": "your_token",
        "GITLAB_TOKEN": "your_token"
      }
    }
  }
}
```

## Troubleshooting

### MCP Server Not Starting

1. **Check binary exists and is executable:**

   ```bash
   ls -la git-pr-mcp
   chmod +x git-pr-mcp  # If needed
   ```

2. **Test server manually:**

   ```bash
   ./git-pr-mcp --help
   ```

3. **Check IDE logs** for MCP-related errors
4. **Use absolute paths** in configuration
5. **Restart IDE** after configuration changes

### Tools Not Working

1. **Verify environment variables:**

   ```bash
   echo $GITHUB_TOKEN
   git-pr-cli validate --check-auth
   ```

2. **Check repositories:**

   ```bash
   git-pr-cli validate --check-repos
   ```

3. **Test CLI directly:**

   ```bash
   git-pr-cli check --help
   ```

### Common Issues

- **Permission denied**: Run `chmod +x git-pr-mcp`
- **Command not found**: Use absolute paths in configuration
- **Environment variables missing**: Set in config or launch IDE from terminal
- **Configuration errors**: Validate JSON syntax with `jq . config.json`

## Security

- MCP server runs locally and executes commands in your project directory
- Environment variables containing tokens are processed securely
- All operations use the same safety features as the CLI tool
- Dry-run capabilities allow safe previewing of actions

## Development

To extend the MCP server:

1. **Edit source files** in `cmd/git-pr-mcp/` and `internal/mcp/`
2. **Rebuild:**

   ```bash
   make build
   # or
   go build -o git-pr-mcp ./cmd/git-pr-mcp
   ```

3. **Test changes:** Restart your IDE to pick up the new binary

## Support

- See the main [README](../../README.md) for general usage
- Check [Configuration Reference](../user-guide/configuration.md) for config options
- File issues on the GitHub repository
