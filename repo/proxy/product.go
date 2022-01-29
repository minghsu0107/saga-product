package proxy

import (
	"context"
	"strconv"

	conf "github.com/minghsu0107/saga-product/config"
	domain_model "github.com/minghsu0107/saga-product/domain/model"
	"github.com/minghsu0107/saga-product/infra/cache"
	"github.com/minghsu0107/saga-product/pkg"
	"github.com/minghsu0107/saga-product/repo"
	"github.com/sirupsen/logrus"
)

var (
	productBloomFilter = "productbloom"
	dummyItem          = "dummy"
)

// ProductRepoCache interface
type ProductRepoCache interface {
	CheckProduct(ctx context.Context, cartItem *domain_model.CartItem) (*repo.ProductStatus, error)
	ListProducts(ctx context.Context, offset, size int) (*[]repo.ProductCatalog, error)
	GetProductDetail(ctx context.Context, productID uint64) (*repo.ProductDetail, error)
	GetProductInventory(ctx context.Context, productID uint64) (int64, error)
	CreateProduct(ctx context.Context, product *domain_model.Product) (uint64, error)
	UpdateProductInventory(ctx context.Context, idempotencyKey uint64, purchasedItems *[]domain_model.PurchasedItem) error
	RollbackProductInventory(ctx context.Context, idempotencyKey uint64) error
}

// ProductRepoCacheImpl implementation
type ProductRepoCacheImpl struct {
	productRepo repo.ProductRepository
	lc          cache.LocalCache
	rc          cache.RedisCache
	useBloom    bool
	logger      *logrus.Entry
}

func NewProductRepoCache(config *conf.Config, repo repo.ProductRepository, lc cache.LocalCache, rc cache.RedisCache) (ProductRepoCache, error) {
	useBloom := config.RedisConfig.Bloom.Activate
	productRepoCache := ProductRepoCacheImpl{
		productRepo: repo,
		lc:          lc,
		rc:          rc,
		useBloom:    useBloom,
		logger:      config.Logger.ContextLogger.WithField("type", "cache:ProductRepoCache"),
	}
	if useBloom {
		ctx := context.Background()
		exist, err := rc.BFExist(ctx, productBloomFilter, dummyItem)
		if err != nil {
			return nil, err
		}
		if !exist {
			if err = rc.BFInsert(ctx, productBloomFilter, config.RedisConfig.Bloom.ErrorRate, config.RedisConfig.Bloom.Capacity, dummyItem); err != nil {
				return nil, err
			}
			productRepoCache.logger.Infof("bloom filter already created: key = %s", productBloomFilter)
		} else {
			productRepoCache.logger.Infof("bloom filter already exists: key = %s", productBloomFilter)
		}
	}
	return &productRepoCache, nil
}

