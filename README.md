# HTTP Reverse Proxy

A basic HTTP reverse proxy server written in Go with features for load balancing, fault tolerance, and monitoring.

## Getting Started

```bash
git clone <repository-url>
cd reverse_proxy
go mod download
go run cmd/main.go [flags]
```

### Available Flags

- `-listen`: Listen address (default `":3000"`)
- `-read-timeout`: Read timeout (default `5s`)
- `-write-timeout`: Write timeout (default `10s`)
- `-rate-limit`: Requests per second (default `100`)
- `-backends`: Comma-separated backend URLs (default `"http://localhost:8080"`)

#### Example of flags:
```bash
go run cmd/main.go --listen=:3000 --rate-limit=200 --backends="http://localhost:8080,http://localhost:8081"
```

### API Endpoints

- `/`: Main proxy endpoint
- `/health`: Health check endpoint
- `/metrics`: Monitoring metrics

### Testing Scenario

#### 1. Start Backend Servers

Terminal 1 - First backend
```bash
go run test_server/main.go --port=8080 --message="Response from server 1"
```

Terminal 2 - Second backend
```bash
go run test_server/main.go --port=8081 --message="Response from server 2"
```

#### 2. Start Proxy Server

Terminal 3 - Proxy server
```bash
go run cmd/main.go --backends="http://localhost:8080,http://localhost:8081" --rate-limit=20
```

#### 3. Test Features

#### Load Balancing

Terminal 4 - Client Side
```bash
# In terminal 4 run this command 2 times
curl http://localhost:3000/
```
> Observe: Requests are being spread evenly amongst the two backend servers.

#### Rate Limiting

Trigger rate limit by exceeding 20 requests per second:
```bash
# In terminal 4 run this command
for i in {1..40}; do curl -i http://localhost:3000/; done
```
> Observe: You should see "429 Too Many Requests" responses.

#### Circuit Breaker

The circuit breaker opens after 5 consecutive failures and stays open for 10 seconds before allowing retry attempts.

Stop one backend server:
```bash
# In Terminal 2 (second backend on 8081):
# Press Ctrl+C to stop the server
# In Terminal 4:
# Send 10 individual requests
curl http://localhost:3000/
```
> Observe: You should recieve the message "Error forwarding request" for half of the requests
```bash
# Run this command 2 more time
curl http://localhost:3000/
```
> Observe: The proxy now routes requests to the remaining live server instead of failing. After a 10-second cooldown period, the proxy will attempt to check the status of the previously faulty server and reintegrate it if it is back online.

#### Check Metrics

Monitor the proxy's metrics:
```bash
# In terminal 4 run this command
curl http://localhost:3000/metrics
```
Sample response:
```bash
{
  "errors": 3,
  "requests": 15,
  "responses": 12,
}
```

#### Health Check

```bash
# In terminal 4 run this command
curl http://localhost:3000/health
```
Sample response:
```bash
{
  "status": "UP",
  "uptime": "2m5s"
}
```

## Resources & References

This implementation was built using:

- Official Go Documentation
  - [net/http](https://pkg.go.dev/net/http) for proxy and server functionality
  - [sync](https://pkg.go.dev/sync) for concurrent access patterns 
  - [time](https://pkg.go.dev/time) for timeout handling
  - [encoding/json](https://pkg.go.dev/encoding/json) for metrics endpoints

- Design Patterns & Architecture
  - [System Design Primer](https://github.com/donnemartin/system-design-primer) for reference architecture:
    - [Load Balancing](https://github.com/donnemartin/system-design-primer#load-balancer)
    - [Reverse Proxy](https://github.com/donnemartin/system-design-primer?tab=readme-ov-file#reverse-proxy-web-server)
  - [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)
  - [Rate Limiting Patterns](https://arpitbhayani.me/blogs/sliding-window-ratelimiter/) for sliding window implementation

- Go Learning Resources
  - [Go by Example](https://gobyexample.com/) for practical patterns
  - [Effective Go](https://go.dev/doc/effective_go) for best practices
  - [Go Project Layout](https://github.com/golang-standards/project-layout) for structure

## Design Decisions

1. **Architecture Components**
- Modular package structure separating core concerns:
  - `proxy` - Core proxy logic
  - `loadbalancer` - Round-robin load balancing
  - `middleware` - Pluggable middleware chain
  - `metrics` - Request monitoring
  - `config` - Configuration management

2. **Middleware Pattern**
- Used the Chain pattern for composable middleware
- Implemented critical middleware components:
  - Rate limiting with sliding window
  - Circuit breaker for fault tolerance 
  - Request logging for observability

3. **Failure Handling**
- Circuit breaker with 3 states (Open/Closed/Half-Open)
- 5 consecutive failures trigger circuit open
- 10-second timeout before attempting recovery
- Consistent error handling through errors.HandleError

4. **Monitoring**
- Health check endpoint for uptime monitoring
- Metrics endpoint tracking requests/errors/responses
- Structured logging of requests and errors

## Limitations

1. **Persistence**
- In-memory storage for metrics and circuit breaker state
- State lost on restart
- No persistent configuration management

2. **Scalability**
- Only one proxy server instance, no coordination between multiple instances
- Rate limiting is done locally on the proxy server, not shared across multiple servers
- Simple round-robin load balancing without complex logic

3. **Security**
- No TLS/HTTPS support
- No authentication/authorization
- Limited request validation

4. **Operational**
- No graceful shutdown handling
- Limited configuration options
- Basic monitoring capabilities

## Scaling Considerations

To scale this proxy for production use, several key enhancements would be needed:

### Distributed State Management
- Replace in-memory storage with a distributed data store for:
  - Coordinated rate limiting across instances
  - Shared circuit breaker state
  - Centralized metrics collection
- Use a distributed configuration store for dynamic settings

### High Availability & Reliability 
- Deploy multiple proxy instances behind a load balancer
- Implement consistent hashing for efficient backend selection
- Use service discovery to automatically manage backend servers
- Set up connection pooling and proper timeout settings
- Implement graceful shutdown and automated failover

### Monitoring & Observability
- Add standardized metrics collection 
- Set up centralized logging
- Add detailed health checks
- Monitor backend service health

### Performance Optimization
- Enable request caching where applicable
- Implement TCP keepalive connections
- Add proper retry strategies
- Optimize timeout configurations

## Security Enhancements

### Transport Security
- Enable HTTPS with TLS
- Implement mutual TLS authentication
- Add proper certificate management

### Authentication & Authorization
- Add API key validation
- Implement role-based access control
- Rate limit per client/token

### Request Protection
- Validate and sanitize all requests
- Set request size limits 
- Configure allowed HTTP methods and content types
- Add security headers

### Infrastructure
- Run as unprivileged user
- Enable audit logging
- Restrict network access
- Regular security updates


## Happy proxying! 🚀
