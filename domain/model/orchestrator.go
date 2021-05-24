package model

// OrchestratorPayload value object
type OrchestratorPayload struct {
	IdempotencyKey uint64
	Purchase       *Purchase
}
