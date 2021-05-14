package product

import (
	"context"

	"github.com/minghsu0107/saga-product/domain/model"
)

// ProductService interface
type ProductService interface {
	CheckProducts(ctx context.Context, cartItems *[]model.CartItem) (*[]model.ProductStatus, error)
	ListProducts(ctx context.Context) (*[]model.ProductCatalog, error)
	GetProducts(ctx context.Context, productIDs []uint64) (*[]model.Product, error)
	CreateProduct(ctx context.Context, product *model.Product) error
}

// SagaProductService interface
type SagaProductService interface {
	UpdateProductInventory(ctx context.Context, idempotencyKey uint64, purchasedItems *[]model.PurchasedItem) error
	RollbackProductInventory(ctx context.Context, idempotencyKey uint64) error
}
