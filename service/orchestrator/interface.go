package orchestrator

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/minghsu0107/saga-product/domain/model"
	"go.opencensus.io/trace"
)

// OrchestratorService interface
type OrchestratorService interface {
	StartTransaction(sc trace.SpanContext, purchase *model.Purchase, correlationID string) error
	HandleReply(sc trace.SpanContext, msg *message.Message, correlationID string) error
}
