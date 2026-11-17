package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.String("port", "8080", "port number")
	flag.Parse()
	fmt.Printf("Listening on %s \n", *port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, "Hello from backend :%s\n", *port)
		if err != nil {
			log.Printf("write error %v", err)
		}
	})
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
