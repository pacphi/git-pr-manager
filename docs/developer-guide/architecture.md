# Architecture Overview

Git PR CLI is built with a modular Go architecture that separates concerns and enables extensibility.

## System Architecture

```text
┌─────────────────┐    ┌─────────────────┐
│   git-pr-cli    │    │   git-pr-mcp    │
│   (CLI App)     │    │  (MCP Server)   │
└─────────────────┘    └─────────────────┘
         │                       │
         └───────────┬───────────┘
                     │
    ┌────────────────┴────────────────┐
    │         Shared Core             │
    │      (pkg/ modules)             │
    └─────────────────────────────────┘
```

## Module Structure

### Entry Points (`cmd/`)

#### CLI Application (`cmd/git-pr-cli/`)

- **Purpose**: Command-line interface for direct user interaction
- **Architecture**: Cobra-based CLI with subcommands
- **Features**: Interactive prompts, colored output, progress indicators

#### MCP Server (`cmd/git-pr-mcp/`)

- **Purpose**: Model Context Protocol server for AI assistant integration
- **Architecture**: MCP protocol server with tools and resources
- **Features**: Natural language processing, structured responses

### Shared Libraries (`pkg/`)

#### Configuration (`pkg/config/`)

```go
// Core configuration management
type Config struct {
    Auth          AuthConfig                 `yaml:"auth"`
    PRFilters     PRFilters                 `yaml:"pr_filters"`
    Repositories  map[string][]Repository   `yaml:"repositories"`
    Notifications NotificationConfig        `yaml:"notifications"`
    Settings      Settings                  `yaml:"settings"`
}
```

**Responsibilities:**

- YAML parsing with validation
- Environment variable resolution
- Configuration merging and inheritance
- Backup and restore functionality

#### Provider Abstraction (`pkg/providers/`)

**Common Interface (`pkg/providers/common/`)**

```go
type Provider interface {
    GetRepositories(ctx context.Context) ([]Repository, error)
    GetPullRequests(ctx context.Context, repo Repository) ([]PullRequest, error)
    MergePullRequest(ctx context.Context, repo Repository, pr PullRequest) error
    ValidateAccess(ctx context.Context, repo Repository) error
}
```

**Implementation Structure:**

```text
pkg/providers/
├── common/           # Shared interfaces and utilities
├── github/          # GitHub API integration
├── gitlab/          # GitLab API integration
└── bitbucket/       # Bitbucket API integration
```

#### Business Logic Modules

**PR Processing (`pkg/pr/`)**

- PR discovery across providers
- Filtering logic (actors, labels, age)
- Status evaluation and readiness
- Concurrent processing with error handling

**Merge Execution (`pkg/merge/`)**

- Strategy implementation (squash, merge, rebase)
- Provider-specific merge operations
- Pre-merge validation
- Post-merge notifications

**Notifications (`pkg/notifications/`)**

- Slack webhook integration
- SMTP email notifications
- Template-based messaging
- Delivery confirmation

**Setup Wizard (`pkg/wizard/`)**

- Interactive repository discovery
- Provider authentication testing
- Configuration generation
- Preview and validation

#### Utilities (`pkg/utils/`)

**HTTP Client (`pkg/utils/http.go`)**

```go
type Client struct {
    client      *resty.Client
    rateLimiter *rate.Limiter
    logger      *logrus.Logger
}
```

**Retry Logic (`pkg/utils/retry.go`)**

```go
func WithRetry(ctx context.Context, maxAttempts int, operation func() error) error
```

**Logging (`pkg/utils/logger.go`)**

- Structured logging with Logrus
- Context-aware logging
- Configurable formats (text, JSON)

### Internal Modules (`internal/`)

#### CLI Implementation (`internal/cli/`)

**Command Structure (`internal/cli/commands/`)**

```go
// Each command implements this pattern
func NewCheckCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "check",
        Short: "Check pull request status",
        RunE:  runCheck,
    }
    return cmd
}
```

