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
	"golang.org/x/time/rate"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config %v", err)
	}

	// Время подключения к БД не должно ограничивать жизнь фоновой синхронизации.
	appContext, cancelApp := context.WithCancel(context.Background())
	defer cancelApp()

	ctx, cancel := context.WithTimeout(appContext, 5*time.Second)
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

	// Один лимит на поставщика, а не на воркера; запас в один токен исключает стартовый всплеск.
	supplierLimiter := rate.NewLimiter(rate.Limit(5), 1)
	supplierClient := supplier.NewClient(httpClient, cfg.SupplierAPI, supplierLimiter)
	syncService := service.NewSyncService(supplierClient, productRepository)
	syncManager := service.NewSyncManager(syncService)
	syncHandler := handler.NewSyncHandler(syncManager, appContext)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("GET /balance", balanceHandler.GetBalance)
	mux.HandleFunc("GET /products", catalogHandler.GetProducts)
	mux.HandleFunc("POST /sync", syncHandler.StartSync)
	mux.HandleFunc("GET /sync/status", syncHandler.GetStatus)

	err = http.ListenAndServe(cfg.HTTPAddr, mux)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
