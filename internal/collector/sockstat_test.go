package collector

import (
	"context"
	"path/filepath"
	"testing"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestSockstat(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)

	procPath, _ := filepath.Abs("testdata/proc")
	c, err := NewSockstat(procPath)
	if err != nil {
		t.Fatalf("failed to create sockstat collector: %v", err)
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

	// Fixture Sockstat: used=100, TCP inuse=10, UDP inuse=5

	m := findMetric("sockets.used")
	if m.Name != "" {
		gauge, ok := m.Data.(metricdata.Gauge[int64])
		if !ok {
			t.Errorf("sockets.used is not Gauge[int64], got %T", m.Data)
		} else {
			if len(gauge.DataPoints) > 0 && gauge.DataPoints[0].Value != 100 {
				t.Errorf("sockets.used = %d, want 100", gauge.DataPoints[0].Value)
			}
		}
	} else {
		t.Error("sockets.used not found")
	}
}
