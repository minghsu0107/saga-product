package orchestrator

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/minghsu0107/saga-product/dep"
	log "github.com/sirupsen/logrus"
)

func RunOrchestratorServer(app string) {
	server, err := dep.InitializeOrchestratorServer()
	if err != nil {
		log.Fatal(err)
	}

	errs := make(chan error, 1)
	go func() {
		errs <- server.Run()
	}()

	// catch shutdown
	done := make(chan bool, 1)
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig

		// graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.GracefulStop(ctx, done)
	}()

	err = <-errs
	if err != nil {
		log.Fatal(err)
	}

	// wait for graceful shutdown
	<-done
}
