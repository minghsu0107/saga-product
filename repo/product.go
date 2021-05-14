package repo

import (
	domain_model "github.com/minghsu0107/saga-product/domain/model"
	//"github.com/minghsu0107/saga-product/infra/db/model"
	"context"

	"gorm.io/gorm"
)

// ProductRepository is the product repository interface
type ProductRepository interface {
	CheckProduct(ctx context.Context, cartItem *domain_model.CartItem) (*ProductStatus, error)
	ListProducts(ctx context.Context, offset, size int) (*[]ProductCatalog, error)
	GetProductDetail(ctx context.Context, productID uint64) (*ProductDetail, error)
	GetProductInventory(ctx context.Context, productID uint64) (int64, error)
	CreateProduct(ctx context.Context, product *domain_model.Product) error
	UpdateProductInventory(ctx context.Context, idempotencyKey uint64, purchasedItems *[]domain_model.PurchasedItem) error
}

// ProductStatus select schema
type ProductStatus struct {
	ProductID uint64
	Exist     bool
}

// ProductCatalog select schema
type ProductCatalog struct {
	ID        uint64
	Name      string
	Inventory int64
}

// ProductDetail select schema
type ProductDetail struct {
	Name        string
	Description string
	BrandName   string
}

// ProductRepositoryImpl implements ProductRepository interface
type ProductRepositoryImpl struct {
	db *gorm.DB
}

// NewProductRepository is the factory of ProductRepository
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &ProductRepositoryImpl{
		db: db,
	}
}

// CheckProducts method
func (repo *ProductRepositoryImpl) CheckProduct(ctx context.Context, cartItem *domain_model.CartItem) (*ProductStatus, error) {
	return nil, nil
}

// ListProducts method
func (repo *ProductRepositoryImpl) ListProducts(ctx context.Context, offset, size int) (*[]ProductCatalog, error) {
	return nil, nil
}

// GetProductDetails method
func (repo *ProductRepositoryImpl) GetProductDetail(ctx context.Context, productID uint64) (*ProductDetail, error) {
	return nil, nil
}

// GetProductInventory method
func (repo *ProductRepositoryImpl) GetProductInventory(ctx context.Context, productID uint64) (int64, error) {
	return 0, nil
}

// CreateProduct method
func (repo *ProductRepositoryImpl) CreateProduct(ctx context.Context, product *domain_model.Product) error {
	return nil
}

// UpdateProductInventory method
func (repo *ProductRepositoryImpl) UpdateProductInventory(ctx context.Context, idempotencyKey uint64, purchasedItems *[]domain_model.PurchasedItem) error {
	return nil
}
