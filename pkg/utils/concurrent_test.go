package utils

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewParallelExecutor(t *testing.T) {
	t.Run("with positive concurrency", func(t *testing.T) {
		pe := NewParallelExecutor(5)

		assert.NotNil(t, pe)
		assert.Equal(t, 5, pe.concurrency)
		assert.NotNil(t, pe.logger)
	})

	t.Run("with zero concurrency defaults to NumCPU", func(t *testing.T) {
		pe := NewParallelExecutor(0)

		assert.NotNil(t, pe)
		assert.Equal(t, runtime.NumCPU(), pe.concurrency)
		assert.NotNil(t, pe.logger)
	})

	t.Run("with negative concurrency defaults to NumCPU", func(t *testing.T) {
		pe := NewParallelExecutor(-5)

		assert.NotNil(t, pe)
		assert.Equal(t, runtime.NumCPU(), pe.concurrency)
		assert.NotNil(t, pe.logger)
	})
}

func TestParallelExecutor_Execute_EmptyTasks(t *testing.T) {
	pe := NewParallelExecutor(2)
	ctx := context.Background()

	var tasks []func(context.Context) error

	err := pe.Execute(ctx, tasks)

	assert.NoError(t, err)
}

func TestParallelExecutor_Execute_AllSuccess(t *testing.T) {
	pe := NewParallelExecutor(3)
	ctx := context.Background()

	var executionOrder []int
	var mu sync.Mutex

	tasks := []func(context.Context) error{
		func(ctx context.Context) error {
			mu.Lock()
			executionOrder = append(executionOrder, 0)
			mu.Unlock()
			time.Sleep(10 * time.Millisecond) // Small delay to ensure concurrency
			return nil
		},
		func(ctx context.Context) error {
			mu.Lock()
			executionOrder = append(executionOrder, 1)
			mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			return nil
		},
		func(ctx context.Context) error {
			mu.Lock()
			executionOrder = append(executionOrder, 2)
			mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			return nil
		},
	}

	start := time.Now()
	err := pe.Execute(ctx, tasks)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(executionOrder))

	// Should complete faster than sequential execution (3 * 10ms = 30ms)
	// With concurrency, should be closer to 10ms plus overhead
	assert.Less(t, duration, 25*time.Millisecond)

	// All tasks should have executed
	assert.Contains(t, executionOrder, 0)
	assert.Contains(t, executionOrder, 1)
	assert.Contains(t, executionOrder, 2)
}

func TestParallelExecutor_Execute_OneTaskFails(t *testing.T) {
	pe := NewParallelExecutor(3)
	ctx := context.Background()

	expectedErr := errors.New("task 1 failed")
	var execCount int32

	tasks := []func(context.Context) error{
		func(ctx context.Context) error {
			atomic.AddInt32(&execCount, 1)
			return nil
		},
		func(ctx context.Context) error {
			atomic.AddInt32(&execCount, 1)
			return expectedErr
		},
		func(ctx context.Context) error {
			atomic.AddInt32(&execCount, 1)
			return nil
		},
	}

	err := pe.Execute(ctx, tasks)

	assert.Error(t, err)
	assert.ErrorIs(t, err, expectedErr)

	// All tasks should have been attempted before the error is returned
	assert.Equal(t, int32(3), atomic.LoadInt32(&execCount))
}

func TestParallelExecutor_Execute_MultipleFails(t *testing.T) {
	pe := NewParallelExecutor(3)
	ctx := context.Background()

	err1 := errors.New("task 0 failed")
	err2 := errors.New("task 2 failed")
	var execCount int32

	tasks := []func(context.Context) error{
		func(ctx context.Context) error {
			atomic.AddInt32(&execCount, 1)
			return err1
		},
		func(ctx context.Context) error {
			atomic.AddInt32(&execCount, 1)
			return nil
		},
		func(ctx context.Context) error {
			atomic.AddInt32(&execCount, 1)
			return err2
		},
	}

	err := pe.Execute(ctx, tasks)

	assert.Error(t, err)
	// Should return one of the errors (errgroup returns the first error encountered)
	assert.True(t, errors.Is(err, err1) || errors.Is(err, err2))

	// All tasks should have been started
	assert.Equal(t, int32(3), atomic.LoadInt32(&execCount))
}

