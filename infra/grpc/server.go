package grpc

// Server interface
type Server interface {
	Run() error
	GracefulStop()
}
