package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vitalmaksim/products/model"
)

type PurchaseRepository struct {
	pool *pgxpool.Pool
}

func NewPurchaseRepository(pool *pgxpool.Pool) *PurchaseRepository {
	return &PurchaseRepository{pool: pool}
}

func (r *PurchaseRepository) CreatePurchase(ctx context.Context, userID, productID int64) (model.Purchase, error) {
	// Списание и покупка должны стать видны вместе либо вместе откатиться.
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return model.Purchase{}, fmt.Errorf("begin purchase transaction: %w", err)
	}
	defer func() {
		// Отмена HTTP-запроса не должна мешать освободить транзакцию и ее блокировки.
		rollbackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		// После Commit Rollback безопасно вернет pgx.ErrTxClosed.
		_ = tx.Rollback(rollbackCtx)
	}()

	const productQuery = `SELECT price, active FROM products WHERE id = $1 FOR SHARE`
	var price int64
	var active bool
	// FOR SHARE удерживает цену и active неизменными до конца покупки, не мешая другим покупателям.
	err = tx.QueryRow(ctx, productQuery, productID).Scan(&price, &active)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Purchase{}, model.ErrProductNotFound
	}
	if err != nil {
		return model.Purchase{}, fmt.Errorf("get purchase product: %w", err)
	}
	if !active {
		return model.Purchase{}, model.ErrProductInactive
	}

	const userQuery = `SELECT id FROM users WHERE id = $1 FOR UPDATE`
	var lockedUserID int64
	// Фиксируем существование пользователя, чтобы отсутствие строки при списании означало нехватку денег.
	err = tx.QueryRow(ctx, userQuery, userID).Scan(&lockedUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Purchase{}, model.ErrUserNotFound
	}
	if err != nil {
		return model.Purchase{}, fmt.Errorf("lock purchase user: %w", err)
	}

	const debitQuery = `
		UPDATE users
		SET balance = balance - $2
		WHERE id = $1 AND balance >= $2
		RETURNING balance
	`
	var balance int64
	// Блокировка строки сериализует списания, а условие не позволяет уйти в минус.
	err = tx.QueryRow(ctx, debitQuery, userID, price).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Purchase{}, model.ErrInsufficientFunds
	}
	if err != nil {
		return model.Purchase{}, fmt.Errorf("debit purchase balance: %w", err)
	}

	const insertQuery = `
		INSERT INTO purchases (user_id, product_id, price)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, product_id, price, created_at
	`
	var purchase model.Purchase
	// Цена покупки сохраняется отдельно: будущая синхронизация каталога не изменит сумму списания.
	err = tx.QueryRow(ctx, insertQuery, userID, productID, price).Scan(
		&purchase.ID, &purchase.UserID, &purchase.ProductID, &purchase.Price, &purchase.CreatedAt,
	)
	if err != nil {
		return model.Purchase{}, fmt.Errorf("insert purchase: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Purchase{}, fmt.Errorf("commit purchase transaction: %w", err)
	}
	return purchase, nil
}
