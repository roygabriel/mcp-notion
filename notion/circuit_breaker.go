package notion

import (
	"errors"
	"sync"
	"time"
)

// CircuitState represents the current state of the circuit breaker.
type CircuitState int

const (
	// CircuitClosed allows requests through normally.
	CircuitClosed CircuitState = iota
	// CircuitOpen blocks all requests.
	CircuitOpen
	// CircuitHalfOpen allows a single probe request through.
	CircuitHalfOpen
)

// ErrCircuitOpen is returned when the circuit breaker is open and rejecting requests.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// CircuitBreaker implements a three-state circuit breaker pattern for the Notion API client.
// It tracks consecutive failures and opens the circuit when the failure threshold is reached,
// preventing further requests until a reset timeout elapses.
type CircuitBreaker struct {
	mu           sync.Mutex
	failures     int
	threshold    int
	resetTimeout time.Duration
	lastFailure  time.Time
	state        CircuitState
	now          func() time.Time // injectable clock for testing
}

// NewCircuitBreaker creates a new CircuitBreaker with a failure threshold of 5
// and a reset timeout of 30 seconds.
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		threshold:    5,
		resetTimeout: 30 * time.Second,
		state:        CircuitClosed,
		now:          time.Now,
	}
}

// Allow checks whether a request is permitted. It returns true if the circuit is
// closed or if the reset timeout has elapsed (transitioning to half-open).
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if cb.now().Sub(cb.lastFailure) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		// Only one probe request at a time in half-open
		return true
	default:
		return true
	}
}

// RecordSuccess records a successful request, resetting the failure count
// and transitioning the circuit back to closed.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = CircuitClosed
}

// RecordFailure records a failed request. If the failure count reaches the
// threshold, the circuit transitions to the open state.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = cb.now()

	if cb.failures >= cb.threshold {
		cb.state = CircuitOpen
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
