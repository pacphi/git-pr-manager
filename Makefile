# Multi-Gitter Pull-Request Automation Makefile
# Provides convenient commands for managing PRs across multiple repositories

.PHONY: help check-deps check-prs merge-prs status dry-run install install-macos install-linux setup clean validate test

# Default configuration file
CONFIG_FILE ?= config.yaml

# Detect OS
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	OS := macos
else ifeq ($(UNAME_S),Linux)
	OS := linux
else
	OS := unknown
endif

# Environment variables for CLI tools
export GITHUB_TOKEN ?=
export GITLAB_TOKEN ?=
export BITBUCKET_USERNAME ?=
export BITBUCKET_APP_PASSWORD ?=
export BITBUCKET_WORKSPACE ?=
export SLACK_WEBHOOK_URL ?=

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "Multi-Gitter Pull-Request Automation Commands (macOS/Linux):"
	@echo "Platform: $(OS)"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(BLUE)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "First-time setup (Easy):"
	@echo "  1. make setup-full                # Complete automated setup with wizard"
	@echo ""
	@echo "First-time setup (Manual):"
	@echo "  1. make setup-config              # Copy config.sample to config.yaml"
	@echo "  2. Edit config.yaml               # Add your repositories and tokens"
	@echo "  3. make validate                  # Check your configuration"
	@echo ""
	@echo "Configuration wizard:"
	@echo "  make setup-wizard                 # Interactive repository discovery wizard"
	@echo "  make wizard-preview               # Preview what would be configured"
	@echo "  make wizard-additive              # Add repositories to existing config"
	@echo ""
	@echo "Environment Variables:"
	@echo "  CONFIG_FILE         Configuration file (default: config.yaml)"
	@echo "  GITHUB_TOKEN        GitHub personal access token"
	@echo "  GITLAB_TOKEN        GitLab personal access token"
	@echo "  BITBUCKET_USERNAME  Bitbucket username"
	@echo "  BITBUCKET_APP_PASSWORD Bitbucket app password"
	@echo "  SLACK_WEBHOOK_URL   Slack webhook URL for notifications"
	@echo ""
	@echo "Examples:"
	@echo "  make check-prs                    # Check PR status with default config"
	@echo "  make merge-prs                    # Merge ready PRs"
	@echo "  make dry-run                      # Show what would be merged"
	@echo "  make CONFIG_FILE=custom.yaml status  # Use custom config file"

check-deps: ## Check for required dependencies
	@echo "$(BLUE)[INFO]$(NC) Checking dependencies..."
ifeq ($(OS),macos)
	@command -v yq >/dev/null 2>&1 || { echo "$(RED)[ERROR]$(NC) yq is required but not installed. Run: brew install yq"; exit 1; }
	@command -v jq >/dev/null 2>&1 || { echo "$(RED)[ERROR]$(NC) jq is required but not installed. Run: brew install jq"; exit 1; }
	@command -v gh >/dev/null 2>&1 || { echo "$(YELLOW)[WARNING]$(NC) GitHub CLI (gh) not found. Install with: brew install gh"; }
else ifeq ($(OS),linux)
	@command -v yq >/dev/null 2>&1 || { echo "$(RED)[ERROR]$(NC) yq is required but not installed. Run: make install-linux"; exit 1; }
	@command -v jq >/dev/null 2>&1 || { echo "$(RED)[ERROR]$(NC) jq is required but not installed. Run: make install-linux"; exit 1; }
	@command -v gh >/dev/null 2>&1 || { echo "$(YELLOW)[WARNING]$(NC) GitHub CLI (gh) not found. Install from: https://cli.github.com/manual/installation"; }
else
	@command -v yq >/dev/null 2>&1 || { echo "$(RED)[ERROR]$(NC) yq is required but not installed. See: https://github.com/mikefarah/yq#install"; exit 1; }
	@command -v jq >/dev/null 2>&1 || { echo "$(RED)[ERROR]$(NC) jq is required but not installed. Install with your package manager"; exit 1; }
	@command -v gh >/dev/null 2>&1 || { echo "$(YELLOW)[WARNING]$(NC) GitHub CLI (gh) not found. See: https://cli.github.com/manual/installation"; }
