.PHONY: help build test clean install deps fmt lint vet deadcode gitleaks ci cross-compile release dev
.PHONY: check merge setup validate config stats test-system run-cli run-mcp

# Build variables
BINARY_CLI := git-pr-cli
BINARY_MCP := git-pr-mcp
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commitSHA=$(COMMIT_SHA)
GOFLAGS := -v

# Configuration
CONFIG_FILE := config.yaml
LOG_DIR := logs

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

# Include environment variables from .env file if it exists
ifneq (,$(wildcard .env))
    include .env
    export
endif

# Default target
default: help

## help: Display this help message
help:
	@echo "Git PR CLI - Available Commands:"
	@echo ""
	@echo "${GREEN}Build and Development:${NC}"
	@grep -E '^## (build|test|clean|run)' Makefile | sed 's/##//' | column -t -s ':'
	@echo ""
	@echo "${GREEN}Setup and Installation:${NC}"
	@grep -E '^## (install|setup|deps)' Makefile | sed 's/##//' | column -t -s ':' | head -12
	@echo ""
	@echo "${GREEN}Core Operations:${NC}"
	@grep -E '^## (check|merge|validate|watch)' Makefile | sed 's/##//' | column -t -s ':'
	@echo ""
	@echo "${GREEN}Quality and CI:${NC}"
	@grep -E '^## (fmt|lint|vet|gitleaks|ci)' Makefile | sed 's/##//' | column -t -s ':'
	@echo ""
	@echo "${GREEN}Distribution:${NC}"
	@grep -E '^## (cross-compile|release)' Makefile | sed 's/##//' | column -t -s ':'
	@echo ""

## build: Build both CLI and MCP server binaries
build:
	@echo "${BLUE}Building $(BINARY_CLI) version $(VERSION)...${NC}"
	@go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_CLI) ./cmd/git-pr-cli
	@echo "${GREEN}Build complete: $(BINARY_CLI)${NC}"
	@echo "${BLUE}Building $(BINARY_MCP) version $(VERSION)...${NC}"
	@go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_MCP) ./cmd/git-pr-mcp
	@echo "${GREEN}Build complete: $(BINARY_MCP)${NC}"

## test: Run all tests with coverage
test:
	@echo "${BLUE}Running tests...${NC}"
	@go test -v -cover ./...

## test-coverage: Generate detailed coverage report
test-coverage:
	@echo "${BLUE}Running tests with coverage...${NC}"
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "${GREEN}Coverage report generated: coverage.html${NC}"

