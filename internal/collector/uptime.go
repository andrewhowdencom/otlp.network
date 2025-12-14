package collector

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const (
	instrumentationName = "github.com/andrewhowdencom/otlp.network/internal/collector"
)

// Uptime collects the uptime of the application.
type Uptime struct {
	startTime time.Time
	meter     metric.Meter
}

// NewUptime creates a new Uptime collector.
func NewUptime() *Uptime {
	return &Uptime{
		startTime: time.Now(),
		meter:     otel.Meter(instrumentationName),
	}
}

// Start registers the uptime metric callback.
func (u *Uptime) Start(ctx context.Context) error {
	uptimeCounter, err := u.meter.Float64ObservableCounter(
		"uptime",
		metric.WithDescription("The uptime of the application in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	_, err = u.meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		o.ObserveFloat64(uptimeCounter, time.Since(u.startTime).Seconds())
		return nil
	}, uptimeCounter)

	return err
}
