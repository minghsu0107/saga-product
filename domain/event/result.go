package event

import (
	"time"
)

var (
	StepUpdateProductInventory = "UPDATE_PRODUCT_INVENTORY"
	StepCreateOrder            = "CREATE_ORDER"
	StepCreatePayment          = "CREATE_PAYMENT"

	StatusExecute        = "STATUS_EXUCUTE"
	StatusSucess         = "STATUS_SUCCESS"
	StatusFailed         = "STATUS_FAILED"
	StatusRollbacked     = "STATUS_ROLLBACKED"
	StatusRollbackFailed = "STATUS_ROLLBACK_FAIL"
)

// PurchaseResult event
type PurchaseResult struct {
	CustomerID uint64
	PurchaseID uint64
	Step       string
	Status     string
	Timestamp  time.Time
}
