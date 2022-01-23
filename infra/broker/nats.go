package broker

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill-nats/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/components/metrics"
	"github.com/ThreeDotsLabs/watermill/message"
	conf "github.com/minghsu0107/saga-product/config"
	stan "github.com/nats-io/stan.go"
	prom "github.com/prometheus/client_golang/prometheus"
)

type NATSPublisher message.Publisher
type NATSSubscriber message.Subscriber

var (
	TxPublisher  NATSPublisher
	TxSubscriber NATSSubscriber
)

// NewNATSPublisher returns a NATS publisher for event streaming
func NewNATSPublisher(config *conf.Config) (NATSPublisher, error) {
	var err error
	TxPublisher, err = nats.NewStreamingPublisher(
		nats.StreamingPublisherConfig{
			ClusterID: config.NATS.ClusterID,
			ClientID:  config.NATS.ClientID + "_publisher",
			StanOptions: []stan.Option{
				stan.NatsURL(config.NATS.URL),
			},
			Marshaler: nats.GobMarshaler{},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	registry, ok := prom.DefaultGatherer.(*prom.Registry)
	if !ok {
		return nil, fmt.Errorf("prometheus type casting error")
	}
	metricsBuilder := metrics.NewPrometheusMetricsBuilder(registry, config.App, "pubsub")
	TxPublisher, err = metricsBuilder.DecoratePublisher(TxPublisher)
	if err != nil {
		return nil, err
	}
	return TxPublisher, nil
}

// NewNATSSubscriber returns a NATS subscriber for event streaming
func NewNATSSubscriber(config *conf.Config) (NATSSubscriber, error) {
	var err error
	TxSubscriber, err = nats.NewStreamingSubscriber(
		nats.StreamingSubscriberConfig{
			ClusterID: config.NATS.ClusterID,
			ClientID:  config.NATS.ClientID + "_subscriber",

			QueueGroup:       config.NATS.QueueGroup,
			DurableName:      config.NATS.DurableName,
			SubscribersCount: config.NATS.SubscriberCount,
			StanOptions: []stan.Option{
				stan.NatsURL(config.NATS.URL),
			},
			Unmarshaler: nats.GobMarshaler{},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	registry, ok := prom.DefaultGatherer.(*prom.Registry)
	if !ok {
		return nil, fmt.Errorf("prometheus type casting error")
	}
	metricsBuilder := metrics.NewPrometheusMetricsBuilder(registry, config.App, "pubsub")
	TxSubscriber, err = metricsBuilder.DecorateSubscriber(TxSubscriber)
	if err != nil {
		return nil, err
	}
	return TxSubscriber, nil
}
