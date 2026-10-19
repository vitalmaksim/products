package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vitalmaksim/products/internal/config"
	"github.com/vitalmaksim/products/internal/handler"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("cannot connect to psql %v", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("unable to ping psql %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health)

	err = http.ListenAndServe(cfg.HTTPAddr, mux)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
