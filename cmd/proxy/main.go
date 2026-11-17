package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/Felgas0/go-reverse-proxy/internal/proxy"
)

func main2() {
	u, err := url.Parse("http://localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	rel, err := u.Parse("/")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(rel)

	revProxy := httputil.NewSingleHostReverseProxy(rel)

	log.Fatal(http.ListenAndServe(":8081", revProxy))
}

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

	pool := proxy.NewServerPool(backends)
	go pool.StartHealthCheck(5 * time.Second)
	log.Fatal(http.ListenAndServe(":8080", pool))
}
