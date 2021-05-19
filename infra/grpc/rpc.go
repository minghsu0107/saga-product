package grpc

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	pb "github.com/minghsu0107/saga-pb"
	"github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/domain/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (srv *Server) CheckProducts(ctx context.Context, req *pb.CheckProductsRequest) (*pb.CheckProductsResponse, error) {
	var cartItems []model.CartItem
	pbCartItems := req.CartItems
	for _, pbCartItem := range pbCartItems {
		cartItems = append(cartItems, model.CartItem{
			ProductID: pbCartItem.ProductId,
			Amount:    pbCartItem.Amount,
		})
	}
	productStatuses, err := srv.productSvc.CheckProducts(ctx, &cartItems)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("internal error: %v", err),
		)
	}
	var pbProductStatuses []*pb.ProductStatus
	for _, status := range *productStatuses {
		pbProductStatuses = append(pbProductStatuses, &pb.ProductStatus{
			ProductId: status.ProductID,
			Status:    getPbProductStatus(status.Status),
		})
	}
	return &pb.CheckProductsResponse{
		ProductStatuses: pbProductStatuses,
	}, nil
}

func (srv *Server) GetProducts(ctx context.Context, req *pb.GetProductsRequest) (*pb.Products, error) {
	productIDs := req.ProductId
	products, err := srv.productSvc.GetProducts(ctx, productIDs)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("internal error: %v", err),
		)
	}
	var pbProducts []*pb.Product
	for _, product := range *products {
		pbProducts = append(pbProducts, &pb.Product{
			ProductId:   product.ID,
			ProductName: product.Detail.Name,
			Description: product.Detail.Description,
			BrandName:   product.Detail.BrandName,
			Inventory:   product.Inventory,
			Price:       product.Detail.Price,
		})
	}
	return &pb.Products{
		Products: pbProducts,
	}, nil
}

func (srv *Server) UpdateProductInventory(ctx context.Context, req *pb.UpdateProductInventoryCmd) (*pb.GeneralResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md.Get(config.IdempotencyKeyHeader)) == 0 {
		return nil, status.Error(codes.Internal, "error parsing metadata")
	}
	idempotencyKey, err := strconv.ParseUint(md.Get(config.IdempotencyKeyHeader)[0], 10, 64)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("internal error: %v", err),
		)
	}

	pbPurchasedItems := req.PurchasedItems
	var purchasedItems []model.PurchasedItem
	for _, pbPurchasedItem := range pbPurchasedItems {
		purchasedItems = append(purchasedItems, model.PurchasedItem{
			ProductID: pbPurchasedItem.ProductId,
			Amount:    pbPurchasedItem.Amount,
		})
	}

	err = srv.sagaProductSvc.UpdateProductInventory(ctx, idempotencyKey, &purchasedItems)
	if err != nil {
		return &pb.GeneralResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time2pbTimestamp(time.Now()),
		}, nil
	}
	return &pb.GeneralResponse{
		Success:   true,
		Error:     "",
		Timestamp: time2pbTimestamp(time.Now()),
	}, nil
}

func (srv *Server) RollbackProductInventory(ctx context.Context, req *pb.RollbackProductInventoryCmd) (*pb.GeneralResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md.Get(config.IdempotencyKeyHeader)) == 0 {
		return nil, status.Error(codes.Internal, "error parsing metadata")
	}
	idempotencyKey, err := strconv.ParseUint(md.Get(config.IdempotencyKeyHeader)[0], 10, 64)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("internal error: %v", err),
		)
	}

	err = srv.sagaProductSvc.RollbackProductInventory(ctx, idempotencyKey)
	if err != nil {
		return &pb.GeneralResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time2pbTimestamp(time.Now()),
		}, nil
	}
	return &pb.GeneralResponse{
		Success:   true,
		Error:     "",
		Timestamp: time2pbTimestamp(time.Now()),
	}, nil
}

func getPbProductStatus(status model.Status) pb.Status {
	switch status {
	case model.ProductOk:
		return pb.Status_STATUS_OK
	case model.ProductNotFound:
		return pb.Status_STATUS_NOT_FOUND
	}
	return pb.Status_STATUS_NOT_FOUND
}

func time2pbTimestamp(now time.Time) *timestamp.Timestamp {
	s := int64(now.Second())
	n := int32(now.Nanosecond())
	return &timestamp.Timestamp{Seconds: s, Nanos: n}
}
