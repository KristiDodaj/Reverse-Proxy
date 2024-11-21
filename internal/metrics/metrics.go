// Metrics provides request tracking and monitoring capabilities for the proxy server.
package metrics

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

// Metrics struct tracks the counts of requests, responses and errors processed by the proxy.
type Metrics struct {
	RequestCount  uint64
	ResponseCount uint64
	ErrorCount    uint64
}

// New creates and returns a new Metrics instance with all counters initialized to zero.
//
// Returns:
//   - *Metrics: A new metrics tracker instance
func New() *Metrics {
	return &Metrics{}
}

// IncRequest atomically increments the request counter by one.
func (m *Metrics) IncRequest() { atomic.AddUint64(&m.RequestCount, 1) }

// IncResponse atomically increments the response counter by one.
func (m *Metrics) IncResponse() { atomic.AddUint64(&m.ResponseCount, 1) }

// IncError atomically increments the error counter by one.
func (m *Metrics) IncError() { atomic.AddUint64(&m.ErrorCount, 1) }

// Handler returns an HTTP handler that exposes the current metrics as JSON.
// The metrics endpoint returns counts for total requests, responses and errors.
//
// Returns:
//   - http.Handler: Handler that serves the metrics data
func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stats := map[string]uint64{
			"requests":  m.RequestCount,
			"responses": m.ResponseCount,
			"errors":    m.ErrorCount,
		}
		json.NewEncoder(w).Encode(stats)
	})
}
