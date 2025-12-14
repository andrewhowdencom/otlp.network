package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/procfs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Sockstat collector exposes socket statistics.
type Sockstat struct {
	meter metric.Meter
	fs    procfs.FS
}

// NewSockstat creates a new Sockstat collector.
func NewSockstat() (*Sockstat, error) {
	fs, err := procfs.NewFS("/proc")
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}

	return &Sockstat{
		meter: otel.Meter("github.com/andrewhowdencom/otlp.network/internal/collector"),
		fs:    fs,
	}, nil
}

// Start registers the Sockstat metrics callbacks.
func (c *Sockstat) Start(ctx context.Context) error {
	used, err := c.meter.Int64ObservableGauge(
		"sockets.used",
		metric.WithDescription("Total number of used sockets"),
	)
	if err != nil {
		return err
	}

	tcpInUse, err := c.meter.Int64ObservableGauge(
		"sockets.tcp.inuse",
		metric.WithDescription("Number of TCP sockets in use"),
	)
	if err != nil {
		return err
	}

	udpInUse, err := c.meter.Int64ObservableGauge(
		"sockets.udp.inuse",
		metric.WithDescription("Number of UDP sockets in use"),
	)
	if err != nil {
		return err
	}

	_, err = c.meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		stats, err := c.fs.NetSockstat()
		if err != nil {
			return fmt.Errorf("failed to read net sockstat: %w", err)
		}

		// Used is typically in 'Sockets' struct or top level.
		// procfs NetSockstat has "Used" field pointer or int?
		// It has Used *int.

		if stats.Used != nil {
			o.ObserveInt64(used, int64(*stats.Used))
		}

		// Iterate protocols
		for _, proto := range stats.Protocols {
			if proto.Protocol == "TCP" {
				o.ObserveInt64(tcpInUse, int64(proto.InUse))
			}
			if proto.Protocol == "UDP" {
				o.ObserveInt64(udpInUse, int64(proto.InUse))
			}
		}

		return nil
	}, used, tcpInUse, udpInUse)

	return err
}
