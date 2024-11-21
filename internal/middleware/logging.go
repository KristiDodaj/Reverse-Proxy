// Logging middleware logs request method, path, status code, and duration
package middleware

import (
	"log"
	"net/http"
	"time"
)

// logFormat defines the format string for logging request information
const logFormat = "method=%s path=%s status=%d duration=%v"

// Logging creates middleware that logs request method, path, status code, and duration
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &ResponseWriter{ResponseWriter: w}

		next.ServeHTTP(wrapped, r)

		log.Printf(logFormat,
			r.Method,
			r.URL.Path,
			wrapped.StatusCode,
			time.Since(start),
		)
	})
}
