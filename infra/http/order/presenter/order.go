package presenter

// DetailedOrder response payload
type DetailedOrder struct {
	ID             uint64          `json:"id"`
	PurchasedItems []PurchasedItem `json:"purchased_items"`
}

// PurchasedItem payload
type PurchasedItem struct {
	ProductID   uint64 `json:"product_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	BrandName   string `json:"brand_name"`
	Price       int64  `json:"price"`
	Amount      int64  `json:"amount"`
}
