package orchestrator

import (
	"encoding/json"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	pb "github.com/minghsu0107/saga-pb"
	"github.com/minghsu0107/saga-product/domain/event"
	"github.com/minghsu0107/saga-product/domain/model"
	"github.com/minghsu0107/saga-product/pkg"
)

func decodeCreatePurchaseResponse(payload message.Payload) (*model.CreatePurchaseResponse, error) {
	var resp pb.CreatePurchaseResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		return nil, err
	}
	purchaseID := resp.PurchaseId
	pbPurchasedItems := resp.Purchase.Order.PurchasedItems
	var purchasedItems []model.PurchasedItem
	for _, pbPurchasedItem := range pbPurchasedItems {
		purchasedItems = append(purchasedItems, model.PurchasedItem{
			ProductID: pbPurchasedItem.ProductId,
			Amount:    pbPurchasedItem.Amount,
		})
	}

	return &model.CreatePurchaseResponse{
		Purchase: &model.Purchase{
			ID: purchaseID,
			Order: &model.Order{
				ID:             purchaseID,
				CustomerID:     resp.Purchase.Order.CustomerId,
				PurchasedItems: &purchasedItems,
			},
			Payment: &model.Payment{
				ID:           purchaseID,
				CurrencyCode: resp.Purchase.Payment.CurrencyCode,
				Amount:       resp.Purchase.Payment.Amount,
			},
		},
		Success: resp.Success,
		Error:   resp.Error,
	}, nil
}

func decodeRollbackResponse(payload message.Payload) (*model.RollbackResponse, error) {
	var resp pb.RollbackResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		return nil, err
	}
	return &model.RollbackResponse{
		CustomerID: resp.CustomerId,
		PurchaseID: resp.PurchaseId,
		Success:    resp.Success,
		Error:      resp.Error,
	}, nil
}

func encodeDomainPurchase(purchase *model.Purchase) *pb.CreatePurchaseCmd {
	var pbPurchasedItems []*pb.PurchasedItem
	for _, purchasedItem := range *purchase.Order.PurchasedItems {
		pbPurchasedItems = append(pbPurchasedItems, &pb.PurchasedItem{
			ProductId: purchasedItem.ProductID,
			Amount:    purchasedItem.Amount,
		})
	}
	cmd := &pb.CreatePurchaseCmd{
		PurchaseId: purchase.ID,
		Purchase: &pb.Purchase{
			Order: &pb.Order{
				CustomerId:     purchase.Order.CustomerID,
				PurchasedItems: pbPurchasedItems,
			},
			Payment: &pb.Payment{
				CurrencyCode: purchase.Payment.CurrencyCode,
				Amount:       purchase.Payment.Amount,
			},
		},
		Timestamp: pkg.Time2pbTimestamp(time.Now()),
	}
	return cmd
}

func encodeDomainPurchaseResult(purchaseResult *event.PurchaseResult) *pb.PurchaseResult {
	return &pb.PurchaseResult{
		CustomerId: purchaseResult.CustomerID,
		PurchaseId: purchaseResult.PurchaseID,
		Step:       getPbPurchaseStep(purchaseResult.Step),
		Status:     getPbPurchaseStatus(purchaseResult.Status),
		Timestamp:  pkg.Time2pbTimestamp(time.Now()),
	}
}

func getPbPurchaseStep(step string) pb.PurchaseStep {
	switch step {
	case event.StepUpdateProductInventory:
		return pb.PurchaseStep_STEP_UPDATE_PRODUCT_INVENTORY
	case event.StepCreateOrder:
		return pb.PurchaseStep_STEP_CREATE_ORDER
	case event.StepCreatePayment:
		return pb.PurchaseStep_STEP_CREATE_PAYMENT
	}
	return -1
}

func getPbPurchaseStatus(status string) pb.PurchaseStatus {
	switch status {
	case event.StatusExecute:
		return pb.PurchaseStatus_STATUS_EXUCUTE
	case event.StatusSucess:
		return pb.PurchaseStatus_STATUS_SUCCESS
	case event.StatusFailed:
		return pb.PurchaseStatus_STATUS_FAILED
	case event.StatusRollbacked:
		return pb.PurchaseStatus_STATUS_ROLLBACKED
	case event.StatusRollbackFailed:
		return pb.PurchaseStatus_STATUS_ROLLBACK_FAIL
	}
	return -1
}
