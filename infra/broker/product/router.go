package product

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	pb "github.com/minghsu0107/saga-pb"
	conf "github.com/minghsu0107/saga-product/config"

	"github.com/minghsu0107/saga-product/infra/broker"
	"github.com/minghsu0107/saga-product/pkg"
	"github.com/minghsu0107/saga-product/service/product"
	"go.opencensus.io/trace"
)

// SagaProductHandler handler
type SagaProductHandler struct {
	svc product.SagaProductService
}

// UpdateProductInventory handler
func (h *SagaProductHandler) UpdateProductInventory(msg *message.Message) ([]*message.Message, error) {
	var sc trace.SpanContext
	if err := json.Unmarshal([]byte(msg.Metadata.Get(conf.SpanContextKey)), &sc); err != nil {
		return nil, err
	}
	_, span := trace.StartSpanWithRemoteParent(context.Background(), "event.UpdateProductInventory", sc)
	defer span.End()

	purchase, pbPurchase, err := broker.DecodeCreatePurchaseCmd(msg.Payload)
	if err != nil {
		return nil, err
	}
	reply := pb.CreatePurchaseResponse{
		PurchaseId: purchase.ID,
		Purchase:   pbPurchase,
	}
	err = h.svc.UpdateProductInventory(context.Background(), purchase.ID, purchase.Order.PurchasedItems)
	if err != nil {
		reply.Success = false
		reply.Error = err.Error()
	} else {
		reply.Success = true
		reply.Error = ""
	}
	reply.Timestamp = pkg.Time2pbTimestamp(time.Now())

	payload, err := json.Marshal(&reply)
	if err != nil {
		return nil, err
	}
	var replyMsgs []*message.Message
	replyMsg := message.NewMessage(watermill.NewUUID(), payload)
	if err := broker.SetSpanContext(replyMsg, span); err != nil {
		return nil, err
	}
	replyMsg.Metadata.Set(conf.HandlerHeader, conf.UpdateProductInventoryHandler)
	replyMsgs = append(replyMsgs, replyMsg)
	return replyMsgs, nil
}

// RollbackProductInventory handler
func (h *SagaProductHandler) RollbackProductInventory(msg *message.Message) ([]*message.Message, error) {
	var sc trace.SpanContext
	if err := json.Unmarshal([]byte(msg.Metadata.Get(conf.SpanContextKey)), &sc); err != nil {
		return nil, err
	}
	_, span := trace.StartSpanWithRemoteParent(context.Background(), "event.RollbackProductInventory", sc)
	defer span.End()

	var cmd pb.RollbackCmd
	if err := json.Unmarshal(msg.Payload, &cmd); err != nil {
		return nil, err
	}

	reply := pb.RollbackResponse{
		CustomerId: cmd.CustomerId,
		PurchaseId: cmd.PurchaseId,
	}
	err := h.svc.RollbackProductInventory(context.Background(), cmd.PurchaseId)
	if err != nil {
		reply.Success = false
		reply.Error = err.Error()
	} else {
		reply.Success = true
		reply.Error = ""
	}
	reply.Timestamp = pkg.Time2pbTimestamp(time.Now())

	payload, err := json.Marshal(&reply)
	if err != nil {
		return nil, err
	}
	var replyMsgs []*message.Message
	replyMsg := message.NewMessage(watermill.NewUUID(), payload)
	if err := broker.SetSpanContext(replyMsg, span); err != nil {
		return nil, err
	}
	replyMsg.Metadata.Set(conf.HandlerHeader, conf.RollbackProductInventoryHandler)
	replyMsgs = append(replyMsgs, replyMsg)
	return replyMsgs, nil
}

// ProductEventRouter implementation
type ProductEventRouter struct {
	router             *message.Router
	sagaProductHandler *SagaProductHandler
	subscriber         message.Subscriber
	publisher          message.Publisher
}

// NewProductEventRouter factory
func NewProductEventRouter(config *conf.Config, sagaProductSvc product.SagaProductService, subscriber message.Subscriber, publisher message.Publisher) (broker.EventRouter, error) {
	router, err := broker.InitializeRouter(config.App)
	if err != nil {
		return nil, err
	}
	sagaProductHandler := SagaProductHandler{
		svc: sagaProductSvc,
	}
	return &ProductEventRouter{
		router:             router,
		sagaProductHandler: &sagaProductHandler,
		subscriber:         subscriber,
		publisher:          publisher,
	}, nil
}

func (r *ProductEventRouter) RegisterHandlers() {
	r.router.AddHandler(
		"sagaproduct_update_product_inventory_handler",
		conf.UpdateProductInventoryTopic,
		r.subscriber,
		conf.ReplyTopic,
		r.publisher,
		r.sagaProductHandler.UpdateProductInventory,
	)
	r.router.AddHandler(
		"sagaproduct_rollback_product_inventory_handler",
		conf.RollbackProductInventoryTopic,
		r.subscriber,
		conf.ReplyTopic,
		r.publisher,
		r.sagaProductHandler.RollbackProductInventory,
	)
}

func (r *ProductEventRouter) Run() error {
	r.RegisterHandlers()
	return r.router.Run(context.Background())
}

func (r *ProductEventRouter) GracefulStop() error {
	return r.router.Close()
}
