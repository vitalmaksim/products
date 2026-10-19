package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetBalance(ctx context.Context, user_id int64) (int64, error) {
	const query = `SELECT balance FROM users WHERE id = $1`

	var balance int64
	err := r.pool.QueryRow(ctx, query, user_id).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("user not found %v", err)
	}
	if err != nil {
		return 0, fmt.Errorf("user balance not found %v", err)
	}

	return balance, nil
}
