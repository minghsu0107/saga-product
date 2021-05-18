package repo

import "errors"

var (
	// ErrInsuffientInventory is insufficient inventory error
	ErrInsuffientInventory = errors.New("insufficient inventory")
)
