package product

import (
	"context"

	"github.com/minghsu0107/saga-product/domain/model"
)

type ProductService interface {
	CheckProducts(ctx context.Context, cartItems *[]model.CartItem) (*[]model.ProductStatus, error)
	ListProducts(ctx context.Context) (*[]model.ProductCatalog, error)
	GetProductDetails(ctx context.Context, productIDs []uint64) (*[]model.Product, error)
}

type SagaProductService interface {
	UpdateProductInventory(ctx context.Context, idempotencyKey string, purchasedItems *[]model.PurchasedItem) error
	RollbackProductInventory(ctx context.Context, idempotencyKey string) error
}
