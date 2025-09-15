# completion - Generate Shell Autocompletion

The `completion` command generates shell autocompletion scripts for Git PR CLI, enabling tab completion for commands, flags, and arguments in your shell.

## Synopsis

```bash
git-pr-cli completion [shell] [flags]
```

## Description

The completion command generates autocompletion scripts for popular shells:

- **bash**: Bourne Again SHell (most common Linux/macOS shell)
- **zsh**: Z Shell (macOS default since Catalina, popular Linux shell)
- **fish**: Friendly Interactive SHell (modern shell with enhanced features)
- **powershell**: Microsoft PowerShell (Windows and cross-platform)

Once installed, autocompletion provides:

- Command name completion
- Flag and option completion
- Argument suggestions
- File and directory path completion
- Smart context-aware suggestions

## Options

```bash
      --no-descriptions  Disable completion descriptions
  -h, --help             help for completion
```

## Supported Shells

### Bash Completion

```bash
# Generate bash completion script
git-pr-cli completion bash

# Install for current user (Linux)
git-pr-cli completion bash > ~/.local/share/bash-completion/completions/git-pr-cli

# Install for current user (macOS with Homebrew)
git-pr-cli completion bash > $(brew --prefix)/etc/bash_completion.d/git-pr-cli

# Install system-wide (requires sudo)
sudo git-pr-cli completion bash > /etc/bash_completion.d/git-pr-cli
```

**Manual Installation:**

```bash
# Add to your ~/.bashrc or ~/.bash_profile
echo 'source <(git-pr-cli completion bash)' >>~/.bashrc

# Reload your shell
source ~/.bashrc
```

### Zsh Completion

```bash
# Generate zsh completion script
git-pr-cli completion zsh

# Install for current user
git-pr-cli completion zsh > "${fpath[1]}/_git-pr-cli"

# Or add to your ~/.zshrc
echo 'source <(git-pr-cli completion zsh)' >>~/.zshrc

# For Oh My Zsh users
git-pr-cli completion zsh > ~/.oh-my-zsh/completions/_git-pr-cli
```

**Setup for Zsh:**

```bash
# Ensure completion system is initialized in ~/.zshrc
autoload -U compinit
compinit

# Reload your shell
exec zsh
```

### Fish Completion

```bash
# Generate fish completion script
git-pr-cli completion fish

# Install for current user
git-pr-cli completion fish > ~/.config/fish/completions/git-pr-cli.fish

# System-wide installation (requires sudo)
sudo git-pr-cli completion fish > /usr/share/fish/completions/git-pr-cli.fish
```

### PowerShell Completion

```powershell
# Generate PowerShell completion script
git-pr-cli completion powershell

# Add to your PowerShell profile
git-pr-cli completion powershell | Out-String | Invoke-Expression

# Or save to profile permanently
git-pr-cli completion powershell >> $PROFILE
```

## Installation Examples

### Automatic Installation Script

Create an installation script for easy setup:

