# go-reverse-proxy

## Proximately working tree

go-reverse-proxy/
├── cmd/
│   └── proxy/
│       └── main.go          # Entrypoint, CLI flags, graceful shutdown
├── internal/
│   ├── proxy/
│   │   ├── balancer.go      # Backend pool, Round-Robin & atomic counters
│   │   └── health.go        # Background active health checker
│   └── ratelimit/
│       └── limiter.go       # Per-IP Token Bucket limiter + background cleanup
├── testserver/
│   └── main.go              # Dummy backend servers to test load balancing locally
├── go.mod
└── README.md
