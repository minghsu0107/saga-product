package repo

import (
	"context"
	"sort"
	"testing"
	"time"

	conf "github.com/minghsu0107/saga-product/config"
	grpc_order "github.com/minghsu0107/saga-product/infra/grpc/order"
	"github.com/minghsu0107/saga-product/pkg"

	domain_model "github.com/minghsu0107/saga-product/domain/model"
	"github.com/minghsu0107/saga-product/infra/db/model"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	productRepo ProductRepository
	orderRepo   OrderRepository
	paymentRepo PaymentRepository
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
	config := conf.Config{
		ServiceOptions: &conf.ServiceOptions{
			Rps:     100,
			Timeout: time.Minute,
		},
	}
	productRepo = NewProductRepository(db, sf)
	orderRepo = NewOrderRepository(&config, new(grpc_order.ProductConn), db)
	paymentRepo = NewPaymentRepository(db)
	db.Migrator().DropTable(&model.Product{}, &model.Idempotency{}, &model.Order{}, &model.Payment{})
	db.AutoMigrate(&model.Product{}, &model.Idempotency{}, &model.Order{}, &model.Payment{})
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
	var _ = Describe("product repo", func() {
		products := []domain_model.Product{
			{
				Detail: &domain_model.ProductDetail{
					Name:        "first",
					Description: "first product",
					BrandName:   "test",
					Price:       100,
				},
				Inventory: 10,
			},
			{
				Detail: &domain_model.ProductDetail{
					Name:        "second",
					Description: "second product",
					BrandName:   "test",
					Price:       200,
				},
				Inventory: 10,
			},
		}
		var productCatalogs []ProductCatalog
		var cartItems []domain_model.CartItem
		var purchasedItems []domain_model.PurchasedItem
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
					Price:       products[i].Detail.Price,
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
			var _ = It("should fail if inventory is not enough", func() {
				idempotencyKey = 2
				for _, cartItem := range cartItems {
					productID := cartItem.ProductID
					purchasedItems = append(purchasedItems, domain_model.PurchasedItem{
						ProductID: productID,
						Amount:    1000,
					})
				}
				err := productRepo.UpdateProductInventory(context.Background(), idempotencyKey, &purchasedItems)
				Expect(err).To(Equal(ErrInsuffientInventory))
			})
			var _ = It("should rollback inventory", func() {
				idempotencyKey = 1
				rollbacked, idempotencies, err := productRepo.RollbackProductInventory(context.Background(), idempotencyKey)
				Expect(err).To(BeNil())
				Expect(rollbacked).To(BeFalse())

				var realIdempotencies []domain_model.Idempotency
				for _, cartItem := range cartItems {
					productID := cartItem.ProductID
					amount := cartItem.Amount
					realIdempotencies = append(realIdempotencies, domain_model.Idempotency{
						ID:        idempotencyKey,
						ProductID: productID,
						Amount:    amount,
					})
				}
				Expect(idempotencies).To(Equal(&realIdempotencies))
			})
			var _ = It("should not rollback inventory again", func() {
				idempotencyKey = 1
				rollbacked, _, err := productRepo.RollbackProductInventory(context.Background(), idempotencyKey)
				Expect(err).To(BeNil())
				Expect(rollbacked).To(BeTrue())
			})
		})
	})
	var _ = Describe("order repo", func() {
		var orderID uint64 = 1
		order := domain_model.Order{
			ID:         orderID,
			CustomerID: 3,
			PurchasedItems: &[]domain_model.PurchasedItem{
				{
					ProductID: 1,
					Amount:    5,
				},
				{
					ProductID: 2,
					Amount:    5,
				},
			},
		}
		var _ = It("should create order", func() {
			err := orderRepo.CreateOrder(context.Background(), &order)
			Expect(err).To(BeNil())
		})
		var _ = It("should retrieve order", func() {
			retrievedOrder, err := orderRepo.GetOrder(context.Background(), orderID)
			Expect(err).To(BeNil())
			Expect(retrievedOrder).To(Equal(&order))
		})
		var _ = It("should delete order", func() {
			err := orderRepo.DeleteOrder(context.Background(), orderID)
			Expect(err).To(BeNil())

			_, err = orderRepo.GetOrder(context.Background(), orderID)
			Expect(err).NotTo(BeNil())
		})
	})
	var _ = Describe("payment repo", func() {
		var paymentID uint64 = 1
		payment := domain_model.Payment{
			ID:           paymentID,
			CustomerID:   3,
			CurrencyCode: "NT",
			Amount:       100,
		}
		var _ = It("should create payment", func() {
			err := paymentRepo.CreatePayment(context.Background(), &payment)
			Expect(err).To(BeNil())
		})
		var _ = It("should retrieve payment", func() {
			retrievedPayment, err := paymentRepo.GetPayment(context.Background(), paymentID)
			Expect(err).To(BeNil())
			Expect(retrievedPayment).To(Equal(&payment))
		})
		var _ = It("should delete payment", func() {
			err := paymentRepo.DeletePayment(context.Background(), paymentID)
			Expect(err).To(BeNil())

			_, err = paymentRepo.GetPayment(context.Background(), paymentID)
			Expect(err).NotTo(BeNil())
		})
	})
})
