package orchestrator

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	conf "github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/infra/broker"
	"github.com/minghsu0107/saga-product/service/orchestrator"
	"go.opencensus.io/trace"
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
	var sc trace.SpanContext
	if err := json.Unmarshal([]byte(msg.Metadata.Get(conf.SpanContextKey)), &sc); err != nil {
		return err
	}
	return h.svc.StartTransaction(sc, purchase, correlationID)
}

func (h *OrchestratorHandler) HandleReply(msg *message.Message) error {
	correlationID := msg.Metadata.Get(middleware.CorrelationIDMetadataKey)
	var sc trace.SpanContext
	json.Unmarshal([]byte(msg.Metadata.Get(conf.SpanContextKey)), &sc)
	if err := json.Unmarshal([]byte(msg.Metadata.Get(conf.SpanContextKey)), &sc); err != nil {
		return err
	}
	return h.svc.HandleReply(sc, msg, correlationID)
}

// OrchestratorEventRouter implementation
type OrchestratorEventRouter struct {
	router              *message.Router
	orchestratorHandler *OrchestratorHandler
	subscriber          message.Subscriber
}

// NewOrchestratorEventRouter factory
func NewOrchestratorEventRouter(config *conf.Config, orchestratorSvc orchestrator.OrchestratorService, subscriber message.Subscriber) (broker.EventRouter, error) {
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
		subscriber:          subscriber,
	}, nil
}

func (r *OrchestratorEventRouter) RegisterHandlers() {
	r.router.AddNoPublisherHandler(
		"sagaorchestrator_start_transaction_handler",
		conf.PurchaseTopic,
		r.subscriber,
		r.orchestratorHandler.StartTransaction,
	)
	r.router.AddNoPublisherHandler(
		"sagaorchestrator_handle_reply_handler",
		conf.ReplyTopic,
		r.subscriber,
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
