package collector

import (
	"context"
	"path/filepath"
	"testing"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestTCP(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)

	procPath, _ := filepath.Abs("testdata/proc")
	c, err := NewTCP(procPath)
	if err != nil {
		t.Fatalf("failed to create tcp collector: %v", err)
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

	// Fixture TCP: ActiveOpens=10, PassiveOpens=5, CurrEstab=2, RetransSegs=5
	// tcp.connection.current = 2
	// tcp.connection.total = 15
	// tcp.retransmit = 5

	// Check tcp.connection.current
	m := findMetric("tcp.connection.current")
	if m.Name != "" {
		gauge, ok := m.Data.(metricdata.Gauge[int64]) // Gauge? It was ObservableGauge which produces Gauge data
		if !ok {
			// Maybe Sum if it's monotonic? No, current is Gauge.
			// Wait, ObservableGauge produces Gauge or Sum?
			// It produces Gauge.
			t.Errorf("tcp.connection.current is not Gauge[int64], got %T", m.Data)
		} else {
			if len(gauge.DataPoints) > 0 && gauge.DataPoints[0].Value != 2 {
				t.Errorf("tcp.connection.current = %d, want 2", gauge.DataPoints[0].Value)
			}
		}
	} else {
		t.Error("tcp.connection.current not found")
	}

	// Check tcp.connection.total
	m = findMetric("tcp.connection.total") // ObservableCounter -> Sum
	if m.Name != "" {
		sum, ok := m.Data.(metricdata.Sum[int64])
		if !ok {
			t.Errorf("tcp.connection.total is not Sum[int64], got %T", m.Data)
		} else {
			if len(sum.DataPoints) > 0 && sum.DataPoints[0].Value != 15 {
				t.Errorf("tcp.connection.total = %d, want 15", sum.DataPoints[0].Value)
			}
		}
	} else {
		t.Error("tcp.connection.total not found")
	}
}
