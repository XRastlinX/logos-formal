package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// PIC-4 Core Types representing the separation of interpretation (Artifact) and authorization (Permit)

type Artifact struct {
	PlannerID string
	TaskType  string
	TargetID  string
	Limits    map[string]float64
}

type Evidence struct {
	ArtifactRoot string
	ValidatorID  string
	AttesterID   string
	Timestamp    string
}

type Permit struct {
	ArtifactRoot string
	TargetID     string
	IssuerID     string
	Signature    string // Simulated crypto signature
	Nonce        string
}

// Helper to simulate cryptographic hashing
func HashArtifact(a Artifact) string {
	record := fmt.Sprintf("%s|%s|%s|%v", a.PlannerID, a.TaskType, a.TargetID, a.Limits)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}
