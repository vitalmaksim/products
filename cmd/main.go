package main

import (
	"fmt"
	"net/http"

	"github.com/vitalmaksim/products/internal/handler"
)

func main() {
	fmt.Println("test")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health)

	http.ListenAndServe(":8081", mux)
}
