package broker

// EventRouter interface
type EventRouter interface {
	RegisterHandlers()
	Run() error
	GracefulStop() error
}
