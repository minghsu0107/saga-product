package payment

import (
	"context"

	"github.com/minghsu0107/saga-product/domain/model"
)

// PaymentService interface
type PaymentService interface {
	GetPayment(ctx context.Context, paymentID uint64) (*model.Payment, error)
}

// SagaPaymentService interface
type SagaPaymentService interface {
	CreatePayment(ctx context.Context, payment *model.Payment) error
	RollbackPayment(ctx context.Context, paymentID uint64) error
}
