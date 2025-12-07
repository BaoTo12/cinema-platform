package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"time"

	"cinemaos-backend/internal/pkg/logger"

	"go.uber.org/zap"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed   State = iota // Normal operation
	StateOpen                  // Failing, reject requests
	StateHalfOpen              // Testing if service recovered
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Common errors
var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// Config holds circuit breaker configuration
type Config struct {
	Name             string        // Name for logging
	MaxFailures      int           // Number of failures before opening
	Timeout          time.Duration // How long to wait before trying again
	MaxHalfOpenCalls int           // Max calls in half-open state
}

// DefaultConfig returns sensible defaults
func DefaultConfig(name string) Config {
	return Config{
		Name:             name,
		MaxFailures:      5,
		Timeout:          30 * time.Second,
		MaxHalfOpenCalls: 3,
	}
}

// CircuitBreaker prevents cascading failures
type CircuitBreaker struct {
	config    Config
	state     State
	failures  int
	successes int
	lastError error
	openedAt  time.Time
	mu        sync.RWMutex
	logger    *logger.Logger
}

// New creates a new circuit breaker
func New(config Config, log *logger.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
		logger: log,
	}
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	if !cb.canExecute() {
		cb.logger.Warn("circuit breaker rejected request",
			zap.String("name", cb.config.Name),
			zap.String("state", cb.state.String()),
		)
		return ErrCircuitOpen
	}

	// Execute the function
	err := fn(ctx)

	// Record the result
	cb.recordResult(err)

	return err
}

// canExecute checks if a request can proceed
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		// Check if timeout has passed
		if time.Since(cb.openedAt) > cb.config.Timeout {
			cb.toHalfOpen()
			return true
		}
		return false

	case StateHalfOpen:
		// Allow limited requests in half-open state
		return cb.successes+cb.failures < cb.config.MaxHalfOpenCalls
	}

	return false
}

// recordResult records success or failure
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastError = err

		switch cb.state {
		case StateClosed:
			if cb.failures >= cb.config.MaxFailures {
				cb.toOpen()
			}
		case StateHalfOpen:
			// Any failure in half-open goes back to open
			cb.toOpen()
		}
	} else {
		cb.successes++

		switch cb.state {
		case StateHalfOpen:
			// Enough successes in half-open closes the circuit
			if cb.successes >= cb.config.MaxHalfOpenCalls {
				cb.toClosed()
			}
		case StateClosed:
			// Reset failure count on success
			cb.failures = 0
		}
	}
}

func (cb *CircuitBreaker) toOpen() {
	cb.state = StateOpen
	cb.openedAt = time.Now()
	cb.logger.Warn("circuit breaker opened",
		zap.String("name", cb.config.Name),
		zap.Int("failures", cb.failures),
		zap.Error(cb.lastError),
	)
}

func (cb *CircuitBreaker) toHalfOpen() {
	cb.state = StateHalfOpen
	cb.failures = 0
	cb.successes = 0
	cb.logger.Info("circuit breaker half-open",
		zap.String("name", cb.config.Name),
	)
}

func (cb *CircuitBreaker) toClosed() {
	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.logger.Info("circuit breaker closed",
		zap.String("name", cb.config.Name),
	)
}

// State returns the current state
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Failures returns the current failure count
func (cb *CircuitBreaker) Failures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

// LastError returns the last error that occurred
func (cb *CircuitBreaker) LastError() error {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.lastError
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.lastError = nil
	cb.logger.Info("circuit breaker manually reset",
		zap.String("name", cb.config.Name),
	)
}
