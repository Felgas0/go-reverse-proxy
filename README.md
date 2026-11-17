# go-reverse-proxy

> An HTTP reverse proxy and load balancer built from scratch to learn Go, concurrency patterns, and network internals.

## Project Structure

```text
go-reverse-proxy/
├── cmd/
│   ├── backend/
│   │   └── main.go          # Dummy backend servers to test load balancing locally
│   └── proxy/
│       └── main.go          # Entrypoint, CLI flags, graceful shutdown
├── internal/
│   ├── proxy/
│   │   ├── balancer.go      # Backend pool, Round-Robin & atomic counters
│   │   └── health.go        # Background active health checker
│   └── ratelimit/
│       └── limiter.go       # Per-IP Token Bucket limiter + background cleanup
├── go.mod
└── README.md
```
