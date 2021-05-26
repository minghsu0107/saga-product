package product

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/minghsu0107/saga-product/domain/model"
	common_presenter "github.com/minghsu0107/saga-product/infra/http/presenter"
	"github.com/minghsu0107/saga-product/infra/http/product/presenter"
	"github.com/minghsu0107/saga-product/service/product"
)

// Router wraps http handlers
type Router struct {
	productSvc     product.ProductService
	sagaProductSvc product.SagaProductService
}

// NewRouter is a factory for router instance
func NewRouter(productSvc product.ProductService, sagaProductSvc product.SagaProductService) *Router {
	return &Router{
		productSvc:     productSvc,
		sagaProductSvc: sagaProductSvc,
	}
}

// ListProducts endpoint
func (r *Router) ListProducts(c *gin.Context) {
	var pagination presenter.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response(c, http.StatusBadRequest, common_presenter.ErrInvalidParam)
		return
	}
	catalogs, err := r.productSvc.ListProducts(c.Request.Context(), pagination.Offset, pagination.Size)
	switch err {
	case nil:
		var productCatalogs []presenter.Productcatalog
		for _, catalog := range *catalogs {
			productCatalogs = append(productCatalogs, presenter.Productcatalog{
				ID:        catalog.ID,
				Name:      catalog.Name,
				Inventory: catalog.Inventory,
				Price:     catalog.Price,
			})
		}
		c.JSON(http.StatusOK, &presenter.ProductCatalogs{
			Catalogs: productCatalogs,
		})
	default:
		response(c, http.StatusInternalServerError, common_presenter.ErrServer)
		return
	}
}

//  GetProducts endpoint
func (r *Router) GetProduct(c *gin.Context) {
	id := c.Param("id")
	productID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		response(c, http.StatusBadRequest, common_presenter.ErrInvalidParam)
		return
	}
	products, err := r.productSvc.GetProducts(c.Request.Context(), []uint64{productID})
	switch err {
	case nil:
		if len(*products) != 1 {
			response(c, http.StatusInternalServerError, common_presenter.ErrServer)
			return
		}
		product := (*products)[0]
		c.JSON(http.StatusOK, &presenter.Product{
			ID:          product.ID,
			Name:        product.Detail.Name,
			Description: product.Detail.Description,
			BrandName:   product.Detail.BrandName,
			Price:       product.Detail.Price,
			Inventory:   product.Inventory,
		})
	default:
		response(c, http.StatusInternalServerError, common_presenter.ErrServer)
		return
	}
}

// CreateProducts endpoint
func (r *Router) CreateProduct(c *gin.Context) {
	var product presenter.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		response(c, http.StatusBadRequest, common_presenter.ErrInvalidParam)
		return
	}
	err := r.productSvc.CreateProduct(c.Request.Context(), &model.Product{
		Detail: &model.ProductDetail{
			Name:        product.Name,
			Description: product.Description,
			BrandName:   product.BrandName,
			Price:       product.Price,
		},
		Inventory: product.Inventory,
	})
	switch err {
	case nil:
		c.JSON(http.StatusCreated, common_presenter.OkMsg)
		return
	default:
		response(c, http.StatusInternalServerError, common_presenter.ErrServer)
		return
	}
}

func response(c *gin.Context, httpCode int, err error) {
	message := err.Error()
	c.JSON(httpCode, common_presenter.ErrResponse{
		Message: message,
	})
}
