// Errors provides custom HTTP error handling functionality for the reverse proxy.
package errors

import (
	"log"
	"net/http"
)

// HTTPError represents an HTTP error response with status code and message.
type HTTPError struct {
	Status  int
	Message string
}

// HandleError writes an HTTP error response and logs the error details.
// It ensures consistent error handling across the application.
//
// Parameters:
//   - w: The response writer to send the error response
//   - err: The HTTPError containing status code and message
//   - logger: Logger instance to record the error
func HandleError(w http.ResponseWriter, err HTTPError, logger *log.Logger) {
	logger.Printf("Error: %s (status: %d)", err.Message, err.Status)
	http.Error(w, err.Message, err.Status)
}
