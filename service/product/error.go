package product

import "errors"

var (
	// ErrInsuffientInventory is insufficient inventory error
	ErrInsuffientInventory = errors.New("insufficient inventory")
	// ErrInvalidIdempotency is invalid idempotency error
	ErrInvalidIdempotency = errors.New("invalid idempotency")
	// ErrProductNotFound is product not found error
	ErrProductNotFound = errors.New("product not found")
)
