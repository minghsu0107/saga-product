package product

import (
	"net"

	infra_grpc "github.com/minghsu0107/saga-product/infra/grpc"
	log "github.com/sirupsen/logrus"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	pb "github.com/minghsu0107/saga-pb"
	"github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/service/product"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// ProductServer implementation
type ProductServer struct {
	Port           string
	s              *grpc.Server
	productSvc     product.ProductService
	sagaProductSvc product.SagaProductService
}

// NewProductServer is the factory of product server
func NewProductServer(config *config.Config, productSvc product.ProductService, sagaProductSvc product.SagaProductService) infra_grpc.Server {
	srv := &ProductServer{
		Port:           config.GRPCPort,
		productSvc:     productSvc,
		sagaProductSvc: sagaProductSvc,
	}

	srv.s = infra_grpc.Initialize(config.OcAgentHost, config.Logger.ContextLogger)
	pb.RegisterProductServiceServer(srv.s, srv)

	grpc_prometheus.Register(srv.s)
	reflection.Register(srv.s)
	return srv
}

// Run method starts the grpc server
func (srv *ProductServer) Run() error {
	addr := "0.0.0.0:" + srv.Port
	log.Infoln("grpc server listening on ", addr)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	if err := srv.s.Serve(lis); err != nil {
		return err
	}
	return nil
}

// GracefulStop stops grpc server gracefully
func (srv *ProductServer) GracefulStop() {
	srv.s.GracefulStop()
}
