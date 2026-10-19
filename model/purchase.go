package model

import "time"

type Purchase struct {
	ID        int64
	UserID    int64
	ProductID int64
	Price     int64
	CreatedAt time.Time
}
