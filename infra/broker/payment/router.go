package payment

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
	"github.com/minghsu0107/saga-product/service/payment"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

// SagaPaymentHandler handler
type SagaPaymentHandler struct {
	svc payment.SagaPaymentService
}

// CreatePayment handler
func (h *SagaPaymentHandler) CreatePayment(msg *message.Message) ([]*message.Message, error) {
	sc, _ := propagation.FromBinary([]byte(msg.Metadata.Get(conf.SpanContextKey)))
	_, span := trace.StartSpanWithRemoteParent(context.Background(), "event.CreatePayment", sc)
	defer span.End()

	purchase, pbPurchase, err := broker.DecodeCreatePurchaseCmd(msg.Payload)
	if err != nil {
		return nil, err
	}
	reply := pb.CreatePurchaseResponse{
		PurchaseId: purchase.ID,
		Purchase:   pbPurchase,
	}
	err = h.svc.CreatePayment(context.Background(), purchase.Payment)
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
	broker.SetSpanContext(replyMsg, span)
	replyMsg.Metadata.Set(conf.HandlerHeader, conf.CreatePaymentHandler)
	replyMsgs = append(replyMsgs, replyMsg)
	return replyMsgs, nil
}

func (h *SagaPaymentHandler) RollbackPayment(msg *message.Message) ([]*message.Message, error) {
	sc, _ := propagation.FromBinary([]byte(msg.Metadata.Get(conf.SpanContextKey)))
	_, span := trace.StartSpanWithRemoteParent(context.Background(), "event.RollbackPayment", sc)
	defer span.End()

	var cmd pb.RollbackCmd
	if err := json.Unmarshal(msg.Payload, &cmd); err != nil {
		return nil, err
	}

	reply := pb.RollbackResponse{
		CustomerId: cmd.CustomerId,
		PurchaseId: cmd.PurchaseId,
	}
	err := h.svc.RollbackPayment(context.Background(), cmd.PurchaseId)
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
	broker.SetSpanContext(replyMsg, span)
	replyMsg.Metadata.Set(conf.HandlerHeader, conf.RollbackPaymentHandler)
	replyMsgs = append(replyMsgs, replyMsg)
	return replyMsgs, nil
}

// PaymentEventRouter implementation
type PaymentEventRouter struct {
	router             *message.Router
	sagaPaymentHandler *SagaPaymentHandler
	txSubscriber       broker.NATSSubscriber
	txPublisher        broker.NATSPublisher
}

// NewPaymentEventRouter factory
func NewPaymentEventRouter(config *conf.Config, sagaPaymentSvc payment.SagaPaymentService, txSubscriber broker.NATSSubscriber, txPublisher broker.NATSPublisher) (broker.EventRouter, error) {
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
		txSubscriber:       txSubscriber,
		txPublisher:        txPublisher,
	}, nil
}

func (r *PaymentEventRouter) RegisterHandlers() {
	r.router.AddHandler(
		"sagapayment_create_payment_handler",
		conf.CreatePaymentTopic,
		r.txSubscriber,
		conf.ReplyTopic,
		r.txPublisher,
		r.sagaPaymentHandler.CreatePayment,
	)
	r.router.AddHandler(
		"sagapayment_rollback_payment_handler",
		conf.RollbackPaymentTopic,
		r.txSubscriber,
		conf.ReplyTopic,
		r.txPublisher,
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
