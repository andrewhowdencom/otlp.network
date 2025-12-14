package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/procfs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Wifi collector exposes wireless interface statistics.
type Wifi struct {
	meter metric.Meter
	fs    procfs.FS
}

// NewWifi creates a new Wifi collector.
func NewWifi() (*Wifi, error) {
	fs, err := procfs.NewFS("/proc")
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}

	return &Wifi{
		meter: otel.Meter("github.com/andrewhowdencom/otlp.network/internal/collector"),
		fs:    fs,
	}, nil
}

// Start registers the wifi metrics callbacks.
func (c *Wifi) Start(ctx context.Context) error {
	signalGauge, err := c.meter.Float64ObservableGauge(
		"wifi.signal",
		metric.WithDescription("Wifi signal level (dBm)"),
		metric.WithUnit("dBm"),
	)
	if err != nil {
		return err
	}

	qualityGauge, err := c.meter.Float64ObservableGauge(
		"wifi.quality",
		metric.WithDescription("Wifi link quality"),
	)
	if err != nil {
		return err
	}

	_, err = c.meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		stats, err := c.fs.Wireless()
		if err != nil {
			return nil
		}

		for _, iface := range stats {
			attrs := metric.WithAttributes(attribute.String("interface", iface.Name))

			// procfs Wireless struct fields:
			// Name, Status, QualityLink, QualityLevel, QualityNoise...

			o.ObserveFloat64(signalGauge, float64(iface.QualityLevel), attrs)
			o.ObserveFloat64(qualityGauge, float64(iface.QualityLink), attrs)
		}
		return nil
	}, signalGauge, qualityGauge)

	return err
}
