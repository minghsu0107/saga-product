package http

// Router wraps http handlers
type Router struct {
}

// NewRouter is a factory for router instance
func NewRouter() *Router {
	return &Router{}
}
