package repo

import (
	"context"
	"time"

	pb "github.com/minghsu0107/saga-pb"
	conf "github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/domain/model"
	"github.com/minghsu0107/saga-product/infra/grpc/auth"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/ratelimit"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

// AuthRepository is the auth repository interface
type AuthRepository interface {
	Auth(ctx context.Context, accessToken string) (*model.AuthResult, error)
}

// AuthRepositoryImpl is the implementation of AuthRepository
type AuthRepositoryImpl struct {
	auth endpoint.Endpoint
}

// NewAuthRepository is the factory of AuthRepository
func NewAuthRepository(conn *auth.AuthConn, config *conf.Config) AuthRepository {
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), config.ServiceOptions.Rps))

	var options []grpctransport.ClientOption

	var auth endpoint.Endpoint
	{
		svcName := "auth.AuthService"

		auth = grpctransport.NewClient(
			conn.Conn,
			svcName,
			"Auth",
			encodeGRPCRequest,
			decodeGRPCResponse,
			&pb.AuthResponse{},
			append(options, grpctransport.ClientBefore(grpctransport.SetRequestHeader(ServiceNameHeader, svcName)))...,
		).Endpoint()
		auth = limiter(auth)
		auth = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "auth",
			Timeout: config.ServiceOptions.Timeout,
		}))(auth)
	}

	return &AuthRepositoryImpl{
		auth: auth,
	}
}

// Auth method implements AuthRepository interface
func (repo *AuthRepositoryImpl) Auth(ctx context.Context, accessToken string) (*model.AuthResult, error) {
	res, err := repo.auth(ctx, &pb.AuthPayload{
		AccessToken: accessToken,
	})
	if err != nil {
		return nil, err
	}
	response := res.(*pb.AuthResponse)
	return &model.AuthResult{
		CustomerID: response.CustomerId,
		Expired:    response.Expired,
	}, nil
}
