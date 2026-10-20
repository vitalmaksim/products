package model

import "errors"

var ErrUserNotFound = errors.New("user not found")

var (
	ErrProductNotFound       = errors.New("product not found")
	ErrProductInactive       = errors.New("product inactive")
	ErrInsufficientFunds     = errors.New("insufficient funds")
	ErrInvalidUserID         = errors.New("invalid user id")
	ErrInvalidProductID      = errors.New("invalid product id")
	ErrInvalidIdempotencyKey = errors.New("invalid idempotency key")
	ErrIdempotencyConflict   = errors.New("idempotency key used for another product")
)
