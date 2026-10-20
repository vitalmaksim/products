package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vitalmaksim/products/model"
)

type PurchaseRepository struct {
	pool *pgxpool.Pool
}

func NewPurchaseRepository(pool *pgxpool.Pool) *PurchaseRepository {
	return &PurchaseRepository{pool: pool}
}

func (r *PurchaseRepository) CreatePurchase(ctx context.Context, userID, productID int64, idempotencyKey string) (model.Purchase, bool, error) {
	// Списание и покупка должны стать видны вместе либо вместе откатиться.
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return model.Purchase{}, false, fmt.Errorf("begin purchase transaction: %w", err)
	}
	defer func() {
		// Отмена HTTP-запроса не должна мешать освободить транзакцию и ее блокировки.
		rollbackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		// После Commit Rollback безопасно вернет pgx.ErrTxClosed.
		_ = tx.Rollback(rollbackCtx)
	}()

	// SELECT перед INSERT не защитил бы от гонки: ключ резервирует уникальный индекс.
	// Резервирование в этой же транзакции откатывается вместе со списанием при любой ошибке.
	const reserveQuery = `
		INSERT INTO purchase_requests (user_id, idempotency_key, product_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, idempotency_key) DO NOTHING
		RETURNING user_id
	`
	var reservedUserID int64
	err = tx.QueryRow(ctx, reserveQuery, userID, idempotencyKey, productID).Scan(&reservedUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		// INSERT мог ждать чужой COMMIT. Отдельный SELECT в READ COMMITTED уже видит его результат.
		const requestQuery = `
			SELECT product_id, purchase_id FROM purchase_requests
			WHERE user_id = $1 AND idempotency_key = $2
		`
		var savedProductID int64
		var purchaseID *int64
		if err := tx.QueryRow(ctx, requestQuery, userID, idempotencyKey).Scan(&savedProductID, &purchaseID); err != nil {
			return model.Purchase{}, false, fmt.Errorf("get existing purchase request: %w", err)
		}
		if savedProductID != productID {
			return model.Purchase{}, false, model.ErrIdempotencyConflict
		}
		if purchaseID == nil {
			return model.Purchase{}, false, fmt.Errorf("existing purchase request has no purchase")
		}

		// Повтор возвращает сохраненную покупку: текущая цена, active и баланс больше не проверяются.
		const purchaseQuery = `
			SELECT id, user_id, product_id, price, created_at FROM purchases
			WHERE id = $1 AND user_id = $2 AND product_id = $3
		`
		var purchase model.Purchase
		if err := tx.QueryRow(ctx, purchaseQuery, *purchaseID, userID, productID).Scan(
			&purchase.ID, &purchase.UserID, &purchase.ProductID, &purchase.Price, &purchase.CreatedAt,
		); err != nil {
			return model.Purchase{}, false, fmt.Errorf("get existing purchase: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return model.Purchase{}, false, fmt.Errorf("commit purchase replay: %w", err)
		}
		return purchase, false, nil
	}
	if err != nil {
		// Внешние ключи проверяются при резервировании, до обычной проверки пользователя и товара.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			switch pgErr.ConstraintName {
			case "purchase_requests_user_id_fkey":
				return model.Purchase{}, false, model.ErrUserNotFound
			case "purchase_requests_product_id_fkey":
				return model.Purchase{}, false, model.ErrProductNotFound
			}
		}
		return model.Purchase{}, false, fmt.Errorf("reserve purchase request: %w", err)
	}

	const productQuery = `SELECT price, active FROM products WHERE id = $1 FOR SHARE`
	var price int64
	var active bool
	// FOR SHARE удерживает цену и active неизменными до конца покупки, не мешая другим покупателям.
	err = tx.QueryRow(ctx, productQuery, productID).Scan(&price, &active)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Purchase{}, false, model.ErrProductNotFound
	}
	if err != nil {
		return model.Purchase{}, false, fmt.Errorf("get purchase product: %w", err)
	}
	if !active {
		return model.Purchase{}, false, model.ErrProductInactive
	}

	const userQuery = `SELECT id FROM users WHERE id = $1 FOR NO KEY UPDATE`
	var lockedUserID int64
	// NO KEY UPDATE сериализует списания, но совместим с KEY SHARE от внешнего ключа резервации.
	// FOR UPDATE здесь мог бы создать deadlock при двух разных ключах одного пользователя.
	err = tx.QueryRow(ctx, userQuery, userID).Scan(&lockedUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Purchase{}, false, model.ErrUserNotFound
	}
	if err != nil {
		return model.Purchase{}, false, fmt.Errorf("lock purchase user: %w", err)
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
		return model.Purchase{}, false, model.ErrInsufficientFunds
	}
	if err != nil {
		return model.Purchase{}, false, fmt.Errorf("debit purchase balance: %w", err)
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
		return model.Purchase{}, false, fmt.Errorf("insert purchase: %w", err)
	}
	const linkQuery = `
		UPDATE purchase_requests SET purchase_id = $3
		WHERE user_id = $1 AND idempotency_key = $2 AND purchase_id IS NULL
		RETURNING purchase_id
	`
	// Сначала связываем резервирование с покупкой и только затем фиксируем все изменения.
	var linkedPurchaseID int64
	if err := tx.QueryRow(ctx, linkQuery, userID, idempotencyKey, purchase.ID).Scan(&linkedPurchaseID); err != nil {
		return model.Purchase{}, false, fmt.Errorf("link purchase request: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Purchase{}, false, fmt.Errorf("commit purchase transaction: %w", err)
	}
	return purchase, true, nil
}
