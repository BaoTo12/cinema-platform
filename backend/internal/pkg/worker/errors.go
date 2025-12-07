package worker

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrPoolNotRunning  = errors.New("worker pool is not running")
	ErrPoolShutdown    = errors.New("worker pool is shutting down")
	ErrShutdownTimeout = errors.New("shutdown timed out")
	ErrQueueFull       = errors.New("job queue is full")
)

// JobPanicError represents a panic that occurred during job execution
type JobPanicError struct {
	JobID string
	Panic interface{}
}

func (e *JobPanicError) Error() string {
	return fmt.Sprintf("job %s panicked: %v", e.JobID, e.Panic)
}
