package model

import "time"

type Purchase struct {
	ID int64
	User
	Product
	CreatedAt time.Time
}
