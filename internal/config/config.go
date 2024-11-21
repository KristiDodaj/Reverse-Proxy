// Config provides configuration management for the reverse proxy server
package config

import (
	"flag"
	"strings"
	"time"
)

// Config holds all configuration parameters for the reverse proxy server.
// It includes network settings, timeouts, rate limiting, and backend server list.
type Config struct {
	ListenAddr   string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	RateLimit    int
	Backends     []string
}

// ParseFlags initializes and returns a Config struct with values from command line flags.
// Supported flags:
//   - listen: Listen address (default ":3000")
//   - read-timeout: Maximum duration for reading request (default 5s)
//   - write-timeout: Maximum duration for writing response (default 10s)
//   - rate-limit: Maximum requests per second (default 100)
//   - backends: Comma-separated list of backend server URLs (default "http://localhost:8080")
func ParseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ListenAddr, "listen", ":3000", "Listen address")
	flag.DurationVar(&cfg.ReadTimeout, "read-timeout", 5*time.Second, "Read timeout")
	flag.DurationVar(&cfg.WriteTimeout, "write-timeout", 10*time.Second, "Write timeout")
	flag.IntVar(&cfg.RateLimit, "rate-limit", 100, "Requests per second limit")
	backends := flag.String("backends", "http://localhost:8080", "Comma-separated backend servers")

	flag.Parse()
	cfg.Backends = strings.Split(*backends, ",")
	return cfg
}
