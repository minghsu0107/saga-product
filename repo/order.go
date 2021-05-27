package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/ratelimit"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	pb "github.com/minghsu0107/saga-pb"
	conf "github.com/minghsu0107/saga-product/config"
	domain_model "github.com/minghsu0107/saga-product/domain/model"
	"github.com/minghsu0107/saga-product/infra/db/model"
	grpc_order "github.com/minghsu0107/saga-product/infra/grpc/order"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

// OrderRepository interface
type OrderRepository interface {
	GetOrder(ctx context.Context, orderID uint64) (*domain_model.Order, error)
	GetDetailedPurchasedItems(ctx context.Context, purchasedItems *[]domain_model.PurchasedItem) (*[]domain_model.DetailedPurchasedItem, error)
	ExistOrder(ctx context.Context, orderID uint64) (bool, error)
	CreateOrder(ctx context.Context, order *domain_model.Order) error
	DeleteOrder(ctx context.Context, orderID uint64) error
}

// OrderRepositoryImpl implementation
type OrderRepositoryImpl struct {
	db          *gorm.DB
	getProducts endpoint.Endpoint
}

// NewOrderRepository factory
func NewOrderRepository(config *conf.Config, conn *grpc_order.ProductConn, db *gorm.DB) OrderRepository {
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), config.ServiceOptions.Rps))

	var options []grpctransport.ClientOption

	var getProducts endpoint.Endpoint
	{
		svcName := "product.ProductService"
		getProducts = grpctransport.NewClient(
			conn.Conn,
			svcName,
			"GetProducts",
			encodeGRPCRequest,
			decodeGRPCResponse,
			&pb.CheckProductsResponse{},
			append(options, grpctransport.ClientBefore(grpctransport.SetRequestHeader(ServiceNameHeader, svcName)))...,
		).Endpoint()
		getProducts = limiter(getProducts)
		getProducts = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "product",
			Timeout: config.ServiceOptions.Timeout,
		}))(getProducts)
	}
	return &OrderRepositoryImpl{
		db:          db,
		getProducts: getProducts,
	}
}

// GetOrder get an order
func (repo *OrderRepositoryImpl) GetOrder(ctx context.Context, orderID uint64) (*domain_model.Order, error) {
	var orders []model.Order
	if err := repo.db.Model(&model.Order{}).Select("id", "product_id", "amount", "customer_id").Where("id = ?", orderID).Order("product_id").Find(&orders).WithContext(ctx).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("order not found; order ID: %v", orderID)
		}
		return nil, err
	}
	if len(orders) == 0 {
		return nil, fmt.Errorf("order not found; order ID: %v", orderID)
	}
	var purchasedItems []domain_model.PurchasedItem
	for _, order := range orders {
		purchasedItems = append(purchasedItems, domain_model.PurchasedItem{
			ProductID: order.ProductID,
			Amount:    order.Amount,
		})
	}
	return &domain_model.Order{
		ID:             orders[0].ID,
		CustomerID:     orders[0].CustomerID,
		PurchasedItems: &purchasedItems,
	}, nil
}

// GetDetailedPurchasedItems get detailed purchased items
func (repo *OrderRepositoryImpl) GetDetailedPurchasedItems(ctx context.Context, purchasedItems *[]domain_model.PurchasedItem) (*[]domain_model.DetailedPurchasedItem, error) {
	var productIDs []uint64
	for _, purchasedItem := range *purchasedItems {
		productIDs = append(productIDs, purchasedItem.ProductID)
	}
	res, err := repo.getProducts(ctx, &pb.GetProductsRequest{
		ProductId: productIDs,
	})
	if err != nil {
		return nil, err
	}
	pbProducts := res.(*pb.Products)
	var detailedPurchasedItems []domain_model.DetailedPurchasedItem
	for i, pbProduct := range pbProducts.Products {
		detailedPurchasedItems = append(detailedPurchasedItems, domain_model.DetailedPurchasedItem{
			ProductID:   pbProduct.ProductId,
			Name:        pbProduct.ProductName,
			Description: pbProduct.Description,
			BrandName:   pbProduct.BrandName,
			Price:       pbProduct.Price,
			Amount:      (*purchasedItems)[i].Amount,
		})
	}
	return &detailedPurchasedItems, nil
}

// ExistOrder checks whether an order exists
func (repo *OrderRepositoryImpl) ExistOrder(ctx context.Context, orderID uint64) (bool, error) {
	var order model.Order
	if err := repo.db.Model(&model.Order{}).Select("id").Where("id = ?", orderID).First(&order).WithContext(ctx).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CreateOrder creates an order
func (repo *OrderRepositoryImpl) CreateOrder(ctx context.Context, order *domain_model.Order) error {
	id := order.ID
	customerID := order.CustomerID
	var entries []model.Order
	for _, purchasedItem := range *order.PurchasedItems {
		entries = append(entries, model.Order{
			ID:         id,
			ProductID:  purchasedItem.ProductID,
			Amount:     purchasedItem.Amount,
			CustomerID: customerID,
		})
	}
	if err := repo.db.Create(&entries).WithContext(ctx).Error; err != nil {
		return err
	}
	return nil
}

// DeleteOrder deletes an order
func (repo *OrderRepositoryImpl) DeleteOrder(ctx context.Context, orderID uint64) error {
	if err := repo.db.Exec("DELETE FROM orders where id = ?", orderID).Error; err != nil {
		return err
	}
	return nil
}
