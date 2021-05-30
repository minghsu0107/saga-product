package order

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
	"github.com/minghsu0107/saga-product/service/order"
	"go.opencensus.io/trace"
)

// SagaOrderHandler handler
type SagaOrderHandler struct {
	svc order.SagaOrderService
}

// CreateOrder handler
func (h *SagaOrderHandler) CreateOrder(msg *message.Message) ([]*message.Message, error) {
	childCtx, span := trace.StartSpan(msg.Context(), "event.CreateOrder")
	defer span.End()

	purchase, pbPurchase, err := broker.DecodeCreatePurchaseCmd(msg.Payload)
	if err != nil {
		return nil, err
	}
	reply := pb.CreatePurchaseResponse{
		PurchaseId: purchase.ID,
		Purchase:   pbPurchase,
	}
	err = h.svc.CreateOrder(context.Background(), purchase.Order)
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
	replyMsg.SetContext(childCtx)
	replyMsg.Metadata.Set(conf.HandlerHeader, conf.CreateOrderHandler)
	replyMsgs = append(replyMsgs, replyMsg)
	return replyMsgs, nil
}

func (h *SagaOrderHandler) RollbackOrder(msg *message.Message) ([]*message.Message, error) {
	childCtx, span := trace.StartSpan(msg.Context(), "event.RollbackOrder")
	defer span.End()

	var cmd pb.RollbackCmd
	if err := json.Unmarshal(msg.Payload, &cmd); err != nil {
		return nil, err
	}

	reply := pb.RollbackResponse{
		CustomerId: cmd.CustomerId,
		PurchaseId: cmd.PurchaseId,
	}
	err := h.svc.RollbackOrder(context.Background(), cmd.PurchaseId)
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
	replyMsg.SetContext(childCtx)
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
		conf.CreateOrderTopic,
		r.subscriber,
		conf.ReplyTopic,
		r.publisher,
		r.sagaOrderHandler.CreateOrder,
	)
	r.router.AddHandler(
		"sagaorder_rollback_order_handler",
		conf.RollbackOrderTopic,
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
