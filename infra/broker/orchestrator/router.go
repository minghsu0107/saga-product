package orchestrator

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill/message"
	pb "github.com/minghsu0107/saga-pb"
	conf "github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/domain/model"
	"github.com/minghsu0107/saga-product/infra/broker"
	"github.com/minghsu0107/saga-product/service/orchestrator"
)

// OrchestratorHandler handler
type OrchestratorHandler struct {
	svc        orchestrator.OrchestratorService
	subscriber message.Subscriber
}

func (h *OrchestratorHandler) HandlePurchase(msg *message.Message) error {
	var cmd pb.CreatePurchase
	if err := json.Unmarshal(msg.Payload, &cmd); err != nil {
		return err
	}
	return h.svc.PlaySaga(context.Background(), createPurchaseCmd2domainPurchase(&cmd), h.subscriber)
}

// OrchestratorEventRouter implementation
type OrchestratorEventRouter struct {
	router              *message.Router
	orchestratorHandler *OrchestratorHandler
	subscriber          message.Subscriber
}

// NewOrchestratorEventRouter factory
func NewOrchestratorEventRouter(config *conf.Config, orchestratorSvc orchestrator.OrchestratorService, subscriber message.Subscriber, publisher message.Publisher) (broker.EventRouter, error) {
	router, err := broker.InitializeRouter(config.App)
	if err != nil {
		return nil, err
	}
	orchestratorHandler := OrchestratorHandler{
		svc:        orchestratorSvc,
		subscriber: subscriber,
	}
	return &OrchestratorEventRouter{
		router:              router,
		orchestratorHandler: &orchestratorHandler,
		subscriber:          subscriber,
	}, nil
}

func (r *OrchestratorEventRouter) RegisterHandlers() {
	r.router.AddNoPublisherHandler(
		"sagaorchestrator_purchase_handler",
		conf.PurchaseTopic,
		r.subscriber,
		r.orchestratorHandler.HandlePurchase,
	)
}

func (r *OrchestratorEventRouter) Run() error {
	r.RegisterHandlers()
	return r.router.Run(context.Background())
}

func (r *OrchestratorEventRouter) GracefulStop() error {
	return r.router.Close()
}

func createPurchaseCmd2domainPurchase(cmd *pb.CreatePurchase) *model.Purchase {
	var amount int64 = 0
	var purchasedItems []model.PurchasedItem
	for _, pbPurchasedItem := range cmd.Purchase.Order.PurchasedItems {
		purchasedItems = append(purchasedItems, model.PurchasedItem{
			ProductID: pbPurchasedItem.ProductId,
			Amount:    pbPurchasedItem.Amount,
		})
		amount += pbPurchasedItem.Amount
	}
	order := model.Order{
		CustomerID:     cmd.Purchase.Order.CustomerId,
		PurchasedItems: &purchasedItems,
	}
	payment := model.Payment{
		CurrencyCode: cmd.Purchase.Payment.CurrencyCode,
		Amount:       amount,
	}
	return &model.Purchase{
		Order:   &order,
		Payment: &payment,
	}
}
