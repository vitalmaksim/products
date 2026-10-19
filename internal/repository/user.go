package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vitalmaksim/products/model"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetBalance(ctx context.Context, userID int64) (int64, error) {
	const query = `SELECT balance FROM users WHERE id = $1`

	var balance int64
	err := r.pool.QueryRow(ctx, query, userID).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("user %d, %w", userID, model.ErrUserNotFound)
	}
	if err != nil {
		return 0, fmt.Errorf("user balance not found %w", err)
	}

	return balance, nil
}
