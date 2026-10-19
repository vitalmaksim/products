package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vitalmaksim/products/model"
)

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) GetProducts(ctx context.Context, limit, offset int) ([]model.Product, error) {
	const query = "SELECT * FROM products ORDER BY id LIMIT $1 OFFSET $2"

	var products []model.Product
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	products, err = pgx.CollectRows(rows, pgx.RowToStructByName[model.Product])
	if err != nil {
		return nil, err
	}

	return products, nil
}
