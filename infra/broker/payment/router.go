package payment

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
	"github.com/minghsu0107/saga-product/service/payment"
)

// SagaPaymentHandler handler
type SagaPaymentHandler struct {
	svc payment.SagaPaymentService
}

// CreatePayment handler
func (h *SagaPaymentHandler) CreatePayment(msg *message.Message) ([]*message.Message, error) {
	var cmd pb.CreatePaymentCmd
	if err := json.Unmarshal(msg.Payload, &cmd); err != nil {
		return nil, err
	}

	var reply pb.GeneralResponse
	err := h.svc.CreatePayment(context.Background(), createPaymentCmd2domainPayment(&cmd))
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
	replyMsg.Metadata.Set(conf.HandlerHeader, conf.CreatePaymentHandler)
	replyMsgs = append(replyMsgs, replyMsg)
	return replyMsgs, nil
}

func (h *SagaPaymentHandler) RollbackPayment(msg *message.Message) ([]*message.Message, error) {
	var cmd pb.RollbackPaymentCmd
	if err := json.Unmarshal(msg.Payload, &cmd); err != nil {
		return nil, err
	}

	var reply pb.GeneralResponse
	err := h.svc.RollbackPayment(context.Background(), cmd.PaymentId)
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
	replyMsg.Metadata.Set(conf.HandlerHeader, conf.RollbackPaymentHandler)
	replyMsgs = append(replyMsgs, replyMsg)
	return replyMsgs, nil
}

// PaymentEventRouter implementation
type PaymentEventRouter struct {
	router             *message.Router
	sagaPaymentHandler *SagaPaymentHandler
	subscriber         message.Subscriber
	publisher          message.Publisher
}

// NewPaymentEventRouter factory
func NewPaymentEventRouter(config *conf.Config, sagaPaymentSvc payment.SagaPaymentService, subscriber message.Subscriber, publisher message.Publisher) (broker.EventRouter, error) {
	router, err := broker.InitializeRouter(config.App)
	if err != nil {
		return nil, err
	}
	sagaPaymentHandler := SagaPaymentHandler{
		svc: sagaPaymentSvc,
	}
	return &PaymentEventRouter{
		router:             router,
		sagaPaymentHandler: &sagaPaymentHandler,
		subscriber:         subscriber,
		publisher:          publisher,
	}, nil
}

func (r *PaymentEventRouter) RegisterHandlers() {
	r.router.AddHandler(
		"sagapayment_create_payment_handler",
		conf.PaymentTopic,
		r.subscriber,
		conf.ReplyTopic,
		r.publisher,
		r.sagaPaymentHandler.CreatePayment,
	)
	r.router.AddHandler(
		"sagapayment_rollback_payment_handler",
		conf.PaymentTopic,
		r.subscriber,
		conf.ReplyTopic,
		r.publisher,
		r.sagaPaymentHandler.RollbackPayment,
	)
}

func (r *PaymentEventRouter) Run() error {
	r.RegisterHandlers()
	return r.router.Run(context.Background())
}

func (r *PaymentEventRouter) GracefulStop() error {
	return r.router.Close()
}

func createPaymentCmd2domainPayment(cmd *pb.CreatePaymentCmd) *model.Payment {
	var amount int64 = 0
	for _, pbPurchasedItem := range cmd.Purchase.Order.PurchasedItems {
		amount += pbPurchasedItem.Amount
	}
	return &model.Payment{
		ID:           cmd.PaymentId,
		CurrencyCode: cmd.Purchase.Payment.CurrencyCode,
		Amount:       amount,
	}
}
