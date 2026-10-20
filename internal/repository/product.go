package repository

import (
	"context"
	"fmt"
	"strings"

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

func (r *ProductRepository) UpsertProducts(
	ctx context.Context,
	products []model.Product,
) error {
	if len(products) == 0 {
		return nil
	}

	const fieldsCount = 4

	placeholders := make([]string, 0, len(products))
	args := make([]any, 0, len(products)*fieldsCount)

	for i, product := range products {
		n := i * fieldsCount

		placeholders = append(
			placeholders,
			fmt.Sprintf(
				"($%d, $%d, $%d, $%d)",
				n+1, n+2, n+3, n+4,
			),
		)

		args = append(
			args,
			product.ExternalID,
			product.Name,
			product.Price,
			product.Description,
		)
	}

	query := fmt.Sprintf(`
        INSERT INTO products (external_id, name, price, description)
        VALUES %s
        ON CONFLICT (external_id)
        DO UPDATE SET
            name = EXCLUDED.name,
            price = EXCLUDED.price
    `, strings.Join(placeholders, ", "))

	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("upsert products: %w", err)
	}

	return nil
}
