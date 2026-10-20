package handler

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/vitalmaksim/products/internal/httpx"
	"github.com/vitalmaksim/products/internal/service"
)

type SyncHandler struct {
	syncManager *service.SyncManager
	appContext  context.Context
}

func NewSyncHandler(syncManager *service.SyncManager, appContext context.Context) *SyncHandler {
	return &SyncHandler{syncManager: syncManager, appContext: appContext}
}

func (h *SyncHandler) StartSync(w http.ResponseWriter, r *http.Request) {
	// Контекст запроса отменяется после ответа, а принятая фоновая работа должна продолжаться.
	err := h.syncManager.Start(h.appContext)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSyncAlreadyRunning):
			httpx.Error(w, http.StatusConflict, "sync already running")
		case errors.Is(err, service.ErrSyncShuttingDown), errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			httpx.Error(w, http.StatusServiceUnavailable, "sync unavailable")
		default:
			log.Printf("start sync: %v", err)
			httpx.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// Быстрая синхронизация может уже закончиться, поэтому здесь подтверждаем только прием задачи.
	httpx.JSON(w, http.StatusAccepted, map[string]string{"message": "sync accepted"})
}

func (h *SyncHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, h.syncManager.GetStatus())
}
