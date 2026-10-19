package service

import (
	"context"
	"fmt"

	"github.com/vitalmaksim/products/model"
)

type ProductRepository interface {
	GetProducts(ctx context.Context, limit, offset int) ([]model.Product, error)
}

type CatalogService struct {
	productRepository ProductRepository
}

func NewCatalogService(productRepository ProductRepository) *CatalogService {
	return &CatalogService{productRepository: productRepository}
}

func (s *CatalogService) GetProducts(ctx context.Context, page, limit int) ([]model.Product, error) {
	offset := limit * (page - 1)

	products, err := s.productRepository.GetProducts(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("unable to get products %w", err)
	}

	return products, nil
}
