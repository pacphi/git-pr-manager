# IntelliJ IDEA / JetBrains IDEs MCP Integration

IntelliJ IDEA and other JetBrains IDEs (WebStorm, PyCharm, GoLand, etc.) support MCP through built-in AI assistant integration starting from version 2024.3.

## Prerequisites

1. Complete the [MCP server setup](../README.md#quick-start)
2. **IntelliJ IDEA 2024.3+** or any JetBrains IDE with AI Assistant plugin
3. **AI Assistant plugin** enabled (usually pre-installed in recent versions)

## Configuration

### Method 1: IDE Settings UI

1. Open IntelliJ IDEA / JetBrains IDE
2. Go to **File** → **Settings** (Windows/Linux) or **IntelliJ IDEA** → **Preferences** (macOS)
3. Navigate to **Tools** → **AI Assistant** → **MCP Servers**
4. Click **"+"** to add new MCP server:
   - **Name**: `git-pr-automation`
   - **Command**: `/absolute/path/to/git-pr-mcp`
   - **Working Directory**: `/absolute/path/to/your/project`
   - **Arguments**: (leave empty)

### Method 2: Configuration File

Create or edit the MCP configuration file:

**Location:**
- **Windows**: `%APPDATA%\JetBrains\<IDE><Version>\mcp-servers.xml`
- **macOS**: `~/Library/Application Support/JetBrains/<IDE><Version>/mcp-servers.xml`
- **Linux**: `~/.config/JetBrains/<IDE><Version>/mcp-servers.xml`

**Configuration:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<mcpServers>
  <server name="git-pr-automation">
    <command>/absolute/path/to/git-pr-mcp</command>
    <workingDirectory>/absolute/path/to/your/project</workingDirectory>
    <arguments></arguments>
    <environment>
      <variable name="GITHUB_TOKEN" value="${GITHUB_TOKEN}" />
      <variable name="GITLAB_TOKEN" value="${GITLAB_TOKEN}" />
      <variable name="BITBUCKET_USERNAME" value="${BITBUCKET_USERNAME}" />
      <variable name="BITBUCKET_APP_PASSWORD" value="${BITBUCKET_APP_PASSWORD}" />
    </environment>
  </server>
</mcpServers>
```

### Method 3: Project-Specific Configuration

Add to your project's `.idea/mcp-servers.xml`:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<mcpServers>
  <server name="git-pr-automation">
    <command>./git-pr-mcp</command>
    <workingDirectory>.</workingDirectory>
  </server>
</mcpServers>
```

## Environment Variables

### Option 1: Launch from Terminal (Recommended)
```bash
# Set environment variables
export GITHUB_TOKEN="ghp_your_token_here"
export GITLAB_TOKEN="glpat_your_token_here"
export BITBUCKET_USERNAME="your_username"
export BITBUCKET_APP_PASSWORD="your_app_password"

# Launch IDE from terminal
idea /path/to/your/project    # IntelliJ IDEA
webstorm /path/to/project     # WebStorm
pycharm /path/to/project      # PyCharm
# etc.
```

### Option 2: IDE Environment Configuration
1. Go to **Settings** → **Build, Execution, Deployment** → **Environment Variables**
2. Add environment variables:
   - `GITHUB_TOKEN=ghp_...`
   - `GITLAB_TOKEN=glpat_...`
   - `BITBUCKET_USERNAME=username`
   - `BITBUCKET_APP_PASSWORD=password`

### Option 3: Run Configuration Environment
For project-specific environment variables:
1. Go to **Run** → **Edit Configurations**
2. Add environment variables in the **Environment Variables** section

## Usage

Once configured and IDE is restarted:

1. **Open AI Assistant** (usually **Alt+Enter** → "AI Actions" or dedicated AI panel)
2. **Start a conversation** with the AI assistant
3. **Use natural language** to interact with Git PR automation

### Example Conversations

**Check PR Status:**
> "Can you check all pull requests across my configured repositories and show me which ones are ready to merge?"

**Merge Management:**
> "Show me all dependabot PRs that are ready to merge, then merge them after I review the list"

**Repository Analysis:**
> "What repositories do I have configured for PR automation and what's their current status?"

**Configuration Management:**
> "Validate my Git PR automation configuration and show me any issues that need to be fixed"


## Verification

Test the MCP integration:

1. **Check AI Assistant Connection:**
   > "Do you have access to Git PR automation tools? List what you can do."

2. **Test Basic Commands:**
   > "Show me the help documentation for checking pull requests"

3. **Verify Configuration:**
   > "Can you validate my current Git PR configuration?"

## Troubleshooting

### MCP Server Not Starting

1. **Check Binary Path and Permissions:**
   ```bash
   ls -la /path/to/git-pr-mcp
   chmod +x /path/to/git-pr-mcp
   ```

2. **Verify Configuration:**
   - Go to **Settings** → **Tools** → **AI Assistant** → **MCP Servers**
   - Check if the server is listed and shows "Connected" status

3. **Check IDE Logs:**
   - Go to **Help** → **Show Log in Finder/Explorer**
   - Look for `idea.log` and search for MCP-related errors

### Environment Variables Issues

1. **Test Environment Variables:**
   ```bash
   echo $GITHUB_TOKEN
   ./git-pr-mcp --help
   ```

2. **Check IDE Environment:**
   - Go to **Help** → **Debug Log Settings**
   - Add `#com.intellij.mcp` to see MCP-specific logs

### AI Assistant Not Responding

1. **Test CLI Directly:**
   ```bash
   cd /path/to/project
   ./git-pr-cli validate --check-auth
   ```

2. **Restart AI Assistant:**
   - Disable and re-enable the AI Assistant plugin
   - Restart the IDE

3. **Check Plugin Status:**
   - Go to **Settings** → **Plugins**
   - Ensure AI Assistant plugin is enabled and updated

## Advanced Configuration

### Custom Server Arguments
```xml
<server name="git-pr-automation">
  <command>/path/to/git-pr-mcp</command>
  <arguments>--log-level debug --config custom-config.yaml</arguments>
  <environment>
    <variable name="LOG_LEVEL" value="debug" />
  </environment>
</server>
```

### Multiple Projects Configuration
```xml
<mcpServers>
  <server name="git-pr-work">
    <command>/path/to/git-pr-mcp</command>
    <workingDirectory>/path/to/work-project</workingDirectory>
  </server>
  <server name="git-pr-personal">
    <command>/path/to/git-pr-mcp</command>
    <workingDirectory>/path/to/personal-project</workingDirectory>
  </server>
</mcpServers>
```

### Performance Tuning
```xml
<server name="git-pr-automation">
  <command>/path/to/git-pr-mcp</command>
  <environment>
    <variable name="LOG_LEVEL" value="warn" />
    <variable name="MAX_CONCURRENT_REQUESTS" value="3" />
    <variable name="REQUEST_TIMEOUT" value="30s" />
  </environment>
</server>
```

## JetBrains IDE Specific Features

### Code Generation
> "Generate a GitHub Actions workflow that uses git-pr-cli for automated PR merging"

### Project Structure Analysis
> "Analyze my project structure and suggest the best Git PR automation configuration"

### Integration with Version Control
- Use with built-in Git tools for enhanced workflow
- Combine with IDE's branch management features

### Run Configurations
Create custom run configurations for Git PR CLI commands:
1. **Run** → **Edit Configurations** → **+** → **Shell Script**
2. Set script path to `git-pr-cli` with desired arguments

## IDE-Specific Locations

### IntelliJ IDEA
- Config: `~/.config/JetBrains/IntelliJIdea2024.3/mcp-servers.xml`
- Logs: `~/.cache/JetBrains/IntelliJIdea2024.3/log/idea.log`

### WebStorm
- Config: `~/.config/JetBrains/WebStorm2024.3/mcp-servers.xml`
- Logs: `~/.cache/JetBrains/WebStorm2024.3/log/idea.log`

### PyCharm
- Config: `~/.config/JetBrains/PyCharm2024.3/mcp-servers.xml`
- Logs: `~/.cache/JetBrains/PyCharm2024.3/log/idea.log`

## Best Practices

1. **Use absolute paths** for reliable cross-session execution
2. **Test CLI independently** before configuring MCP
3. **Launch from terminal** for proper environment variable inheritance
4. **Monitor IDE performance** with large repository sets
5. **Use specific queries** rather than broad requests for better AI responses
6. **Keep IDE updated** for latest MCP features and improvements

## Common Issues

- **"AI Assistant not available"**: Ensure plugin is installed and enabled
- **"MCP server connection failed"**: Check paths, permissions, and logs
- **"Environment variables not found"**: Launch IDE from terminal or set in IDE
- **"Commands timing out"**: Consider performance tuning for large repo sets
- **XML configuration errors**: Validate XML syntax and structure

## Security Considerations

- Environment variables are handled securely by the IDE
- MCP server runs locally within IDE process sandbox
- All operations maintain CLI tool's built-in safety features
- Consider using IDE's password manager for sensitive tokens