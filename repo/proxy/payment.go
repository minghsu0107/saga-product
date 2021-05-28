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

// PaymentRepoCache interface
type PaymentRepoCache interface {
	GetPayment(ctx context.Context, paymentID uint64) (*domain_model.Payment, error)
	ExistPayment(ctx context.Context, paymentID uint64) (bool, error)
	CreatePayment(ctx context.Context, payment *domain_model.Payment) error
	DeletePayment(ctx context.Context, paymentID uint64) error
}

// PaymentRepoCacheImpl implementation
type PaymentRepoCacheImpl struct {
	paymentRepo repo.PaymentRepository
	rc          cache.RedisCache
	logger      *logrus.Entry
}

// NewPaymentRepoCache factory
func NewPaymentRepoCache(config *conf.Config, repo repo.PaymentRepository, rc cache.RedisCache) PaymentRepoCache {
	return &PaymentRepoCacheImpl{
		paymentRepo: repo,
		rc:          rc,
		logger:      config.Logger.ContextLogger.WithField("type", "cache:PaymentRepoCache"),
	}
}

func (c *PaymentRepoCacheImpl) GetPayment(ctx context.Context, paymentID uint64) (*domain_model.Payment, error) {
	payment := &domain_model.Payment{}
	key := pkg.Join("payment:", strconv.FormatUint(paymentID, 10))

	ok, err := c.rc.Get(ctx, key, payment)
	if ok && err == nil {
		return payment, nil
	}

	payment, err = c.paymentRepo.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	c.logError(c.rc.Set(ctx, key, payment))
	return payment, nil
}

func (c *PaymentRepoCacheImpl) ExistPayment(ctx context.Context, paymentID uint64) (bool, error) {
	payment := &domain_model.Payment{}
	key := pkg.Join("payment:", strconv.FormatUint(paymentID, 10))

	ok, err := c.rc.Get(ctx, key, payment)
	if ok && err == nil {
		return true, nil
	}
	return c.paymentRepo.ExistPayment(ctx, paymentID)
}

func (c *PaymentRepoCacheImpl) CreatePayment(ctx context.Context, payment *domain_model.Payment) error {
	return c.paymentRepo.CreatePayment(ctx, payment)
}

func (c *PaymentRepoCacheImpl) DeletePayment(ctx context.Context, paymentID uint64) error {
	err := c.paymentRepo.DeletePayment(ctx, paymentID)
	if err != nil {
		return err
	}
	key := pkg.Join("payment:", strconv.FormatUint(paymentID, 10))
	c.logError(c.rc.Delete(ctx, key))
	return nil
}

func (c *PaymentRepoCacheImpl) logError(err error) {
	if err == nil {
		return
	}
	c.logger.Error(err.Error())
}
