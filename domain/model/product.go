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

// ProductStatus enumeration
type ProductStatus int

const (
	// ProductOk is ok status
	ProductOk ProductStatus = iota
	// ProductNotFound is not found status
	ProductNotFound
)
