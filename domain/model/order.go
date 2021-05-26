package model

// Order entity
type Order struct {
	ID             uint64
	CustomerID     uint64
	PurchasedItems *[]PurchasedItem
}

// DetailedOrder value object
type DetailedOrder struct {
	ID                     uint64
	CustomerID             uint64
	DetailedPurchasedItems *[]DetailedPurchasedItem
}

// DetailedPurchasedItem value object
type DetailedPurchasedItem struct {
	ProductID   uint64
	Name        string
	Description string
	BrandName   string
	Price       int64
	Amount      int64
}
