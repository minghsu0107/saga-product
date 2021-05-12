package infra

import (
	"context"

	infra_cache "github.com/minghsu0107/saga-product/infra/cache"
	infra_grpc "github.com/minghsu0107/saga-product/infra/grpc"
	infra_http "github.com/minghsu0107/saga-product/infra/http"
	log "github.com/sirupsen/logrus"
)

// Server wraps http and grpc server
type Server struct {
	HTTPServer   *infra_http.Server
	GRPCServer   *infra_grpc.Server
	CacheCleaner infra_cache.LocalCacheCleaner
}

func NewServer(httpServer *infra_http.Server, grpcServer *infra_grpc.Server, cacheCleaner infra_cache.LocalCacheCleaner) *Server {
	return &Server{
		HTTPServer:   httpServer,
		GRPCServer:   grpcServer,
		CacheCleaner: cacheCleaner,
	}
}

// Run server
func (s *Server) Run() error {
	errs := make(chan error, 1)
	go func() {
		errs <- s.HTTPServer.Run()
	}()
	go func() {
		errs <- s.GRPCServer.Run()
	}()
	go func() {
		errs <- s.CacheCleaner.SubscribeInvalidationEvent()
	}()
	err := <-errs
	if err != nil {
		return err
	}
	return nil
}

// GracefulStop server
func (s *Server) GracefulStop(ctx context.Context, done chan bool) {
	errs := make(chan error, 1)
	go func() {
		errs <- s.HTTPServer.GracefulStop(ctx)
	}()
	go func() {
		s.GRPCServer.GracefulStop()
	}()
	go func() {
		s.CacheCleaner.Close()
	}()
	err := <-errs
	if err != nil {
		log.Error(err)
	}
	log.Info("gracefully shutdowned")
	done <- true
}
