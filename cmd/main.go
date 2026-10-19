package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/vitalmaksim/products/internal/handler"
)

func main() {
	fmt.Println("test")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health)

	err := http.ListenAndServe(":8081", mux)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
