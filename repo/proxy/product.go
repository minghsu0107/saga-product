package proxy

import (
	"context"

	domain_model "github.com/minghsu0107/saga-product/domain/model"
	"github.com/minghsu0107/saga-product/infra/cache"
	"github.com/minghsu0107/saga-product/repo"
)

// ProductRepoCache interface
type ProductRepoCache interface {
	CheckProduct(ctx context.Context, cartItem *domain_model.CartItem) (*repo.ProductStatus, error)
	ListProducts(ctx context.Context, offset, size int) (*[]repo.ProductCatalog, error)
	GetProductDetail(ctx context.Context, productID uint64) (*repo.ProductDetail, error)
	GetProductInventory(ctx context.Context, productID uint64) (int64, error)
	CreateProduct(ctx context.Context, product *domain_model.Product) error
	UpdateProductInventory(ctx context.Context, idempotencyKey uint64, purchasedItems *[]domain_model.PurchasedItem) error
}

// ProductRepoCacheImpl implementation
type ProductRepoCacheImpl struct {
	productRepo repo.ProductRepository
	lc          cache.LocalCache
	rc          cache.RedisCache
}

func NewProductRepoCache(repo repo.ProductRepository, lc cache.LocalCache, rc cache.RedisCache) ProductRepoCache {
	return &ProductRepoCacheImpl{
		productRepo: repo,
		lc:          lc,
		rc:          rc,
	}
}

func (c *ProductRepoCacheImpl) CheckProduct(ctx context.Context, cartItem *domain_model.CartItem) (*repo.ProductStatus, error) {
	return nil, nil
}

func (c *ProductRepoCacheImpl) ListProducts(ctx context.Context, offset, size int) (*[]repo.ProductCatalog, error) {
	return c.productRepo.ListProducts(ctx, offset, size)
}

func (c *ProductRepoCacheImpl) GetProductDetail(ctx context.Context, productID uint64) (*repo.ProductDetail, error) {
	return nil, nil
}

func (c *ProductRepoCacheImpl) GetProductInventory(ctx context.Context, productID uint64) (int64, error) {
	return 0, nil
}

func (c *ProductRepoCacheImpl) CreateProduct(ctx context.Context, product *domain_model.Product) error {
	return c.productRepo.CreateProduct(ctx, product)
}

func (c *ProductRepoCacheImpl) UpdateProductInventory(ctx context.Context, idempotencyKey uint64, purchasedItems *[]domain_model.PurchasedItem) error {
	return nil
}
