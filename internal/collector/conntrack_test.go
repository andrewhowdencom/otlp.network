package collector

import (
	"context"
	"path/filepath"
	"testing"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestConntrack(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)

	procPath, _ := filepath.Abs("testdata/proc")
	c, err := NewConntrack(procPath)
	if err != nil {
		t.Fatalf("failed to create conntrack collector: %v", err)
	}

	if err := c.Start(context.Background()); err != nil {
		t.Fatalf("failed to start collector: %v", err)
	}

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("failed to collect metrics: %v", err)
	}

	if len(rm.ScopeMetrics) == 0 {
		t.Fatal("no scope metrics found")
	}
	metrics := rm.ScopeMetrics[0].Metrics

	findMetric := func(name string) metricdata.Metrics {
		for _, m := range metrics {
			if m.Name == name {
				return m
			}
		}
		return metricdata.Metrics{}
	}

	// Fixture: count=123, max=65536

	// Check conntrack.entries (Gauge)
	m := findMetric("conntrack.entries")
	if m.Name != "" {
		gauge, ok := m.Data.(metricdata.Gauge[int64])
		if !ok {
			t.Errorf("conntrack.entries is not Gauge[int64], got %T", m.Data)
		} else {
			if len(gauge.DataPoints) > 0 && gauge.DataPoints[0].Value != 123 {
				t.Errorf("conntrack.entries = %d, want 123", gauge.DataPoints[0].Value)
			}
		}
	} else {
		t.Error("conntrack.entries not found")
	}

	// Check conntrack.limit (Gauge)
	m = findMetric("conntrack.limit")
	if m.Name != "" {
		gauge, ok := m.Data.(metricdata.Gauge[int64])
		if !ok {
			t.Errorf("conntrack.limit is not Gauge[int64], got %T", m.Data)
		} else {
			if len(gauge.DataPoints) > 0 && gauge.DataPoints[0].Value != 65536 {
				t.Errorf("conntrack.limit = %d, want 65536", gauge.DataPoints[0].Value)
			}
		}
	} else {
		t.Error("conntrack.limit not found")
	}
}
