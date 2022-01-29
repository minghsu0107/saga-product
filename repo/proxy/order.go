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

var orderBloomFilter = "orderbloom"

// OrderRepoCache interface
type OrderRepoCache interface {
	GetOrder(ctx context.Context, orderID uint64) (*domain_model.Order, error)
	GetDetailedPurchasedItems(ctx context.Context, purchasedItems *[]domain_model.PurchasedItem) (*[]domain_model.DetailedPurchasedItem, error)
	CreateOrder(ctx context.Context, order *domain_model.Order) error
	DeleteOrder(ctx context.Context, orderID uint64) error
}

// OrderRepoCacheImpl implementation
type OrderRepoCacheImpl struct {
	orderRepo repo.OrderRepository
	rc        cache.RedisCache
	useBloom  bool
	logger    *logrus.Entry
}

// NewOrderRepoCache factory
func NewOrderRepoCache(config *conf.Config, repo repo.OrderRepository, rc cache.RedisCache) (OrderRepoCache, error) {
	useBloom := config.RedisConfig.Bloom.Activate
	orderRepoCache := OrderRepoCacheImpl{
		orderRepo: repo,
		rc:        rc,
		useBloom:  useBloom,
		logger:    config.Logger.ContextLogger.WithField("type", "cache:OrderRepoCache"),
	}
	if useBloom {
		ctx := context.Background()
		exist, err := rc.BFExist(ctx, orderBloomFilter, "dummy")
		if err != nil {
			return nil, err
		}
		if !exist {
			if err := rc.BFInsert(ctx, orderBloomFilter, config.RedisConfig.Bloom.ErrorRate, config.RedisConfig.Bloom.Capacity, dummyItem); err != nil {
				return nil, err
			}
			orderRepoCache.logger.Infof("bloom filter created: key = %s", orderBloomFilter)
		} else {
			orderRepoCache.logger.Infof("bloom filter already exists: key = %s", orderBloomFilter)
		}
	}
	return &orderRepoCache, nil
}

func (c *OrderRepoCacheImpl) GetOrder(ctx context.Context, orderID uint64) (*domain_model.Order, error) {
	if c.useBloom {
		exist, err := c.rc.BFExist(ctx, orderBloomFilter, orderID)
		c.logError(err)
		if !exist && err == nil {
			return nil, repo.ErrOrderNotFound
		}
	}

	order := &domain_model.Order{}
	key := pkg.Join("order:", strconv.FormatUint(orderID, 10))

	ok, err := c.rc.Get(ctx, key, order)
	if ok && err == nil {
		return order, nil
	}

	order, err = c.orderRepo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	c.logError(c.rc.Set(ctx, key, order))
	return order, nil
}

func (c *OrderRepoCacheImpl) GetDetailedPurchasedItems(ctx context.Context, purchasedItems *[]domain_model.PurchasedItem) (*[]domain_model.DetailedPurchasedItem, error) {
	return c.orderRepo.GetDetailedPurchasedItems(ctx, purchasedItems)
}

func (c *OrderRepoCacheImpl) CreateOrder(ctx context.Context, order *domain_model.Order) error {
	if c.useBloom {
		c.logError(c.rc.BFAdd(ctx, orderBloomFilter, order.ID))
	}
	return c.orderRepo.CreateOrder(ctx, order)
}

func (c *OrderRepoCacheImpl) DeleteOrder(ctx context.Context, orderID uint64) error {
	err := c.orderRepo.DeleteOrder(ctx, orderID)
	if err != nil {
		return err
	}
	key := pkg.Join("order:", strconv.FormatUint(orderID, 10))
	c.logError(c.rc.Delete(ctx, key))
	return nil
}

func (c *OrderRepoCacheImpl) logError(err error) {
	if err == nil {
		return
	}
	c.logger.Error(err.Error())
}
