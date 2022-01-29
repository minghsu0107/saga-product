package repo

import "errors"

var (
	// ErrInsuffientInventory is insufficient inventory error
	ErrInsuffientInventory = errors.New("insufficient inventory")
	// ErrInvalidIdempotency is invalid idempotency error
	ErrInvalidIdempotency = errors.New("invalid idempotency")
	// ErrProductNotFound is product not found error
	ErrProductNotFound = errors.New("product not found")
	// ErrOrderNotFound is order not found error
	ErrOrderNotFound = errors.New("order not found")
	// ErrPaymentNotFound is payment not found error
	ErrPaymentNotFound = errors.New("payment not found")
)
