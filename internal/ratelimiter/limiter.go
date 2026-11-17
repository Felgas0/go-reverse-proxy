package ratelimiter

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type ClientBucket struct {
	tokens         float64
	lastRefillTime time.Time
}

type RateLimiter struct {
	mutex      sync.Mutex
	clients    map[string]*ClientBucket
	maxTokens  float64
	refillRate float64
}

func NewRateLimiter(maxTokens float64, refillRate float64) *RateLimiter {
	return &RateLimiter{
		clients:    map[string]*ClientBucket{},
		maxTokens:  maxTokens,
		refillRate: refillRate,
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	clientBucket, exists := rl.clients[ip]
	if !exists {
		rl.clients[ip] = &ClientBucket{
			tokens: rl.maxTokens - 1,
			lastRefillTime: time.Now(),
		}
		return true
	}

	elapsed := time.Since(clientBucket.lastRefillTime).Seconds()
	clientBucket.tokens = min(rl.maxTokens, clientBucket.tokens+(elapsed*rl.refillRate))
	clientBucket.lastRefillTime = time.Now()

	if clientBucket.tokens >= 1 {
		clientBucket.tokens -= 1
		return true
	}
	return false
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		if !rl.Allow(ip) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
