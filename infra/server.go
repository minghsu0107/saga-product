package infra

import (
	"context"

	infra_broker "github.com/minghsu0107/saga-product/infra/broker"
	infra_cache "github.com/minghsu0107/saga-product/infra/cache"
	infra_grpc "github.com/minghsu0107/saga-product/infra/grpc"
	grpc_auth "github.com/minghsu0107/saga-product/infra/grpc/auth"
	grpc_order "github.com/minghsu0107/saga-product/infra/grpc/order"
	infra_http "github.com/minghsu0107/saga-product/infra/http"
	infra_observe "github.com/minghsu0107/saga-product/infra/observe"
	log "github.com/sirupsen/logrus"
)

// ProductServer wrapper
type ProductServer struct {
	HTTPServer  infra_http.Server
	GRPCServer  infra_grpc.Server
	EventRouter infra_broker.EventRouter
	ObsInjector *infra_observe.ObservabilityInjector
}

// OrderServer wrapper
type OrderServer struct {
	HTTPServer  infra_http.Server
	EventRouter infra_broker.EventRouter
	ObsInjector *infra_observe.ObservabilityInjector
}

// PaymentServer wrapper
type PaymentServer struct {
	HTTPServer  infra_http.Server
	EventRouter infra_broker.EventRouter
	ObsInjector *infra_observe.ObservabilityInjector
}

// OrchestratorServer wrapper
type OrchestratorServer struct {
	EventRouter infra_broker.EventRouter
	ObsInjector *infra_observe.ObservabilityInjector
}

// NewProductServer factory
func NewProductServer(httpServer infra_http.Server, grpcServer infra_grpc.Server, eventRouter infra_broker.EventRouter, obsInjector *infra_observe.ObservabilityInjector) *ProductServer {
	return &ProductServer{
		HTTPServer:  httpServer,
		GRPCServer:  grpcServer,
		EventRouter: eventRouter,
		ObsInjector: obsInjector,
	}
}

// Run server
func (s *ProductServer) Run() error {
	if err := s.ObsInjector.Register(); err != nil {
		return err
	}
	go func() {
		err := s.HTTPServer.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		err := s.GRPCServer.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		err := s.EventRouter.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()
	return nil
}

// GracefulStop server
func (s *ProductServer) GracefulStop(ctx context.Context, done chan bool) {
	err := s.HTTPServer.GracefulStop(ctx)
	if err != nil {
		log.Error(err)
	}
	s.GRPCServer.GracefulStop()
	err = s.EventRouter.GracefulStop()
	if err != nil {
		log.Error(err)
	}

	if infra_observe.TracerProvider != nil {
		err = infra_observe.TracerProvider.Shutdown(ctx)
		if err != nil {
			log.Error(err)
		}
	}
	if err = infra_cache.RedisClient.Close(); err != nil {
		log.Error(err)
	}
	if err = infra_broker.TxPublisher.Close(); err != nil {
		log.Error(err)
	}
	if err = infra_broker.TxSubscriber.Close(); err != nil {
		log.Error(err)
	}

	log.Info("gracefully shutdowned")
	done <- true
}

// NewOrderServer factory
func NewOrderServer(httpServer infra_http.Server, eventRouter infra_broker.EventRouter, obsInjector *infra_observe.ObservabilityInjector) *OrderServer {
	return &OrderServer{
		HTTPServer:  httpServer,
		EventRouter: eventRouter,
		ObsInjector: obsInjector,
	}
}

// Run server
func (s *OrderServer) Run() error {
	if err := s.ObsInjector.Register(); err != nil {
		return err
	}
	go func() {
		err := s.HTTPServer.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		err := s.EventRouter.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()
	return nil
}

// GracefulStop server
func (s *OrderServer) GracefulStop(ctx context.Context, done chan bool) {
	err := s.HTTPServer.GracefulStop(ctx)
	if err != nil {
		log.Error(err)
	}
	err = s.EventRouter.GracefulStop()
	if err != nil {
		log.Error(err)
	}

	if infra_observe.TracerProvider != nil {
		err = infra_observe.TracerProvider.Shutdown(ctx)
		if err != nil {
			log.Error(err)
		}
	}
	if err = infra_cache.RedisClient.Close(); err != nil {
		log.Error(err)
	}
	if err = infra_broker.TxPublisher.Close(); err != nil {
		log.Error(err)
	}
	if err = infra_broker.TxSubscriber.Close(); err != nil {
		log.Error(err)
	}
	if err = grpc_auth.AuthClientConn.Conn.Close(); err != nil {
		log.Error(err)
	}
	if err = grpc_order.ProductClientConn.Conn.Close(); err != nil {
		log.Error(err)
	}

	log.Info("gracefully shutdowned")
	done <- true
}

// NewPaymentServer factory
func NewPaymentServer(httpServer infra_http.Server, eventRouter infra_broker.EventRouter, obsInjector *infra_observe.ObservabilityInjector) *PaymentServer {
	return &PaymentServer{
		HTTPServer:  httpServer,
		EventRouter: eventRouter,
		ObsInjector: obsInjector,
	}
}

// Run server
func (s *PaymentServer) Run() error {
	if err := s.ObsInjector.Register(); err != nil {
		return err
	}
	go func() {
		err := s.HTTPServer.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		err := s.EventRouter.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()
	return nil
}

// GracefulStop server
func (s *PaymentServer) GracefulStop(ctx context.Context, done chan bool) {
	err := s.HTTPServer.GracefulStop(ctx)
	if err != nil {
		log.Error(err)
	}
	err = s.EventRouter.GracefulStop()
	if err != nil {
		log.Error(err)
	}

	if infra_observe.TracerProvider != nil {
		err = infra_observe.TracerProvider.Shutdown(ctx)
		if err != nil {
			log.Error(err)
		}
	}
	if err = infra_cache.RedisClient.Close(); err != nil {
		log.Error(err)
	}
	if err = infra_broker.TxPublisher.Close(); err != nil {
		log.Error(err)
	}
	if err = infra_broker.TxSubscriber.Close(); err != nil {
		log.Error(err)
	}
	if err = grpc_auth.AuthClientConn.Conn.Close(); err != nil {
		log.Error(err)
	}

	log.Info("gracefully shutdowned")
	done <- true
}

// NewOrchestratorServer factory
func NewOrchestratorServer(eventRouter infra_broker.EventRouter, obsInjector *infra_observe.ObservabilityInjector) *OrchestratorServer {
	return &OrchestratorServer{
		EventRouter: eventRouter,
		ObsInjector: obsInjector,
	}
}

// Run server
func (s *OrchestratorServer) Run() error {
	if err := s.ObsInjector.Register(); err != nil {
		return err
	}
	go func() {
		err := s.EventRouter.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()
	return nil
}

// GracefulStop server
func (s *OrchestratorServer) GracefulStop(ctx context.Context, done chan bool) {
	err := s.EventRouter.GracefulStop()
	if err != nil {
		log.Error(err)
	}

	if infra_observe.TracerProvider != nil {
		err = infra_observe.TracerProvider.Shutdown(ctx)
		if err != nil {
			log.Error(err)
		}
	}
	if err = infra_broker.TxPublisher.Close(); err != nil {
		log.Error(err)
	}
	if err = infra_broker.ResultPublisher.Close(); err != nil {
		log.Error(err)
	}
	if err = infra_broker.TxSubscriber.Close(); err != nil {
		log.Error(err)
	}

	log.Info("gracefully shutdowned")
	done <- true
}
