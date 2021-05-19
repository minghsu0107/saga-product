package model

// Product data model
type Product struct {
	ID          uint64 `gorm:"primaryKey"`
	Name        string `gorm:"type:varchar(256);not null"`
	Description string `gorm:"type:text;not null"`
	BrandName   string `gorm:"type:varchar(256);not null"`
	Inventory   int64  `gorm:"not null"`
	Price       int64  `gorm:"not null"`
	UpdatedAt   int64  `gorm:"autoUpdateTime:milli"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli"`
}

// Idempotency data model
type Idempotency struct {
	ID        uint64 `gorm:"primaryKey"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
}
