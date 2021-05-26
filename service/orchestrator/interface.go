package orchestrator

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/minghsu0107/saga-product/domain/model"
)

// OrchestratorService interface
type OrchestratorService interface {
	PlaySaga(ctx context.Context, purchase *model.Purchase, subscriber message.Subscriber) error
}
