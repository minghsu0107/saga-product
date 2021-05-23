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
		log.Fatal("invalid app name. Should be one of 'product', 'order', or 'payment'")
	}
}
