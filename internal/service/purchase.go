package service

import (
	"context"
	"fmt"

	"github.com/vitalmaksim/products/model"
)

type PurchaseRepository interface {
	CreatePurchase(ctx context.Context, userID, productID int64) (model.Purchase, error)
}

type PurchaseService struct {
	purchaseRepository PurchaseRepository
}

func NewPurchaseService(purchaseRepository PurchaseRepository) *PurchaseService {
	return &PurchaseService{purchaseRepository: purchaseRepository}
}

func (s *PurchaseService) CreatePurchase(ctx context.Context, userID, productID int64) (model.Purchase, error) {
	if userID <= 0 {
		return model.Purchase{}, model.ErrInvalidUserID
	}
	if productID <= 0 {
		return model.Purchase{}, model.ErrInvalidProductID
	}
	purchase, err := s.purchaseRepository.CreatePurchase(ctx, userID, productID)
	if err != nil {
		return model.Purchase{}, fmt.Errorf("create purchase: %w", err)
	}
	return purchase, nil
}
