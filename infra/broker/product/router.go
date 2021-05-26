package product

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	pb "github.com/minghsu0107/saga-pb"
	conf "github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/domain/model"
	"github.com/minghsu0107/saga-product/infra/broker"
	"github.com/minghsu0107/saga-product/pkg"
	"github.com/minghsu0107/saga-product/service/product"
)

// SagaProductHandler handler
type SagaProductHandler struct {
	svc product.SagaProductService
}

// UpdateProductInventory handler
func (h *SagaProductHandler) UpdateProductInventory(msg *message.Message) ([]*message.Message, error) {
	var cmd pb.UpdateProductInventoryCmd
	if err := json.Unmarshal(msg.Payload, &cmd); err != nil {
		return nil, err
	}
	idempotencyKey, err := strconv.ParseUint(msg.Metadata.Get(conf.IdempotencyKeyHeader), 10, 64)
	if err != nil {
		return nil, err
	}
	pbPurchasedItems := cmd.PurchasedItems
	var purchasedItems []model.PurchasedItem
	for _, pbPurchasedItem := range pbPurchasedItems {
		purchasedItems = append(purchasedItems, model.PurchasedItem{
			ProductID: pbPurchasedItem.ProductId,
			Amount:    pbPurchasedItem.Amount,
		})
	}

	var reply pb.GeneralResponse
	err = h.svc.UpdateProductInventory(context.Background(), idempotencyKey, &purchasedItems)
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
	replyMsg.Metadata.Set(conf.HandlerHeader, conf.UpdateProductInventoryHandler)
	replyMsgs = append(replyMsgs, replyMsg)
	return replyMsgs, nil
}

// RollbackProductInventory handler
func (h *SagaProductHandler) RollbackProductInventory(msg *message.Message) ([]*message.Message, error) {
	var cmd pb.RollbackProductInventoryCmd
	if err := json.Unmarshal(msg.Payload, &cmd); err != nil {
		return nil, err
	}
	idempotencyKey, err := strconv.ParseUint(msg.Metadata.Get(conf.IdempotencyKeyHeader), 10, 64)
	if err != nil {
		return nil, err
	}

	var reply pb.GeneralResponse
	err = h.svc.RollbackProductInventory(context.Background(), idempotencyKey)
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
	replyTopic         string
	productTopic       string
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
		replyTopic:         conf.ReplyTopic,
		productTopic:       conf.ProductTopic,
	}, nil
}

func (r *ProductEventRouter) RegisterHandlers() {
	r.router.AddHandler(
		"sagaproduct_update_product_inventory_handler",
		r.productTopic,
		r.subscriber,
		r.replyTopic,
		r.publisher,
		r.sagaProductHandler.UpdateProductInventory,
	)
	r.router.AddHandler(
		"sagaproduct_rollback_product_inventory_handler",
		r.productTopic,
		r.subscriber,
		r.replyTopic,
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