endif
	@command -v curl >/dev/null 2>&1 || { echo "$(RED)[ERROR]$(NC) curl is required but not installed."; exit 1; }
	@echo "$(GREEN)[SUCCESS]$(NC) All required dependencies are installed"

validate: check-deps ## Validate configuration file
	@echo "$(BLUE)[INFO]$(NC) Validating configuration file: $(CONFIG_FILE)"
	@if [ ! -f "$(CONFIG_FILE)" ]; then \
		echo "$(RED)[ERROR]$(NC) Configuration file not found: $(CONFIG_FILE)"; \
		echo "$(YELLOW)[SETUP]$(NC) Please run: make setup-config"; \
		echo "$(YELLOW)[SETUP]$(NC) This will copy config.sample to config.yaml for you to customize"; \
		exit 1; \
	fi
	@yq eval '.' $(CONFIG_FILE) >/dev/null 2>&1 || { echo "$(RED)[ERROR]$(NC) Invalid YAML syntax in $(CONFIG_FILE)"; exit 1; }
	@echo "$(GREEN)[SUCCESS]$(NC) Configuration file is valid"

check-prs: validate ## Check PR status across all repositories
	@echo "$(BLUE)[INFO]$(NC) Checking PR status across repositories..."
	@CONFIG_FILE=$(CONFIG_FILE) ./check-prs.sh

check-prs-json: validate ## Check PR status and output in JSON format
	@echo "$(BLUE)[INFO]$(NC) Checking PR status (JSON output)..."
	@CONFIG_FILE=$(CONFIG_FILE) OUTPUT_FORMAT=json ./check-prs.sh

status: check-prs ## Alias for check-prs

merge-prs: validate ## Merge all ready PRs
	@echo "$(BLUE)[INFO]$(NC) Merging ready PRs across repositories..."
	@CONFIG_FILE=$(CONFIG_FILE) ./merge-prs.sh

dry-run: validate ## Show what PRs would be merged without actually merging
	@echo "$(BLUE)[INFO]$(NC) Dry run - showing what would be merged..."
	@CONFIG_FILE=$(CONFIG_FILE) DRY_RUN=true ./merge-prs.sh

force-merge: validate ## Force merge PRs even if they appear not mergeable
	@echo "$(YELLOW)[WARNING]$(NC) Force merging PRs..."
	@CONFIG_FILE=$(CONFIG_FILE) FORCE=true ./merge-prs.sh

install: ## Install required dependencies (auto-detect platform)
ifeq ($(OS),macos)
	@$(MAKE) install-macos
else ifeq ($(OS),linux)
	@$(MAKE) install-linux
else
	@echo "$(RED)[ERROR]$(NC) Unsupported operating system. Please install dependencies manually:"
	@echo "  - jq: JSON processor"
	@echo "  - yq: YAML processor (https://github.com/mikefarah/yq)"
	@echo "  - gh: GitHub CLI (optional, https://cli.github.com/)"
endif

install-macos: ## Install required dependencies on macOS
	@echo "$(BLUE)[INFO]$(NC) Installing dependencies on macOS..."
	@if command -v brew >/dev/null 2>&1; then \
		brew install yq jq gh; \
		echo "$(GREEN)[SUCCESS]$(NC) Dependencies installed successfully"; \
	else \
		echo "$(RED)[ERROR]$(NC) Homebrew not found. Please install Homebrew first: https://brew.sh"; \
		exit 1; \
	fi

