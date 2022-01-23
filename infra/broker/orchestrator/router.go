package orchestrator

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	conf "github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/infra/broker"
	"github.com/minghsu0107/saga-product/service/orchestrator"
	"go.opencensus.io/trace/propagation"
)

// OrchestratorHandler handler
type OrchestratorHandler struct {
	svc orchestrator.OrchestratorService
}

// StartTransaction starts the transaction
func (h *OrchestratorHandler) StartTransaction(msg *message.Message) error {
	purchase, _, err := broker.DecodeCreatePurchaseCmd(msg.Payload)
	if err != nil {
		return err
	}
	correlationID := msg.Metadata.Get(middleware.CorrelationIDMetadataKey)
	sc, _ := propagation.FromBinary([]byte(msg.Metadata.Get(conf.SpanContextKey)))
	return h.svc.StartTransaction(sc, purchase, correlationID)
}

func (h *OrchestratorHandler) HandleReply(msg *message.Message) error {
	correlationID := msg.Metadata.Get(middleware.CorrelationIDMetadataKey)
	sc, _ := propagation.FromBinary([]byte(msg.Metadata.Get(conf.SpanContextKey)))
	return h.svc.HandleReply(sc, msg, correlationID)
}

// OrchestratorEventRouter implementation
type OrchestratorEventRouter struct {
	router              *message.Router
	orchestratorHandler *OrchestratorHandler
	txSubscriber        broker.NATSSubscriber
}

// NewOrchestratorEventRouter factory
func NewOrchestratorEventRouter(config *conf.Config, orchestratorSvc orchestrator.OrchestratorService, txSubscriber broker.NATSSubscriber) (broker.EventRouter, error) {
	router, err := broker.InitializeRouter(config.App)
	if err != nil {
		return nil, err
	}
	orchestratorHandler := OrchestratorHandler{
		svc: orchestratorSvc,
	}
	return &OrchestratorEventRouter{
		router:              router,
		orchestratorHandler: &orchestratorHandler,
		txSubscriber:        txSubscriber,
	}, nil
}

func (r *OrchestratorEventRouter) RegisterHandlers() {
	r.router.AddNoPublisherHandler(
		"sagaorchestrator_start_transaction_handler",
		conf.PurchaseTopic,
		r.txSubscriber,
		r.orchestratorHandler.StartTransaction,
	)
	r.router.AddNoPublisherHandler(
		"sagaorchestrator_handle_reply_handler",
		conf.ReplyTopic,
		r.txSubscriber,
		r.orchestratorHandler.HandleReply,
	)
}

func (r *OrchestratorEventRouter) Run() error {
	r.RegisterHandlers()
	return r.router.Run(context.Background())
}

func (r *OrchestratorEventRouter) GracefulStop() error {
	return r.router.Close()
}
