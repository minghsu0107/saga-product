package main

import (
	"log"
	"os"

	cmd "github.com/minghsu0107/saga-product/cmd"
)

var (
	app = os.Getenv("APP")
)

func main() {
	switch app {
	case "product":
		cmd.RunProductServer(app)
	default:
		log.Fatalf("invalid app name: %s. Should be one of 'product', 'order', 'payment', or 'orchestrator'", app)
	}
}
