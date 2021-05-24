package model

// Purchase entity
type Purchase struct {
	Order   *Order
	Payment *Payment
}
