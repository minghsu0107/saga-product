package model

// Order entity
type Order struct {
	ID             uint64
	CustomerID     uint64
	PurchasedItems *[]PurchasedItem
}
