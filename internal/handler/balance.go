package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/vitalmaksim/products/internal/httpx"
	"github.com/vitalmaksim/products/model"
)

type BalanceService interface {
	GetBalance(ctx context.Context, userId int64) (int64, error)
}

type BalanceHandler struct {
	balance BalanceService
}

func NewBalanceHandler(balance BalanceService) *BalanceHandler {
	return &BalanceHandler{balance: balance}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.Header.Get("X-User-Id"), 10, 64)
	if err != nil || userID <= 0 {
		httpx.Error(w, http.StatusBadRequest, "invalid user given")
		return
	}

	balance, err := h.balance.GetBalance(r.Context(), userID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			httpx.Error(w, http.StatusNotFound, "user not found")
			return
		}

		httpx.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]int64{
		"balance": balance,
	})
}
