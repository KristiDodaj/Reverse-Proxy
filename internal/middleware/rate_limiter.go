// Rate limiter middleware for enforcing requests per second (RPS) limits.
package middleware

import (
	"log"
	"net/http"
	"reverse_proxy/internal/errors"
	"sync"
	"time"
)

// // RateLimiter implements a sliding window rate limiting algorithm.
// It tracks request timestamps within a sliding window to enforce
// requests per second (RPS) limits.
type RateLimiter struct {
	mu         sync.Mutex
	timestamps []time.Time
	rps        int
}

// NewRateLimiter creates a new rate limiter instance.
// Parameters:
//   - rps: Maximum number of requests allowed per second
//
// Returns:
//   - *RateLimiter: A new rate limiter configured with the specified RPS
func NewRateLimiter(rps int) *RateLimiter {
	return &RateLimiter{
		timestamps: make([]time.Time, 0, rps),
		rps:        rps,
	}
}

// Allow checks if a new request should be allowed based on the rate limit.
// Uses a sliding window of 1 second to determine if the request can proceed.
// Returns:
//   - bool: true if request is allowed, false if rate limit exceeded
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	window := now.Add(-time.Second)

	// Remove timestamps older than 1 second from the window
	valid := 0
	for _, ts := range rl.timestamps {
		if ts.After(window) {
			rl.timestamps[valid] = ts
			valid++
		}
	}
	rl.timestamps = rl.timestamps[:valid]

	// Allow request if under RPS limit
	if len(rl.timestamps) < rl.rps {
		rl.timestamps = append(rl.timestamps, now)
		return true
	}
	return false
}

// Middleware wraps an http.Handler with rate limiting functionality.
// Parameters:
//   - next: The handler to wrap with rate limiting
//
// Returns:
//   - http.Handler: A new handler that enforces rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.Allow() {
			errors.HandleError(w, errors.HTTPError{
				Status:  http.StatusTooManyRequests,
				Message: "Rate limit exceeded",
			}, log.Default())
			return
		}
		next.ServeHTTP(w, r)
	})
}
