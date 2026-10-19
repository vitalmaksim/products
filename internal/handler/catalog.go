package handler

import (
	"log"
	"net/http"

	"github.com/vitalmaksim/products/internal/httpx"
	"github.com/vitalmaksim/products/internal/service"
)

type CatalogHandler struct {
	catalogService *service.CatalogService
}

func NewCatalogHandler(catalogService *service.CatalogService) *CatalogHandler {
	return &CatalogHandler{catalogService: catalogService}
}

func (h *CatalogHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	page, err := httpx.GetIntQueryParam(r, "page", 1, 1)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	limit, err := httpx.GetIntQueryParam(r, "limit", 20, 1)
	if err != nil || limit > 100 {
		httpx.Error(w, http.StatusBadRequest, "invalid limit")
		return
	}

	rows, err := h.catalogService.GetProducts(r.Context(), page, limit)
	if err != nil {
		log.Printf("%v", err)

		httpx.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"items": rows,
		"page":  page,
		"limit": limit,
	})
}
