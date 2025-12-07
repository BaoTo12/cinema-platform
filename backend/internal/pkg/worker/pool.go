package worker

import (
	"context"
	"sync"
	"time"

	"cinemaos-backend/internal/pkg/logger"

	"go.uber.org/zap"
)

// Job represents a unit of work to be processed
type Job struct {
	ID      string
	Type    string
	Payload interface{}
	Handler func(ctx context.Context, payload interface{}) error
}

// Result represents the outcome of a job
type Result struct {
	JobID   string
	Success bool
	Error   error
	Time    time.Duration
}

// Pool is a worker pool for processing jobs concurrently
type Pool struct {
	name       string
	workers    int
	jobs       chan Job
	results    chan Result
	logger     *logger.Logger
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	isRunning  bool
	mu         sync.RWMutex
}

// NewPool creates a new worker pool
// workers: number of concurrent workers
// queueSize: size of the job queue buffer
func NewPool(name string, workers, queueSize int, log *logger.Logger) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		name:    name,
		workers: workers,
		jobs:    make(chan Job, queueSize),
		results: make(chan Result, queueSize),
		logger:  log,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start starts all workers in the pool
func (p *Pool) Start() {
	p.mu.Lock()
	if p.isRunning {
		p.mu.Unlock()
		return
	}
	p.isRunning = true
	p.mu.Unlock()

	p.logger.Info("starting worker pool",
		zap.String("pool", p.name),
		zap.Int("workers", p.workers),
	)

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker is the goroutine that processes jobs
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	p.logger.Debug("worker started",
		zap.String("pool", p.name),
		zap.Int("worker_id", id),
	)

	for {
		select {
		case <-p.ctx.Done():
			p.logger.Debug("worker shutting down",
				zap.String("pool", p.name),
				zap.Int("worker_id", id),
			)
			return

		case job, ok := <-p.jobs:
			if !ok {
				return // Channel closed
			}

			start := time.Now()

			// Execute the job handler
			err := p.executeJob(job)

			result := Result{
				JobID:   job.ID,
				Success: err == nil,
				Error:   err,
				Time:    time.Since(start),
			}

			// Non-blocking send to results channel
			select {
			case p.results <- result:
			default:
				// Results channel is full, log and discard
				p.logger.Warn("results channel full, discarding result",
					zap.String("job_id", job.ID),
				)
			}
		}
	}
}

// executeJob runs a job with panic recovery
func (p *Pool) executeJob(job Job) (err error) {
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error("job panicked",
				zap.String("job_id", job.ID),
				zap.String("job_type", job.Type),
				zap.Any("panic", r),
			)
			err = &JobPanicError{JobID: job.ID, Panic: r}
		}
	}()

	// Execute with context
	return job.Handler(p.ctx, job.Payload)
}

// Submit submits a job to the pool
// Returns false if the queue is full
func (p *Pool) Submit(job Job) bool {
	p.mu.RLock()
	if !p.isRunning {
		p.mu.RUnlock()
		return false
	}
	p.mu.RUnlock()

	select {
	case p.jobs <- job:
		p.logger.Debug("job submitted",
			zap.String("job_id", job.ID),
			zap.String("job_type", job.Type),
		)
		return true
	default:
		// Queue is full
		p.logger.Warn("job queue full, rejecting job",
			zap.String("job_id", job.ID),
		)
		return false
	}
}

// SubmitWait submits a job and blocks until accepted or context is cancelled
func (p *Pool) SubmitWait(ctx context.Context, job Job) error {
	p.mu.RLock()
	if !p.isRunning {
		p.mu.RUnlock()
		return ErrPoolNotRunning
	}
	p.mu.RUnlock()

	select {
	case p.jobs <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-p.ctx.Done():
		return ErrPoolShutdown
	}
}

// Results returns the results channel for consuming job outcomes
func (p *Pool) Results() <-chan Result {
	return p.results
}

// Stop gracefully shuts down the worker pool
func (p *Pool) Stop(timeout time.Duration) error {
	p.mu.Lock()
	if !p.isRunning {
		p.mu.Unlock()
		return nil
	}
	p.isRunning = false
	p.mu.Unlock()

	p.logger.Info("stopping worker pool",
		zap.String("pool", p.name),
		zap.Duration("timeout", timeout),
	)

	// Signal workers to stop
	p.cancel()

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.logger.Info("worker pool stopped gracefully",
			zap.String("pool", p.name),
		)
		return nil
	case <-time.After(timeout):
		p.logger.Warn("worker pool timed out during shutdown",
			zap.String("pool", p.name),
		)
		return ErrShutdownTimeout
	}
}

// QueueSize returns the current number of pending jobs
func (p *Pool) QueueSize() int {
	return len(p.jobs)
}

// IsRunning returns whether the pool is currently running
func (p *Pool) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isRunning
}
