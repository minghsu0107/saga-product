package product

import (
	"context"

	conf "github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/domain/model"
	"github.com/minghsu0107/saga-product/repo"
	"github.com/minghsu0107/saga-product/repo/proxy"
	log "github.com/sirupsen/logrus"
)

// ProductServiceImpl implementation
type ProductServiceImpl struct {
	productRepo proxy.ProductRepoCache
	logger      *log.Entry
}

// NewProductService is the factory of ProductService
func NewProductService(config *conf.Config, productRepo proxy.ProductRepoCache) ProductService {
	return &ProductServiceImpl{
		productRepo: productRepo,
		logger: config.Logger.ContextLogger.WithFields(log.Fields{
			"type": "service:CustomerService",
		}),
	}
}

func (svc *ProductServiceImpl) CheckProducts(ctx context.Context, cartItems *[]model.CartItem) (*[]model.ProductStatus, error) {
	var productStatuses []model.ProductStatus
	for _, cartItem := range *cartItems {
		status, err := svc.productRepo.CheckProduct(ctx, &cartItem)
		if err != nil {
			svc.logger.Error(err)
			return nil, err
		}
		productStatuses = append(productStatuses, *mapProductStatus(status))
	}
	return &productStatuses, nil
}

func (svc *ProductServiceImpl) ListProducts(ctx context.Context, offset, size int) (*[]model.ProductCatalog, error) {
	repoProductCatalogs, err := svc.productRepo.ListProducts(ctx, offset, size)
	if err != nil {
		svc.logger.Error(err)
		return nil, err
	}
	var productCatalogs []model.ProductCatalog
	for _, repoProductCatalog := range *repoProductCatalogs {
		productCatalogs = append(productCatalogs, model.ProductCatalog{
			ID:        repoProductCatalog.ID,
			Name:      repoProductCatalog.Name,
			Inventory: repoProductCatalog.Inventory,
		})
	}
	return &productCatalogs, nil
}

func (svc *ProductServiceImpl) GetProducts(ctx context.Context, productIDs []uint64) (*[]model.Product, error) {
	var products []model.Product
	for _, productID := range productIDs {
		productDetail, err := svc.productRepo.GetProductDetail(ctx, productID)
		if err != nil {
			svc.logger.Error(err)
			return nil, err
		}
		inventory, err := svc.productRepo.GetProductInventory(ctx, productID)
		if err != nil {
			svc.logger.Error(err)
			return nil, err
		}
		products = append(products, model.Product{
			ID: productID,
			Detail: &model.ProductDetail{
				Name:        productDetail.Name,
				Description: productDetail.Description,
				BrandName:   productDetail.BrandName,
			},
			Inventory: inventory,
		})
	}
	return &products, nil
}

func (svc *ProductServiceImpl) CreateProduct(ctx context.Context, product *model.Product) error {
	err := svc.productRepo.CreateProduct(ctx, product)
	if err != nil {
		svc.logger.Error(err)
		return err
	}
	return nil
}

func mapProductStatus(status *repo.ProductStatus) *model.ProductStatus {
	productStatus := &model.ProductStatus{
		ProductID: status.ProductID,
	}
	if status.Exist {
		productStatus.Status = model.ProductOk
	} else {
		productStatus.Status = model.ProductNotFound
	}
	return productStatus
}

// SagaProductServiceImpl implementation
type SagaProductServiceImpl struct {
	productRepo proxy.ProductRepoCache
	logger      *log.Entry
}

// NewSagaProductService is the factory of ProductService
func NewSagaProductService(config *conf.Config, productRepo proxy.ProductRepoCache) SagaProductService {
	return &SagaProductServiceImpl{
		productRepo: productRepo,
		logger: config.Logger.ContextLogger.WithFields(log.Fields{
			"type": "service:CustomerService",
		}),
	}
}

func (svc *SagaProductServiceImpl) UpdateProductInventory(ctx context.Context, idempotencyKey uint64, purchasedItems *[]model.PurchasedItem) error {
	err := svc.productRepo.UpdateProductInventory(ctx, idempotencyKey, purchasedItems)
	if err != nil {
		if err != repo.ErrInsuffientInventory {
			svc.logger.Error(err)
			return err
		}
		return ErrInsuffientInventory
	}
	return nil
}
func (svc *SagaProductServiceImpl) RollbackProductInventory(ctx context.Context, idempotencyKey uint64) error {
	svc.logger.Infof("rollback product inventory; idempotency key: %v\n", idempotencyKey)
	return nil
}
