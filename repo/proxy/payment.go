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

var (
	paymentBloomFilter  = "paymentbloom"
	paymentCuckooFilter = "paymentcuckoo"
)

// PaymentRepoCache interface
type PaymentRepoCache interface {
	GetPayment(ctx context.Context, paymentID uint64) (*domain_model.Payment, error)
	CreatePayment(ctx context.Context, payment *domain_model.Payment) error
	DeletePayment(ctx context.Context, paymentID uint64) error
}

// PaymentRepoCacheImpl implementation
type PaymentRepoCacheImpl struct {
	paymentRepo repo.PaymentRepository
	rc          cache.RedisCache
	useCuckoo   bool
	logger      *logrus.Entry
}

// NewPaymentRepoCache factory
func NewPaymentRepoCache(config *conf.Config, repo repo.PaymentRepository, rc cache.RedisCache) (PaymentRepoCache, error) {
	useCuckoo := config.RedisConfig.UseCuckoo
	paymentRepoCache := PaymentRepoCacheImpl{
		paymentRepo: repo,
		rc:          rc,
		useCuckoo:   useCuckoo,
		logger:      config.Logger.ContextLogger.WithField("type", "cache:PaymentRepoCache"),
	}
	ctx := context.Background()
	var exist bool
	var err error
	if useCuckoo {
		exist, err = rc.CFExist(ctx, paymentCuckooFilter, dummyItem)
		if err != nil {
			return nil, err
		}
		if !exist {
			if err = rc.CFReserve(ctx, paymentCuckooFilter, config.RedisConfig.Cuckoo.Capacity, config.RedisConfig.Cuckoo.BucketSize, config.RedisConfig.Cuckoo.MaxIterations); err != nil {
				return nil, err
			}
			if err = rc.CFAdd(ctx, paymentCuckooFilter, dummyItem); err != nil {
				return nil, err
			}
			paymentRepoCache.logger.Infof("cuckoo filter created: key = %s", paymentCuckooFilter)
		} else {
			paymentRepoCache.logger.Infof("cuckoo filter already exists: key = %s", paymentCuckooFilter)
		}
	} else {
		exist, err = rc.BFExist(ctx, paymentBloomFilter, dummyItem)
		if err != nil {
			return nil, err
		}
		if !exist {
			if err = rc.BFInsert(ctx, paymentBloomFilter, config.RedisConfig.Bloom.ErrorRate, config.RedisConfig.Bloom.Capacity, dummyItem); err != nil {
				return nil, err
			}
			paymentRepoCache.logger.Infof("bloom filter created: key = %s", paymentBloomFilter)
		} else {
			paymentRepoCache.logger.Infof("bloom filter already exists: key = %s", paymentBloomFilter)
		}
	}
	return &paymentRepoCache, nil
}

func (c *PaymentRepoCacheImpl) GetPayment(ctx context.Context, paymentID uint64) (*domain_model.Payment, error) {
	if c.useCuckoo {
		exist, err := c.rc.CFExist(ctx, paymentCuckooFilter, paymentID)
		c.logError(err)
		if !exist && err == nil {
			return nil, repo.ErrPaymentNotFound
		}
	} else {
		exist, err := c.rc.BFExist(ctx, paymentBloomFilter, paymentID)
		c.logError(err)
		if !exist && err == nil {
			return nil, repo.ErrPaymentNotFound
		}
	}

	payment := &domain_model.Payment{}
	key := pkg.Join("payment:", strconv.FormatUint(paymentID, 10))

	ok, err := c.rc.Get(ctx, key, payment)
	c.logError(err)
	if ok && err == nil {
		return payment, nil
	}

	payment, err = c.paymentRepo.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	c.logError(c.rc.Set(ctx, key, payment))
	c.logError(c.rc.BFAdd(ctx, paymentBloomFilter, paymentID))
	return payment, nil
}

func (c *PaymentRepoCacheImpl) CreatePayment(ctx context.Context, payment *domain_model.Payment) error {
	if c.useCuckoo {
		c.logError(c.rc.CFAdd(ctx, paymentCuckooFilter, payment.ID))
	} else {
		c.logError(c.rc.BFAdd(ctx, paymentBloomFilter, payment.ID))
	}
	return c.paymentRepo.CreatePayment(ctx, payment)
}

func (c *PaymentRepoCacheImpl) DeletePayment(ctx context.Context, paymentID uint64) error {
	err := c.paymentRepo.DeletePayment(ctx, paymentID)
	if err != nil {
		return err
	}
	key := pkg.Join("payment:", strconv.FormatUint(paymentID, 10))
	c.logError(c.rc.Delete(ctx, key))
	if c.useCuckoo {
		c.logError(c.rc.CFDel(ctx, paymentCuckooFilter, paymentID))
	}
	return nil
}

func (c *PaymentRepoCacheImpl) logError(err error) {
	if err == nil {
		return
	}
	c.logger.Error(err.Error())
}
