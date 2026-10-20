package repository

import (
	"context"
	"math/rand/v2"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vitalmaksim/products/model"
)

func TestProductRepository_UpsertProducts(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create test database pool: %v", err)
	}
	t.Cleanup(pool.Close)

	var databaseName string
	if err := pool.QueryRow(ctx, "SELECT current_database()").Scan(&databaseName); err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	// Проверяем реальную БД на сервере, а не просто наличие слова test в URL.
	if !strings.HasSuffix(databaseName, "_test") {
		t.Fatalf("refusing to write to database %q: name must end with _test", databaseName)
	}
	if _, err := pool.Exec(ctx, "SELECT id, external_id, name, description, price, active, created_at FROM products LIMIT 0"); err != nil {
		t.Fatalf("test database must have applied migrations: %v", err)
	}

	externalID := rand.Int64N(1<<62) + 1
	var exists bool
	if err := pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM products WHERE external_id = $1)", externalID).Scan(&exists); err != nil {
		t.Fatalf("check test external id: %v", err)
	}
	if exists {
		t.Fatal("test external id already exists; refusing to overwrite another product")
	}
	product := model.Product{ExternalID: externalID, Name: "upsert test product", Price: 300}
	t.Cleanup(func() {
		// Cleanup выполняется перед pool.Close (порядок LIFO), даже если проверка завершится через Fatal.
		cleanupCtx, cancelCleanup := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelCleanup()
		if _, err := pool.Exec(cleanupCtx, "DELETE FROM products WHERE external_id = $1 AND name = $2", externalID, product.Name); err != nil {
			t.Errorf("delete test product: %v", err)
		}
	})

	repository := NewProductRepository(pool)
	if err := repository.UpsertProducts(ctx, []model.Product{product}); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	var price int64
	if err := pool.QueryRow(ctx, "SELECT price FROM products WHERE external_id = $1", externalID).Scan(&price); err != nil {
		t.Fatalf("read inserted product: %v", err)
	}
	if price != 300 {
		t.Fatalf("inserted price: got %d, want 300", price)
	}
	if _, err := pool.Exec(ctx, "UPDATE products SET active = true WHERE external_id = $1", externalID); err != nil {
		t.Fatalf("activate test product: %v", err)
	}

	product.Price = 500
	if err := repository.UpsertProducts(ctx, []model.Product{product}); err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM products WHERE external_id = $1", externalID).Scan(&count); err != nil {
		t.Fatalf("count products: %v", err)
	}
	if count != 1 {
		t.Fatalf("product count: got %d, want 1", count)
	}
	var active bool
	if err := pool.QueryRow(ctx, "SELECT price, active FROM products WHERE external_id = $1", externalID).Scan(&price, &active); err != nil {
		t.Fatalf("read updated product: %v", err)
	}
	if price != 500 {
		t.Errorf("updated price: got %d, want 500", price)
	}
	if !active {
		t.Error("second upsert reset active: got false, want true")
	}
}
