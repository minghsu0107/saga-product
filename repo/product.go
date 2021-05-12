package repo

import (
	//domain_model "github.com/minghsu0107/saga-product/domain/model"
	//"github.com/minghsu0107/saga-product/infra/db/model"
	"gorm.io/gorm"
)

// ProductRepository is the product repository interface
type ProductRepository interface {
}

// ProductRepositoryImpl implements ProductRepository interface
type ProductRepositoryImpl struct {
	db *gorm.DB
}

// NewProductRepository is the factory of ProductRepository
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &ProductRepositoryImpl{
		db: db,
	}
}
