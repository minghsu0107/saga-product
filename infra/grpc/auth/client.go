package auth

import (
	conf "github.com/minghsu0107/saga-product/config"
	infra_grpc "github.com/minghsu0107/saga-product/infra/grpc"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	// AuthClientConn grpc connection
	AuthClientConn *AuthConn
)

// AuthConn is a wrapper for Auth grpc connection
type AuthConn struct {
	Conn *grpc.ClientConn
}

// NewAuthConn returns a grpc client connection for AuthRepository
func NewAuthConn(config *conf.Config) (*AuthConn, error) {
	log.Info("connecting to grpc auth service...")
	conn, err := infra_grpc.InitializeClient(config.RPCEndpoints.AuthSvcHost, config.OcAgentHost)
	if err != nil {
		return nil, err
	}
	AuthClientConn = &AuthConn{
		Conn: conn,
	}
	return AuthClientConn, nil
}
