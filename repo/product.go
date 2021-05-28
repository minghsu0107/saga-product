package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"

	domain_model "github.com/minghsu0107/saga-product/domain/model"
	"github.com/minghsu0107/saga-product/infra/db/model"
	"github.com/minghsu0107/saga-product/pkg"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProductRepository is the product repository interface
type ProductRepository interface {
	CheckProduct(ctx context.Context, cartItem *domain_model.CartItem) (*ProductStatus, error)
	ListProducts(ctx context.Context, offset, size int) (*[]ProductCatalog, error)
	GetProductDetail(ctx context.Context, productID uint64) (*ProductDetail, error)
	GetProductInventory(ctx context.Context, productID uint64) (int64, error)
	CreateProduct(ctx context.Context, product *domain_model.Product) error
	UpdateProductInventory(ctx context.Context, idempotencyKey uint64, purchasedItems *[]domain_model.PurchasedItem) error
	RollbackProductInventory(ctx context.Context, idempotencyKey uint64) (bool, *[]domain_model.Idempotency, error)
}

// ProductStatus select schema
type ProductStatus struct {
	ProductID uint64
	Exist     bool
}

type productCheck struct {
	ID uint64
}

type productInventory struct {
	Inventory int64
}

// ProductCatalog select schema
type ProductCatalog struct {
	ID        uint64
	Name      string
	Inventory int64
	Price     int64
}

// ProductDetail select schema
type ProductDetail struct {
	Name        string
	Description string
	BrandName   string
	Price       int64
}

// ProductRepositoryImpl implements ProductRepository interface
type ProductRepositoryImpl struct {
	db *gorm.DB
	sf pkg.IDGenerator
}

// NewProductRepository is the factory of ProductRepository
func NewProductRepository(db *gorm.DB, sf pkg.IDGenerator) ProductRepository {
	return &ProductRepositoryImpl{
		db: db,
		sf: sf,
	}
}

// CheckProducts method
func (repo *ProductRepositoryImpl) CheckProduct(ctx context.Context, cartItem *domain_model.CartItem) (*ProductStatus, error) {
	var check productCheck
	productID := cartItem.ProductID
	if err := repo.db.Model(&model.Product{}).Select("id").Where("id = ?", productID).First(&check).WithContext(ctx).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &ProductStatus{
				ProductID: productID,
				Exist:     false,
			}, nil
		}
		return nil, err
	}
	return &ProductStatus{
		ProductID: productID,
		Exist:     true,
	}, nil
}

// ListProducts method
func (repo *ProductRepositoryImpl) ListProducts(ctx context.Context, offset, size int) (*[]ProductCatalog, error) {
	var catalogs []ProductCatalog
	if err := paginate(repo.db, offset, size).Model(&model.Product{}).Select("id", "name", "inventory", "price").Find(&catalogs).Error; err != nil {
		return nil, err
	}
	return &catalogs, nil
}

// GetProductDetails method
func (repo *ProductRepositoryImpl) GetProductDetail(ctx context.Context, productID uint64) (*ProductDetail, error) {
	var productDetail ProductDetail
	if err := repo.db.Model(&model.Product{}).Select("name", "description", "brand_name", "price").Where("id = ?", productID).First(&productDetail).WithContext(ctx).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("product not found; product ID: %v", productID)
		}
		return nil, err
	}
	return &productDetail, nil
}

// GetProductInventory method
func (repo *ProductRepositoryImpl) GetProductInventory(ctx context.Context, productID uint64) (int64, error) {
	var productInventory productInventory
	if err := repo.db.Model(&model.Product{}).Select("inventory").Where("id = ?", productID).First(&productInventory).WithContext(ctx).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("product not found; product ID: %v", productID)
		}
		return 0, err
	}
	return productInventory.Inventory, nil
}

