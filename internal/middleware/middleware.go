// Middleware provides HTTP middleware components for the reverse proxy
package middleware

import (
	"net/http"
)

// Middleware defines a function type that wraps an http.Handler with additional functionality
type Middleware func(http.Handler) http.Handler

// Chain applies multiple middleware functions to a handler in reverse order.
// The first middleware in the slice will be the outermost (first to handle the request).
func Chain(h http.Handler, middleware ...Middleware) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}
