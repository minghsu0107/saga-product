package http

import (
	"github.com/gin-gonic/gin"
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

}

//  GetProducts endpoint
func (r *Router) GetProduct(c *gin.Context) {

}

// CreateProducts endpoint
func (r *Router) CreateProduct(c *gin.Context) {

}
