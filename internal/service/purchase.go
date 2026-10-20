package service

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/vitalmaksim/products/model"
)

type PurchaseRepository interface {
	CreatePurchase(ctx context.Context, userID, productID int64, idempotencyKey string) (model.Purchase, bool, error)
}

type PurchaseService struct {
	purchaseRepository PurchaseRepository
}

func NewPurchaseService(purchaseRepository PurchaseRepository) *PurchaseService {
	return &PurchaseService{purchaseRepository: purchaseRepository}
}

// Второй результат равен true только при создании новой покупки, а не при повторе запроса.
func (s *PurchaseService) CreatePurchase(ctx context.Context, userID, productID int64, idempotencyKey string) (model.Purchase, bool, error) {
	if userID <= 0 {
		return model.Purchase{}, false, model.ErrInvalidUserID
	}
	if productID <= 0 {
		return model.Purchase{}, false, model.ErrInvalidProductID
	}
	// VARCHAR ограничивает символы, а не байты; сам ключ не нормализуем, чтобы не смешивать разные ключи.
	if strings.TrimSpace(idempotencyKey) == "" || !utf8.ValidString(idempotencyKey) ||
		utf8.RuneCountInString(idempotencyKey) > 128 || strings.ContainsRune(idempotencyKey, '\x00') {
		return model.Purchase{}, false, model.ErrInvalidIdempotencyKey
	}
	purchase, created, err := s.purchaseRepository.CreatePurchase(ctx, userID, productID, idempotencyKey)
	if err != nil {
		return model.Purchase{}, false, fmt.Errorf("create purchase: %w", err)
	}
	return purchase, created, nil
}
