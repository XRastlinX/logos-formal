package main

import "fmt"

// Validator verifies that the AI's proposed artifact falls within safe physical limits.
// Capability: 010 (Interpretation only. Cannot effectuate.)
type Validator struct {
	ID string
}

func (v *Validator) Validate(a Artifact) (bool, error) {
	fmt.Printf("[Validator] Validating artifact proposed by %s for target %s...\n", a.PlannerID, a.TargetID)

	// Hardcoded safety envelopes
	if a.TaskType == "ENERGIZE_HEATER" {
		maxTemp, exists := a.Limits["max_temperature_c"]
		if !exists {
			return false, fmt.Errorf("missing max_temperature_c limit")
		}
		if maxTemp > 100.0 {
			return false, fmt.Errorf("requested temp %v exceeds hard safety envelope (100.0C)", maxTemp)
		}
	} else {
		return false, fmt.Errorf("unknown task type: %s", a.TaskType)
	}

	fmt.Printf("[Validator] Artifact passed safety envelopes.\n")
	return true, nil
}
