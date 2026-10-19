package main

import (
	"log"
	"net/http"

	"github.com/vitalmaksim/products/internal/config"
	"github.com/vitalmaksim/products/internal/handler"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health)

	err = http.ListenAndServe(cfg.HTTPAddr, mux)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
