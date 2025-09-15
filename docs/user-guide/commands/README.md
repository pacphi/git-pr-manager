# Command Reference

Git PR CLI provides several commands for managing pull requests across multiple repositories.

## Global Flags

These flags are available for all commands:

```bash
  -c, --config string       config file (default is config.yaml)
  -h, --help                help for git-pr-cli
      --log-format string   log format (text, json) (default "text")
      --log-level string    log level (debug, info, warn, error) (default "info")
  -q, --quiet               quiet output
  -v, --verbose             verbose output
      --version             version for git-pr-cli
```

## Commands

### Core Operations

- **[check](check.md)** - Check pull request status across repositories
- **[merge](merge.md)** - Merge ready pull requests
- **[watch](watch.md)** - Continuously monitor pull request status

### Configuration & Setup

- **[setup](setup.md)** - Set up Git PR automation configuration
- **[validate](validate.md)** - Validate configuration and connectivity

### Information

- **[stats](stats.md)** - Show repository and PR statistics

### Utility

- **[completion](completion.md)** - Generate shell autocompletion

## Quick Examples

```bash
# Check all repositories
git-pr-cli check

# Merge with dry-run first
git-pr-cli merge --dry-run
git-pr-cli merge

# Watch continuously
git-pr-cli watch --interval=5m

# Set up new configuration
git-pr-cli setup wizard

# Validate everything
git-pr-cli validate --check-repos

# Get detailed statistics
git-pr-cli stats --detailed
```

## Command Categories

### **Daily Operations**

- `check` - See what's ready to merge
- `merge` - Execute merges
- `stats` - Monitor activity

### **Setup & Configuration**

- `setup wizard` - Initial configuration
- `validate` - Verify setup

### **Monitoring**

- `watch` - Continuous monitoring
- `stats` - Activity analysis

### **Advanced**

- `completion` - Shell integration
- Provider-specific flags and options
