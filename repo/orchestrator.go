package repo

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

// OrchestratorRepository interface
type OrchestratorRepository interface {
	PublishPurchaseResult(ctx context.Context, topic string, msg *message.Message) error
}

// OrchestratorRepositoryImpl implementation
type OrchestratorRepositoryImpl struct {
	publisher message.Publisher
}

// NewOrchestratorRepository factory
func NewOrchestratorRepository(publisher message.Publisher) OrchestratorRepository {
	return &OrchestratorRepositoryImpl{
		publisher: publisher,
	}
}

func (r *OrchestratorRepositoryImpl) PublishPurchaseResult(ctx context.Context, topic string, msg *message.Message) error {
	return r.publisher.Publish(topic, msg)
}
