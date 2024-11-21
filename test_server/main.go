// Main provides a simple HTTP server for testing the reverse proxy.
// It serves customizable responses on a configurable port.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

// main initializes and starts the test HTTP server.
// It performs the following steps:
// 1. Parses command line flags for configuration
// 2. Sets up request handler with custom response
// 3. Starts the server and blocks until it exits
func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "Port to listen on")
	message := flag.String("message", "Hello from backend", "Message to return")
	flag.Parse()

	// Configure request handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Log incoming request details
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)
		// Return custom response with server identity
		fmt.Fprintf(w, "%s:%d - %s\n", r.Host, *port, *message)
	})

	// Start HTTP server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
