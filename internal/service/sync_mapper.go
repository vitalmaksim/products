package service

import (
	"github.com/vitalmaksim/products/internal/supplier"
	"github.com/vitalmaksim/products/model"
)

func mapProducts(items []supplier.ProductDTO) []model.Product {
	res := make([]model.Product, 0, len(items))

	for _, i := range items {
		model := model.Product{
			ExternalID: i.ID,
			Name:       i.Name,
			Price:      i.Price,
		}

		res = append(res, model)
	}

	return res
}
