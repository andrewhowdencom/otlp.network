package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Config holds the configuration for OpenTelemetry.
type Config struct {
	ServiceName    string
	ServiceVersion string
	Endpoint       string
	Insecure       bool
	Interval       time.Duration
}

// Setup initializes the OpenTelemetry SDK with OTLP HTTP and Prometheus exporters.
func Setup(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 1. OTLP HTTP Exporter (Optional)
	var readers []sdkmetric.Reader

	if cfg.Endpoint != "" {
		var otlpOpts []otlpmetrichttp.Option
		otlpOpts = append(otlpOpts, otlpmetrichttp.WithEndpoint(cfg.Endpoint))
		if cfg.Insecure {
			otlpOpts = append(otlpOpts, otlpmetrichttp.WithInsecure())
		}

		otlpExporter, err := otlpmetrichttp.New(ctx, otlpOpts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
		}
		readers = append(readers, sdkmetric.NewPeriodicReader(otlpExporter, sdkmetric.WithInterval(cfg.Interval)))
	}

	// 2. Prometheus Exporter
	promExporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus exporter: %w", err)
	}
	readers = append(readers, promExporter)

	// 3. Meter Provider with readers
	opts := []sdkmetric.Option{
		sdkmetric.WithResource(res),
	}
	for _, r := range readers {
		opts = append(opts, sdkmetric.WithReader(r))
	}

	meterProvider := sdkmetric.NewMeterProvider(opts...)

	otel.SetMeterProvider(meterProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return meterProvider.Shutdown, nil
}
