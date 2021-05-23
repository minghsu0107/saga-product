package db

import (
	"fmt"

	"github.com/minghsu0107/saga-product/infra/db/model"
	"gorm.io/gorm"
)

// Migrator migrates DB schemas on startup
type Migrator struct {
	db  *gorm.DB
	app string
}

// NewMigrator is the factory of Migrator
func NewMigrator(db *gorm.DB, app string) *Migrator {
	return &Migrator{
		db:  db,
		app: app,
	}
}

// Migrate method migrates db schemas
func (m *Migrator) Migrate() error {
	switch m.app {
	case "product":
		return m.db.AutoMigrate(&model.Product{}, &model.Idempotency{})
	}
	return fmt.Errorf("invalid app name")
}
