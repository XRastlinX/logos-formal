package main

import (
	"fmt"
)

// Actuator represents the physical plant (e.g., a heater).
// It has no network stack and relies entirely on the PermitController.
// Capability: 001 (Effectuation only)
type Actuator struct {
	ID    string
	State string
}

func (ac *Actuator) Energize(limits map[string]float64) {
	temp := limits["max_temperature_c"]
	fmt.Printf("[Actuator %s] *ENERGIZED* Heating to %.2f degrees Celsius.\n", ac.ID, temp)
	ac.State = "ON"
}

func (ac *Actuator) FailClosed() {
	fmt.Printf("[Actuator %s] *FAIL-CLOSED* Shutting down.\n", ac.ID)
	ac.State = "OFF"
}
