# go-reverse-proxy

![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)

A production-grade, concurrent HTTP reverse proxy and load balancer built from scratch in Go. Designed to demonstrate concurrency patterns, lock-free primitives, network resilience, and zero-downtime lifecycle management.

---

## Architecture Overview

```text
                        ┌──────────────────────────────────────────────┐
                        │              Client Request                  │
                        └──────────────────────┬───────────────────────┘
                                               │
                                               ▼
                        ┌──────────────────────────────────────────────┐
                        │          Per-IP Rate Limiting                │
                        │   (Token Bucket Algorithm / 429 Guard)       │
                        └──────────────────────┬───────────────────────┘
                                               │ (Allowed)
                                               ▼
                        ┌──────────────────────────────────────────────┐
                        │          Round-Robin Load Balancer           │
                        │       (Atomic Counter / Lock-Free)           │
                        └───────┬──────────────┬──────────────┬────────┘
                                │              │              │
                     ┌──────────┘              │              └──────────┐
                     ▼                         ▼                         ▼
             ┌──────────────┐          ┌──────────────┐          ┌──────────────┐
             │  Backend 1   │          │  Backend 2   │          │  Backend 3   │
             │ :8081 (UP)   │          │ :8082 (UP)   │          │ :8083 (DOWN) │
             └──────────────┘          └──────────────┘          └──────────────┘
                     ▲                         ▲                         ▲
                     │                         │                         │
                     └─────────────────────────┴─────────────────────────┘
                                  Active Health Checks
                       (Periodic TCP Dial / Context-Aware Ticker)
```

---

## Key Features & Concurrency Design

* **Round-Robin Load Balancing**: Dispatches requests evenly across healthy backends using `sync/atomic.Uint64` for thread-safe, lock-free index incrementation.
* **Active Health Checking**: Background goroutine monitors backend node health via TCP connection probes (`net.DialTimeout`), automatically marking unhealthy backends and restoring them when they recover.
* **Per-IP Token Bucket Rate Limiting**: Implements the Token Bucket algorithm with custom mutex locks (`sync.Mutex`), tracking refill timestamps per remote IP and guarding against burst traffic with HTTP `429 Too Many Requests`.
* **Graceful Drain & Shutdown**: Listens for OS termination signals (`SIGINT`, `SIGTERM`) using `signal.NotifyContext`. Background workers cancel instantly, while active HTTP requests are given a bounded 10-second drain window before process exit (`http.Server.Shutdown`).
* **Connection Hardening**: Explicit timeout configurations (`ReadTimeout`, `IdleTimeout`) preventing slowloris socket exhaustion.

---

## Project Structure

```text
go-reverse-proxy/
├── cmd/
│   ├── backend/
│   │   └── main.go          # Mock backend servers for local testing
│   └── proxy/
│       └── main.go          # Application entrypoint, CLI flags & graceful shutdown
├── internal/
│   ├── proxy/
│   │   ├── balancer.go      # Server pool, backend registry & atomic round-robin
│   │   └── health.go        # Context-aware background TCP health checker
│   └── ratelimiter/
│       └── limiter.go       # Per-IP token bucket rate limiter & HTTP middleware
├── go.mod
└── README.md
```

---

## Getting Started

### Prerequisites
* Go 1.21 or higher installed

### 1. Start Mock Backends
Spin up two backend instances on different ports:

```bash
# Terminal 1
go run cmd/backend/main.go -port 8081

# Terminal 2
go run cmd/backend/main.go -port 8082
```

### 2. Start the Reverse Proxy
Run the proxy configured to balance between both backends:

```bash
# Terminal 3
go run cmd/proxy/main.go \
  -port :8080 \
  -url "http://localhost:8081,http://localhost:8082" \
  -rate 5 \
  -burst 10 \
  -interval 3s
```

---

## CLI Configuration Flags

| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `-port` | `string` | `:8080` | Address and port for the proxy to listen on |
| `-url` | `string` | `http://localhost:8080` | Comma-separated list of target backend URLs |
| `-burst` | `float64` | `5` | Maximum token bucket burst capacity per IP |
| `-rate` | `float64` | `1` | Token refill rate per second per IP |
| `-interval`| `duration`| `5s` | Frequency of active TCP backend health checks |

---

## Verification & Testing

### 1. Verify Round-Robin Load Balancing
Send multiple requests and observe alternate responses:

```bash
for i in {1..4}; do curl -s http://localhost:8080; done
```
*Output:*
```text
Hello from backend :8081
Hello from backend :8082
Hello from backend :8081
Hello from backend :8082
```

### 2. Verify Rate Limiting
Flood the proxy to exceed bucket burst capacity:

```bash
for i in {1..8}; do curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080; done
```
*Output:*
```text
200
200
200
200
200
429
429
429
```

### 3. Verify Graceful Shutdown
Simulate a slow request and trigger `Ctrl+C`:
1. Add a temporary `time.Sleep(3 * time.Second)` in `cmd/backend/main.go`.
2. Send a request: `curl -i http://localhost:8080`.
3. Press `Ctrl+C` in the proxy terminal while the request is in-flight.
4. **Result:** The proxy halts new connections, waits for the in-flight request to finish with `200 OK`, exits background health checkers, and cleanly closes.

---

## Benchmarking

Benchmark throughput and latency under load using `hey`:

```bash
hey -n 10000 -c 50 http://localhost:8080/
```
