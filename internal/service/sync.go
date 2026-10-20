package service

import (
	"context"
	"fmt"

	"github.com/vitalmaksim/products/internal/supplier"
	"github.com/vitalmaksim/products/model"
)

type ProductSupplier interface {
	GetProducts(ctx context.Context, page int) (supplier.ProductsResponse, error)
}

type ProductWriter interface {
	UpsertProducts(ctx context.Context, products []model.Product) error
}

type SyncService struct {
	productSupplier ProductSupplier
	productWriter   ProductWriter
}

func NewSyncService(
	productSupplier ProductSupplier,
	productWriter ProductWriter,
) *SyncService {
	return &SyncService{
		productSupplier: productSupplier,
		productWriter:   productWriter,
	}
}

func (s *SyncService) Sync(ctx context.Context) error {
	// 1. Получить первую страницу поставщика
	response, err := s.productSupplier.GetProducts(ctx, 1)
	if err != nil {
		return fmt.Errorf("unable to get first product pages %w", err)
	}

	batch := make([]model.Product, 0, 100)
	batch = append(batch, mapProducts(response.Items)...)
	totalPages := response.TotalPages

	for len(batch) >= 100 {
		err = s.productWriter.UpsertProducts(ctx, batch[:100])
		if err != nil {
			return fmt.Errorf("unable to upsert products %w", err)
		}
		batch = batch[100:]
	}

	// 2. Узнать TotalPages
	for page := 2; page <= totalPages; page++ {
		// 3. Последовательно загрузить остальные страницы
		response, err = s.productSupplier.GetProducts(ctx, page)
		if err != nil {
			return fmt.Errorf("unable to get product page %w %d", err, page)
		}

		// 4. Преобразовать DTO в модели через mapProducts()
		batch = append(batch, mapProducts(response.Items)...)

		// 5. Сохранить товары пачками по 100 через UpsertProducts()
		for len(batch) >= 100 {
			err = s.productWriter.UpsertProducts(ctx, batch[:100])
			if err != nil {
				return fmt.Errorf("unable to upsert products %w", err)
			}
			batch = batch[100:]
		}
	}

	err = s.productWriter.UpsertProducts(ctx, batch)
	if err != nil {
		return fmt.Errorf("unable to upsert products %w", err)
	}

	return nil
}
