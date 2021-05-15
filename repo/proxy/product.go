package proxy

import (
	"context"
	"strconv"

	conf "github.com/minghsu0107/saga-account/config"
	domain_model "github.com/minghsu0107/saga-product/domain/model"
	"github.com/minghsu0107/saga-product/infra/cache"
	"github.com/minghsu0107/saga-product/pkg"
	"github.com/minghsu0107/saga-product/repo"
	"github.com/sirupsen/logrus"
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
	logger      *logrus.Entry
}

// RedisProductInventory stores product inventory in redis
type RedisProductInventory struct {
	Inventory int64 `redis:"inventory"`
}

func NewProductRepoCache(config *conf.Config, repo repo.ProductRepository, lc cache.LocalCache, rc cache.RedisCache) ProductRepoCache {
	return &ProductRepoCacheImpl{
		productRepo: repo,
		lc:          lc,
		rc:          rc,
		logger:      config.Logger.ContextLogger.WithField("type", "cache"),
	}
}

func (c *ProductRepoCacheImpl) CheckProduct(ctx context.Context, cartItem *domain_model.CartItem) (*repo.ProductStatus, error) {
	status := &repo.ProductStatus{}
	key := pkg.Join("productcheck:", strconv.FormatUint(cartItem.ProductID, 10))

	ok, err := c.lc.Get(key, status)
	if ok && err == nil {
		return status, nil
	}

	ok, err = c.rc.Get(ctx, key, status)
	if ok && err == nil {
		c.logError(c.lc.Set(key, status))
		return status, nil
	}

	// get lock (request coalescing)
	mutex := c.rc.GetMutex(pkg.Join("mutex:", key))
	if err := mutex.Lock(); err != nil {
		return nil, err
	}
	defer mutex.Unlock()

	ok, err = c.rc.Get(ctx, key, status)
	if ok && err == nil {
		c.logError(c.lc.Set(key, status))
		return status, nil
	}
	status, err = c.productRepo.CheckProduct(ctx, cartItem)
	if err != nil {
		return nil, err
	}

	c.logError(c.rc.Set(ctx, key, status))
	return status, nil
}

func (c *ProductRepoCacheImpl) ListProducts(ctx context.Context, offset, size int) (*[]repo.ProductCatalog, error) {
	return c.productRepo.ListProducts(ctx, offset, size)
}

func (c *ProductRepoCacheImpl) GetProductDetail(ctx context.Context, productID uint64) (*repo.ProductDetail, error) {
	detail := &repo.ProductDetail{}
	key := pkg.Join("productdetail:", strconv.FormatUint(productID, 10))

	ok, err := c.lc.Get(key, detail)
	if ok && err == nil {
		return detail, nil
	}

	ok, err = c.rc.Get(ctx, key, detail)
	if ok && err == nil {
		c.logError(c.lc.Set(key, detail))
		return detail, nil
	}

	// get lock (request coalescing)
	mutex := c.rc.GetMutex(pkg.Join("mutex:", key))
	if err := mutex.Lock(); err != nil {
		return nil, err
	}
	defer mutex.Unlock()

	ok, err = c.rc.Get(ctx, key, detail)
	if ok && err == nil {
		c.logError(c.lc.Set(key, detail))
		return detail, nil
	}
	detail, err = c.productRepo.GetProductDetail(ctx, productID)
	if err != nil {
		return nil, err
	}

	c.logError(c.rc.Set(ctx, key, detail))
	return detail, nil
}

func (c *ProductRepoCacheImpl) GetProductInventory(ctx context.Context, productID uint64) (int64, error) {
	redisInventory := &RedisProductInventory{}
	key := pkg.Join("productinventory:", strconv.FormatUint(productID, 10))

	ok, err := c.rc.Get(ctx, key, redisInventory)
	if ok && err == nil {
		c.logError(c.lc.Set(key, redisInventory))
		return redisInventory.Inventory, nil
	}

	// get lock (request coalescing)
	mutex := c.rc.GetMutex(pkg.Join("mutex:", key))
	if err := mutex.Lock(); err != nil {
		return 0, err
	}
	defer mutex.Unlock()

	ok, err = c.rc.Get(ctx, key, redisInventory)
	if ok && err == nil {
		return redisInventory.Inventory, nil
	}
	var inventory int64
	inventory, err = c.productRepo.GetProductInventory(ctx, productID)
	if err != nil {
		return 0, err
	}

	c.logError(c.rc.Set(ctx, key, &RedisProductInventory{inventory}))
	return inventory, nil
}

func (c *ProductRepoCacheImpl) CreateProduct(ctx context.Context, product *domain_model.Product) error {
	return c.productRepo.CreateProduct(ctx, product)
}

func (c *ProductRepoCacheImpl) UpdateProductInventory(ctx context.Context, idempotencyKey uint64, purchasedItems *[]domain_model.PurchasedItem) error {
	err := c.productRepo.UpdateProductInventory(ctx, idempotencyKey, purchasedItems)
	if err != nil {
		return err
	}
	var cmds []cache.RedisCmd
	for _, purchasedItem := range *purchasedItems {
		cmds = append(cmds, cache.RedisCmd{
			OpType: cache.INCRBY,
			Payload: cache.RedisIncrByPayload{
				Key: pkg.Join("productinventory:", strconv.FormatUint(purchasedItem.ProductID, 10)),
				Val: purchasedItem.Amount,
			},
		})
	}
	c.logger.Error(c.rc.ExecPipeLine(ctx, &cmds))
	return nil
}

func (c *ProductRepoCacheImpl) logError(err error) {
	if err == nil {
		return
	}
	c.logger.Error(err)
}