**UI Utilities (`internal/cli/ui/`)**

- Progress bars and spinners
- Colored terminal output
- Interactive prompts
- Table formatting

#### MCP Implementation (`internal/mcp/`)

**Tool Handlers (`internal/mcp/tools/`)**

```go
type CheckPRsTool struct {
    executor *executor.Executor
}

func (t *CheckPRsTool) Execute(ctx context.Context, args map[string]interface{}) (*mcp.ToolResult, error)
```

**Resource Providers (`internal/mcp/resources/`)**

- Configuration access
- Repository statistics
- Environment status

## Design Principles

### 1. Interface-Based Design

All major components use interfaces for loose coupling:

```go
// Provider abstraction allows multiple Git services
type Provider interface {
    GetRepositories(ctx context.Context) ([]Repository, error)
    // ...
}

// Notification abstraction supports multiple channels
type Notifier interface {
    Send(ctx context.Context, message Message) error
}
```

### 2. Context-Aware Operations

All operations accept `context.Context` for:

- Cancellation support
- Timeout handling
- Request tracing
- Structured logging

### 3. Error Handling Strategy

```go
// Errors are wrapped with context
func (p *GitHubProvider) GetPullRequests(ctx context.Context, repo Repository) ([]PullRequest, error) {
    prs, err := p.client.GetPRs(repo.Name)
    if err != nil {
        return nil, fmt.Errorf("failed to get PRs for %s: %w", repo.Name, err)
    }
    return prs, nil
}
```

### 4. Concurrent Processing

Uses `golang.org/x/sync/errgroup` for safe concurrency:

```go
func (p *PRProcessor) ProcessRepositories(ctx context.Context, repos []Repository) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(p.config.MaxConcurrentRequests)

    for _, repo := range repos {
        repo := repo // capture loop variable
        g.Go(func() error {
            return p.processRepository(ctx, repo)
        })
    }

    return g.Wait()
}
```

## Data Flow

### 1. Configuration Loading

```text
config.yaml → Environment Variables → Validation → Config Struct
```

### 2. Repository Discovery

```text
Providers → Authentication → Repository List → Filtering → Final Set
```

### 3. PR Processing

```text
Repository → Get PRs → Filter PRs → Check Status → Ready List
```

### 4. Merge Execution

```text
Ready PRs → Pre-merge Validation → Merge → Post-merge Actions → Notifications
```

## Extension Points

### Adding New Providers

1. Implement the `Provider` interface
2. Add configuration schema
3. Register in provider factory
4. Add authentication handling

```go
// pkg/providers/newprovider/client.go
type NewProvider struct {
    client *http.Client
    config NewProviderConfig
}

func (p *NewProvider) GetRepositories(ctx context.Context) ([]Repository, error) {
    // Implementation
}
```

### Adding New Notification Channels

1. Implement the `Notifier` interface
2. Add configuration schema
3. Register in notification manager

```go
// pkg/notifications/teams.go
type TeamsNotifier struct {
    webhookURL string
}

func (t *TeamsNotifier) Send(ctx context.Context, message Message) error {
    // Implementation
}
```

### Adding New Commands

1. Create command in `internal/cli/commands/`
2. Register with root command
3. Add help and flag definitions

## Performance Considerations

### Rate Limiting

- Per-provider rate limiters
- Configurable limits
- Exponential backoff

### Caching

- HTTP response caching
- Repository metadata caching
- Configuration validation caching

### Memory Management

- Streaming large responses
- Bounded concurrent operations
- Garbage collection optimization

## Security

### Token Management

- Environment variable storage only
- No token logging
- Secure HTTP transport

### Input Validation

- Configuration schema validation
- API response validation
- User input sanitization

### Network Security

- TLS/HTTPS enforcement
- Certificate validation
- Proxy support

This architecture provides a solid foundation for maintaining and extending the Git PR automation tool while ensuring reliability, performance, and security.
