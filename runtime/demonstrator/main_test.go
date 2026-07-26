package main

import (
	"strings"
	"testing"
)

func TestPermitVerification(t *testing.T) {
	pc := NewPermitController("TARGET-1")
	
	validArtifact := Artifact{
		PlannerID: "AGENT-1",
		TaskType:  "ENERGIZE_HEATER",
		TargetID:  "TARGET-1",
		Limits: map[string]float64{
			"max_temperature_c": 75.0,
		},
	}
	
	validPermit := Permit{
		ArtifactRoot: HashArtifact(validArtifact),
		TargetID:     "TARGET-1",
		IssuerID:     "ADMIN",
		Signature:    "VALID_SIG",
		Nonce:        "nonce-1",
	}
	
	// Test positive case: valid permit passes
	err := pc.Verify(validPermit, validArtifact)
	if err != nil {
		t.Errorf("expected valid permit to pass verification, got: %v", err)
	}
	
	// Test negative vector: replayed nonce rejected
	err = pc.Verify(validPermit, validArtifact)
	if err == nil {
		t.Errorf("expected replayed nonce to be rejected")
	} else if !strings.Contains(err.Error(), "Replay attack detected") {
		t.Errorf("expected replay attack error, got: %v", err)
	}
	
	// Test negative vector: wrong target ID rejected
	badTargetPermit := validPermit
	badTargetPermit.Nonce = "nonce-2"
	badTargetPermit.TargetID = "TARGET-2"
	err = pc.Verify(badTargetPermit, validArtifact)
	if err == nil {
		t.Errorf("expected wrong target ID to be rejected")
	} else if !strings.Contains(err.Error(), "Permit target mismatch") {
		t.Errorf("expected target mismatch error, got: %v", err)
	}
	
	// Test negative vector: tampered artifact rejected
	tamperedArtifact := validArtifact
	tamperedArtifact.Limits = map[string]float64{"max_temperature_c": 150.0}
	
	badRootPermit := validPermit
	badRootPermit.Nonce = "nonce-3"
	err = pc.Verify(badRootPermit, tamperedArtifact)
	if err == nil {
		t.Errorf("expected tampered artifact to be rejected")
	} else if !strings.Contains(err.Error(), "Artifact root mismatch") {
		t.Errorf("expected artifact root mismatch error, got: %v", err)
	}
}

func TestActuatorFailClosed(t *testing.T) {
	actuator := &Actuator{ID: "ACT-1", State: "ON"}
	
	// Actuator should fail closed when permit is denied
	actuator.FailClosed()
	
	if actuator.State != "OFF" {
		t.Errorf("expected actuator to fail closed (OFF), got state: %s", actuator.State)
	}
}
