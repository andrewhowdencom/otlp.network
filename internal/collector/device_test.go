package collector

import (
	"context"
	"path/filepath"
	"testing"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestDevice(t *testing.T) {
	// Setup Manual Reader and Provider
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	// Set Global Provider (since collector usages otel.Meter)
	otel.SetMeterProvider(provider)

	// Path to testdata
	procPath, _ := filepath.Abs("testdata/proc")

	// Initialize Collector
	c, err := NewDevice(procPath)
	if err != nil {
		t.Fatalf("failed to create device collector: %v", err)
	}

	// Start Collector
	if err := c.Start(context.Background()); err != nil {
		t.Fatalf("failed to start collector: %v", err)
	}

	// Collect Metrics
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("failed to collect metrics: %v", err)
	}

	// Verify Metrics
	if len(rm.ScopeMetrics) == 0 {
		t.Fatal("no scope metrics found")
	}

	metrics := rm.ScopeMetrics[0].Metrics

	// Helper to find metric
	findMetric := func(name string) metricdata.Metrics {
		for _, m := range metrics {
			if m.Name == name {
				return m
			}
		}
		return metricdata.Metrics{}
	}

	// Check device.io
	m := findMetric("device.io")
	if m.Name == "" {
		t.Error("metric device.io not found")
	} else {
		// Assert type is Sum
		sum, ok := m.Data.(metricdata.Sum[int64])
		if !ok {
			t.Errorf("device.io is not Sum[int64], got %T", m.Data)
		} else {
			// Check data points
			// Fixture: eth0 rx=5000, tx=2000
			// We verify at least one point
			foundEth0 := false
			for _, dp := range sum.DataPoints {
				ifv, _ := dp.Attributes.Value("interface")
				dir, _ := dp.Attributes.Value("direction")
				if ifv.AsString() == "eth0" && dir.AsString() == "receive" {
					if dp.Value != 5000 {
						t.Errorf("eth0 receive bytes = %d, want 5000", dp.Value)
					}
					foundEth0 = true
				}
			}
			if !foundEth0 {
				t.Error("eth0 receive data point not found")
			}
		}
	}
}
