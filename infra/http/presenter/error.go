package presenter

import (
	"errors"
)

var (
	// ErrInvalidParam is invalid parameter error
	ErrInvalidParam = errors.New("invalid parameter")
	// ErrUnautorized is unauthorized error
	ErrUnautorized = errors.New("unauthorized")
	// ErrServer is server error
	ErrServer = errors.New("server error")
)

// ErrResponse is the error response type
type ErrResponse struct {
	Message string `json:"msg"`
}
