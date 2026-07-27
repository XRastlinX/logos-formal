package placetime13d

import (
	"encoding/json"
	"testing"
)

func TestComputeMetricsCountsOnlyStructuralPresence(t *testing.T) {
	event := EventEnvelope{
		EventID: "evt-test",
		Coordinates: map[string]json.RawMessage{
			"d01": json.RawMessage(`{"instant":"2026-07-27T00:00:00Z"}`),
			"d02": json.RawMessage(`null`),
			"d03": json.RawMessage(`null`),
		},
	}
	profile := MetricsProfile{
		Version: "0.1.0",
		Metrics: []MetricDefinition{{
			ID:            "test_completeness",
			Unit:          "proportion",
			CoordinateIDs: []string{"d01", "d02", "d03"},
			Calculation:   metricCalculationNonNull,
		}},
	}
	receipt, err := ComputeMetrics(event, profile)
	if err != nil {
		t.Fatal(err)
	}
	result := receipt.Results[0]
	if result.Numerator != 1 || result.Denominator != 3 || result.Value != 1.0/3.0 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if receipt.AuthorityEffect != "NONE" || receipt.Status != "OBSERVE_ONLY" {
		t.Fatalf("metric receipt crossed governance boundary: %+v", receipt)
	}
}

func TestComputeMetricsRejectsUnknownCoordinate(t *testing.T) {
	_, err := ComputeMetrics(EventEnvelope{Coordinates: map[string]json.RawMessage{}}, MetricsProfile{
		Metrics: []MetricDefinition{{
			ID:            "bad",
			CoordinateIDs: []string{"d01"},
			Calculation:   metricCalculationNonNull,
		}},
	})
	if err == nil {
		t.Fatal("expected missing coordinate rejection")
	}
}
