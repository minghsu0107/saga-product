package db

import (
	"github.com/minghsu0107/saga-product/infra/db/model"
	"gorm.io/gorm"
)

// Migrator migrates DB schemas on startup
type Migrator struct {
	db *gorm.DB
}

// NewMigrator is the factory of Migrator
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{
		db: db,
	}
}

// Migrate method migrates db schemas
func (m *Migrator) Migrate() error {
	return m.db.AutoMigrate(&model.Product{}, &model.Idempotency{})
}
