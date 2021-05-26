package model

// CreatePurchaseResponse value object
type CreatePurchaseResponse struct {
	Purchase *Purchase
	Success  bool
	Error    string
}

// RollbackResponse value object
type RollbackResponse struct {
	CustomerID uint64
	PurchaseID uint64
	Success    bool
	Error      string
}
