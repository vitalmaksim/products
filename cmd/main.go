package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() (runErr error) {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Сигнал запускает shutdown, но не отменяет активные HTTP-запросы до начала их завершения.
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	appCtx, cancelApp := context.WithCancel(context.Background())
	defer cancelApp()

	dbCtx, cancelDB := context.WithTimeout(appCtx, 5*time.Second)
	defer cancelDB()

	pool, err := pgxpool.New(dbCtx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}
	var server *http.Server
	var syncManager *service.SyncManager
	var serverErrors chan error
	defer func() {
		// Очистка нужна и при ошибке инициализации, и при неожиданной остановке HTTP.
		runErr = errors.Join(runErr, shutdown(server, serverErrors, syncManager, pool, cancelApp))
	}()

	err = pool.Ping(dbCtx)
	cancelDB()
	if err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	userRepository := repository.NewUserRepository(pool)
	balanceService := service.NewBalanceService(userRepository)
	balanceHandler := handler.NewBalanceHandler(balanceService)

	productRepository := repository.NewProductRepository(pool)
	catalogService := service.NewCatalogService(productRepository)
	catalogHandler := handler.NewCatalogHandler(catalogService)

	purchaseRepository := repository.NewPurchaseRepository(pool)
	purchaseService := service.NewPurchaseService(purchaseRepository)
	purchaseHandler := handler.NewPurchaseHandler(purchaseService)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Один лимит на поставщика, а не на воркера; запас в один токен исключает стартовый всплеск.
	supplierLimiter := rate.NewLimiter(rate.Limit(5), 1)
	supplierClient := supplier.NewClient(httpClient, cfg.SupplierAPI, supplierLimiter)
	syncService := service.NewSyncService(supplierClient, productRepository)
	syncManager = service.NewSyncManager(syncService)
	syncHandler := handler.NewSyncHandler(syncManager, appCtx)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("GET /balance", balanceHandler.GetBalance)
	mux.HandleFunc("GET /products", catalogHandler.GetProducts)
	mux.HandleFunc("POST /purchases", purchaseHandler.CreatePurchase)
	mux.HandleFunc("POST /sync", syncHandler.StartSync)
	mux.HandleFunc("GET /sync/status", syncHandler.GetStatus)

	// appCtx не является родителем HTTP-контекстов: активной покупке даем завершить транзакцию.
	server = &http.Server{Addr: cfg.HTTPAddr, Handler: mux}
	serverErrors = make(chan error, 1)
	go func(result chan<- error) {
		// Буфер позволяет завершить горутину, даже если основной поток уже получил сигнал.
		result <- server.ListenAndServe()
	}(serverErrors)

	select {
	case <-signalCtx.Done():
		return nil
	case err := <-serverErrors:
		// Результат уже получен, поэтому cleanup не должен ждать его повторно.
		serverErrors = nil
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	}
}

func shutdown(server *http.Server, serverErrors <-chan error, syncManager *service.SyncManager, pool *pgxpool.Pool, cancelApp context.CancelFunc) error {
	// Один независимый бюджет для HTTP, синхронизации и закрытия БД, а не по 15 секунд каждому.
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShutdown()

	var shutdownErr error
	httpStopped := true
	if server != nil {
		// Callback вызывается после закрытия слушателей: фоновую работу отменяем, пока HTTP дожидается запросов.
		server.RegisterOnShutdown(cancelApp)
		// Shutdown дает запросам закончиться; Close сразу закрывает соединения и отменяет их контексты.
		if err := server.Shutdown(shutdownCtx); err != nil {
			httpStopped = false
			shutdownErr = fmt.Errorf("shutdown HTTP: %w", err)
			if err := server.Close(); err != nil {
				shutdownErr = errors.Join(shutdownErr, fmt.Errorf("force close HTTP: %w", err))
			}
		}
	}
	// Нужен и при ошибке до запуска сервера; повторный cancel безопасен.
	cancelApp()

	syncStopped := true
	if syncManager != nil {
		if err := syncManager.Shutdown(shutdownCtx); err != nil {
			syncStopped = false
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown sync: %w", err))
		}
	}
	if serverErrors != nil {
		select {
		case err := <-serverErrors:
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				shutdownErr = errors.Join(shutdownErr, fmt.Errorf("serve HTTP: %w", err))
			}
		case <-shutdownCtx.Done():
			httpStopped = false
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("wait HTTP server: %w", shutdownCtx.Err()))
		}
	}
	if !httpStopped || !syncStopped {
		// Close пула под живой работой может зависнуть. Возвращаем ошибку в main для выхода процесса.
		return errors.Join(shutdownErr, errors.New("shutdown incomplete: postgres pool left for process exit"))
	}

	poolClosed := make(chan struct{})
	go func() {
		// Здесь HTTP и синхронизация уже завершены, поэтому SQL больше не должен использовать пул.
		pool.Close()
		close(poolClosed)
	}()
	select {
	case <-poolClosed:
		return shutdownErr
	case <-shutdownCtx.Done():
		// pgxpool.Close не принимает context; не позволяем утечке соединения навечно удержать процесс.
		return errors.Join(shutdownErr, fmt.Errorf("close postgres pool: %w", shutdownCtx.Err()))
	}
}
