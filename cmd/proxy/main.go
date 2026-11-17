package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Felgas0/go-reverse-proxy/internal/proxy"
	"github.com/Felgas0/go-reverse-proxy/internal/ratelimiter"
)


func main() {
	url := flag.String("url", "http://localhost:8080", "backend url's")
	maxTokens := flag.Float64("burst", 5, "set maxTokens")
	refillRate := flag.Float64("rate", 1, "set refillRate")
	port := flag.String("port", ":8080", "Associate a Port")
	interval := flag.Duration("interval", 5*time.Second, "health check interval")

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

	ctxTerminate, stopTerminate := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopTerminate()


	limiter := ratelimiter.NewRateLimiter(*maxTokens, *refillRate)

	pool := proxy.NewServerPool(backends)
	handlerWithLimiter := limiter.Middleware(pool)
	go pool.StartHealthCheck(ctxTerminate, *interval)

	httpServer := &http.Server{
		Addr:        ":" + *port,
		Handler:     handlerWithLimiter,
		ReadTimeout: 2 * time.Second,
		IdleTimeout: 60 * time.Second,
	}

	go func() {
		err := httpServer.ListenAndServe()
		if err != nil {
			if err != http.ErrServerClosed {
				log.Fatalf("ErrServerClosed is expected other error is occurring %s", err)
			}
		}

	}()

	<-ctxTerminate.Done()
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()


	err := httpServer.Shutdown(ctxShutDown)
	if err != nil{
		log.Printf("shutdow error %s", err)
	}
	log.Printf("Server has finished running")

}
