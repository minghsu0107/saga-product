package pkg

import (
	"fmt"
	"net/http"

	conf "github.com/minghsu0107/saga-product/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

var TracerProvider *tracesdk.TracerProvider

type ObservibilityInjector struct {
	promPort  string
	jaegerUrl string
	app       string
}

func NewObservibilityInjector(config *conf.Config) (*ObservibilityInjector, error) {
	promPort := config.PromPort
	jaegerUrl := config.JaegerUrl
	app := config.App

	if app == "" {
		return nil, fmt.Errorf("app name should not be empty")
	}

	return &ObservibilityInjector{
		promPort:  promPort,
		jaegerUrl: jaegerUrl,
		app:       app,
	}, nil
}

func (injector *ObservibilityInjector) Register(errs chan error) {
	if injector.jaegerUrl != "" {
		err := initTracerProvider(injector.jaegerUrl, injector.app)
		if err != nil {
			errs <- err
		}
		otel.SetTracerProvider(TracerProvider)
	}
	if injector.promPort != "" {
		go func() {
			log.Infof("starting prom metrics on PROM_PORT=[%s]", injector.promPort)
			errs <- http.ListenAndServe(fmt.Sprintf(":%s", injector.promPort), promhttp.Handler())
		}()
	}
}

func initTracerProvider(jaegerUrl, serviceName string) error {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerUrl)))
	if err != nil {
		return err
	}
	TracerProvider = tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)
	return nil
}
