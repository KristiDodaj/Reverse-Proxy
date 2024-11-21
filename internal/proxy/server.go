// Package proxy implements a reverse proxy server with load balancing,
// rate limiting, and monitoring capabilities
package proxy

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"reverse_proxy/internal/config"
	"reverse_proxy/internal/errors"
	"reverse_proxy/internal/loadbalancer"
	"reverse_proxy/internal/metrics"
	"reverse_proxy/internal/middleware"
)

// startTime tracks when the server was initialized
var startTime = time.Now()

// Server represents the main reverse proxy server.
// It coordinates all components including configuration, metrics,
// rate limiting, load balancing, and circuit breaker.
type Server struct {
	cfg          *config.Config
	metrics      *metrics.Metrics
	limiter      *middleware.RateLimiter
	lb           *loadbalancer.LoadBalancer
	circuitBreak *middleware.CircuitBreaker
}

// NewServer creates a new reverse proxy server instance.
// Parameters:
//   - cfg: Configuration including listen address, timeouts, and backend servers
//
// Returns:
//   - *Server: Configured server instance with all components initialized
func NewServer(cfg *config.Config) *Server {
	cb := middleware.NewCircuitBreaker()
	return &Server{
		cfg:          cfg,
		metrics:      metrics.New(),
		limiter:      middleware.NewRateLimiter(cfg.RateLimit),
		lb:           loadbalancer.New(cfg.Backends, cb),
		circuitBreak: cb,
	}
}

// Run starts the HTTP server and begins processing requests.
// It configures the server with timeouts and handlers, then
// blocks until the server encounters an error or is shutdown.
// Returns:
//   - error: Any error that caused the server to stop
func (s *Server) Run() error {
	handler := s.createHandler()

	server := &http.Server{
		Addr:         s.cfg.ListenAddr,
		Handler:      handler,
		ReadTimeout:  s.cfg.ReadTimeout,
		WriteTimeout: s.cfg.WriteTimeout,
	}

	return server.ListenAndServe()
}

// ProxyHandler implements the core reverse proxy functionality.
// It forwards requests to backend servers and tracks metrics.
type ProxyHandler struct {
	lb           *loadbalancer.LoadBalancer
	metrics      *metrics.Metrics
	config       *config.Config
	circuitBreak *middleware.CircuitBreaker
}

// ServeHTTP implements http.Handler interface for the proxy.
// It handles incoming requests by:
// 1. Selecting a backend server
// 2. Creating and forwarding the proxy request
// 3. Returning the response to the client
// 4. Tracking metrics for the request
func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.metrics.IncRequest()
	wrapped := &middleware.ResponseWriter{ResponseWriter: w}

	backend := p.lb.Next()
	if backend == "" {
		p.metrics.IncError()
		errors.HandleError(wrapped, errors.HTTPError{
			Status:  http.StatusServiceUnavailable,
			Message: "No backends available",
		}, log.Default())
		return
	}

	// Create proxy request
	targetURL := backend + r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		p.metrics.IncError()
		errors.HandleError(wrapped, errors.HTTPError{
			Status:  http.StatusInternalServerError,
			Message: "Error creating proxy request",
		}, log.Default())
		return
	}

	// Forward request
	client := &http.Client{Timeout: p.config.WriteTimeout}
	resp, err := client.Do(proxyReq)
	if err != nil {
		p.metrics.IncError()
		p.circuitBreak.OnBackendFailure(backend)
		errors.HandleError(wrapped, errors.HTTPError{
			Status:  http.StatusBadGateway,
			Message: "Error forwarding request",
		}, log.Default())
		return
	}
	defer resp.Body.Close()

	p.circuitBreak.OnBackendSuccess(backend)

	if _, err := io.Copy(wrapped, resp.Body); err != nil {
		p.metrics.IncError()
		errors.HandleError(wrapped, errors.HTTPError{
			Status:  http.StatusInternalServerError,
			Message: "Error copying response",
		}, log.Default())
		return
	}

	p.metrics.IncResponse()
}

// createHandler sets up the HTTP request processing pipeline.
// It configures:
// 1. The main proxy handler
// 2. Middleware chain (rate limiting, logging)
// 3. Routes for health checks and metrics
// Returns:
//   - http.Handler: The fully configured request handler
func (s *Server) createHandler() http.Handler {
	proxy := &ProxyHandler{
		lb:           s.lb,
		metrics:      s.metrics,
		config:       s.cfg,
		circuitBreak: s.circuitBreak,
	}

	handler := middleware.Chain(
		proxy,
		s.limiter.Middleware,
		s.circuitBreak.Middleware,
		middleware.Logging,
	)

	mux := http.NewServeMux()
	mux.Handle("/", handler)
	mux.Handle("/health", healthHandler())
	mux.Handle("/metrics", s.metrics.Handler())

	return mux
}

// healthHandler returns an HTTP handler for health checks.
// It responds with server status and uptime information in JSON format.
// Returns:
//   - http.Handler: Health check endpoint handler
func healthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		health := map[string]interface{}{
			"status": "UP",
			"uptime": time.Since(startTime).String(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	})
}