func (c *ProductRepoCacheImpl) CheckProduct(ctx context.Context, cartItem *domain_model.CartItem) (*repo.ProductStatus, error) {
	if c.useBloom {
		exist, err := c.rc.BFExist(ctx, productBloomFilter, cartItem.ProductID)
		c.logError(err)
		if !exist && err == nil {
			return &repo.ProductStatus{
				ProductID: cartItem.ProductID,
				Price:     0,
				Exist:     false,
			}, nil
		}
	}
	status := &repo.ProductStatus{}
	key := pkg.Join("productcheck:", strconv.FormatUint(cartItem.ProductID, 10))

	ok, err := c.lc.Get(key, status)
	c.logError(err)
	if ok && err == nil {
		return status, nil
	}

	ok, err = c.rc.Get(ctx, key, status)
	c.logError(err)
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
	c.logError(err)
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
	if c.useBloom {
		exist, err := c.rc.BFExist(ctx, productBloomFilter, productID)
		c.logError(err)
		if !exist && err == nil {
			return nil, repo.ErrProductNotFound
		}
	}

	detail := &repo.ProductDetail{}
	key := pkg.Join("productdetail:", strconv.FormatUint(productID, 10))

	ok, err := c.lc.Get(key, detail)
	c.logError(err)
	if ok && err == nil {
		return detail, nil
	}

	ok, err = c.rc.Get(ctx, key, detail)
	c.logError(err)
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
	c.logError(err)
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
	if c.useBloom {
		exist, err := c.rc.BFExist(ctx, productBloomFilter, productID)
		c.logError(err)
		if !exist && err == nil {
			return 0, repo.ErrProductNotFound
		}
	}

	var redisInventory int64
	key := pkg.Join("productinventory:", strconv.FormatUint(productID, 10))

	ok, err := c.rc.Get(ctx, key, &redisInventory)
	c.logError(err)
	if ok && err == nil {
		c.logError(c.lc.Set(key, &redisInventory))
		return redisInventory, nil
	}

	// get lock (request coalescing)
	mutex := c.rc.GetMutex(pkg.Join("mutex:", key))
	if err := mutex.Lock(); err != nil {
		return 0, err
	}
	defer mutex.Unlock()

	ok, err = c.rc.Get(ctx, key, &redisInventory)
	c.logError(err)
	if ok && err == nil {
		return redisInventory, nil
	}
	var inventory int64
	inventory, err = c.productRepo.GetProductInventory(ctx, productID)
	if err != nil {
		return 0, err
	}

	c.logError(c.rc.Set(ctx, key, &inventory))
	return inventory, nil
}

func (c *ProductRepoCacheImpl) CreateProduct(ctx context.Context, product *domain_model.Product) (uint64, error) {
	productID, err := c.productRepo.CreateProduct(ctx, product)
	if err != nil {
		return 0, err
	}
	if c.useBloom {
		c.logError(c.rc.BFAdd(ctx, productBloomFilter, productID))
	}
	return productID, nil
}

func (c *ProductRepoCacheImpl) UpdateProductInventory(ctx context.Context, idempotencyKey uint64, purchasedItems *[]domain_model.PurchasedItem) error {
	err := c.productRepo.UpdateProductInventory(ctx, idempotencyKey, purchasedItems)
	if err != nil {
		return err
	}
	var cmds []cache.RedisCmd
	for _, purchasedItem := range *purchasedItems {
		key := pkg.Join("productinventory:", strconv.FormatUint(purchasedItem.ProductID, 10))
		exist, err := c.rc.Exist(ctx, key)
		if exist && err == nil {
			cmds = append(cmds, cache.RedisCmd{
				OpType: cache.INCRBY,
				Payload: cache.RedisIncrByPayload{
					Key: key,
					Val: -purchasedItem.Amount,
				},
			})
		}
	}
	if len(cmds) > 0 {
		c.logError(c.rc.ExecPipeLine(ctx, &cmds))
	}
	return nil
}

// RollbackProductInventory method
func (c *ProductRepoCacheImpl) RollbackProductInventory(ctx context.Context, idempotencyKey uint64) error {
	var err error
	var rollbacked bool
	var idempotencies *[]domain_model.Idempotency
	rollbacked, idempotencies, err = c.productRepo.RollbackProductInventory(ctx, idempotencyKey)
	if err != nil {
		return err
	}
	if rollbacked {
		return nil
	}
	var cmds []cache.RedisCmd
	for _, idempotency := range *idempotencies {
		key := pkg.Join("productinventory:", strconv.FormatUint(idempotency.ProductID, 10))
		exist, err := c.rc.Exist(ctx, key)
		if exist && err == nil {
			cmds = append(cmds, cache.RedisCmd{
				OpType: cache.INCRBY,
				Payload: cache.RedisIncrByPayload{
					Key: key,
					Val: idempotency.Amount,
				},
			})
		}
	}
	if len(cmds) > 0 {
		c.logError(c.rc.ExecPipeLine(ctx, &cmds))
	}
	return nil
}

func (c *ProductRepoCacheImpl) logError(err error) {
	if err == nil {
		return
	}
	c.logger.Error(err.Error())
}
