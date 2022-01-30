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
	orderBloomFilter  = "orderbloom"
	orderCuckooFilter = "ordercuckoo"
)

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
	useCuckoo bool
	logger    *logrus.Entry
}

// NewOrderRepoCache factory
func NewOrderRepoCache(config *conf.Config, repo repo.OrderRepository, rc cache.RedisCache) (OrderRepoCache, error) {
	useCuckoo := config.RedisConfig.UseCuckoo
	orderRepoCache := OrderRepoCacheImpl{
		orderRepo: repo,
		rc:        rc,
		useCuckoo: useCuckoo,
		logger:    config.Logger.ContextLogger.WithField("type", "cache:OrderRepoCache"),
	}
	ctx := context.Background()
	var exist bool
	var err error
	if useCuckoo {
		exist, err = rc.CFExist(ctx, orderCuckooFilter, dummyItem)
		if err != nil {
			return nil, err
		}
		if !exist {
			if err = rc.CFReserve(ctx, orderCuckooFilter, config.RedisConfig.Cuckoo.Capacity, config.RedisConfig.Cuckoo.BucketSize, config.RedisConfig.Cuckoo.MaxIterations); err != nil {
				return nil, err
			}
			if err = rc.CFAdd(ctx, orderCuckooFilter, dummyItem); err != nil {
				return nil, err
			}
			orderRepoCache.logger.Infof("cuckoo filter created: key = %s", orderCuckooFilter)
		} else {
			orderRepoCache.logger.Infof("cuckoo filter already exists: key = %s", orderCuckooFilter)
		}
	} else {
		exist, err = rc.BFExist(ctx, orderBloomFilter, dummyItem)
		if err != nil {
			return nil, err
		}
		if !exist {
			if err = rc.BFInsert(ctx, orderBloomFilter, config.RedisConfig.Bloom.ErrorRate, config.RedisConfig.Bloom.Capacity, dummyItem); err != nil {
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
	if c.useCuckoo {
		exist, err := c.rc.CFExist(ctx, orderCuckooFilter, orderID)
		c.logError(err)
		if !exist && err == nil {
			return nil, repo.ErrOrderNotFound
		}
	} else {
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
	if c.useCuckoo {
		c.logError(c.rc.CFAdd(ctx, orderCuckooFilter, order.ID))
	} else {
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
	if c.useCuckoo {
		c.logError(c.rc.CFDel(ctx, orderCuckooFilter, orderID))
	}
	return nil
}

func (c *OrderRepoCacheImpl) logError(err error) {
	if err == nil {
		return
	}
	c.logger.Error(err.Error())
}
