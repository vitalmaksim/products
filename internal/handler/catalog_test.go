package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vitalmaksim/products/internal/service"
	"github.com/vitalmaksim/products/model"
)

type fakeCatalogRepository struct {
	products []model.Product
	limit    int
	offset   int
}

func (f *fakeCatalogRepository) GetProducts(ctx context.Context, limit, offset int) ([]model.Product, error) {
	f.limit, f.offset = limit, offset
	return f.products, nil
}

func TestGetProductsAPI(t *testing.T) {
	repository := &fakeCatalogRepository{products: []model.Product{
		{ID: 1, Name: "Book", Price: 300},
		{ID: 2, Name: "Game", Price: 500},
	}}
	// Handler и Service настоящие; подменена только граница с PostgreSQL.
	handler := NewCatalogHandler(service.NewCatalogService(repository))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /products", handler.GetProducts)
	request, err := http.NewRequest(http.MethodGet, "/products?page=1&limit=2", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	response := recorder.Result()
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want 200; body: %s", response.StatusCode, recorder.Body.String())
	}

	var body struct {
		Items []model.Product `json:"items"`
		Page  int             `json:"page"`
		Limit int             `json:"limit"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response JSON: %v", err)
	}
	if len(body.Items) != 2 {
		t.Fatalf("items count: got %d, want 2", len(body.Items))
	}
	if body.Page != 1 || body.Limit != 2 {
		t.Errorf("pagination: got page=%d limit=%d, want page=1 limit=2", body.Page, body.Limit)
	}
	if body.Items[0].Name != "Book" {
		t.Errorf("first product name: got %q, want Book", body.Items[0].Name)
	}
	if repository.limit != 2 || repository.offset != 0 {
		t.Errorf("repository arguments: got limit=%d offset=%d, want limit=2 offset=0", repository.limit, repository.offset)
	}
}
