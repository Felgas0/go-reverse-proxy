package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
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
