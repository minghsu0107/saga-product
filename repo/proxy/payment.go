package proxy

import (
	"context"

	domain_model "github.com/minghsu0107/saga-product/domain/model"
)

// PaymentRepoCache interface
type PaymentRepoCache interface {
	GetPayment(ctx context.Context, paymentID uint64) (*domain_model.Payment, error)
	ExistPayment(ctx context.Context, paymentID uint64) (bool, error)
	CreatePayment(ctx context.Context, payment *domain_model.Payment) error
	DeletePayment(ctx context.Context, paymentID uint64) error
}
