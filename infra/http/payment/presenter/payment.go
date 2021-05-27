package presenter

// Payment response payload
type Payment struct {
	ID           uint64 `json:"id"`
	CurrencyCode string `json:"currency_code"`
	Amount       int64  `json:"amount"`
}
