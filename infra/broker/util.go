package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	pb "github.com/minghsu0107/saga-pb"
	conf "github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/domain/model"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	TraceContext        = propagation.TraceContext{}
	TraceparentHeader   = TraceContext.Fields()[0]
	W3CSupportedVersion = 0
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
			CustomerID:   cmd.Purchase.Order.CustomerId,
			CurrencyCode: cmd.Purchase.Payment.CurrencyCode,
			Amount:       cmd.Purchase.Payment.Amount,
		},
	}, cmd.Purchase, nil
}

// SetSpanContext set span context to the message
func SetSpanContext(ctx context.Context, msg *message.Message) {
	msg.Metadata.Set(conf.SpanContextKey, spanContextToW3C(ctx))
}

func spanContextToW3C(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return ""
	}
	// Clear all flags other than the trace-context supported sampling bit.
	flags := sc.TraceFlags() & trace.FlagsSampled
	return fmt.Sprintf("%.2x-%s-%s-%s",
		W3CSupportedVersion,
		sc.TraceID(),
		sc.SpanID(),
		flags)
}
