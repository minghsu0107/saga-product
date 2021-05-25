package broker

import (
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/metrics"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	prom "github.com/prometheus/client_golang/prometheus"
)

// InitializeRouter factory
func InitializeRouter(app string) (*message.Router, error) {
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}
	registry, ok := prom.DefaultGatherer.(*prom.Registry)
	if !ok {
		return nil, fmt.Errorf("prometheus type casting error")
	}
	metricsBuilder := metrics.NewPrometheusMetricsBuilder(registry, app, "pubsub")
	metricsBuilder.AddPrometheusRouterMetrics(router)

	// Router level middleware are executed for every message sent to the router
	router.AddMiddleware(
		// CorrelationID will copy the correlation id from the incoming message's metadata to the produced messages
		middleware.CorrelationID,
		// Timeout makes the handler cancel the incoming message's context after a specified time
		middleware.Timeout(time.Second*15),
		middleware.Recoverer,
	)
	return router, nil
}
