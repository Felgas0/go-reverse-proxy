package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Felgas0/go-reverse-proxy/internal/proxy"
	"github.com/Felgas0/go-reverse-proxy/internal/ratelimiter"
)

const maxTokens float64 = 5
const refillRate float64 = 1


func main() {
	url := flag.String("url", "http://localhost:8080", "backend url's")
	flag.Parse()
	tokens := strings.Split(*url, ",")
	var backends []*proxy.Backend
	for _, u := range tokens {
		fmt.Println(u)
		b, err := proxy.NewBackend(u)
		if err != nil {
			log.Fatal(err)
		}

		backends = append(backends, b)
	}

	limiter := ratelimiter.NewRateLimiter(maxTokens, refillRate)

	pool := proxy.NewServerPool(backends)
	handlerWithLimiter := limiter.Middleware(pool)
	go pool.StartHealthCheck(5 * time.Second)
	log.Fatal(http.ListenAndServe(":8080", handlerWithLimiter))
}
