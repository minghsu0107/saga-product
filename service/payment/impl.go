package payment

import (
	"context"
	"errors"

	conf "github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/domain/model"
	"github.com/minghsu0107/saga-product/repo"
	"github.com/minghsu0107/saga-product/repo/proxy"
	log "github.com/sirupsen/logrus"
)

// PaymentServiceImpl implementation
type PaymentServiceImpl struct {
	paymentRepo proxy.PaymentRepoCache
	logger      *log.Entry
}

// SagaPaymentServiceImpl implementation
type SagaPaymentServiceImpl struct {
	paymentRepo proxy.PaymentRepoCache
	logger      *log.Entry
}

// NewPaymentService factory
func NewPaymentService(config *conf.Config, paymentRepo proxy.PaymentRepoCache) PaymentService {
	return &PaymentServiceImpl{
		paymentRepo: paymentRepo,
		logger: config.Logger.ContextLogger.WithFields(log.Fields{
			"type": "service:PaymentService",
		}),
	}
}

// GetPayment method
func (svc *PaymentServiceImpl) GetPayment(ctx context.Context, customerID, paymentID uint64) (*model.Payment, error) {
	payment, err := svc.paymentRepo.GetPayment(ctx, paymentID)
	if err != nil {
		svc.logger.Error(err.Error())
		if errors.Is(err, repo.ErrPaymentNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	if customerID != payment.CustomerID {
		return nil, ErrUnauthorized
	}
	return payment, nil
}

// NewSagaPaymentService factory
func NewSagaPaymentService(config *conf.Config, paymentRepo proxy.PaymentRepoCache) SagaPaymentService {
	return &SagaPaymentServiceImpl{
		paymentRepo: paymentRepo,
		logger: config.Logger.ContextLogger.WithFields(log.Fields{
			"type": "service:SagaPaymentService",
		}),
	}
}

// CreatePayment method
func (svc *SagaPaymentServiceImpl) CreatePayment(ctx context.Context, payment *model.Payment) error {
	if err := svc.paymentRepo.CreatePayment(ctx, payment); err != nil {
		svc.logger.Error(err.Error())
		return err
	}
	return nil
}

// RollbackPayment method
func (svc *SagaPaymentServiceImpl) RollbackPayment(ctx context.Context, paymentID uint64) error {
	err := svc.paymentRepo.DeletePayment(ctx, paymentID)
	if err != nil {
		svc.logger.Error(err.Error())
		return err
	}
	return nil
}
