package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/procfs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// TCP collector exposes TCP protocol statistics.
type TCP struct {
	meter          metric.Meter
	fs             procfs.FS
	procMountPoint string
}

// NewTCP creates a new TCP collector.
func NewTCP(procMountPoint string) (*TCP, error) {
	fs, err := procfs.NewFS(procMountPoint)
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}

	return &TCP{
		meter:          otel.Meter("github.com/andrewhowdencom/otlp.network/internal/collector"),
		fs:             fs,
		procMountPoint: procMountPoint,
	}, nil
}

// Start registers the TCP metrics callbacks.
func (c *TCP) Start(ctx context.Context) error {
	connectionCurrent, err := c.meter.Int64ObservableGauge(
		"tcp.connection.current",
		metric.WithDescription("Current TCP connections"),
		metric.WithUnit("{connection}"), // Dimensionless mostly
	)
	if err != nil {
		return err
	}

	connectionTotal, err := c.meter.Int64ObservableCounter(
		"tcp.connection.total",
		metric.WithDescription("Total TCP connections opened"),
	)
	if err != nil {
		return err
	}

	retransmit, err := c.meter.Int64ObservableCounter(
		"tcp.retransmit",
		metric.WithDescription("TCP segments retransmitted"),
	)
	if err != nil {
		return err
	}

	_, err = c.meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		snmp, err := readNetSNMP(c.procMountPoint)
		if err != nil {
			return fmt.Errorf("failed to read net snmp: %w", err)
		}

		// SNMP TCP Keys: 'CurrEstab', 'ActiveOpens', 'PassiveOpens', 'RetransSegs'

		if v, ok := snmp.TCP["CurrEstab"]; ok {
			o.ObserveInt64(connectionCurrent, v)
		}

		var total int64
		if v, ok := snmp.TCP["ActiveOpens"]; ok {
			total += v
		}
		if v, ok := snmp.TCP["PassiveOpens"]; ok {
			total += v
		}
		o.ObserveInt64(connectionTotal, total)

		if v, ok := snmp.TCP["RetransSegs"]; ok {
			o.ObserveInt64(retransmit, v)
		}

		return nil
	}, connectionCurrent, connectionTotal, retransmit)

	return err
}
