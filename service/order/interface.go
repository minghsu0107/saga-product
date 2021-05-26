package order

import (
	"context"

	"github.com/minghsu0107/saga-product/domain/model"
)

// OrderService interface
type OrderService interface {
	GetDetailedOrder(ctx context.Context, orderID uint64) (*model.DetailedOrder, error)
}

// SagaOrderService interface
type SagaOrderService interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	RollbackOrder(ctx context.Context, orderID uint64) error
}
