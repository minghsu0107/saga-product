package product

import (
	"context"
	"fmt"

	pb "github.com/minghsu0107/saga-pb"
	"github.com/minghsu0107/saga-product/domain/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (srv *ProductServer) CheckProducts(ctx context.Context, req *pb.CheckProductsRequest) (*pb.CheckProductsResponse, error) {
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
			Price:     status.Price,
			Status:    getPbProductStatus(status.Status),
		})
	}
	return &pb.CheckProductsResponse{
		ProductStatuses: pbProductStatuses,
	}, nil
}

func (srv *ProductServer) GetProducts(ctx context.Context, req *pb.GetProductsRequest) (*pb.Products, error) {
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

func getPbProductStatus(status model.Status) pb.Status {
	switch status {
	case model.ProductOk:
		return pb.Status_STATUS_OK
	case model.ProductNotFound:
		return pb.Status_STATUS_NOT_FOUND
	}
	return pb.Status_STATUS_NOT_FOUND
}
