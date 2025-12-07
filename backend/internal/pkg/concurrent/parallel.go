package concurrent

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

// FetchResult holds a result from a concurrent fetch operation
type FetchResult[T any] struct {
	Value T
	Error error
}

// Parallel executes multiple functions concurrently and returns all results
// If any function returns an error, it cancels the others and returns the first error
func Parallel(ctx context.Context, funcs ...func(ctx context.Context) error) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, fn := range funcs {
		fn := fn // Capture loop variable
		g.Go(func() error {
			return fn(ctx)
		})
	}
	return g.Wait()
}

// ParallelLimit executes functions concurrently with a limit on parallelism
func ParallelLimit(ctx context.Context, limit int, funcs ...func(ctx context.Context) error) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(limit)
	for _, fn := range funcs {
		fn := fn
		g.Go(func() error {
			return fn(ctx)
		})
	}
	return g.Wait()
}

// FanOut runs the same function on multiple inputs concurrently
func FanOut[T any, R any](ctx context.Context, inputs []T, fn func(ctx context.Context, input T) (R, error)) ([]R, error) {
	g, ctx := errgroup.WithContext(ctx)
	results := make([]R, len(inputs))

	for i, input := range inputs {
		i, input := i, input
		g.Go(func() error {
			result, err := fn(ctx, input)
			if err != nil {
				return err
			}
			results[i] = result
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

// Collector collects results from concurrent operations
type Collector[T any] struct {
	results []T
	mu      sync.Mutex
}

// NewCollector creates a new result collector
func NewCollector[T any]() *Collector[T] {
	return &Collector[T]{
		results: make([]T, 0),
	}
}

// Add adds a result to the collector (thread-safe)
func (c *Collector[T]) Add(result T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results = append(c.results, result)
}

// Results returns all collected results
func (c *Collector[T]) Results() []T {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.results
}

// Semaphore is a counting semaphore for limiting concurrency
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore creates a new semaphore with the given capacity
func NewSemaphore(capacity int) *Semaphore {
	return &Semaphore{
		ch: make(chan struct{}, capacity),
	}
}

// Acquire acquires a slot, blocking if none available
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release releases a slot
func (s *Semaphore) Release() {
	<-s.ch
}

// TryAcquire attempts to acquire a slot without blocking
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}
