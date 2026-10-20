package service

import (
	"context"
	"fmt"

	"github.com/vitalmaksim/products/internal/supplier"
	"github.com/vitalmaksim/products/model"
)

const syncBatchSize = 100

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
	return s.SyncConcurrent(ctx)
}

func (s *SyncService) SyncSequential(ctx context.Context) error {
	response, err := s.firstPage(ctx)
	if err != nil {
		return err
	}
	totalPages := response.TotalPages
	batch := make([]model.Product, 0, syncBatchSize)

	for page := 1; page <= totalPages; page++ {
		if page > 1 {
			response, err = s.productSupplier.GetProducts(ctx, page)
			if err != nil {
				return fmt.Errorf("get product page %d: %w", page, err)
			}
		}
		batch, err = s.appendBatch(ctx, batch, response.Items)
		if err != nil {
			return err
		}
	}
	return s.saveBatch(ctx, batch)
}

func (s *SyncService) firstPage(ctx context.Context) (supplier.ProductsResponse, error) {
	if err := ctx.Err(); err != nil {
		return supplier.ProductsResponse{}, fmt.Errorf("get product page 1: %w", err)
	}
	response, err := s.productSupplier.GetProducts(ctx, 1)
	if err != nil {
		return supplier.ProductsResponse{}, fmt.Errorf("get product page 1: %w", err)
	}
	if response.TotalPages < 1 {
		return supplier.ProductsResponse{}, fmt.Errorf("invalid total_pages on product page 1: %d", response.TotalPages)
	}
	return response, nil
}

func (s *SyncService) appendBatch(ctx context.Context, batch []model.Product, items []supplier.ProductDTO) ([]model.Product, error) {
	for len(items) > 0 {
		// Преобразуем только свободную часть пачки, не копируя целиком большую страницу.
		n := min(syncBatchSize-len(batch), len(items))
		batch = append(batch, mapProducts(items[:n])...)
		items = items[n:]
		if len(batch) == syncBatchSize {
			if err := s.saveBatch(ctx, batch); err != nil {
				return nil, err
			}
			batch = batch[:0]
		}
	}
	return batch, nil
}

func (s *SyncService) saveBatch(ctx context.Context, batch []model.Product) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("upsert products: %w", err)
	}
	if len(batch) == 0 {
		return nil
	}
	if err := s.productWriter.UpsertProducts(ctx, batch); err != nil {
		return fmt.Errorf("upsert products: %w", err)
	}
	return nil
}
