package order

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/infra/http/order/presenter"
	common_presenter "github.com/minghsu0107/saga-product/infra/http/presenter"
	ordersvc "github.com/minghsu0107/saga-product/service/order"
)

// Router wraps http handlers
type Router struct {
	orderSvc ordersvc.OrderService
}

// NewRouter is a factory for router instance
func NewRouter(orderSvc ordersvc.OrderService) *Router {
	return &Router{
		orderSvc: orderSvc,
	}
}

// GetDetailedOrder endpoint
func (r *Router) GetDetailedOrder(c *gin.Context) {
	customerID, ok := c.Request.Context().Value(config.CustomerKey).(uint64)
	if !ok {
		response(c, http.StatusUnauthorized, common_presenter.ErrUnauthorized)
		return
	}

	id := c.Param("id")
	orderID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		response(c, http.StatusBadRequest, common_presenter.ErrInvalidParam)
		return
	}

	order, err := r.orderSvc.GetDetailedOrder(c.Request.Context(), customerID, orderID)
	switch err {
	case ordersvc.ErrOrderNotFound:
		response(c, http.StatusNotFound, ordersvc.ErrOrderNotFound)
		return
	case ordersvc.ErrUnauthorized:
		response(c, http.StatusUnauthorized, common_presenter.ErrUnauthorized)
		return
	case nil:
		var purchasedItems []presenter.PurchasedItem
		for _, detailedPurchasedItem := range *order.DetailedPurchasedItems {
			purchasedItems = append(purchasedItems, presenter.PurchasedItem{
				ProductID:   detailedPurchasedItem.ProductID,
				Name:        detailedPurchasedItem.Name,
				Description: detailedPurchasedItem.Description,
				BrandName:   detailedPurchasedItem.BrandName,
				Price:       detailedPurchasedItem.Price,
				Amount:      detailedPurchasedItem.Amount,
			})
		}
		c.JSON(http.StatusOK, &presenter.DetailedOrder{
			ID:             order.ID,
			PurchasedItems: purchasedItems,
		})
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
