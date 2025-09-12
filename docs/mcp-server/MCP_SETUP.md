# Multi-Gitter PR Automation MCP Server Setup Guide

This guide explains how to set up and use the Multi-Gitter PR Automation MCP (Model Context Protocol) server with various IDEs and AI coding assistants.

## Overview

The Multi-Gitter PR Automation MCP server provides a natural language interface to the existing Makefile-based PR automation tool. It allows you to interact with the tool through AI assistants in your IDE using conversational commands like "check all PRs" or "merge ready dependabot updates."

**Architecture:**

The MCP server acts as a **tool provider** that exposes automation capabilities to your IDE's AI assistant. The server itself doesn't make LLM calls - instead, your IDE's AI (Claude, GPT, etc.) calls the MCP server's tools based on your natural language requests.

## Features

### Available Tools

The MCP server exposes the following tools that map to Makefile targets:

#### Setup & Configuration

- `setup_repositories` - Run the interactive setup wizard to configure repositories automatically
- `validate_config` - Validate the configuration file and check for common issues
- `backup_restore_config` - Backup or restore configuration files

#### PR Management

- `check_pull_requests` - Check pull request status across all configured repositories
- `merge_pull_requests` - Merge ready pull requests across configured repositories
- `watch_repositories` - Continuously monitor PR status (limited in MCP)

#### Repository Tools

- `get_repository_stats` - Get statistics about configured repositories
- `test_notifications` - Test Slack and email notification configuration
- `lint_scripts` - Lint shell scripts using shellcheck

#### Utility Tools

- `check_dependencies` - Check if required dependencies are installed
- `install_dependencies` - Install required dependencies automatically

### Available Resources

The server provides contextual information through resources:

- `config://current` - Current YAML configuration file
- `stats://repositories` - Repository statistics by provider
- `makefile://targets` - Available Makefile targets with descriptions
- `env://status` - Environment variables and dependencies status

## Common Prerequisites

Before setting up any IDE, complete these common steps:

1. **Build the MCP Server**

   ```bash
   cd mcp-server
   go build -o multi-gitter-pr-a8n-mcp .
   ```

2. **Set Up Environment Variables** (same as for the original tool)

   ```bash
   export GITHUB_TOKEN="your_github_token"
   export GITLAB_TOKEN="your_gitlab_token"
   export BITBUCKET_USERNAME="your_bitbucket_username"
   export BITBUCKET_APP_PASSWORD="your_bitbucket_app_password"
   # Optional
   export SLACK_WEBHOOK_URL="your_slack_webhook_url"
   ```

3. **Install Dependencies**

   ```bash
   make install  # or make install-macos / make install-linux
   ```

4. **Configure Repositories**

   ```bash
   make setup-wizard  # Interactive configuration
   # or
   make setup-config  # Copy sample and edit manually
   ```

## IDE-Specific Setup

Choose your IDE for detailed setup instructions:

- **[Claude Code](CLAUDE_CODE.md)** - Native MCP support
- **[Claude Desktop](CLAUDE_DESKTOP.md)** - Standalone Claude app
- **[Cursor](CURSOR.md)** - AI-first code editor
- **[VSCode](VSCODE.md)** - Visual Studio Code with MCP extensions
- **[Zed](ZED.md)** - High-performance editor
- **[IntelliJ IDEA / JetBrains IDEs](INTELLIJ.md)** - Built-in MCP support (2025.2+)
- **[Windsurf](WINDSURF.md)** - AI-powered development environment

## Configuration File Templates

Pre-configured IDE setup files are available in the `samples/` directory:

- `claude-desktop.json` - Claude Desktop configuration
- `cursor-mcp.json` - Cursor MCP configuration
- `vscode-settings.json` - VSCode workspace settings
- `zed-mcp.json` - Zed configuration
- `intellij-mcp.xml` - IntelliJ IDEA MCP configuration
- `windsurf-mcp.json` - Windsurf configuration

Copy the appropriate file to your IDE's configuration location as described in the IDE-specific guides.

## Common Configuration Pattern

Most IDEs use variations of this JSON structure (except IntelliJ which uses XML):

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

**Key Configuration Notes:**

