package order

import (
	"context"

	conf "github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/domain/model"
	"github.com/minghsu0107/saga-product/repo/proxy"
	log "github.com/sirupsen/logrus"
)

// OrderServiceImpl implementation
type OrderServiceImpl struct {
	orderRepo proxy.OrderRepoCache
	logger    *log.Entry
}

// SagaOrderServiceImpl implementation
type SagaOrderServiceImpl struct {
	orderRepo proxy.OrderRepoCache
	logger    *log.Entry
}

// NewOrderService factory
func NewOrderService(config *conf.Config, orderRepo proxy.OrderRepoCache) OrderService {
	return &OrderServiceImpl{
		orderRepo: orderRepo,
		logger: config.Logger.ContextLogger.WithFields(log.Fields{
			"type": "service:OrderService",
		}),
	}
}

// GetOrder method
func (svc *OrderServiceImpl) GetDetailedOrder(ctx context.Context, orderID uint64) (*model.DetailedOrder, error) {
	order, err := svc.orderRepo.GetOrder(ctx, orderID)
	if err != nil {
		svc.logger.Error(err)
		return nil, err
	}
	var productIDs []uint64
	for _, purchasedItem := range *order.PurchasedItems {
		productIDs = append(productIDs, purchasedItem.ProductID)
	}
	detailedPurchasedItems, err := svc.orderRepo.GetDetailedPurchasedItems(ctx, &productIDs)
	if err != nil {
		svc.logger.Error(err)
		return nil, err
	}
	return &model.DetailedOrder{
		ID:                     order.ID,
		CustomerID:             order.CustomerID,
		DetailedPurchasedItems: detailedPurchasedItems,
	}, nil
}

// NewSagaOrderService factory
func NewSagaOrderService(config *conf.Config, orderRepo proxy.OrderRepoCache) SagaOrderService {
	return &SagaOrderServiceImpl{
		orderRepo: orderRepo,
		logger: config.Logger.ContextLogger.WithFields(log.Fields{
			"type": "service:SagaOrderService",
		}),
	}
}

// CreateOrder method
func (svc *SagaOrderServiceImpl) CreateOrder(ctx context.Context, order *model.Order) error {
	if err := svc.orderRepo.CreateOrder(ctx, order); err != nil {
		svc.logger.Error(err)
		return err
	}
	return nil
}

// RollbackOrder method
func (svc *SagaOrderServiceImpl) RollbackOrder(ctx context.Context, orderID uint64) error {
	exist, err := svc.orderRepo.ExistOrder(ctx, orderID)
	if err != nil {
		svc.logger.Error(err)
		return err
	}
	if exist {
		return nil
	}
	err = svc.orderRepo.DeleteOrder(ctx, orderID)
	if err != nil {
		svc.logger.Error(err)
		return err
	}
	return nil
}
