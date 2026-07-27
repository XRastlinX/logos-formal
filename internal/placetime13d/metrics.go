package placetime13d

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

const metricCalculationNonNull = "NON_NULL_COUNT_OVER_DECLARED_COUNT"

type MetricsProfile struct {
	Schema          string             `json:"schema"`
	Version         string             `json:"version"`
	Status          string             `json:"status"`
	AuthorityEffect string             `json:"authorityEffect"`
	Metrics         []MetricDefinition `json:"metrics"`
}

type MetricDefinition struct {
	ID                string   `json:"id"`
	Numerator         string   `json:"numerator"`
	Denominator       string   `json:"denominator"`
	Unit              string   `json:"unit"`
	Scope             string   `json:"scope"`
	ObservationWindow string   `json:"observationWindow"`
	Uncertainty       string   `json:"uncertainty"`
	Source            string   `json:"source"`
	CoordinateIDs     []string `json:"coordinateIds"`
	Calculation       string   `json:"calculation"`
}

type MetricResult struct {
	MetricID        string  `json:"metricId"`
	Numerator       int     `json:"numerator"`
	Denominator     int     `json:"denominator"`
	Value           float64 `json:"value"`
	Unit            string  `json:"unit"`
	AuthorityEffect string  `json:"authorityEffect"`
}

type MetricReceipt struct {
	Schema          string         `json:"schema"`
	EventID         string         `json:"eventId"`
	ProfileVersion  string         `json:"profileVersion"`
	Results         []MetricResult `json:"results"`
	Status          string         `json:"status"`
	AuthorityEffect string         `json:"authorityEffect"`
	NonClaim        string         `json:"nonClaim"`
}

func LoadMetricsProfile(path string) (MetricsProfile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return MetricsProfile{}, err
	}
	var profile MetricsProfile
	if err := json.Unmarshal(content, &profile); err != nil {
		return MetricsProfile{}, fmt.Errorf("decode metrics profile: %w", err)
	}
	if profile.Status != "PROPOSED" || profile.AuthorityEffect != "NONE" {
		return MetricsProfile{}, fmt.Errorf("metrics profile must remain PROPOSED with authorityEffect NONE")
	}
	for _, definition := range profile.Metrics {
		if definition.Calculation != metricCalculationNonNull {
			return MetricsProfile{}, fmt.Errorf("metric %s uses unsupported calculation %s", definition.ID, definition.Calculation)
		}
		if len(definition.CoordinateIDs) == 0 {
			return MetricsProfile{}, fmt.Errorf("metric %s has no coordinate IDs", definition.ID)
		}
	}
	return profile, nil
}

func ComputeMetrics(event EventEnvelope, profile MetricsProfile) (MetricReceipt, error) {
	results := make([]MetricResult, 0, len(profile.Metrics))
	for _, definition := range profile.Metrics {
		present := 0
		for _, coordinateID := range definition.CoordinateIDs {
			value, exists := event.Coordinates[coordinateID]
			if !exists {
				return MetricReceipt{}, fmt.Errorf("metric %s references missing coordinate %s", definition.ID, coordinateID)
			}
			if !isJSONNull(value) {
				present++
			}
		}
		denominator := len(definition.CoordinateIDs)
		results = append(results, MetricResult{
			MetricID:        definition.ID,
			Numerator:       present,
			Denominator:     denominator,
			Value:           float64(present) / float64(denominator),
			Unit:            definition.Unit,
			AuthorityEffect: "NONE",
		})
	}
	return MetricReceipt{
		Schema:          "urn:logos-formal:placetime-13d-metric-receipt:v0.1",
		EventID:         event.EventID,
		ProfileVersion:  profile.Version,
		Results:         results,
		Status:          "OBSERVE_ONLY",
		AuthorityEffect: "NONE",
		NonClaim:        "Structural completeness only; this receipt does not measure truth, quality, validity, or authority.",
	}, nil
}

func isJSONNull(value json.RawMessage) bool {
	return len(value) == 0 || bytes.Equal(bytes.TrimSpace(value), []byte("null"))
}
