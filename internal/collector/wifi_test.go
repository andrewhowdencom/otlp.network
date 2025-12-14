package collector

import (
	"context"
	"path/filepath"
	"testing"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestWifi(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)

	procPath, _ := filepath.Abs("testdata/proc")
	c, err := NewWifi(procPath)
	if err != nil {
		t.Fatalf("failed to create wifi collector: %v", err)
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

	// Fixture wlan0: level=-40 (signal), link=50 (quality)

	// Check wifi.signal (Gauge)
	m := findMetric("wifi.signal")
	if m.Name == "" {
		t.Error("wifi.signal not found")
	} else {
		gauge, ok := m.Data.(metricdata.Gauge[float64])
		if !ok {
			t.Errorf("wifi.signal is not Gauge[float64], got %T", m.Data)
		} else {
			found := false
			for _, dp := range gauge.DataPoints {
				ifv, _ := dp.Attributes.Value("interface")
				if ifv.AsString() == "wlan0" {
					if dp.Value != -40.0 {
						t.Errorf("wlan0 signal = %f, want -40.0", dp.Value)
					}
					found = true
				}
			}
			if !found {
				t.Error("wlan0 signal data point not found")
			}
		}
	}
}
