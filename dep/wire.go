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
	"github.com/minghsu0107/saga-product/infra/cache"
	"github.com/minghsu0107/saga-product/infra/db"
	infra_grpc "github.com/minghsu0107/saga-product/infra/grpc"
	infra_http "github.com/minghsu0107/saga-product/infra/http"
	infra_observe "github.com/minghsu0107/saga-product/infra/observe"
	"github.com/minghsu0107/saga-product/pkg"
)

func InitializeProductServer() (*infra.Server, error) {
	wire.Build(
		conf.NewConfig,

		infra.NewServer,
		infra_http.NewServer,
		infra_http.NewEngine,
		infra_http.NewRouter,

		infra_grpc.NewGRPCServer,

		infra_observe.NewObservibilityInjector,

		db.NewDatabaseConnection,

		cache.NewLocalCache,
		cache.NewRedisClient,
		cache.NewRedisCache,

		proxy.NewProductRepoCache,

		product.NewProductService,
		product.NewSagaProductService,

		repo.NewProductRepository,

		pkg.NewSonyFlake,
	)
	return &infra.Server{}, nil
}

func InitializeMigrator(app string) (*db.Migrator, error) {
	wire.Build(
		conf.NewConfig,
		db.NewDatabaseConnection,
		db.NewMigrator,
	)
	return &db.Migrator{}, nil
}
