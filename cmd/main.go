// Main provides the entry point for the reverse proxy server.
// It handles configuration parsing and server initialization.
package main

import (
	"log"
	"reverse_proxy/internal/config"
	"reverse_proxy/internal/proxy"
)

// main initializes and starts the reverse proxy server.
// It performs the following steps:
// 1. Parses command line flags for configuration
// 2. Creates a new proxy server instance
// 3. Starts the server and blocks until it exits
// 4. Logs any fatal errors that occur
func main() {
	cfg := config.ParseFlags()
	server := proxy.NewServer(cfg)
	log.Printf("Starting proxy server on %s", cfg.ListenAddr)
	log.Fatal(server.Run())
}