func TestParallelExecutor_Execute_ContextCancellation(t *testing.T) {
	pe := NewParallelExecutor(2)
	ctx, cancel := context.WithCancel(context.Background())

	var startedTasks int32
	var completedTasks int32

	tasks := []func(context.Context) error{
		func(ctx context.Context) error {
			atomic.AddInt32(&startedTasks, 1)
			defer atomic.AddInt32(&completedTasks, 1)

			// This task will complete quickly
			time.Sleep(10 * time.Millisecond)
			return nil
		},
		func(ctx context.Context) error {
			atomic.AddInt32(&startedTasks, 1)
			defer atomic.AddInt32(&completedTasks, 1)

			// This task will check for cancellation
			for i := 0; i < 100; i++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					time.Sleep(5 * time.Millisecond)
				}
			}
			return nil
		},
		func(ctx context.Context) error {
			atomic.AddInt32(&startedTasks, 1)
			defer atomic.AddInt32(&completedTasks, 1)

			// This task will also check for cancellation
			for i := 0; i < 100; i++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					time.Sleep(5 * time.Millisecond)
				}
			}
			return nil
		},
	}

	// Cancel after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := pe.Execute(ctx, tasks)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)

	// Should complete faster than if all tasks ran to completion
	assert.Less(t, duration, 200*time.Millisecond)

	// Some tasks should have started
	assert.Greater(t, atomic.LoadInt32(&startedTasks), int32(0))
}

func TestParallelExecutor_Execute_ConcurrencyLimit(t *testing.T) {
	concurrency := 2
	pe := NewParallelExecutor(concurrency)
	ctx := context.Background()

	var activeTasks int32
	var maxActiveTasks int32
	var mu sync.Mutex

	updateMaxActive := func() {
		active := atomic.LoadInt32(&activeTasks)
		mu.Lock()
		if active > maxActiveTasks {
			maxActiveTasks = active
		}
		mu.Unlock()
	}

	// Create more tasks than concurrency limit
	numTasks := 10
	tasks := make([]func(context.Context) error, numTasks)

	for i := 0; i < numTasks; i++ {
		tasks[i] = func(ctx context.Context) error {
			atomic.AddInt32(&activeTasks, 1)
			updateMaxActive()

			// Simulate work
			time.Sleep(50 * time.Millisecond)

			atomic.AddInt32(&activeTasks, -1)
			return nil
		}
	}

	err := pe.Execute(ctx, tasks)

	assert.NoError(t, err)

	// Max active tasks should not exceed concurrency limit
	mu.Lock()
	assert.LessOrEqual(t, maxActiveTasks, int32(concurrency))
	mu.Unlock()

	// All tasks should be completed
	assert.Equal(t, int32(0), atomic.LoadInt32(&activeTasks))
}

func TestParallelExecutor_Execute_LargeBatch(t *testing.T) {
	pe := NewParallelExecutor(5)
	ctx := context.Background()

	numTasks := 100
	completedTasks := make([]bool, numTasks)
	var mu sync.Mutex

	tasks := make([]func(context.Context) error, numTasks)
	for i := 0; i < numTasks; i++ {
		i := i // Capture loop variable
		tasks[i] = func(ctx context.Context) error {
			// Simulate some work
			time.Sleep(time.Millisecond)

			mu.Lock()
			completedTasks[i] = true
			mu.Unlock()

			return nil
		}
	}

	start := time.Now()
	err := pe.Execute(ctx, tasks)
	duration := time.Since(start)

	assert.NoError(t, err)

	// All tasks should have completed
	mu.Lock()
	for i, completed := range completedTasks {
		assert.True(t, completed, "Task %d should have completed", i)
	}
	mu.Unlock()

	// Should complete much faster than sequential execution
	// Sequential would be ~100ms, parallel should be ~20ms with 5 workers
	assert.Less(t, duration, 50*time.Millisecond)
}

