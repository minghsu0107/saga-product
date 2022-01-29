package payment

import (
	"errors"
)

var (
	// ErrUnauthorized is unauthorized error
	ErrUnauthorized = errors.New("unauthorized")
	// ErrPaymentNotFound is payment not found error
	ErrPaymentNotFound = errors.New("payment not found")
)
