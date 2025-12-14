package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/procfs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// UDP collector exposes UDP protocol statistics.
type UDP struct {
	meter          metric.Meter
	fs             procfs.FS
	procMountPoint string
}

// NewUDP creates a new UDP collector.
func NewUDP(procMountPoint string) (*UDP, error) {
	fs, err := procfs.NewFS(procMountPoint)
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}

	return &UDP{
		meter:          otel.Meter("github.com/andrewhowdencom/otlp.network/internal/collector"),
		fs:             fs,
		procMountPoint: procMountPoint,
	}, nil
}

// Start registers the UDP metrics callbacks.
func (c *UDP) Start(ctx context.Context) error {
	packets, err := c.meter.Int64ObservableCounter(
		"udp.packets",
		metric.WithDescription("UDP packets statistics"),
	)
	if err != nil {
		return err
	}

	drops, err := c.meter.Int64ObservableCounter(
		"udp.drops",
		metric.WithDescription("UDP drops statistics"),
	)
	if err != nil {
		return err
	}

	_, err = c.meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		snmp, err := readNetSNMP(c.procMountPoint)
		if err != nil {
			return fmt.Errorf("failed to read net snmp: %w", err)
		}

		// SNMP UDP Keys: InDatagrams, OutDatagrams, InErrors, NoPorts, RcvbufErrors

		if v, ok := snmp.UDP["InDatagrams"]; ok {
			o.ObserveInt64(packets, v, metric.WithAttributes(attribute.String("direction", "in"), attribute.String("type", "datagrams")))
		}
		if v, ok := snmp.UDP["OutDatagrams"]; ok {
			o.ObserveInt64(packets, v, metric.WithAttributes(attribute.String("direction", "out"), attribute.String("type", "datagrams")))
		}
		if v, ok := snmp.UDP["InErrors"]; ok {
			o.ObserveInt64(packets, v, metric.WithAttributes(attribute.String("type", "errors")))
		}

		if v, ok := snmp.UDP["NoPorts"]; ok {
			o.ObserveInt64(drops, v, metric.WithAttributes(attribute.String("reason", "no_port")))
		}
		if v, ok := snmp.UDP["RcvbufErrors"]; ok {
			o.ObserveInt64(drops, v, metric.WithAttributes(attribute.String("reason", "rcv_buf")))
		}

		return nil
	}, packets, drops)

	return err
}
