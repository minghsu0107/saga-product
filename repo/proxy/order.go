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

// OrderRepoCache interface
type OrderRepoCache interface {
	GetOrder(ctx context.Context, orderID uint64) (*domain_model.Order, error)
	GetDetailedPurchasedItems(ctx context.Context, purchasedItems *[]domain_model.PurchasedItem) (*[]domain_model.DetailedPurchasedItem, error)
	ExistOrder(ctx context.Context, orderID uint64) (bool, error)
	CreateOrder(ctx context.Context, order *domain_model.Order) error
	DeleteOrder(ctx context.Context, orderID uint64) error
}

// OrderRepoCacheImpl implementation
type OrderRepoCacheImpl struct {
	orderRepo repo.OrderRepository
	rc        cache.RedisCache
	logger    *logrus.Entry
}

// NewOrderRepoCache factory
func NewOrderRepoCache(config *conf.Config, repo repo.OrderRepository, rc cache.RedisCache) OrderRepoCache {
	return &OrderRepoCacheImpl{
		orderRepo: repo,
		rc:        rc,
		logger:    config.Logger.ContextLogger.WithField("type", "cache:OrderRepoCache"),
	}
}

func (c *OrderRepoCacheImpl) GetOrder(ctx context.Context, orderID uint64) (*domain_model.Order, error) {
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

func (c *OrderRepoCacheImpl) ExistOrder(ctx context.Context, orderID uint64) (bool, error) {
	order := &domain_model.Order{}
	key := pkg.Join("order:", strconv.FormatUint(orderID, 10))

	ok, err := c.rc.Get(ctx, key, order)
	if ok && err == nil {
		return true, nil
	}
	return c.orderRepo.ExistOrder(ctx, orderID)
}

func (c *OrderRepoCacheImpl) CreateOrder(ctx context.Context, order *domain_model.Order) error {
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
