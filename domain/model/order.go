package model

// Order entity
type Order struct {
	CustomerID     uint64
	PurchasedItems *[]PurchasedItem
}
