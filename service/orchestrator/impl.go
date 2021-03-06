package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	pb "github.com/minghsu0107/saga-pb"
	conf "github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/domain/event"
	"github.com/minghsu0107/saga-product/domain/model"
	"github.com/minghsu0107/saga-product/infra/broker"
	"github.com/minghsu0107/saga-product/pkg"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

type MessageType int

const (
	TX_MSG MessageType = iota
	RESULT_MSG
)

// OrchestratorServiceImpl implementation
type OrchestratorServiceImpl struct {
	txPublisher     broker.NATSPublisher
	resultPublisher broker.RedisPublisher
	logger          *log.Entry
}

// NewOrchestratorService factory
func NewOrchestratorService(config *conf.Config, txPublisher broker.NATSPublisher, resultPublisher broker.RedisPublisher) (OrchestratorService, error) {
	return &OrchestratorServiceImpl{
		txPublisher:     txPublisher,
		resultPublisher: resultPublisher,
		logger: config.Logger.ContextLogger.WithFields(log.Fields{
			"type": "service:OrchestratorService",
		}),
	}, nil
}

// StartTransaction starts the first transaction, which is UpdateProductInventory
func (svc *OrchestratorServiceImpl) StartTransaction(parentCtx context.Context, purchase *model.Purchase, correlationID string) error {
	tr := otel.Tracer("startTransaction")
	ctx, span := tr.Start(parentCtx, "event.StartTransaction")
	defer span.End()

	cmd := encodeDomainPurchase(purchase)
	payload, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	svc.publishPurchaseResult(ctx, &event.PurchaseResult{
		CustomerID: purchase.Order.CustomerID,
		PurchaseID: purchase.ID,
		Step:       event.StepUpdateProductInventory,
		Status:     event.StatusExecute,
	}, correlationID)
	middleware.SetCorrelationID(correlationID, msg)
	svc.logger.Infof("update product inventory %v", purchase.ID)
	return svc.publishMessage(ctx, conf.UpdateProductInventoryTopic, msg, TX_MSG)
}

// HandleReply handles reply events
func (svc *OrchestratorServiceImpl) HandleReply(parentCtx context.Context, msg *message.Message, correlationID string) error {
	tr := otel.Tracer("handleReply")
	ctx, span := tr.Start(parentCtx, "event.HandleReply")
	defer span.End()

	handler := msg.Metadata.Get(conf.HandlerHeader)
	switch handler {
	case conf.UpdateProductInventoryHandler:
		resp, err := decodeCreatePurchaseResponse(msg.Payload)
		if err != nil {
			return err
		}
		if resp.Success {
			return svc.createOrder(ctx, resp.Purchase, correlationID)
		}
		svc.logger.Error(resp.Error)
		return svc.rollbackProductInventory(ctx, resp.Purchase.Order.CustomerID, resp.Purchase.ID, correlationID)
	case conf.RollbackProductInventoryHandler:
		resp, err := decodeRollbackResponse(msg.Payload)
		if err != nil {
			return err
		}
		svc.publishRollbackResult(ctx, event.StepUpdateProductInventory, resp, correlationID)
	case conf.CreateOrderHandler:
		resp, err := decodeCreatePurchaseResponse(msg.Payload)
		if err != nil {
			return err
		}
		if resp.Success {
			return svc.createPayment(ctx, resp.Purchase, correlationID)
		}
		svc.logger.Error(resp.Error)
		return svc.rollbackFromOrder(ctx, resp.Purchase.Order.CustomerID, resp.Purchase.ID, correlationID)
	case conf.RollbackOrderHandler:
		resp, err := decodeRollbackResponse(msg.Payload)
		if err != nil {
			return err
		}
		svc.publishRollbackResult(ctx, event.StepCreateOrder, resp, correlationID)
	case conf.CreatePaymentHandler:
		resp, err := decodeCreatePurchaseResponse(msg.Payload)
		if err != nil {
			return err
		}
		if resp.Success {
			svc.publishPurchaseResult(ctx, &event.PurchaseResult{
				CustomerID: resp.Purchase.Order.CustomerID,
				PurchaseID: resp.Purchase.ID,
				Step:       event.StepCreatePayment,
				Status:     event.StatusSucess,
			}, correlationID)

			return nil
		}
		svc.logger.Error(resp.Error)
		return svc.rollbackFromPayment(ctx, resp.Purchase.Order.CustomerID, resp.Purchase.ID, correlationID)
	case conf.RollbackPaymentHandler:
		resp, err := decodeRollbackResponse(msg.Payload)
		if err != nil {
			return err
		}
		svc.publishRollbackResult(ctx, event.StepCreatePayment, resp, correlationID)
	default:
		return fmt.Errorf("unkown handler: %s", handler)
	}
	return nil
}

