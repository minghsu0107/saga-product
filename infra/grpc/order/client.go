package order

import (
	conf "github.com/minghsu0107/saga-product/config"
	infra_grpc "github.com/minghsu0107/saga-product/infra/grpc"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	// ProductClientConn grpc connection
	ProductClientConn *ProductConn
)

// ProductConn is a wrapper for Product grpc connection
type ProductConn struct {
	Conn *grpc.ClientConn
}

// NewProductConn returns a grpc client connection for ProductRepository
func NewProductConn(config *conf.Config) (*ProductConn, error) {
	log.Info("connecting to grpc product service...")
	conn, err := infra_grpc.InitializeClient(config.RPCEndpoints.ProductSvcHost)
	if err != nil {
		return nil, err
	}
	ProductClientConn = &ProductConn{
		Conn: conn,
	}
	return ProductClientConn, nil
}
