package http

import "context"

// Server interface
type Server interface {
	RegisterRoutes()
	Run() error
	GracefulStop(ctx context.Context) error
}