```bash
#!/bin/bash
# install-completion.sh

set -e

SHELL_NAME=$(basename "$SHELL")

case "$SHELL_NAME" in
    bash)
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS
            COMPLETION_DIR="$(brew --prefix 2>/dev/null)/etc/bash_completion.d"
            if [ ! -d "$COMPLETION_DIR" ]; then
                echo "Installing bash-completion via Homebrew..."
                brew install bash-completion
            fi
        else
            # Linux
            COMPLETION_DIR="$HOME/.local/share/bash-completion/completions"
            mkdir -p "$COMPLETION_DIR"
        fi

        echo "Installing bash completion for git-pr-cli..."
        git-pr-cli completion bash > "$COMPLETION_DIR/git-pr-cli"
        echo "Completion installed! Restart your shell or run: source ~/.bashrc"
        ;;

    zsh)
        COMPLETION_DIR="$HOME/.config/zsh/completions"
        mkdir -p "$COMPLETION_DIR"

        echo "Installing zsh completion for git-pr-cli..."
        git-pr-cli completion zsh > "$COMPLETION_DIR/_git-pr-cli"

        # Add to fpath in .zshrc if not already present
        if ! grep -q "$COMPLETION_DIR" ~/.zshrc 2>/dev/null; then
            echo "fpath=($COMPLETION_DIR \$fpath)" >> ~/.zshrc
            echo "autoload -U compinit && compinit" >> ~/.zshrc
        fi

        echo "Completion installed! Restart your shell or run: exec zsh"
        ;;

    fish)
        COMPLETION_DIR="$HOME/.config/fish/completions"
        mkdir -p "$COMPLETION_DIR"

        echo "Installing fish completion for git-pr-cli..."
        git-pr-cli completion fish > "$COMPLETION_DIR/git-pr-cli.fish"
        echo "Completion installed! Restart your shell or run: exec fish"
        ;;

    *)
        echo "Unsupported shell: $SHELL_NAME"
        echo "Supported shells: bash, zsh, fish"
        echo "For PowerShell, run: git-pr-cli completion powershell | Out-String | Invoke-Expression"
        exit 1
        ;;
esac
```

### Docker/Container Usage

```bash
# Generate completion inside container
docker run --rm my-git-pr-cli completion bash

# Mount and install on host
docker run --rm -v ~/.local/share/bash-completion/completions:/completions \
  my-git-pr-cli sh -c 'git-pr-cli completion bash > /completions/git-pr-cli'
```

## Completion Features

### Command Completion

Tab completion works for all main commands:

```bash
git-pr-cli <TAB>
# Shows: check, completion, help, merge, setup, stats, test, validate, watch

git-pr-cli ch<TAB>
# Completes to: git-pr-cli check

git-pr-cli s<TAB>
# Shows: setup, stats (multiple matches)
```

### Flag Completion

Tab completion works for command flags:

```bash
git-pr-cli check --<TAB>
# Shows: --config, --debug, --dry-run, --help, --output, --provider, --repos

git-pr-cli merge --dry<TAB>
# Completes to: git-pr-cli merge --dry-run
```

### Argument Completion

Smart completion for command arguments:

```bash
git-pr-cli --config <TAB>
# Shows available .yaml files in current directory

git-pr-cli check --provider <TAB>
# Shows: bitbucket, github, gitlab

git-pr-cli stats --format <TAB>
# Shows: csv, json, text, yaml
```

### File Path Completion

Standard file/directory completion:

```bash
git-pr-cli --config my-<TAB>
# Completes file paths starting with "my-"

git-pr-cli validate --check-repos repo<TAB>
# Smart completion for repository names from config
```

## Advanced Configuration

### Custom Completion Functions

For advanced users, you can customize completion behavior:

```bash
# Add custom completion for repository names
_git_pr_cli_custom() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    case ${prev} in
        --repos)
            # Custom repository completion from config file
            opts=$(git-pr-cli completion --list-repos 2>/dev/null || echo "")
            COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
            return 0
            ;;
    esac
}

# Register custom completion
complete -F _git_pr_cli_custom git-pr-cli
```

### Completion Performance

For large configurations, completion can be optimized:

```bash
# Disable expensive completions
git-pr-cli completion bash --no-descriptions > ~/.local/share/bash-completion/completions/git-pr-cli

# Or use static completion without dynamic lookups
export GIT_PR_CLI_COMPLETION_DISABLE_DYNAMIC=1
```

## Troubleshooting

### Completion Not Working

#### Check Installation

```bash
# Verify completion script exists
ls -la ~/.local/share/bash-completion/completions/git-pr-cli

# Test completion manually
source ~/.local/share/bash-completion/completions/git-pr-cli
```

#### Shell Configuration

**Bash:**