- Use **absolute paths** for production setups
- Ensure the binary has execute permissions (`chmod +x`)
- Set `cwd` to your project root directory
- Restart your IDE after configuration changes

## LLM Provider Configuration

**Important:** The MCP server itself doesn't need LLM configuration - it only provides tools to your IDE's AI assistant.

**LLM configuration is handled by your IDE:**

- **Claude Desktop**: Uses Anthropic's Claude models (configured in Claude Desktop settings)
- **Cursor**: Uses OpenAI GPT models by default (configured in Cursor settings)
- **VSCode**: Depends on AI extension (GitHub Copilot, Claude extension, etc.)
- **Other IDEs**: Each has its own AI provider configuration

The MCP server works with any LLM that your IDE supports - it's provider-agnostic.

## Usage Examples

Once configured, you can interact with the tool through natural language in your AI assistant:

### Basic Commands

- "Check the status of all pull requests"
- "Merge all ready dependabot PRs"
- "Show me repository statistics"
- "Validate my configuration"
- "Install missing dependencies"

### Advanced Usage

- "Check PRs only for GitHub repositories"
- "Do a dry run of merging PRs to see what would happen"
- "Test my notification setup"
- "Show me what environment variables are missing"

### With Parameters

- "Check PRs with JSON output format"
- "Backup my current configuration"
- "Install dependencies for Linux platform"

## Common Troubleshooting

### MCP Server Not Starting

1. **Check the binary exists and is executable:**

   ```bash
   ls -la mcp-server/multi-gitter-pr-a8n-mcp
   chmod +x mcp-server/multi-gitter-pr-a8n-mcp  # If needed
   ```

2. **Test the server manually:**

   ```bash
   cd /path/to/your/project
   ./mcp-server/multi-gitter-pr-a8n-mcp
   ```

3. **Use absolute paths:** Relative paths can cause issues in IDE configurations
4. **Check working directory:** Ensure `cwd` points to project root

### Tools Not Working

1. **Verify environment variables:**

   ```bash
   echo $GITHUB_TOKEN
   echo $GITLAB_TOKEN
   echo $BITBUCKET_USERNAME
   ```

2. **Check dependencies:**

   ```bash
   make check-deps
   ```

3. **Validate configuration:**

   ```bash
   make validate
   ```

### IDE Not Recognizing MCP Server

1. **Check IDE MCP support:** Ensure your IDE version supports MCP
2. **Restart IDE:** Always restart after adding MCP configuration
3. **Check IDE logs:** Look for MCP-related errors in IDE logs
4. **Verify JSON syntax:** Use `jq . config.json` to validate configuration files

### Platform-Specific Issues

#### macOS

- Ensure Xcode Command Line Tools are installed
- Use full paths starting with `/Users/`

#### Linux

- Verify execute permissions: `chmod +x multi-gitter-pr-a8n-mcp`
- Check if binary is compatible with your Linux distribution

### Environment Variable Issues

**How Environment Variables Work:**
The MCP server inherits all environment variables from its parent process and passes them to the underlying `make` commands. This means your GitHub, GitLab, and other tokens are automatically available.

**Common Solutions:**

1. **Launch IDE from terminal** (most reliable):
   ```bash
   export GITHUB_TOKEN="your_token"
   export GITLAB_TOKEN="your_token"
   code .  # or your IDE command
   ```

2. **Set in IDE configuration** (see IDE-specific guides for syntax):
   - Most IDEs support `env` blocks in MCP configuration
   - Avoids terminal dependency but requires per-IDE setup

3. **Use system environment** rather than shell-specific variables:
   - Add to `~/.profile` or system environment settings
   - Ensures availability across all applications

## Security Considerations

- The MCP server runs locally and executes commands in your project directory
- Environment variables containing tokens are accessed but not logged
- All operations use the same safety features as the original Makefile tool
- Consider running in a separate directory if concerned about command execution

## Development

To modify or extend the MCP server:

1. **Edit Go source files** in `mcp-server/`
2. **Rebuild:**

   ```bash
   cd mcp-server
   go build -o multi-gitter-pr-a8n-mcp .
   ```

3. **Test:** Restart your IDE/AI assistant to pick up changes

## Support

- See the main project README for general usage
- Check `docs/TROUBLESHOOTING.md` for common issues
- File issues on the project GitHub repository
