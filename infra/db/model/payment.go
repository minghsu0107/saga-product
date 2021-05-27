package model

// Payment data model
type Payment struct {
	ID           uint64 `gorm:"primaryKey"`
	CustomerID   uint64 `gorm:"index;not null"`
	CurrencyCode string `gorm:"not null"`
	Amount       int64  `gorm:"not null"`
	UpdatedAt    int64  `gorm:"autoUpdateTime:milli"`
	CreatedAt    int64  `gorm:"autoCreateTime:milli"`
}
