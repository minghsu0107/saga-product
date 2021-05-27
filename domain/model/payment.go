package model

// payment value object
type Payment struct {
	ID           uint64
	CustomerID   uint64
	CurrencyCode string
	Amount       int64
}
