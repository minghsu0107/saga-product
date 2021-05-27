package main

import (
	"log"
	"os"

	"github.com/minghsu0107/saga-product/cmd/orchestrator"
	"github.com/minghsu0107/saga-product/cmd/order"
	"github.com/minghsu0107/saga-product/cmd/payment"
	"github.com/minghsu0107/saga-product/cmd/product"
)

var (
	app = os.Getenv("APP")
)

func main() {
	switch app {
	case "product":
		product.RunProductServer(app)
	case "order":
		order.RunOrderServer(app)
	case "payment":
		payment.RunPaymentServer(app)
	case "orchestrator":
		orchestrator.RunOrchestratorServer(app)
	default:
		log.Fatalf("invalid app name: %s. Should be one of 'product', 'order', 'payment', or 'orchestrator'", app)
	}
}