```bash
# Ensure bash-completion is sourced in ~/.bashrc
if [ -f /etc/bash_completion ]; then
    . /etc/bash_completion
fi

# Or on macOS with Homebrew
if [ -f $(brew --prefix)/etc/bash_completion ]; then
    . $(brew --prefix)/etc/bash_completion
fi
```

**Zsh:**

```bash
# Ensure completion system is initialized in ~/.zshrc
autoload -U compinit
compinit -u

# Check if completion directory is in fpath
echo $fpath | grep -q completions || echo "Completion directory not in fpath"
```

**Fish:**

```bash
# Check fish completion directories
fish -c 'echo $fish_complete_path'

# Refresh completions
fish -c 'fish_update_completions'
```

### Slow Completion

```bash
# Disable dynamic completion features
export GIT_PR_CLI_COMPLETION_TIMEOUT=1

# Use simplified completion
git-pr-cli completion bash --no-descriptions
```

### Permission Issues

```bash
# Fix ownership of completion files
sudo chown $USER:$USER ~/.local/share/bash-completion/completions/git-pr-cli

# Fix permissions
chmod 644 ~/.local/share/bash-completion/completions/git-pr-cli
```

## Testing Completion

### Manual Testing

```bash
# Test basic completion
git-pr-cli <TAB><TAB>

# Test flag completion
git-pr-cli check --<TAB><TAB>

# Test argument completion
git-pr-cli --config <TAB><TAB>
```

### Automated Testing

```bash
#!/bin/bash
# test-completion.sh

source ~/.local/share/bash-completion/completions/git-pr-cli

# Test command completion
COMP_WORDS=(git-pr-cli "")
COMP_CWORD=1
_git_pr_cli

if [[ " ${COMPREPLY[@]} " =~ " check " ]]; then
    echo "✅ Command completion working"
else
    echo "❌ Command completion failed"
fi

# Test flag completion
COMP_WORDS=(git-pr-cli check "--")
COMP_CWORD=2
_git_pr_cli

if [[ " ${COMPREPLY[@]} " =~ " --dry-run " ]]; then
    echo "✅ Flag completion working"
else
    echo "❌ Flag completion failed"
fi
```

## Integration Examples

### IDE Integration

Many IDEs can use shell completion for their integrated terminals:

**VS Code:**

- Terminal uses system shell completion automatically
- Ensure completion is installed for your default shell

**IntelliJ IDEA:**

- Built-in terminal respects shell completion
- Configure shell path in Terminal settings

### CI/CD Integration

```yaml
# .github/workflows/test.yml
- name: Test CLI completion
  run: |
    ./git-pr-cli completion bash > /tmp/completion
    source /tmp/completion
    # Test completion functionality
```

## Best Practices

1. **Install Early**: Set up completion during initial CLI installation
2. **Update Regularly**: Regenerate completion scripts after CLI updates
3. **Test Installation**: Verify completion works after installation
4. **Shell-Specific**: Use appropriate completion for your shell
5. **Performance**: Disable expensive features for large configurations

## Related Commands

Completion enhances the experience of all Git PR CLI commands:

- `git-pr-cli help` - View command help with easier navigation
- All commands benefit from tab completion of flags and arguments
- File path completion works with `--config` flags across all commands

## Examples of Completion in Action

### Basic Usage

```bash
$ git-pr-cli <TAB>
check      completion help       merge      setup      stats      test       validate   watch

$ git-pr-cli check --<TAB>
--config       --debug        --dry-run      --help         --output       --provider     --repos        --show-details --show-status

$ git-pr-cli check --output <TAB>
json  text  yaml
```

### Advanced Usage

```bash
$ git-pr-cli merge --repos <TAB>
owner/repo1  owner/repo2  group/project1  workspace/repo1

$ git-pr-cli validate --provider <TAB>
bitbucket  github  gitlab

$ git-pr-cli stats --period <TAB>
1d   7d   30d  90d  all
```

The completion system makes Git PR CLI more user-friendly and reduces the need to remember exact command syntax and available options.
