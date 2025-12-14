package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/procfs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Softnet collector exposes softnet statistics (kernel packet processing).
type Softnet struct {
	meter metric.Meter
	fs    procfs.FS
}

// NewSoftnet creates a new Softnet collector.
func NewSoftnet() (*Softnet, error) {
	fs, err := procfs.NewFS("/proc")
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}

	return &Softnet{
		meter: otel.Meter("github.com/andrewhowdencom/otlp.network/internal/collector"),
		fs:    fs,
	}, nil
}

// Start registers the Softnet metrics callbacks.
func (c *Softnet) Start(ctx context.Context) error {
	processed, err := c.meter.Int64ObservableCounter(
		"softnet.processed",
		metric.WithDescription("Number of packets processed by softnet"),
	)
	if err != nil {
		return err
	}

	dropped, err := c.meter.Int64ObservableCounter(
		"softnet.dropped",
		metric.WithDescription("Number of packets dropped by softnet"),
	)
	if err != nil {
		return err
	}

	squeezed, err := c.meter.Int64ObservableCounter(
		"softnet.squeezed",
		metric.WithDescription("Number of times softnet ran out of quota (time squeezed)"),
	)
	if err != nil {
		return err
	}

	_, err = c.meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		stats, err := c.fs.NetSoftnetStat()
		if err != nil {
			return fmt.Errorf("failed to read net softnet stats: %w", err)
		}

		// Softnet stats are per-CPU. We should probably sum them up for a global view,
		// or expose per cpu?
		// Global view is usually sufficient for "is my network stack overloaded?".
		var totalProcessed, totalDropped, totalSqueezed int64

		for _, cpu := range stats {
			totalProcessed += int64(cpu.Processed)
			totalDropped += int64(cpu.Dropped)
			totalSqueezed += int64(cpu.TimeSqueezed)
		}

		o.ObserveInt64(processed, totalProcessed)
		o.ObserveInt64(dropped, totalDropped)
		o.ObserveInt64(squeezed, totalSqueezed)

		return nil
	}, processed, dropped, squeezed)

	return err
}
