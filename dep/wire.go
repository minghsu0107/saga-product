//+build wireinject

// The build tag makes sure the stub is not built in the final build.
package dep

import (
	"github.com/google/wire"
	conf "github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/repo"
	"github.com/minghsu0107/saga-product/repo/proxy"
	"github.com/minghsu0107/saga-product/service/orchestrator"
	"github.com/minghsu0107/saga-product/service/order"
	"github.com/minghsu0107/saga-product/service/payment"
	"github.com/minghsu0107/saga-product/service/product"

	"github.com/minghsu0107/saga-product/infra"
	"github.com/minghsu0107/saga-product/infra/broker"
	infra_broker_orchestrator "github.com/minghsu0107/saga-product/infra/broker/orchestrator"
	infra_broker_order "github.com/minghsu0107/saga-product/infra/broker/order"
	infra_broker_payment "github.com/minghsu0107/saga-product/infra/broker/payment"
	infra_broker_product "github.com/minghsu0107/saga-product/infra/broker/product"
	"github.com/minghsu0107/saga-product/infra/cache"
	"github.com/minghsu0107/saga-product/infra/db"
	infra_grpc_auth "github.com/minghsu0107/saga-product/infra/grpc/auth"
	infra_grpc_order "github.com/minghsu0107/saga-product/infra/grpc/order"
	infra_grpc_product "github.com/minghsu0107/saga-product/infra/grpc/product"
	"github.com/minghsu0107/saga-product/infra/http/middleware"
	infra_http_order "github.com/minghsu0107/saga-product/infra/http/order"
	infra_http_payment "github.com/minghsu0107/saga-product/infra/http/payment"
	infra_http_product "github.com/minghsu0107/saga-product/infra/http/product"
	infra_observe "github.com/minghsu0107/saga-product/infra/observe"
	"github.com/minghsu0107/saga-product/pkg"
)

func InitializeProductServer() (*infra.ProductServer, error) {
	wire.Build(
		conf.NewConfig,

		infra.NewProductServer,
		infra_http_product.NewProductServer,
		infra_http_product.NewEngine,
		infra_http_product.NewRouter,

		infra_grpc_product.NewProductServer,

		infra_broker_product.NewProductEventRouter,

		infra_observe.NewObservibilityInjector,

		db.NewDatabaseConnection,

		broker.NewNATSPublisher,
		broker.NewNATSSubscriber,

		cache.NewLocalCache,
		cache.NewRedisClient,
		cache.NewRedisCache,

		proxy.NewProductRepoCache,

		product.NewProductService,
		product.NewSagaProductService,

		repo.NewProductRepository,

		pkg.NewSonyFlake,
	)
	return &infra.ProductServer{}, nil
}

func InitializeOrderServer() (*infra.OrderServer, error) {
	wire.Build(
		conf.NewConfig,

		infra.NewOrderServer,
		infra_http_order.NewOrderServer,
		infra_http_order.NewEngine,
		infra_http_order.NewRouter,

		middleware.NewJWTAuthChecker,

		infra_grpc_order.NewProductConn,
		infra_grpc_auth.NewAuthConn,

		infra_broker_order.NewOrderEventRouter,

		infra_observe.NewObservibilityInjector,

		db.NewDatabaseConnection,

		broker.NewNATSPublisher,
		broker.NewNATSSubscriber,

		cache.NewRedisClient,
		cache.NewRedisCache,

		proxy.NewOrderRepoCache,

		order.NewOrderService,
		order.NewSagaOrderService,

		repo.NewOrderRepository,
		repo.NewAuthRepository,
	)
	return &infra.OrderServer{}, nil
}

func InitializePaymentServer() (*infra.PaymentServer, error) {
	wire.Build(
		conf.NewConfig,

		infra.NewPaymentServer,
		infra_http_payment.NewPaymentServer,
		infra_http_payment.NewEngine,
		infra_http_payment.NewRouter,

		middleware.NewJWTAuthChecker,

		infra_grpc_auth.NewAuthConn,

		infra_broker_payment.NewPaymentEventRouter,

		infra_observe.NewObservibilityInjector,

		db.NewDatabaseConnection,

		broker.NewNATSPublisher,
		broker.NewNATSSubscriber,

		cache.NewRedisClient,
		cache.NewRedisCache,

		proxy.NewPaymentRepoCache,

		payment.NewPaymentService,
		payment.NewSagaPaymentService,

		repo.NewPaymentRepository,
		repo.NewAuthRepository,
	)
	return &infra.PaymentServer{}, nil
}

func InitializeOrchestratorServer() (*infra.OrchestratorServer, error) {
	wire.Build(
		conf.NewConfig,

		infra.NewOrchestratorServer,

		infra_broker_orchestrator.NewOrchestratorEventRouter,

		infra_observe.NewObservibilityInjector,

		broker.NewNATSPublisher,
		broker.NewNATSSubscriber,
		broker.NewRedisPublisher,

		orchestrator.NewOrchestratorService,
	)
	return &infra.OrchestratorServer{}, nil
}

func InitializeMigrator(app string) (*db.Migrator, error) {
	wire.Build(
		conf.NewConfig,
		db.NewDatabaseConnection,
		db.NewMigrator,
	)
	return &db.Migrator{}, nil
}
