package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/procfs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Conntrack collector exposes connection tracking statistics.
type Conntrack struct {
	meter          metric.Meter
	fs             procfs.FS
	procMountPoint string
}

// NewConntrack creates a new Conntrack collector.
func NewConntrack(procMountPoint string) (*Conntrack, error) {
	fs, err := procfs.NewFS(procMountPoint)
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}

	return &Conntrack{
		meter:          otel.Meter("github.com/andrewhowdencom/otlp.network/internal/collector"),
		fs:             fs,
		procMountPoint: procMountPoint,
	}, nil
}

// Start registers the Conntrack metrics callbacks.
func (c *Conntrack) Start(ctx context.Context) error {
	entries, err := c.meter.Int64ObservableGauge(
		"conntrack.entries",
		metric.WithDescription("Number of entries in conntrack table"),
	)
	if err != nil {
		return err
	}

	limit, err := c.meter.Int64ObservableGauge(
		"conntrack.limit",
		metric.WithDescription("Limit of entries in conntrack table"),
	)
	if err != nil {
		return err
	}

	_, err = c.meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		stats, err := c.fs.ConntrackStat()
		if err != nil {
			// Often returns error if module not loaded.
			// Try reading sysctl directly if needed, but procfs wraps this IIRC.
			// Actually procfs.ConntrackStat reads /proc/net/stat/nf_conntrack usually.
			// But global count is in /proc/sys/net/netfilter/nf_conntrack_count.
			// procfs.ConntrackStat returns per-cpu stats usually.
			// Let's check if we can get the global count easier?
			// The Plan "Source" listed /proc/sys/...
			// procfs has NetStat which might have it? No.
			// Let's try to read simple files if procfs struct is complex/per-cpu.
			// But wait, user requested "Conntrack".
			// Let's implement reading the sys files manually if procfs is overkill or wrong.
			// Actually, `procfs` does NOT seem to expose `nf_conntrack_count` easily in a single call in older versions,
			// but recent ones might.
			// Let's use standard Go file reading for the global count since it's just a number in a file.
			return c.readSysctl(o, entries, limit)
		}

		// But ConntrackStat is a list.
		// Actually, `nf_conntrack_count` is the authoritative global count.
		// Let's try to use the sysctl method for simplicity and accuracy.
		// Re-reading plan: "/proc/sys/net/netfilter/nf_conntrack_count".

		_ = stats // unused if we use sysctl
		return c.readSysctl(o, entries, limit)
	}, entries, limit)

	return err
}

func (c *Conntrack) readSysctl(o metric.Observer, entries, limit metric.Int64ObservableGauge) error {
	// Read count
	count, err := readFileInt(c.procMountPoint + "/sys/net/netfilter/nf_conntrack_count")
	if err == nil {
		o.ObserveInt64(entries, count)
	}

	// Read max
	max, err := readFileInt(c.procMountPoint + "/sys/net/netfilter/nf_conntrack_max")
	if err == nil {
		o.ObserveInt64(limit, max)
	}

	// We suppress errors because on some systems (e.g. containers) this might not be available
	return nil
}
