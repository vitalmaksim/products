package model

import "time"

type Product struct {
	ID          int64     `db:"id" json:"id"`
	ExternalID  int64     `db:"external_id" json:"external_id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Price       int64     `db:"price" json:"price"`
	Active      bool      `db:"active" json:"active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