install-linux: ## Install required dependencies on Linux
	@echo "$(BLUE)[INFO]$(NC) Installing dependencies on Linux..."
	@if command -v apt-get >/dev/null 2>&1; then \
		echo "$(BLUE)[INFO]$(NC) Using apt-get (Debian/Ubuntu)..."; \
		sudo apt-get update && sudo apt-get install -y jq curl; \
		sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64; \
		sudo chmod +x /usr/local/bin/yq; \
		echo "$(GREEN)[SUCCESS]$(NC) Dependencies installed successfully"; \
	elif command -v yum >/dev/null 2>&1; then \
		echo "$(BLUE)[INFO]$(NC) Using yum (RHEL/CentOS)..."; \
		sudo yum install -y jq curl; \
		sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64; \
		sudo chmod +x /usr/local/bin/yq; \
		echo "$(GREEN)[SUCCESS]$(NC) Dependencies installed successfully"; \
	elif command -v dnf >/dev/null 2>&1; then \
		echo "$(BLUE)[INFO]$(NC) Using dnf (Fedora)..."; \
		sudo dnf install -y jq curl; \
		sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64; \
		sudo chmod +x /usr/local/bin/yq; \
		echo "$(GREEN)[SUCCESS]$(NC) Dependencies installed successfully"; \
	elif command -v pacman >/dev/null 2>&1; then \
		echo "$(BLUE)[INFO]$(NC) Using pacman (Arch)..."; \
		sudo pacman -S --noconfirm jq yq curl; \
		echo "$(GREEN)[SUCCESS]$(NC) Dependencies installed successfully"; \
	else \
		echo "$(RED)[ERROR]$(NC) No supported package manager found. Please install manually:"; \
		echo "  - jq: JSON processor"; \
		echo "  - yq: YAML processor (https://github.com/mikefarah/yq)"; \
		echo "  - curl: Command line HTTP client"; \
		exit 1; \
	fi

setup-config: ## Copy config.sample to config.yaml for customization
	@if [ -f "config.yaml" ]; then \
		echo "$(YELLOW)[WARNING]$(NC) config.yaml already exists"; \
		echo "$(BLUE)[INFO]$(NC) Use 'make backup-config' to backup current config first if needed"; \
	else \
		echo "$(BLUE)[INFO]$(NC) Copying config.sample to config.yaml..."; \
		cp config.sample config.yaml; \
		echo "$(GREEN)[SUCCESS]$(NC) Configuration file created: config.yaml"; \
		echo "$(YELLOW)[NEXT]$(NC) Edit config.yaml to add your repositories and tokens"; \
		echo "$(YELLOW)[NEXT]$(NC) Then run: make validate"; \
	fi

setup: install setup-config ## Setup dependencies, config, and authenticate with Git providers
	@echo "$(BLUE)[INFO]$(NC) Setting up authentication..."
	@if command -v gh >/dev/null 2>&1; then \
		echo "$(BLUE)[INFO]$(NC) Authenticating with GitHub..."; \
		gh auth login --web; \
	fi
	@echo ""
	@echo "$(YELLOW)[NOTE]$(NC) Please set the following environment variables:"
	@echo "  export GITHUB_TOKEN=your_github_token"
	@echo "  export GITLAB_TOKEN=your_gitlab_token"
	@echo "  export BITBUCKET_USERNAME=your_bitbucket_username"
	@echo "  export BITBUCKET_APP_PASSWORD=your_bitbucket_app_password"
	@echo ""
	@echo "Or add them to your ~/.bashrc, ~/.zshrc, or ~/.profile"

test-notifications: ## Test Slack and email notification setup
	@echo "$(BLUE)[INFO]$(NC) Testing notification setup..."
	@./test-notifications.sh

