package utils

import (
	"context"
	"runtime"

	"golang.org/x/sync/errgroup"
)

// WorkerPool has been removed as it was over-engineered for this use case.
// Use ParallelExecutor or Batch for most parallel processing needs.

// ParallelExecutor executes functions in parallel with controlled concurrency
type ParallelExecutor struct {
	concurrency int
	logger      *Logger
}

// NewParallelExecutor creates a new parallel executor
func NewParallelExecutor(concurrency int) *ParallelExecutor {
	if concurrency <= 0 {
		concurrency = runtime.NumCPU()
	}

	return &ParallelExecutor{
		concurrency: concurrency,
		logger:      GetGlobalLogger(),
	}
}

// Execute executes functions in parallel and collects results
func (pe *ParallelExecutor) Execute(ctx context.Context, tasks []func(context.Context) error) error {
	if len(tasks) == 0 {
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(pe.concurrency)

	for i, task := range tasks {
		i, task := i, task // Capture loop variables
		g.Go(func() error {
			pe.logger.Debugf("Starting task %d", i)
			err := task(ctx)
			if err != nil {
				pe.logger.WithError(err).Errorf("Task %d failed", i)
			} else {
				pe.logger.Debugf("Task %d completed", i)
			}
			return err
		})
	}

	return g.Wait()
}

// Batch has been simplified. For batch processing needs,
// use ParallelExecutor with chunked data or implement
// specific batch logic in the calling code.
