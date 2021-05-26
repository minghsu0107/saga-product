//+build wireinject

// The build tag makes sure the stub is not built in the final build.
package dep

import (
	"github.com/google/wire"
	conf "github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/repo"
	"github.com/minghsu0107/saga-product/repo/proxy"
	"github.com/minghsu0107/saga-product/service/product"

	"github.com/minghsu0107/saga-product/infra"
	"github.com/minghsu0107/saga-product/infra/broker"
	infra_broker_product "github.com/minghsu0107/saga-product/infra/broker/product"
	"github.com/minghsu0107/saga-product/infra/cache"
	"github.com/minghsu0107/saga-product/infra/db"
	infra_grpc_product "github.com/minghsu0107/saga-product/infra/grpc/product"
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

func InitializeMigrator(app string) (*db.Migrator, error) {
	wire.Build(
		conf.NewConfig,
		db.NewDatabaseConnection,
		db.NewMigrator,
	)
	return &db.Migrator{}, nil
}
