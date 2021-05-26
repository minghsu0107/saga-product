package repo

import (
	"context"

	domain_model "github.com/minghsu0107/saga-product/domain/model"
)

// PaymentRepository interface
type PaymentRepository interface {
	GetPayment(ctx context.Context, paymentID uint64) (*domain_model.Payment, error)
	ExistPayment(ctx context.Context, paymentID uint64) bool
	CreatePayment(ctx context.Context, payment *domain_model.Payment) error
	DeletePayment(ctx context.Context, paymentID uint64) error
}
