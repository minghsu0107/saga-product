package model

// Purchase entity
type Purchase struct {
	ID      uint64
	Order   *Order
	Payment *Payment
}
