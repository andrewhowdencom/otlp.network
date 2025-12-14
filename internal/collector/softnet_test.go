package collector

import (
	"context"
	"path/filepath"
	"testing"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestSoftnet(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)

	procPath, _ := filepath.Abs("testdata/proc")
	c, err := NewSoftnet(procPath)
	if err != nil {
		t.Fatalf("failed to create softnet collector: %v", err)
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

	// Fixture Softnet: 2 CPUs. Each: processed=100 (0x64), dropped=1 (0x1), squeezed=2 (0x2)
	// Total: processed=200, dropped=2, squeezed=4

	// Check softnet.processed (Sum)
	m := findMetric("softnet.processed")
	if m.Name != "" {
		sum, ok := m.Data.(metricdata.Sum[int64])
		if !ok {
			t.Errorf("softnet.processed is not Sum[int64], got %T", m.Data)
		} else {
			if len(sum.DataPoints) > 0 && sum.DataPoints[0].Value != 200 {
				t.Errorf("softnet.processed = %d, want 200", sum.DataPoints[0].Value)
			}
		}
	} else {
		t.Error("softnet.processed not found")
	}
}
