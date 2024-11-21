// Loadbalancer implements a round-robin load balancing algorithm
// for distributing requests across multiple backend servers.
package loadbalancer

import (
	"reverse_proxy/internal/middleware"
	"sync/atomic"
)

// LoadBalancer is a struct that manages the distribution of requests across multiple backend servers
type LoadBalancer struct {
	backends []string
	current  uint64
	cb       *middleware.CircuitBreaker
}

// New creates a new LoadBalancer instance with the provided backend server URLs and a circuit breaker.
// It initializes the round-robin counter to 0.
//
// Parameters:
//   - backends: A slice of backend server URLs to distribute requests across
//   - cb: A circuit breaker instance to check the availability of backends
//
// Returns:
//   - *LoadBalancer: A new load balancer instance
func New(backends []string, cb *middleware.CircuitBreaker) *LoadBalancer {
	return &LoadBalancer{
		backends: backends,
		current:  0,
		cb:       cb,
	}
}

// Next returns the URL of the next backend server in round-robin order.
// It uses atomic operations to safely increment the counter across multiple goroutines.
// It also checks the circuit breaker to ensure the backend is available.
//
// Returns:
//   - string: The URL of the next backend server, or empty string if no backends are available
func (lb *LoadBalancer) Next() string {
	if len(lb.backends) == 0 {
		return ""
	}

	start := atomic.AddUint64(&lb.current, 1)
	for i := 0; i < len(lb.backends); i++ {
		idx := (start + uint64(i)) % uint64(len(lb.backends))
		backend := lb.backends[idx]
		if !lb.cb.IsBackendOpen(backend) {
			return backend
		}
	}
	return ""
}
