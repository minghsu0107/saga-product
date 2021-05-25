package model

// Order data model
type Order struct {
	ID         uint64 `gorm:"primaryKey"`
	ProductID  uint64 `gorm:"primaryKey"`
	Amount     int64  `gorm:"not null"`
	CustomerID uint64 `gorm:"not null"`
	UpdatedAt  int64  `gorm:"autoUpdateTime:milli"`
	CreatedAt  int64  `gorm:"autoCreateTime:milli"`
}
