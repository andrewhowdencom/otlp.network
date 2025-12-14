package collector

import (
	"context"
	"fmt"
	"strings"

	"github.com/prometheus/procfs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Device collector exposes network interface statistics.
type Device struct {
	meter metric.Meter
	fs    procfs.FS
}

// NewDevice creates a new Device collector.
func NewDevice(procMountPoint string) (*Device, error) {
	fs, err := procfs.NewFS(procMountPoint)
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}

	return &Device{
		meter: otel.Meter("github.com/andrewhowdencom/otlp.network/internal/collector"),
		fs:    fs,
	}, nil
}

// Start registers the device metrics callbacks.
func (c *Device) Start(ctx context.Context) error {
	ioMetric, err := c.meter.Int64ObservableCounter(
		"device.io",
		metric.WithDescription("Network interface I/O"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return err
	}

	packetsMetric, err := c.meter.Int64ObservableCounter(
		"device.packets",
		metric.WithDescription("Network interface packets"),
		metric.WithUnit("{packets}"),
	)
	if err != nil {
		return err
	}

	errorsMetric, err := c.meter.Int64ObservableCounter(
		"device.errors",
		metric.WithDescription("Network interface errors"),
		metric.WithUnit("{errors}"),
	)
	if err != nil {
		return err
	}

	droppedMetric, err := c.meter.Int64ObservableCounter(
		"device.dropped",
		metric.WithDescription("Network interface dropped packets"),
		metric.WithUnit("{packets}"),
	)
	if err != nil {
		return err
	}

	_, err = c.meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		stats, err := c.fs.NetDev()
		if err != nil {
			return fmt.Errorf("failed to read net dev stats: %w", err)
		}

		for _, iface := range stats {
			// RX
			rxAttrs := metric.WithAttributes(
				attribute.String("interface", iface.Name),
				attribute.String("direction", "receive"),
			)
			o.ObserveInt64(ioMetric, int64(iface.RxBytes), rxAttrs)
			o.ObserveInt64(packetsMetric, int64(iface.RxPackets), rxAttrs)
			o.ObserveInt64(errorsMetric, int64(iface.RxErrors), rxAttrs)
			o.ObserveInt64(droppedMetric, int64(iface.RxDropped), rxAttrs)

			// TX
			txAttrs := metric.WithAttributes(
				attribute.String("interface", iface.Name),
				attribute.String("direction", "transmit"),
			)
			o.ObserveInt64(ioMetric, int64(iface.TxBytes), txAttrs)
			o.ObserveInt64(packetsMetric, int64(iface.TxPackets), txAttrs)
			o.ObserveInt64(errorsMetric, int64(iface.TxErrors), txAttrs)
			o.ObserveInt64(droppedMetric, int64(iface.TxDropped), txAttrs)
		}
		return nil
	}, ioMetric, packetsMetric, errorsMetric, droppedMetric)

	return err
}

func (c *Device) isLoopback(name string) bool {
	// TODO: Filter loopback if requested? Implementation plan said "maybe filtering lo".
	// For now keeping it simple as per plan "starting with all".
	return strings.HasPrefix(name, "lo")
}
