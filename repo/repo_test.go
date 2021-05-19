package repo

import (
	"context"
	"sort"
	"testing"

	"github.com/minghsu0107/saga-product/pkg"

	domain_model "github.com/minghsu0107/saga-product/domain/model"
	"github.com/minghsu0107/saga-product/infra/db/model"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	productRepo ProductRepository
	sf          pkg.IDGenerator
)

func TestRepo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "repo suite")
}

var _ = BeforeSuite(func() {
	InitDB()
	var err error
	sf, err = pkg.NewSonyFlake()
	if err != nil {
		panic(err)
	}
	productRepo = NewProductRepository(db, sf)
	db.Migrator().DropTable(&model.Product{}, &model.Idempotency{})
	db.AutoMigrate(&model.Product{}, &model.Idempotency{})
})

var _ = AfterSuite(func() {
	db.Migrator().DropTable(&model.Product{}, &model.Idempotency{})
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()
})

var _ = Describe("test repo", func() {
	products := []domain_model.Product{
		{
			Detail: &domain_model.ProductDetail{
				Name:        "first",
				Description: "first product",
				BrandName:   "test",
			},
			Inventory: 10,
		},
		{
			Detail: &domain_model.ProductDetail{
				Name:        "second",
				Description: "second product",
				BrandName:   "test",
			},
			Inventory: 10,
		},
	}
	var productCatalogs []ProductCatalog
	var cartItems []domain_model.CartItem
	var purchasedItems []domain_model.PurchasedItem
	var _ = Describe("product repo", func() {
		var _ = It("should create products", func() {
			for _, product := range products {
				err := productRepo.CreateProduct(context.Background(), &product)
				Expect(err).To(BeNil())
			}
		})
		var _ = It("should list proudcts", func() {
			offset := 0
			size := 100
			catalogs, err := productRepo.ListProducts(context.Background(), offset, size)
			Expect(err).To(BeNil())
			Expect(len(*catalogs)).To(Equal(len(products)))

			productCatalogs = *catalogs
			sort.Slice(productCatalogs, func(i, j int) bool { return productCatalogs[i].ID < productCatalogs[j].ID })

		})
		var _ = It("should check product", func() {
			for _, catalog := range productCatalogs {
				cartItems = append(cartItems, domain_model.CartItem{
					ProductID: catalog.ID,
					Amount:    1,
				})
			}
			status, err := productRepo.CheckProduct(context.Background(), &cartItems[0])
			Expect(err).To(BeNil())
			Expect(status).To(Equal(&ProductStatus{
				ProductID: cartItems[0].ProductID,
				Exist:     true,
			}))

			var fakeID uint64 = 1
			status, err = productRepo.CheckProduct(context.Background(), &domain_model.CartItem{ProductID: fakeID})
			Expect(err).To(BeNil())
			Expect(status).To(Equal(&ProductStatus{
				ProductID: fakeID,
				Exist:     false,
			}))
		})
		var _ = It("should get product detail", func() {
			for i, productCatalog := range productCatalogs {
				detail, err := productRepo.GetProductDetail(context.Background(), productCatalog.ID)
				Expect(err).To(BeNil())
				Expect(detail).To(Equal(&ProductDetail{
					Name:        products[i].Detail.Name,
					Description: products[i].Detail.Description,
					BrandName:   products[i].Detail.BrandName,
				}))
			}
		})
		var _ = It("should get product inventory", func() {
			for i, productCatalog := range productCatalogs {
				inventory, err := productRepo.GetProductInventory(context.Background(), productCatalog.ID)
				Expect(err).To(BeNil())
				Expect(inventory).To(Equal(products[i].Inventory))
			}
		})
		var _ = Describe("update inventory case", func() {
			var idempotencyKey uint64 = 1
			var _ = It("should update product inventory", func() {
				for _, cartItem := range cartItems {
					productID := cartItem.ProductID
					amount := cartItem.Amount
					purchasedItems = append(purchasedItems, domain_model.PurchasedItem{
						ProductID: productID,
						Amount:    amount,
					})
				}
				err := productRepo.UpdateProductInventory(context.Background(), idempotencyKey, &purchasedItems)
				Expect(err).To(BeNil())

				for i, productCatalog := range productCatalogs {
					inventory, err := productRepo.GetProductInventory(context.Background(), productCatalog.ID)
					Expect(err).To(BeNil())
					Expect(inventory).To(Equal(productCatalog.Inventory - purchasedItems[i].Amount))
				}
			})
			var _ = It("should not violate idempotency when updating product inventory again", func() {
				err := productRepo.UpdateProductInventory(context.Background(), idempotencyKey, &purchasedItems)
				Expect(err).To(Equal(ErrInvalidIdempotency))
			})
		})
	})
})
