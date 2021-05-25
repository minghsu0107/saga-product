package infra

import (
	"context"

	infra_broker "github.com/minghsu0107/saga-product/infra/broker"
	infra_grpc "github.com/minghsu0107/saga-product/infra/grpc"
	infra_http "github.com/minghsu0107/saga-product/infra/http"
	infra_observe "github.com/minghsu0107/saga-product/infra/observe"
	log "github.com/sirupsen/logrus"
)

// Server wraps http and grpc server
type Server struct {
	HTTPServer  infra_http.Server
	GRPCServer  infra_grpc.Server
	EventRouter infra_broker.EventRouter
	ObsInjector *infra_observe.ObservibilityInjector
}

func NewServer(httpServer infra_http.Server, grpcServer infra_grpc.Server, eventRouter infra_broker.EventRouter, obsInjector *infra_observe.ObservibilityInjector) *Server {
	return &Server{
		HTTPServer:  httpServer,
		GRPCServer:  grpcServer,
		EventRouter: eventRouter,
		ObsInjector: obsInjector,
	}
}

// Run server
func (s *Server) Run() error {
	errs := make(chan error, 1)
	s.ObsInjector.Register(errs)
	go func() {
		errs <- s.HTTPServer.Run()
	}()
	go func() {
		errs <- s.GRPCServer.Run()
	}()
	go func() {
		errs <- s.EventRouter.Run()
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
		errs <- s.EventRouter.GracefulStop()
	}()
	err := <-errs
	if err != nil {
		log.Error(err)
	}
	log.Info("gracefully shutdowned")
	done <- true
}
