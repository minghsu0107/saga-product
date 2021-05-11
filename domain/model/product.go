package model

// Product entity
type Product struct {
	ID          uint64
	Name        string
	Description string
	BrandName   string
	Inventory   int64
}

// CartItem value object
type CartItem struct {
	ProductID uint64
	Amount    int64
}

// PurchasedItem value object
type PurchasedItem struct {
	ProductID uint64
	Amount    int64
}

// Status enumeration
type Status int

const (
	// ProductOk is ok status
	ProductOk Status = iota
	// ProductNotFound is not found status
	ProductNotFound
)

type ProductStatus struct {
	ProductID uint64
	Status    Status
}

type ProductCatalog struct {
	ID        uint64
	Name      string
	Inventory int64
}
