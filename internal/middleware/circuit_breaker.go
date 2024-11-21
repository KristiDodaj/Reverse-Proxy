// Circuit breaker middleware to handle service failures and prevent cascading failures
package middleware

import (
	"log"
	"net/http"
	"reverse_proxy/internal/errors"
	"sync"
	"time"
)

// CircuitState represents the possible states of the circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern to handle service failures
type CircuitBreaker struct {
	mu               sync.RWMutex
	backends         map[string]*backendState
	failureThreshold int
	timeout          time.Duration
}

// backendState represents the state of a single backend server in the circuit breaker
type backendState struct {
	state           CircuitState
	failureCount    int
	lastFailureTime time.Time
}

// NewCircuitBreaker creates a new circuit breaker with default settings
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		backends:         make(map[string]*backendState),
		failureThreshold: 5,
		timeout:          10 * time.Second,
	}
}

// OnBackendFailure handles a failed request by incrementing the failure counter
// and potentially opening the circuit if the threshold is reached.
func (cb *CircuitBreaker) OnBackendFailure(backend string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state, exists := cb.backends[backend]
	if !exists {
		state = &backendState{state: StateClosed}
		cb.backends[backend] = state
	}

	state.failureCount++
	state.lastFailureTime = time.Now()

	if state.failureCount >= cb.failureThreshold {
		state.state = StateOpen
	}
}

// OnBackendSuccess handles a successful request by resetting the failure counter
// and potentially closing the circuit if in half-open state.
func (cb *CircuitBreaker) OnBackendSuccess(backend string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state, exists := cb.backends[backend]
	if !exists {
		return
	}

	state.failureCount = 0
	if state.state == StateHalfOpen {
		state.state = StateClosed
	}
}

// IsBackendOpen checks if the circuit is currently open (failing).
// If the timeout period has elapsed, transitions to half-open state.
// Returns:
//   - bool: true if circuit is open and requests should be blocked
func (cb *CircuitBreaker) IsBackendOpen(backend string) bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	state, exists := cb.backends[backend]
	if !exists {
		cb.mu.RUnlock()
		cb.mu.Lock()
		cb.backends[backend] = &backendState{state: StateClosed}
		cb.mu.Unlock()
		cb.mu.RLock()
		return false
	}

	if state.state == StateOpen {
		if time.Since(state.lastFailureTime) > cb.timeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			state.state = StateHalfOpen
			cb.mu.Unlock()
			cb.mu.RLock()
			return false
		}
		return true
	}
	return false
}

// Middleware wraps an http.Handler with circuit breaker functionality.
// Blocks requests when the circuit is open, allows a single request through
// when half-open, and tracks success/failure of requests to manage circuit state.
// Parameters:
//   - next: The handler to wrap with circuit breaking
//
// Returns:
//   - http.Handler: A new handler that implements circuit breaking
func (cb *CircuitBreaker) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backend := r.URL.Host
		if cb.IsBackendOpen(backend) {
			errors.HandleError(w, errors.HTTPError{
				Status:  http.StatusServiceUnavailable,
				Message: "Circuit breaker is open",
			}, log.Default())
			return
		}
		wrapped := &ResponseWriter{ResponseWriter: w}
		next.ServeHTTP(wrapped, r)

		if wrapped.StatusCode >= 500 {
			cb.OnBackendFailure(backend)
		} else {
			cb.OnBackendSuccess(backend)
		}
	})
}
