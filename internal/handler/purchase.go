package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/vitalmaksim/products/internal/httpx"
	"github.com/vitalmaksim/products/internal/service"
	"github.com/vitalmaksim/products/model"
)

type PurchaseHandler struct {
	purchaseService *service.PurchaseService
}

func NewPurchaseHandler(purchaseService *service.PurchaseService) *PurchaseHandler {
	return &PurchaseHandler{purchaseService: purchaseService}
}

func (h *PurchaseHandler) CreatePurchase(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.Header.Get("X-User-ID"), 10, 64)
	if err != nil || userID <= 0 {
		httpx.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var request struct {
		ProductID int64 `json:"product_id"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	// Decode читает только одно значение: запрещаем второй JSON и мусор после тела.
	if err := decoder.Decode(new(any)); err != io.EOF {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	purchase, created, err := h.purchaseService.CreatePurchase(r.Context(), userID, request.ProductID, r.Header.Get("Idempotency-Key"))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidIdempotencyKey):
			httpx.Error(w, http.StatusBadRequest, "idempotency key must contain 1 to 128 characters")
		case errors.Is(err, model.ErrIdempotencyConflict):
			httpx.Error(w, http.StatusConflict, "idempotency key used for another product")
		case errors.Is(err, model.ErrInvalidUserID), errors.Is(err, model.ErrInvalidProductID):
			httpx.Error(w, http.StatusBadRequest, "invalid user id or product id")
		case errors.Is(err, model.ErrUserNotFound):
			httpx.Error(w, http.StatusNotFound, "user not found")
		case errors.Is(err, model.ErrProductNotFound):
			httpx.Error(w, http.StatusNotFound, "product not found")
		case errors.Is(err, model.ErrProductInactive):
			httpx.Error(w, http.StatusConflict, "product inactive")
		case errors.Is(err, model.ErrInsufficientFunds):
			httpx.Error(w, http.StatusConflict, "insufficient funds")
		default:
			log.Printf("create purchase: %v", err)
			httpx.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	httpx.JSON(w, status, purchase)
}
