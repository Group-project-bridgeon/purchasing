package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestMetrics_RecordHTTPRequest(t *testing.T) {
	RecordHTTPRequest("POST", "/api/v1/purchase-orders", 201, 45*time.Millisecond)

	metric := &dto.Metric{}
	err := HTTPRequestsTotal.WithLabelValues("POST", "/api/v1/purchase-orders", "201").Write(metric)
	if err != nil {
		t.Fatalf("failed to read metric: %v", err)
	}

	if val := metric.GetCounter().GetValue(); val < 1 {
		t.Errorf("expected counter >= 1, got %v", val)
	}
}

func TestMetrics_RecordOutbox(t *testing.T) {
	RecordOutboxEvent("dispatched")
	RecordOutboxEvent("failed")
	RecordOutboxBatchDuration(12 * time.Millisecond)

	metric := &dto.Metric{}
	err := OutboxEventsTotal.WithLabelValues("dispatched").Write(metric)
	if err != nil {
		t.Fatalf("failed to read dispatched metric: %v", err)
	}
	if val := metric.GetCounter().GetValue(); val < 1 {
		t.Errorf("expected dispatched counter >= 1, got %v", val)
	}

	err = OutboxEventsTotal.WithLabelValues("failed").Write(metric)
	if err != nil {
		t.Fatalf("failed to read failed metric: %v", err)
	}
	if val := metric.GetCounter().GetValue(); val < 1 {
		t.Errorf("expected failed counter >= 1, got %v", val)
	}
}

func TestDBPoolCollector_NilSafe(t *testing.T) {
	collector := RegisterDBPoolCollector(nil)
	if collector == nil {
		t.Fatal("expected non-nil collector")
	}

	// Calling second time shouldn't panic or error
	collector2 := RegisterDBPoolCollector(nil)
	if collector2 != collector {
		t.Errorf("expected same collector instance")
	}

	descCh := make(chan *prometheus.Desc, 10)
	collector.Describe(descCh)
	if len(descCh) != 4 {
		t.Errorf("expected 4 descriptions, got %d", len(descCh))
	}

	metricCh := make(chan prometheus.Metric, 10)
	collector.Collect(metricCh)
	if len(metricCh) != 0 {
		t.Errorf("expected 0 metrics collected from nil pool, got %d", len(metricCh))
	}
}
