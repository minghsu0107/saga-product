package event

import (
	"time"
)

// PurchaseResult event
type PurchaseResult struct {
	PurchaseID uint64
	Step       string
	Status     string
	Timestamp  time.Time
}