func (svc *OrchestratorServiceImpl) rollbackProductInventory(ctx context.Context, customerID, purchaseID uint64, correlationID string) error {
	svc.logger.Infof("rollback product inventory %v", purchaseID)
	svc.publishPurchaseResult(ctx, &event.PurchaseResult{
		CustomerID: customerID,
		PurchaseID: purchaseID,
		Step:       event.StepUpdateProductInventory,
		Status:     event.StatusFailed,
	}, correlationID)

	cmd := &pb.RollbackCmd{
		PurchaseId: purchaseID,
		Timestamp:  pkg.Time2pbTimestamp(time.Now()),
	}
	payload, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	middleware.SetCorrelationID(correlationID, msg)

	svc.publishPurchaseResult(ctx, &event.PurchaseResult{
		CustomerID: customerID,
		PurchaseID: purchaseID,
		Step:       event.StepUpdateProductInventory,
		Status:     event.StatusRollbacked,
	}, correlationID)

	return svc.publishMessage(ctx, conf.RollbackProductInventoryTopic, msg, TX_MSG)
}

func (svc *OrchestratorServiceImpl) rollbackFromOrder(ctx context.Context, customerID, purchaseID uint64, correlationID string) error {
	svc.publishPurchaseResult(ctx, &event.PurchaseResult{
		CustomerID: customerID,
		PurchaseID: purchaseID,
		Step:       event.StepCreateOrder,
		Status:     event.StatusFailed,
	}, correlationID)

	var err error
	svc.publishPurchaseResult(ctx, &event.PurchaseResult{
		CustomerID: customerID,
		PurchaseID: purchaseID,
		Step:       event.StepCreateOrder,
		Status:     event.StatusRollbacked,
	}, correlationID)
	if err = svc.rollbackOrder(ctx, customerID, purchaseID, correlationID); err != nil {
		svc.logger.Error(err)
	}

	svc.publishPurchaseResult(ctx, &event.PurchaseResult{
		CustomerID: customerID,
		PurchaseID: purchaseID,
		Step:       event.StepUpdateProductInventory,
		Status:     event.StatusRollbacked,
	}, correlationID)
	if err = svc.rollbackProductInventory(ctx, customerID, purchaseID, correlationID); err != nil {
		svc.logger.Error(err)
	}
	return err
}

func (svc *OrchestratorServiceImpl) rollbackFromPayment(ctx context.Context, customerID, purchaseID uint64, correlationID string) error {
	svc.publishPurchaseResult(ctx, &event.PurchaseResult{
		CustomerID: customerID,
		PurchaseID: purchaseID,
		Step:       event.StepCreatePayment,
		Status:     event.StatusFailed,
	}, correlationID)

	var err error
	svc.publishPurchaseResult(ctx, &event.PurchaseResult{
		CustomerID: customerID,
		PurchaseID: purchaseID,
		Step:       event.StepCreatePayment,
		Status:     event.StatusRollbacked,
	}, correlationID)
	if err = svc.rollbackPayment(ctx, customerID, purchaseID, correlationID); err != nil {
		svc.logger.Error(err)
	}

	svc.publishPurchaseResult(ctx, &event.PurchaseResult{
		CustomerID: customerID,
		PurchaseID: purchaseID,
		Step:       event.StepCreateOrder,
		Status:     event.StatusRollbacked,
	}, correlationID)
	if err = svc.rollbackOrder(ctx, customerID, purchaseID, correlationID); err != nil {
		svc.logger.Error(err)
	}

	svc.publishPurchaseResult(ctx, &event.PurchaseResult{
		CustomerID: customerID,
		PurchaseID: purchaseID,
		Step:       event.StepUpdateProductInventory,
		Status:     event.StatusRollbacked,
	}, correlationID)
	if err = svc.rollbackProductInventory(ctx, customerID, purchaseID, correlationID); err != nil {
		svc.logger.Error(err)
	}
	return err
}