// CreateProduct method
func (repo *ProductRepositoryImpl) CreateProduct(ctx context.Context, product *domain_model.Product) error {
	sonyflakeID, err := repo.sf.NextID()
	if err != nil {
		return err
	}
	if err := repo.db.Create(&model.Product{
		ID:          sonyflakeID,
		Name:        product.Detail.Name,
		Description: product.Detail.Description,
		BrandName:   product.Detail.BrandName,
		Inventory:   product.Inventory,
		Price:       product.Detail.Price,
	}).WithContext(ctx).Error; err != nil {
		return err
	}
	return nil
}

// UpdateProductInventory method
func (repo *ProductRepositoryImpl) UpdateProductInventory(ctx context.Context, idempotencyKey uint64, purchasedItems *[]domain_model.PurchasedItem) error {
	var err error
	var idempotency model.Idempotency
	err = repo.db.Model(&model.Idempotency{}).Where("id = ?", idempotencyKey).First(&idempotency).Error
	if err == nil {
		return ErrInvalidIdempotency
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	sort.Slice(*purchasedItems, func(i, j int) bool { return (*purchasedItems)[i].ProductID < (*purchasedItems)[j].ProductID })
	tx := repo.db.Begin(&sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	for _, purchasedItem := range *purchasedItems {
		var productInventory productInventory
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.Product{}).Select("inventory").Where("id = ?", purchasedItem.ProductID).First(&productInventory).Error; err != nil {
			tx.Rollback()
			return err
		}
		if productInventory.Inventory < purchasedItem.Amount {
			tx.Rollback()
			return ErrInsuffientInventory
		}
		if err := tx.Model(&model.Product{}).Where("id = ?", purchasedItem.ProductID).Update("inventory", gorm.Expr("inventory - ?", purchasedItem.Amount)).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	var idempotencies []model.Idempotency
	for _, purchasedItem := range *purchasedItems {
		idempotencies = append(idempotencies, model.Idempotency{
			ID:         idempotencyKey,
			ProductID:  purchasedItem.ProductID,
			Amount:     purchasedItem.Amount,
			Rollbacked: false,
		})
	}
	if err := tx.Model(&model.Idempotency{}).Create(&idempotencies).WithContext(ctx).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// RollbackProductInventory method
func (repo *ProductRepositoryImpl) RollbackProductInventory(ctx context.Context, idempotencyKey uint64) (bool, *[]domain_model.Idempotency, error) {
	var idempotencies []model.Idempotency
	if err := repo.db.Model(&model.Idempotency{}).Select("product_id", "amount", "rollbacked").Where("id = ?", idempotencyKey).Order("product_id").Find(&idempotencies).Error; err != nil {
		return false, nil, err
	}
	if len(idempotencies) == 0 {
		return false, nil, fmt.Errorf("idempotency kwy not found: %v", idempotencyKey)
	}
	if idempotencies[0].Rollbacked {
		return true, nil, nil
	}

	tx := repo.db.Begin(&sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return false, nil, err
	}

	for _, idempotency := range idempotencies {
		var productInventory productInventory
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.Product{}).Select("inventory").Where("id = ?", idempotency.ProductID).First(&productInventory).Error; err != nil {
			tx.Rollback()
			return false, nil, err
		}
		if err := tx.Model(&model.Product{}).Where("id = ?", idempotency.ProductID).Update("inventory", gorm.Expr("inventory + ?", idempotency.Amount)).Error; err != nil {
			tx.Rollback()
			return false, nil, err
		}
	}
	if err := tx.Model(&model.Idempotency{}).Where("id = ?", idempotencyKey).Update("rollbacked", true).WithContext(ctx).Error; err != nil {
		tx.Rollback()
		return false, nil, err
	}
	var domainIdempotencies []domain_model.Idempotency
	for _, idempotency := range idempotencies {
		domainIdempotencies = append(domainIdempotencies, domain_model.Idempotency{
			ID:        idempotencyKey,
			ProductID: idempotency.ProductID,
			Amount:    idempotency.Amount,
		})
	}
	return false, &domainIdempotencies, tx.Commit().Error
}

func paginate(db *gorm.DB, offset, size int) *gorm.DB {
	if offset < 0 {
		offset = 0
	}

	if size < 0 {
		size = 0
	}

	if size > 500 {
		size = 500
	}
	return db.Offset(offset).Limit(size)
}