test: validate ## Run basic tests to verify functionality
	@echo "$(BLUE)[INFO]$(NC) Running basic functionality tests..."
	@echo "Testing configuration parsing..."
	@yq '.repositories.github[0].name' $(CONFIG_FILE) >/dev/null 2>&1 && echo "$(GREEN)✓$(NC) GitHub repos configured" || echo "$(YELLOW)!$(NC) No GitHub repos configured"
	@yq '.repositories.gitlab[0].name' $(CONFIG_FILE) >/dev/null 2>&1 && echo "$(GREEN)✓$(NC) GitLab repos configured" || echo "$(YELLOW)!$(NC) No GitLab repos configured"
	@yq '.repositories.bitbucket[0].name' $(CONFIG_FILE) >/dev/null 2>&1 && echo "$(GREEN)✓$(NC) Bitbucket repos configured" || echo "$(YELLOW)!$(NC) No Bitbucket repos configured"
	@echo ""
	@echo "Testing scripts..."
	@if ./check-prs.sh --help >/dev/null 2>&1; then \
		echo "$(GREEN)✓$(NC) check-prs.sh script working"; \
	else \
		echo "$(RED)✗$(NC) check-prs.sh script has issues"; \
	fi
	@if ./merge-prs.sh --help >/dev/null 2>&1; then \
		echo "$(GREEN)✓$(NC) merge-prs.sh script working"; \
	else \
		echo "$(RED)✗$(NC) merge-prs.sh script has issues"; \
	fi

clean: ## Clean up temporary files and caches
	@echo "$(BLUE)[INFO]$(NC) Cleaning up..."
	@find . -name "*.tmp" -delete 2>/dev/null || true
	@find . -name ".DS_Store" -delete 2>/dev/null || true
	@echo "$(GREEN)[SUCCESS]$(NC) Cleanup completed"

# Provider-specific commands
check-github: validate ## Check only GitHub repositories
	@echo "$(BLUE)[INFO]$(NC) Checking GitHub repositories only..."
	@CONFIG_FILE=$(CONFIG_FILE) ./check-prs.sh | grep -E "github|REPOSITORY|---"

check-gitlab: validate ## Check only GitLab repositories
	@echo "$(BLUE)[INFO]$(NC) Checking GitLab repositories only..."
	@CONFIG_FILE=$(CONFIG_FILE) ./check-prs.sh | grep -E "gitlab|REPOSITORY|---"

check-bitbucket: validate ## Check only Bitbucket repositories
	@echo "$(BLUE)[INFO]$(NC) Checking Bitbucket repositories only..."
	@CONFIG_FILE=$(CONFIG_FILE) ./check-prs.sh | grep -E "bitbucket|REPOSITORY|---"

# Utility commands
stats: validate ## Show repository statistics
	@echo "$(BLUE)[INFO]$(NC) Repository statistics:"
	@echo "GitHub repos: $$(yq '.repositories.github | length' $(CONFIG_FILE) 2>/dev/null || echo 0)"
	@echo "GitLab repos: $$(yq '.repositories.gitlab | length' $(CONFIG_FILE) 2>/dev/null || echo 0)"
	@echo "Bitbucket repos: $$(yq '.repositories.bitbucket | length' $(CONFIG_FILE) 2>/dev/null || echo 0)"
	@echo "Total repos: $$(( $$(yq '.repositories.github | length' $(CONFIG_FILE) 2>/dev/null || echo 0) + $$(yq '.repositories.gitlab | length' $(CONFIG_FILE) 2>/dev/null || echo 0) + $$(yq '.repositories.bitbucket | length' $(CONFIG_FILE) 2>/dev/null || echo 0) ))"

watch: ## Continuously monitor PR status (refresh every 30 seconds)
	@echo "$(BLUE)[INFO]$(NC) Starting continuous monitoring (Ctrl+C to stop)..."
	@while true; do \
		clear; \
		echo "$(BLUE)[$(shell date)]$(NC) PR Status:"; \
		echo ""; \
		CONFIG_FILE=$(CONFIG_FILE) ./check-prs.sh || true; \
		echo ""; \
		echo "$(YELLOW)[INFO]$(NC) Refreshing in 30 seconds... (Ctrl+C to stop)"; \
		sleep 30; \
	done

# Configuration management
config-template: ## Create a template configuration file
	@echo "$(BLUE)[INFO]$(NC) Creating template configuration file: config-template.yaml"
	@sed 's/owner\/repo[0-9]/your-username\/your-repo/g' $(CONFIG_FILE) > config-template.yaml
