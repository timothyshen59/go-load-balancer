package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	// Identify which backend handled the request
	host := os.Getenv("BACKEND_NAME")
	if host == "" {
		host = "unknown-backend"
	}
	fmt.Fprintf(w, "Hello from %s\n", host)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT env var required, e.g. PORT=8081")
	}

	http.HandleFunc("/", handler)

	log.Printf("Backend %s listening on :%s\n", os.Getenv("BACKEND_NAME"), port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
