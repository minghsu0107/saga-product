package order

import (
	"errors"
)

var (
	// ErrUnauthorized is unauthorized error
	ErrUnauthorized = errors.New("unauthorized")
	// ErrOrderNotFound is order not found error
	ErrOrderNotFound = errors.New("order not found")
)
