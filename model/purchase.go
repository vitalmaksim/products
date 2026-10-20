package model

import "time"

type Purchase struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	ProductID int64     `db:"product_id" json:"product_id"`
	Price     int64     `db:"price" json:"price"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