ifeq ($(OS),macos)
	@sed -i '' 's/group\/project[0-9]/your-group\/your-project/g' config-template.yaml
	@sed -i '' 's/workspace\/repo[0-9]/your-workspace\/your-repo/g' config-template.yaml
else
	@sed -i 's/group\/project[0-9]/your-group\/your-project/g' config-template.yaml
	@sed -i 's/workspace\/repo[0-9]/your-workspace\/your-repo/g' config-template.yaml
endif
	@echo "$(GREEN)[SUCCESS]$(NC) Template created: config-template.yaml"

validate-config: validate ## Validate the configuration file (alias for validate)

# Development and debugging
debug: ## Run in debug mode with verbose output
	@echo "$(BLUE)[INFO]$(NC) Running in debug mode..."
	@set -x; CONFIG_FILE=$(CONFIG_FILE) ./check-prs.sh

lint: ## Lint shell scripts
	@echo "$(BLUE)[INFO]$(NC) Linting shell scripts..."
	@if command -v shellcheck >/dev/null 2>&1; then \
		shellcheck *.sh && echo "$(GREEN)[SUCCESS]$(NC) All scripts passed linting"; \
	else \
		echo "$(YELLOW)[WARNING]$(NC) shellcheck not installed. Install with: brew install shellcheck"; \
	fi

# Multi-gitter integration (if multi-gitter is installed)
multi-gitter-check: ## Check if multi-gitter is available
	@if command -v multi-gitter >/dev/null 2>&1; then \
		echo "$(GREEN)[SUCCESS]$(NC) multi-gitter is installed: $$(multi-gitter --version)"; \
	else \
		echo "$(YELLOW)[WARNING]$(NC) multi-gitter not found. Install from: https://github.com/lindell/multi-gitter"; \
	fi

# Backup and restore
backup-config: ## Backup current configuration
	@echo "$(BLUE)[INFO]$(NC) Backing up configuration..."
	@cp $(CONFIG_FILE) $(CONFIG_FILE).backup.$$(date +%Y%m%d_%H%M%S)
	@echo "$(GREEN)[SUCCESS]$(NC) Configuration backed up"

restore-config: ## Restore configuration from latest backup
	@echo "$(BLUE)[INFO]$(NC) Restoring configuration from latest backup..."
	@if ls $(CONFIG_FILE).backup.* 1> /dev/null 2>&1; then \
		latest=$$(ls -t $(CONFIG_FILE).backup.* | head -1); \
		cp "$$latest" $(CONFIG_FILE); \
		echo "$(GREEN)[SUCCESS]$(NC) Configuration restored from $$latest"; \
	else \
		echo "$(RED)[ERROR]$(NC) No backup files found"; \
		exit 1; \
	fi

# Configuration wizard commands
setup-wizard: ## Run the interactive configuration wizard
	@echo "$(BLUE)[INFO]$(NC) Starting configuration wizard..."
	@CONFIG_FILE=$(CONFIG_FILE) ./setup-wizard.sh

config-wizard: setup-wizard ## Alias for setup-wizard

wizard-preview: ## Run the configuration wizard in preview mode
	@echo "$(BLUE)[INFO]$(NC) Running configuration wizard in preview mode..."
	@CONFIG_FILE=$(CONFIG_FILE) ./setup-wizard.sh --preview

wizard-additive: ## Run the wizard in additive mode (add to existing config)
	@echo "$(BLUE)[INFO]$(NC) Running configuration wizard in additive mode..."
	@CONFIG_FILE=$(CONFIG_FILE) ./setup-wizard.sh --additive

# Enhanced setup with wizard
setup-full: install setup-config setup-wizard ## Complete setup with dependency installation, config creation, and wizard
	@echo "$(GREEN)[SUCCESS]$(NC) Full setup completed!"
	@echo "$(YELLOW)[NEXT]$(NC) Your repositories have been configured automatically"
	@echo "$(YELLOW)[NEXT]$(NC) Run 'make validate' to verify your configuration"