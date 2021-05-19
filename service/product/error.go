package product

import "errors"

var (
	// ErrInsuffientInventory is insufficient inventory error
	ErrInsuffientInventory = errors.New("insufficient inventory")
	// ErrInvalidIdempotency is invalid idempotency error
	ErrInvalidIdempotency = errors.New("invalid idempotency")
)
