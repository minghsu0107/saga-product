package order

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	pb "github.com/minghsu0107/saga-pb"
	conf "github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/domain/model"
	"github.com/minghsu0107/saga-product/infra/broker"
	"github.com/minghsu0107/saga-product/pkg"
	"github.com/minghsu0107/saga-product/service/order"
)

// SagaOrderHandler handler
type SagaOrderHandler struct {
	svc order.SagaOrderService
}

// CreateOrder handler
func (h *SagaOrderHandler) CreateOrder(msg *message.Message) ([]*message.Message, error) {
	var cmd pb.CreateOrderCmd
	if err := json.Unmarshal(msg.Payload, &cmd); err != nil {
		return nil, err
	}
	var purchasedItems []model.PurchasedItem
	for _, pbPurchasedItem := range cmd.Order.PurchasedItems {
		purchasedItems = append(purchasedItems, model.PurchasedItem{
			ProductID: pbPurchasedItem.ProductId,
			Amount:    pbPurchasedItem.Amount,
		})
	}
	order := model.Order{
		ID:             cmd.OrderId,
		CustomerID:     cmd.Order.CustomerId,
		PurchasedItems: &purchasedItems,
	}
	var reply pb.GeneralResponse
	err := h.svc.CreateOrder(context.Background(), &order)
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
	replyMsg.Metadata.Set(conf.HandlerHeader, conf.CreateOrderHandler)
	replyMsgs = append(replyMsgs, replyMsg)
	return replyMsgs, nil
}

func (h *SagaOrderHandler) RollbackOrder(msg *message.Message) ([]*message.Message, error) {
	var cmd pb.RollbackOrderCmd
	if err := json.Unmarshal(msg.Payload, &cmd); err != nil {
		return nil, err
	}

	var reply pb.GeneralResponse
	err := h.svc.RollbackOrder(context.Background(), cmd.OrderId)
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
	replyMsg.Metadata.Set(conf.HandlerHeader, conf.RollbackOrderHandler)
	replyMsgs = append(replyMsgs, replyMsg)
	return replyMsgs, nil
}

// OrderEventRouter implementation
type OrderEventRouter struct {
	router           *message.Router
	sagaOrderHandler *SagaOrderHandler
	subscriber       message.Subscriber
	publisher        message.Publisher
}

// NewOrderEventRouter factory
func NewOrderEventRouter(config *conf.Config, sagaOrderSvc order.SagaOrderService, subscriber message.Subscriber, publisher message.Publisher) (broker.EventRouter, error) {
	router, err := broker.InitializeRouter(config.App)
	if err != nil {
		return nil, err
	}
	sagaOrderHandler := SagaOrderHandler{
		svc: sagaOrderSvc,
	}
	return &OrderEventRouter{
		router:           router,
		sagaOrderHandler: &sagaOrderHandler,
		subscriber:       subscriber,
		publisher:        publisher,
	}, nil
}

func (r *OrderEventRouter) RegisterHandlers() {
	r.router.AddHandler(
		"sagaorder_create_order_handler",
		conf.OrderTopic,
		r.subscriber,
		conf.ReplyTopic,
		r.publisher,
		r.sagaOrderHandler.CreateOrder,
	)
	r.router.AddHandler(
		"sagaorder_rollback_order_handler",
		conf.OrderTopic,
		r.subscriber,
		conf.ReplyTopic,
		r.publisher,
		r.sagaOrderHandler.RollbackOrder,
	)
}

func (r *OrderEventRouter) Run() error {
	r.RegisterHandlers()
	return r.router.Run(context.Background())
}

func (r *OrderEventRouter) GracefulStop() error {
	return r.router.Close()
}
