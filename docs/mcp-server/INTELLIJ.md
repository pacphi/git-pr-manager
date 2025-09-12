# IntelliJ IDEA MCP Integration

This guide covers setting up the Multi-Gitter PR Automation MCP server with IntelliJ IDEA and other JetBrains IDEs.

## Prerequisites

Complete the [common prerequisites](MCP_SETUP.md#common-prerequisites) in the main MCP setup guide first.

**Additional requirement:** IntelliJ IDEA 2025.2+ with built-in MCP support

## Configuration

### Using Pre-configured File

1. **Copy the configuration** from the project:

   ```bash
   cp samples/intellij-mcp.xml .idea/mcp-config.xml
   ```

2. **Update paths** in the XML file to use absolute paths:

   ```xml
   <?xml version="1.0" encoding="UTF-8"?>
   <component name="MCPConfiguration">
     <servers>
       <server name="multi-gitter-pr-automation">
         <command>/absolute/path/to/your/project/mcp-server/multi-gitter-pr-a8n-mcp</command>
         <args></args>
         <cwd>/absolute/path/to/your/project</cwd>
       </server>
     </servers>
   </component>
   ```

### Manual Configuration

1. **Open Settings**: `IntelliJ IDEA > Preferences` (macOS) or `File > Settings` (Linux)

2. **Navigate to MCP Servers**: `Tools > MCP Servers`

3. **Add new server**:
   - **Name**: `multi-gitter-pr-automation`
   - **Command**: `/absolute/path/to/your/project/mcp-server/multi-gitter-pr-a8n-mcp`
   - **Working Directory**: `/absolute/path/to/your/project`
   - **Arguments**: (leave empty)

4. **Apply and restart** IntelliJ IDEA

## Platform-Specific Notes

### macOS

- Use full paths starting with `/Users/yourusername/...`
- Ensure Xcode Command Line Tools are installed for Go builds

### Linux

- Standard Unix paths work fine
- Ensure the binary has execute permissions: `chmod +x multi-gitter-pr-a8n-mcp`

## Usage

### AI Assistant Integration

Once configured, the MCP server integrates with IntelliJ's AI assistant:

1. **Open AI Assistant**: `View > Tool Windows > AI Assistant`
2. **Use natural language commands**:
   - "Check all pull requests in my configured repositories"
   - "Show me repository statistics"
   - "Merge ready dependabot updates with a dry run first"
   - "Validate my PR automation configuration"

### Available Commands

The MCP server provides these tools to the AI assistant:

#### Setup & Configuration

- Setup repositories interactively
- Validate configuration files
- Backup/restore configurations

#### PR Management

- Check pull request status
- Merge ready pull requests
- Monitor repositories continuously

#### Utilities

- Get repository statistics
- Test notifications (Slack/email)
- Check and install dependencies

## Troubleshooting

### MCP Server Not Found

1. **Check IntelliJ version**: Ensure you have 2025.2+ with MCP support
2. **Verify absolute paths**: Relative paths may not work in IDE configurations
3. **Test binary directly**:

   ```bash
   cd /path/to/your/project
   ./mcp-server/multi-gitter-pr-a8n-mcp
   ```

### AI Assistant Not Showing Tools

1. **Restart IntelliJ** after configuration changes
2. **Check IDE logs**: `Help > Show Log in Explorer/Finder`
3. **Look for MCP errors** in the log files

### Permission Issues

**macOS/Linux**:

```bash
chmod +x /path/to/project/mcp-server/multi-gitter-pr-a8n-mcp
```

### Environment Variables Not Available

IntelliJ may not inherit your shell environment. Options:

1. **Set in IDE**: `Run > Edit Configurations > Environment Variables`
2. **Launch from terminal**: Start IntelliJ from terminal to inherit environment
3. **Use IDE environment**: Set tokens in IntelliJ's environment configuration

## Advanced Configuration

### Custom Environment Variables

In the MCP configuration, you can specify environment variables:

```xml
<server name="multi-gitter-pr-automation">
  <command>/path/to/mcp-server/multi-gitter-pr-a8n-mcp</command>
  <environment>
    <env name="LOG_LEVEL" value="debug" />
  </environment>
</server>
```

### Multiple Projects

You can configure different MCP servers for different projects by using project-specific `.idea/mcp-config.xml` files with different working directories.

## Example Usage

Here are example natural language commands you can use with IntelliJ's AI assistant:

### Basic Operations

- "Check the status of all my pull requests"
- "Show me what would happen if I merged ready PRs (dry run)"
- "Merge all approved dependency updates"

### Configuration Management

- "Validate my current PR automation setup"
- "Show me statistics about my configured repositories"
- "Test my Slack notification configuration"

### Advanced Queries

- "Check PRs only for my GitHub repositories"
- "Show me which repositories have the most pending updates"
- "What environment variables do I need to configure?"

## Integration with JetBrains Fleet

Similar configuration applies to JetBrains Fleet when MCP support becomes available. The XML configuration format may differ slightly.

## Next Steps

- Explore the main MCP setup guide for detailed usage patterns
- Configure your `config.yaml` file for your specific repositories
- Set up notifications for Slack or email
- Review the troubleshooting guide for common issues
