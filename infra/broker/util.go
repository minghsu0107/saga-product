package broker

import (
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill/message"
	pb "github.com/minghsu0107/saga-pb"
	"github.com/minghsu0107/saga-product/domain/model"
)

func DecodeCreatePurchaseCmd(payload message.Payload) (*model.Purchase, *pb.Purchase, error) {
	var cmd pb.CreatePurchaseCmd
	if err := json.Unmarshal(payload, &cmd); err != nil {
		return nil, nil, err
	}

	purchaseID := cmd.PurchaseId

	pbPurchasedItems := cmd.Purchase.Order.PurchasedItems
	var purchasedItems []model.PurchasedItem
	for _, pbPurchasedItem := range pbPurchasedItems {
		purchasedItems = append(purchasedItems, model.PurchasedItem{
			ProductID: pbPurchasedItem.ProductId,
			Amount:    pbPurchasedItem.Amount,
		})
	}
	return &model.Purchase{
		ID: purchaseID,
		Order: &model.Order{
			ID:             purchaseID,
			CustomerID:     cmd.Purchase.Order.CustomerId,
			PurchasedItems: &purchasedItems,
		},
		Payment: &model.Payment{
			ID:           purchaseID,
			CurrencyCode: cmd.Purchase.Payment.CurrencyCode,
			Amount:       cmd.Purchase.Payment.Amount,
		},
	}, cmd.Purchase, nil
}