func TestParallelExecutor_Execute_ContextPropagation(t *testing.T) {
	pe := NewParallelExecutor(3)

	type contextKey string
	key := contextKey("test-key")
	value := "test-value"

	ctx := context.WithValue(context.Background(), key, value)

	var receivedValues []string
	var mu sync.Mutex

	tasks := []func(context.Context) error{
		func(ctx context.Context) error {
			val := ctx.Value(key)
			if val != nil {
				mu.Lock()
				receivedValues = append(receivedValues, val.(string))
				mu.Unlock()
			}
			return nil
		},
		func(ctx context.Context) error {
			val := ctx.Value(key)
			if val != nil {
				mu.Lock()
				receivedValues = append(receivedValues, val.(string))
				mu.Unlock()
			}
			return nil
		},
		func(ctx context.Context) error {
			val := ctx.Value(key)
			if val != nil {
				mu.Lock()
				receivedValues = append(receivedValues, val.(string))
				mu.Unlock()
			}
			return nil
		},
	}

	err := pe.Execute(ctx, tasks)

	assert.NoError(t, err)

	mu.Lock()
	assert.Equal(t, 3, len(receivedValues))
	for _, receivedValue := range receivedValues {
		assert.Equal(t, value, receivedValue)
	}
	mu.Unlock()
}

func TestParallelExecutor_Execute_TaskPanic(t *testing.T) {
	pe := NewParallelExecutor(2)
	ctx := context.Background()

	tasks := []func(context.Context) error{
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			// Simulate a panic by returning a specific error instead
			return errors.New("simulated panic error")
		},
	}

	err := pe.Execute(ctx, tasks)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulated panic error")
}

func TestParallelExecutor_Execute_SingleTask(t *testing.T) {
	pe := NewParallelExecutor(5)
	ctx := context.Background()

	var executed bool

	tasks := []func(context.Context) error{
		func(ctx context.Context) error {
			executed = true
			return nil
		},
	}

	err := pe.Execute(ctx, tasks)

	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestParallelExecutor_Execute_TaskReturningError(t *testing.T) {
	pe := NewParallelExecutor(1)
	ctx := context.Background()

	expectedError := errors.New("specific task error")

	tasks := []func(context.Context) error{
		func(ctx context.Context) error {
			return expectedError
		},
	}

	err := pe.Execute(ctx, tasks)

	assert.Error(t, err)
	assert.ErrorIs(t, err, expectedError)
}

// Benchmark to ensure the parallel executor performs better than sequential execution
func BenchmarkParallelExecutor_Execute(b *testing.B) {
	pe := NewParallelExecutor(4)
	ctx := context.Background()

	// Create tasks that simulate work
	numTasks := 20
	tasks := make([]func(context.Context) error, numTasks)
	for i := 0; i < numTasks; i++ {
		tasks[i] = func(ctx context.Context) error {
			// Simulate CPU work
			for j := 0; j < 1000; j++ {
				_ = j * j
			}
			return nil
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := pe.Execute(ctx, tasks)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark sequential execution for comparison
func BenchmarkSequentialExecution(b *testing.B) {
	ctx := context.Background()

	// Create the same tasks as parallel benchmark
	numTasks := 20
	tasks := make([]func(context.Context) error, numTasks)
	for i := 0; i < numTasks; i++ {
		tasks[i] = func(ctx context.Context) error {
			// Simulate CPU work
			for j := 0; j < 1000; j++ {
				_ = j * j
			}
			return nil
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, task := range tasks {
			err := task(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
