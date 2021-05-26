package repo

import (
	"context"

	domain_model "github.com/minghsu0107/saga-product/domain/model"
)

// OrderRepository interface
type OrderRepository interface {
	GetOrder(ctx context.Context, orderID uint64) (*domain_model.Order, error)
	ExistOrder(ctx context.Context, orderID uint64) bool
	CreateOrder(ctx context.Context, order *domain_model.Order) error
	DeleteOrder(ctx context.Context, orderID uint64) error
}
