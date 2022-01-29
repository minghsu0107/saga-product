package payment

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/minghsu0107/saga-product/config"
	"github.com/minghsu0107/saga-product/infra/http/payment/presenter"
	common_presenter "github.com/minghsu0107/saga-product/infra/http/presenter"
	paymentsvc "github.com/minghsu0107/saga-product/service/payment"
)

// Router wraps http handlers
type Router struct {
	paymentSvc paymentsvc.PaymentService
}

// NewRouter is a factory for router instance
func NewRouter(paymentSvc paymentsvc.PaymentService) *Router {
	return &Router{
		paymentSvc: paymentSvc,
	}
}

// GetPayment endpoint
func (r *Router) GetPayment(c *gin.Context) {
	customerID, ok := c.Request.Context().Value(config.CustomerKey).(uint64)
	if !ok {
		response(c, http.StatusUnauthorized, common_presenter.ErrUnautorized)
		return
	}

	id := c.Param("id")
	paymentID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		response(c, http.StatusBadRequest, common_presenter.ErrInvalidParam)
		return
	}

	payment, err := r.paymentSvc.GetPayment(c.Request.Context(), customerID, paymentID)
	switch err {
	case paymentsvc.ErrPaymentNotFound:
		response(c, http.StatusNotFound, paymentsvc.ErrPaymentNotFound)
		return
	case paymentsvc.ErrUnauthorized:
		response(c, http.StatusUnauthorized, common_presenter.ErrUnautorized)
		return
	case nil:
		c.JSON(http.StatusOK, &presenter.Payment{
			ID:           payment.ID,
			CurrencyCode: payment.CurrencyCode,
			Amount:       payment.Amount,
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