func (svc *OrchestratorServiceImpl) createOrder(ctx context.Context, purchase *model.Purchase, correlationID string) error {
	svc.logger.Infof("create order %v", purchase.ID)
	svc.publishPurchaseResult(ctx, &event.PurchaseResult{
		CustomerID: purchase.Order.CustomerID,
		PurchaseID: purchase.ID,
		Step:       event.StepUpdateProductInventory,
		Status:     event.StatusSucess,
	}, correlationID)

	cmd := encodeDomainPurchase(purchase)
	payload, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	middleware.SetCorrelationID(correlationID, msg)

	svc.publishPurchaseResult(ctx, &event.PurchaseResult{
		CustomerID: purchase.Order.CustomerID,
		PurchaseID: purchase.ID,
		Step:       event.StepCreateOrder,
		Status:     event.StatusExecute,
	}, correlationID)

	return svc.publishMessage(ctx, conf.CreateOrderTopic, msg, TX_MSG)
}

func (svc *OrchestratorServiceImpl) rollbackOrder(ctx context.Context, customerID, purchaseID uint64, correlationID string) error {
	svc.logger.Infof("rollback order %v", purchaseID)
	cmd := &pb.RollbackCmd{
		PurchaseId: purchaseID,
		Timestamp:  pkg.Time2pbTimestamp(time.Now()),
	}
	payload, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	middleware.SetCorrelationID(correlationID, msg)
	return svc.publishMessage(ctx, conf.RollbackOrderTopic, msg, TX_MSG)
}

func (svc *OrchestratorServiceImpl) createPayment(ctx context.Context, purchase *model.Purchase, correlationID string) error {
	svc.logger.Infof("create payment %v", purchase.ID)
	svc.publishPurchaseResult(ctx, &event.PurchaseResult{
		CustomerID: purchase.Order.CustomerID,
		PurchaseID: purchase.ID,
		Step:       event.StepCreateOrder,
		Status:     event.StatusSucess,
	}, correlationID)

	cmd := encodeDomainPurchase(purchase)
	payload, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	middleware.SetCorrelationID(correlationID, msg)

	svc.publishPurchaseResult(ctx, &event.PurchaseResult{
		CustomerID: purchase.Order.CustomerID,
		PurchaseID: purchase.ID,
		Step:       event.StepCreatePayment,
		Status:     event.StatusExecute,
	}, correlationID)

	return svc.publishMessage(ctx, conf.CreatePaymentTopic, msg, TX_MSG)
}

func (svc *OrchestratorServiceImpl) rollbackPayment(ctx context.Context, customerID, purchaseID uint64, correlationID string) error {
	svc.logger.Infof("rollback payment %v", purchaseID)
	cmd := &pb.RollbackCmd{
		PurchaseId: purchaseID,
		Timestamp:  pkg.Time2pbTimestamp(time.Now()),
	}
	payload, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	middleware.SetCorrelationID(correlationID, msg)
	return svc.publishMessage(ctx, conf.RollbackPaymentTopic, msg, TX_MSG)
}

func (svc *OrchestratorServiceImpl) publishPurchaseResult(ctx context.Context, purchaseResult *event.PurchaseResult, correlationID string) {
	result := encodeDomainPurchaseResult(purchaseResult)
	payload, err := json.Marshal(result)
	if err != nil {
		svc.logger.Error(err)
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	middleware.SetCorrelationID(correlationID, msg)
	if err := svc.publishMessage(ctx, conf.PurchaseResultTopic, msg, RESULT_MSG); err != nil {
		svc.logger.Error(err)
	}
}

func (svc *OrchestratorServiceImpl) publishMessage(ctx context.Context, topic string, msg *message.Message, messageType MessageType) error {
	broker.SetSpanContext(ctx, msg)
	switch messageType {
	case TX_MSG:
		return svc.txPublisher.Publish(topic, msg)
	case RESULT_MSG:
		return svc.resultPublisher.Publish(topic, msg)
	default:
		return fmt.Errorf("unkown message type: %v", messageType)
	}
}

func (svc *OrchestratorServiceImpl) publishRollbackResult(ctx context.Context, step string, rollbackResponse *model.RollbackResponse, correlationID string) {
	if !rollbackResponse.Success {
		svc.logger.Error(rollbackResponse.Error)
		svc.publishPurchaseResult(ctx, &event.PurchaseResult{
			CustomerID: rollbackResponse.CustomerID,
			PurchaseID: rollbackResponse.PurchaseID,
			Step:       step,
			Status:     event.StatusRollbackFailed,
		}, correlationID)
	}
}
