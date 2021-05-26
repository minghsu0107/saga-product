package orchestrator

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/minghsu0107/saga-product/domain/model"
)

// OrchestratorService interface
type OrchestratorService interface {
	StartTransaction(ctx context.Context, purchase *model.Purchase, correlationID string) error
	HandleReply(ctx context.Context, msg *message.Message, correlationID string) error
}
