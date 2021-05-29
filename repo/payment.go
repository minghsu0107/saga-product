package repo

import (
	"context"
	"errors"
	"fmt"

	domain_model "github.com/minghsu0107/saga-product/domain/model"
	"github.com/minghsu0107/saga-product/infra/db/model"
	"gorm.io/gorm"
)

// PaymentRepository interface
type PaymentRepository interface {
	GetPayment(ctx context.Context, paymentID uint64) (*domain_model.Payment, error)
	CreatePayment(ctx context.Context, payment *domain_model.Payment) error
	DeletePayment(ctx context.Context, paymentID uint64) error
}

// PaymentRepositoryImpl implementation
type PaymentRepositoryImpl struct {
	db *gorm.DB
}

// NewPaymentRepository factory
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &PaymentRepositoryImpl{
		db: db,
	}
}

// GetPayment get an payment
func (repo *PaymentRepositoryImpl) GetPayment(ctx context.Context, paymentID uint64) (*domain_model.Payment, error) {
	var payment model.Payment
	if err := repo.db.Model(&model.Payment{}).Select("id", "customer_id", "currency_code", "amount").Where("id = ?", paymentID).First(&payment).WithContext(ctx).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("payment not found; payment ID: %v", paymentID)
		}
		return nil, err
	}
	return &domain_model.Payment{
		ID:           payment.ID,
		CustomerID:   payment.CustomerID,
		CurrencyCode: payment.CurrencyCode,
		Amount:       payment.Amount,
	}, nil
}

// CreatePayment creates a payment
func (repo *PaymentRepositoryImpl) CreatePayment(ctx context.Context, payment *domain_model.Payment) error {
	if err := repo.db.Create(&model.Payment{
		ID:           payment.ID,
		CustomerID:   payment.CustomerID,
		CurrencyCode: payment.CurrencyCode,
		Amount:       payment.Amount,
	}).WithContext(ctx).Error; err != nil {
		return err
	}
	return nil
}

// DeletePayment deletes an payment
func (repo *PaymentRepositoryImpl) DeletePayment(ctx context.Context, paymentID uint64) error {
	if err := repo.db.Exec("DELETE FROM payments where id = ?", paymentID).Error; err != nil {
		return err
	}
	return nil
}
