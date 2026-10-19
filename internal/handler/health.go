package handler

import (
	"net/http"

	"github.com/vitalmaksim/products/internal/httpx"
)

func Health(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]string{
		"status": "health",
	})
}
