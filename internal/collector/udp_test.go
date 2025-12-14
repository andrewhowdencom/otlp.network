package collector

import (
	"context"
	"path/filepath"
	"testing"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestUDP(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)

	procPath, _ := filepath.Abs("testdata/proc")
	c, err := NewUDP(procPath)
	if err != nil {
		t.Fatalf("failed to create udp collector: %v", err)
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

	// Fixture UDP: InDatagrams=500, OutDatagrams=400, InErrors=5, NoPorts=10, RcvbufErrors=2

	// Check udp.packets (Sum)
	m := findMetric("udp.packets")
	if m.Name != "" {
		sum, ok := m.Data.(metricdata.Sum[int64])
		if !ok {
			t.Errorf("udp.packets is not Sum[int64], got %T", m.Data)
		} else {
			foundIn := false
			for _, dp := range sum.DataPoints {
				dir, _ := dp.Attributes.Value("direction")
				if dir.AsString() == "in" {
					if dp.Value != 500 {
						t.Errorf("udp in packets = %d, want 500", dp.Value)
					}
					foundIn = true
				}
			}
			if !foundIn {
				t.Error("udp in packets data point not found")
			}
		}
	} else {
		t.Error("udp.packets not found")
	}

	// Check udp.drops (Sum)
	m = findMetric("udp.drops")
	if m.Name != "" {
		sum, ok := m.Data.(metricdata.Sum[int64])
		if !ok {
			t.Errorf("udp.drops is not Sum[int64], got %T", m.Data)
		} else {
			foundNoPort := false
			for _, dp := range sum.DataPoints {
				reason, _ := dp.Attributes.Value("reason")
				if reason.AsString() == "no_port" {
					if dp.Value != 10 {
						t.Errorf("udp no_port drops = %d, want 10", dp.Value)
					}
					foundNoPort = true
				}
			}
			if !foundNoPort {
				t.Error("udp no_port drops data point not found")
			}
		}
	} else {
		t.Error("udp.drops not found")
	}
}
