package orchestrator

import (
	"context"

	"github.com/minghsu0107/saga-product/domain/event"
	"github.com/minghsu0107/saga-product/domain/model"
)

// OrchestratorService interface
type OrchestratorService interface {
	StartSaga(ctx context.Context, payload *model.OrchestratorPayload)
	UpdateProductInventory(ctx context.Context, idempotencyKey uint64, purchasedItems *[]model.PurchasedItem) error
	RollbackProductInventory(ctx context.Context, idempotencyKey uint64) error

	CreateOrder(ctx context.Context, order *model.Order) error
	RollbackOrder(ctx context.Context, order_id uint64) error

	CreatePayment(ctx context.Context, payment *model.Payment) error
	RollbackPayment(ctx context.Context, payment_id uint64) error

	PublishPurchaseEvent(ctx context.Context, result *event.PurchaseResult) error
}
