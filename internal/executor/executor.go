package executor

import (
	"context"
	"fmt"

	"github.com/pacphi/git-pr-manager/pkg/config"
	"github.com/pacphi/git-pr-manager/pkg/merge"
	"github.com/pacphi/git-pr-manager/pkg/pr"
	"github.com/pacphi/git-pr-manager/pkg/providers"
	"github.com/pacphi/git-pr-manager/pkg/providers/common"
	"github.com/pacphi/git-pr-manager/pkg/utils"
)

// Executor orchestrates all PR operations
type Executor struct {
	config        *config.Config
	providers     map[string]common.Provider
	prProcessor   *pr.Processor
	mergeExecutor *merge.Executor
	logger        *utils.Logger
}

// New creates a new executor instance
func New(cfg *config.Config) (*Executor, error) {
	logger := utils.GetGlobalLogger()

	// Initialize providers using the factory
	factory := providers.NewFactory(cfg)
	providers, err := factory.CreateProviders()
	if err != nil {
		return nil, fmt.Errorf("failed to create providers: %w", err)
	}

	logger.Debugf("Initialized %d provider(s)", len(providers))
	for name := range providers {
		logger.Debugf("Provider initialized: %s", name)
	}

	// Create processors
	prProcessor := pr.NewProcessor(providers, cfg)
	mergeExecutor := merge.NewExecutor(providers, cfg)

	return &Executor{
		config:        cfg,
		providers:     providers,
		prProcessor:   prProcessor,
		mergeExecutor: mergeExecutor,
		logger:        logger,
	}, nil
}

// ProcessPRs processes PRs across all configured repositories
func (e *Executor) ProcessPRs(ctx context.Context, opts pr.ProcessOptions) ([]pr.ProcessResult, error) {
	return e.prProcessor.ProcessAllPRs(ctx, opts)
}

// MergePRs merges ready PRs
func (e *Executor) MergePRs(ctx context.Context, results []pr.ProcessResult, opts merge.MergeOptions) ([]merge.MergeResult, error) {
	return e.mergeExecutor.MergePRs(ctx, results, opts)
}

// ValidateMergeability validates that PRs can be merged
func (e *Executor) ValidateMergeability(ctx context.Context, results []pr.ProcessResult) error {
	return e.mergeExecutor.ValidateMergeability(ctx, results)
}

// TestAuthentication tests authentication for all configured providers
func (e *Executor) TestAuthentication(ctx context.Context) error {
	for name, provider := range e.providers {
		e.logger.Infof("Testing authentication for %s", name)
		if err := provider.Authenticate(ctx); err != nil {
			return fmt.Errorf("authentication failed for %s: %w", name, err)
		}
		e.logger.Infof("Authentication successful for %s", name)
	}
	return nil
}

// GetProviders returns the configured providers
func (e *Executor) GetProviders() map[string]common.Provider {
	return e.providers
}

// GetConfig returns the configuration
func (e *Executor) GetConfig() *config.Config {
	return e.config
}

// Close closes the executor and cleans up resources
func (e *Executor) Close() error {
	// Cleanup resources if needed
	return nil
}
