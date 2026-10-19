package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vitalmaksim/products/internal/config"
	"github.com/vitalmaksim/products/internal/handler"
	"github.com/vitalmaksim/products/internal/repository"
	"github.com/vitalmaksim/products/internal/service"
	"github.com/vitalmaksim/products/internal/supplier"
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
	defer pool.Close()

	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("unable to ping psql %v", err)
	}

	userRepository := repository.NewUserRepository(pool)
	balanceService := service.NewBalanceService(userRepository)
	balanceHandler := handler.NewBalanceHandler(balanceService)

	productRepository := repository.NewProductRepository(pool)
	catalogService := service.NewCatalogService(productRepository)
	catalogHandler := handler.NewCatalogHandler(catalogService)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	supplierClient := supplier.NewClient(httpClient, cfg.SupplierAPI)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("GET /balance", balanceHandler.GetBalance)
	mux.HandleFunc("GET /products", catalogHandler.GetProducts)

	err = http.ListenAndServe(cfg.HTTPAddr, mux)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
