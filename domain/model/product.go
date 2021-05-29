package model

// Product entity
type Product struct {
	ID        uint64
	Detail    *ProductDetail
	Inventory int64
}

// ProductDetail value object
type ProductDetail struct {
	Name        string
	Description string
	BrandName   string
	Price       int64
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

// ProductStatus value object
type ProductStatus struct {
	ProductID uint64
	Price     int64
	Status    Status
}

// ProductCatalog value object
type ProductCatalog struct {
	ID        uint64
	Name      string
	Inventory int64
	Price     int64
}

// Idempotency entity
type Idempotency struct {
	ID        uint64
	ProductID uint64
	Amount    int64
}