## clean: Remove build artifacts and temporary files
clean:
	@echo "${BLUE}Cleaning build artifacts...${NC}"
	@rm -f $(BINARY_CLI) $(BINARY_MCP)
	@rm -rf bin/ releases/
	@rm -f coverage.out coverage.html
	@rm -rf tmp/ logs/*.log
	@echo "${GREEN}Clean complete${NC}"

## install: Install required dependencies
install:
	@echo "${BLUE}Installing dependencies...${NC}"
	@if command -v brew >/dev/null 2>&1; then \
		make install-macos; \
	elif command -v apt-get >/dev/null 2>&1; then \
		make install-linux-apt; \
	elif command -v yum >/dev/null 2>&1; then \
		make install-linux-yum; \
	elif command -v dnf >/dev/null 2>&1; then \
		make install-linux-dnf; \
	elif command -v pacman >/dev/null 2>&1; then \
		make install-linux-pacman; \
	elif command -v winget >/dev/null 2>&1; then \
		make install-windows-winget; \
	elif command -v choco >/dev/null 2>&1; then \
		make install-windows-choco; \
	elif command -v scoop >/dev/null 2>&1; then \
		make install-windows-scoop; \
	else \
		echo "${RED}Unsupported package manager. Please install jq, yq, and curl manually.${NC}"; \
		echo "For Windows: Install winget, chocolatey, or scoop package manager"; \
		exit 1; \
	fi

## install-macos: Install dependencies on macOS using Homebrew
install-macos:
	@if ! command -v brew >/dev/null 2>&1; then \
		echo "${RED}Homebrew not found. Please install it first: https://brew.sh${NC}"; \
		exit 1; \
	fi
	@echo "${BLUE}Installing dependencies via Homebrew...${NC}"
	@brew install jq yq curl gh || true
	@echo "${GREEN}macOS dependencies installed${NC}"

## install-linux-apt: Install dependencies on Ubuntu/Debian
install-linux-apt:
	@echo "${BLUE}Installing dependencies via apt...${NC}"
	@sudo apt-get update -qq
	@sudo apt-get install -y jq curl wget
	@wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
	@chmod +x /usr/local/bin/yq
	@echo "${GREEN}Linux (apt) dependencies installed${NC}"

## install-linux-yum: Install dependencies on CentOS/RHEL
install-linux-yum:
	@echo "${BLUE}Installing dependencies via yum...${NC}"
	@sudo yum install -y jq curl wget
	@wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
	@chmod +x /usr/local/bin/yq
	@echo "${GREEN}Linux (yum) dependencies installed${NC}"

## install-linux-dnf: Install dependencies on Fedora
install-linux-dnf:
	@echo "${BLUE}Installing dependencies via dnf...${NC}"
	@sudo dnf install -y jq curl wget
	@wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
	@chmod +x /usr/local/bin/yq
	@echo "${GREEN}Linux (dnf) dependencies installed${NC}"

## install-linux-pacman: Install dependencies on Arch Linux
install-linux-pacman:
	@echo "${BLUE}Installing dependencies via pacman...${NC}"
	@sudo pacman -Sy jq curl wget --noconfirm
	@wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
	@chmod +x /usr/local/bin/yq
	@echo "${GREEN}Linux (pacman) dependencies installed${NC}"

## install-windows-winget: Install dependencies on Windows using winget
install-windows-winget:
	@echo "${BLUE}Installing dependencies via winget...${NC}"
	@winget install --id Microsoft.PowerShell --silent
	@winget install --id stedolan.jq --silent
	@winget install --id mikefarah.yq --silent
	@winget install --id cURL.cURL --silent
	@echo "${GREEN}Windows (winget) dependencies installed${NC}"

## install-windows-choco: Install dependencies on Windows using Chocolatey
install-windows-choco:
	@echo "${BLUE}Installing dependencies via chocolatey...${NC}"
	@choco install jq yq curl -y
	@echo "${GREEN}Windows (chocolatey) dependencies installed${NC}"

## install-windows-scoop: Install dependencies on Windows using Scoop
install-windows-scoop:
	@echo "${BLUE}Installing dependencies via scoop...${NC}"
	@scoop install jq yq curl
	@echo "${GREEN}Windows (scoop) dependencies installed${NC}"

## deps: Download and tidy Go dependencies
deps:
	@echo "${BLUE}Downloading Go dependencies...${NC}"
	@go mod download
	@go mod tidy
	@echo "${GREEN}Dependencies updated${NC}"

## deps-upgrade: Upgrade all dependencies to their latest versions
deps-upgrade:
	@echo "${BLUE}Upgrading all dependencies...${NC}"
	@go get -u ./...
	@go mod tidy
	@echo "${GREEN}Dependencies upgraded to latest versions${NC}"

## fmt: Format Go code
fmt:
	@echo "${BLUE}Formatting code...${NC}"
	@go fmt ./...
	@echo "${GREEN}Code formatted${NC}"

## lint: Run Go linter (requires golangci-lint)
lint:
	@echo "${BLUE}Running linter...${NC}"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "${YELLOW}golangci-lint not installed. Install with:${NC}"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## vet: Run Go vet
vet:
	@echo "${BLUE}Running go vet...${NC}"
	@go vet ./...

## deadcode: Run deadcode analysis
deadcode:
	@echo "${BLUE}Running deadcode analysis...${NC}"
	@if command -v deadcode >/dev/null 2>&1; then \
		deadcode -test ./...; \
	else \
		echo "${YELLOW}deadcode not installed. Install with:${NC}"; \
		echo "  go install golang.org/x/tools/cmd/deadcode@latest"; \
	fi

## gitleaks: Run gitleaks secret scanning
gitleaks:
	@echo "${BLUE}Running gitleaks secret scanning...${NC}"
	@if command -v gitleaks >/dev/null 2>&1; then \
		gitleaks dir . --verbose; \
	else \
		echo "${YELLOW}gitleaks not installed. Install with:${NC}"; \
		if command -v brew >/dev/null 2>&1; then \
			echo "  brew install gitleaks"; \
		elif command -v apt-get >/dev/null 2>&1; then \
			echo "  # Download latest release from GitHub:"; \
			echo "  curl -sSfL https://api.github.com/repos/gitleaks/gitleaks/releases/latest | grep browser_download_url | grep linux_x64 | cut -d '\"' -f 4 | xargs curl -sSfL -o gitleaks.tar.gz"; \
			echo "  tar -xzf gitleaks.tar.gz && sudo mv gitleaks /usr/local/bin/"; \
		elif command -v yum >/dev/null 2>&1 || command -v dnf >/dev/null 2>&1; then \
			echo "  # Download latest release from GitHub:"; \
			echo "  curl -sSfL https://api.github.com/repos/gitleaks/gitleaks/releases/latest | grep browser_download_url | grep linux_x64 | cut -d '\"' -f 4 | xargs curl -sSfL -o gitleaks.tar.gz"; \
			echo "  tar -xzf gitleaks.tar.gz && sudo mv gitleaks /usr/local/bin/"; \
		elif command -v pacman >/dev/null 2>&1; then \
			echo "  # Download latest release from GitHub:"; \
			echo "  curl -sSfL https://api.github.com/repos/gitleaks/gitleaks/releases/latest | grep browser_download_url | grep linux_x64 | cut -d '\"' -f 4 | xargs curl -sSfL -o gitleaks.tar.gz"; \
			echo "  tar -xzf gitleaks.tar.gz && sudo mv gitleaks /usr/local/bin/"; \
		elif command -v winget >/dev/null 2>&1; then \
			echo "  winget install --id trufflesecurity.gitleaks"; \
		elif command -v choco >/dev/null 2>&1; then \
			echo "  choco install gitleaks"; \
		elif command -v scoop >/dev/null 2>&1; then \
			echo "  scoop install gitleaks"; \
		else \
			echo "  # Download from releases page: https://github.com/gitleaks/gitleaks/releases"; \
			echo "  # For Windows: Download .exe from releases page"; \
			echo "  # Or use Docker: docker run --rm -v \$$(pwd):/path ghcr.io/gitleaks/gitleaks:latest dir /path"; \
		fi; \
	fi

## ci: Run all CI checks (fmt, vet, lint, deadcode, gitleaks, test)
ci: fmt vet lint deadcode gitleaks test
	@echo "${GREEN}All CI checks completed successfully${NC}"

## run-cli: Build and run the CLI
run-cli: build
	@echo "${BLUE}Running $(BINARY_CLI)...${NC}"
	@./$(BINARY_CLI)

## run-mcp: Build and run the MCP server
run-mcp: build
	@echo "${BLUE}Running $(BINARY_MCP)...${NC}"
	@./$(BINARY_MCP)

## check: Check pull request status across repositories
check: build
	@echo "${BLUE}Checking pull request status...${NC}"
	@mkdir -p $(LOG_DIR)
	@./$(BINARY_CLI) check $(ARGS) | tee $(LOG_DIR)/check-$$(date +%Y%m%d_%H%M%S).log

## merge: Merge ready pull requests
merge: build
	@echo "${BLUE}Merging ready pull requests...${NC}"
	@mkdir -p $(LOG_DIR)
	@./$(BINARY_CLI) merge $(ARGS) | tee $(LOG_DIR)/merge-$$(date +%Y%m%d_%H%M%S).log

## setup: Run setup wizard for configuration
setup: build
	@echo "${BLUE}Running setup wizard...${NC}"
	@./$(BINARY_CLI) setup wizard

## validate: Validate configuration and connectivity
validate: build
	@echo "${BLUE}Validating configuration...${NC}"
	@./$(BINARY_CLI) validate --check-repos

## stats: Show repository and PR statistics
stats: build
	@echo "${BLUE}Generating statistics...${NC}"
	@./$(BINARY_CLI) stats --detailed

## watch: Continuously monitor pull requests
watch: build
	@echo "${BLUE}Starting watch mode...${NC}"
	@./$(BINARY_CLI) watch --interval=30s

## test-system: Test system functionality and integrations
test-system: build
	@echo "${BLUE}Testing system functionality...${NC}"
	@./$(BINARY_CLI) test --notifications

## cross-compile: Build binaries for multiple platforms
cross-compile:
	@echo "${BLUE}Building for multiple platforms...${NC}"
	@mkdir -p bin/

	@echo "Building CLI for Darwin AMD64..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_CLI)-darwin-amd64 ./cmd/git-pr-cli

	@echo "Building CLI for Darwin ARM64..."
	@GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_CLI)-darwin-arm64 ./cmd/git-pr-cli

	@echo "Building CLI for Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_CLI)-linux-amd64 ./cmd/git-pr-cli

	@echo "Building CLI for Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_CLI)-linux-arm64 ./cmd/git-pr-cli

	@echo "Building CLI for Windows AMD64..."
	@GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_CLI)-windows-amd64.exe ./cmd/git-pr-cli

	@echo "Building MCP for Darwin AMD64..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_MCP)-darwin-amd64 ./cmd/git-pr-mcp

	@echo "Building MCP for Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_MCP)-linux-amd64 ./cmd/git-pr-mcp

	@echo "${GREEN}Cross-compilation complete. Binaries in bin/${NC}"

## release: Create release artifacts
release: clean cross-compile
	@echo "${BLUE}Creating release artifacts...${NC}"
	@mkdir -p releases/
	@for file in bin/*; do \
		if [ -f "$$file" ]; then \
			base=$$(basename $$file); \
			tar czf "releases/$$base-$(VERSION).tar.gz" -C bin/ $$base; \
			echo "Created: releases/$$base-$(VERSION).tar.gz"; \
		fi \
	done
	@echo "${GREEN}Release artifacts created in releases/${NC}"

## dev: Development mode with file watching (requires entr)
dev:
	@if command -v entr >/dev/null 2>&1; then \
		find . -name "*.go" | entr -r make run-cli; \
	else \
		echo "${YELLOW}entr not installed. Install with:${NC}"; \
		if command -v brew >/dev/null 2>&1; then \
			echo "  brew install entr"; \
		elif command -v apt-get >/dev/null 2>&1; then \
			echo "  sudo apt-get install entr"; \
		elif command -v yum >/dev/null 2>&1; then \
			echo "  sudo yum install entr"; \
		elif command -v dnf >/dev/null 2>&1; then \
			echo "  sudo dnf install entr"; \
		elif command -v pacman >/dev/null 2>&1; then \
			echo "  sudo pacman -S entr"; \
		elif command -v winget >/dev/null 2>&1; then \
			echo "  # entr not available via winget - use WSL or alternative file watcher"; \
		elif command -v choco >/dev/null 2>&1; then \
			echo "  # entr not available via chocolatey - use WSL or alternative file watcher"; \
		elif command -v scoop >/dev/null 2>&1; then \
			echo "  # entr not available via scoop - use WSL or alternative file watcher"; \
		else \
			echo "  # Platform not detected - install entr manually or use alternative file watcher"; \
		fi; \
		exit 1; \
	fi

# Convenience aliases for backward compatibility
dry-run: ARGS=--dry-run
dry-run: merge

check-json: ARGS=--output json
check-json: check

.DEFAULT_GOAL := help