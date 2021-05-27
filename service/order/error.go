package order

import (
	"errors"
)

var (
	// ErrUnautorized is unauthorized error
	ErrUnautorized = errors.New("unauthorized")
)
