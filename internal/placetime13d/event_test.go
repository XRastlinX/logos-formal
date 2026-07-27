package placetime13d

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidProposalEnvelope(t *testing.T) {
	event := validProposalEvent(t)
	if err := ValidateEvent(event); err != nil {
		t.Fatal(err)
	}
}

func TestMissingDimensionIsRejected(t *testing.T) {
	event := validProposalEvent(t)
	delete(event.Coordinates, "d13")
	assertRejectedWith(t, event, "coordinates contains 12 entries", "coordinates is missing d13")
}

func TestCertifierReferenceDoesNotBecomePermit(t *testing.T) {
	event := validProposalEvent(t)
	event.RequestedEffect = "APPLY"
	event.AuthorityProof = AuthorityProof{
		Presence: "PRESENT",
		Subject:  "did:example:principal",
		Root:     hashPointer("b"),
	}
	event.Coordinates["d11"] = raw(t, map[string]any{
		"presence":   "PRESENT",
		"identifier": "did:example:principal",
	})
	assertRejectedWith(t, event, "APPLY requires permitRef presence PRESENT")
}

func TestLineageCommitmentDoesNotBecomePermit(t *testing.T) {
	event := validProposalEvent(t)
	event.RequestedEffect = "APPLY"
	event.Coordinates["d08"] = raw(t, map[string]any{
		"presence":         "PRESENT",
		"root":             HashRef{Alg: "sha256", Value: strings.Repeat("c", 64)},
		"verificationRule": "test-only",
	})
	assertRejectedWith(t, event, "APPLY requires permitRef presence PRESENT")
}

func TestWitnessRequiresActualGitObjectIDs(t *testing.T) {
	event := validProposalEvent(t)
	event.BindingPhase = "WITNESS"
	assertRejectedWith(t, event, "WITNESS binding requires commitOid and treeOid")
}

func validProposalEvent(t *testing.T) EventEnvelope {
	t.Helper()
	artifact := HashRef{Alg: "sha256", Value: strings.Repeat("a", 64)}
	coordinates := map[string]json.RawMessage{
		"d01": raw(t, map[string]any{
			"instant":       "2026-07-27T12:00:00Z",
			"clockSource":   "test-clock",
			"precisionNs":   1,
			"uncertaintyNs": 0,
		}),
		"d02": raw(t, nil),
		"d03": raw(t, nil),
		"d04": raw(t, nil),
		"d05": raw(t, nil),
		"d06": raw(t, nil),
		"d07": raw(t, nil),
		"d08": raw(t, map[string]any{"presence": "ABSENT"}),
		"d09": raw(t, nil),
		"d10": raw(t, nil),
		"d11": raw(t, map[string]any{"presence": "ABSENT"}),
		"d12": raw(t, nil),
		"d13": raw(t, map[string]any{"presence": "ABSENT"}),
	}
	return EventEnvelope{
		Schema:       EventSchema,
		Version:      "0.1.0",
		EventID:      "evt-test-proposal",
		RegistryRoot: GeneratedRegistryRoot,
		BindingPhase: "PROPOSAL",
		ArtifactRoot: artifact,
		ParentEvents: []string{},
		GitBinding: GitBinding{
			Repository:       "https://example.invalid/repo.git",
			ObjectFormat:     "sha1",
			CommitOID:        nil,
			TreeOID:          nil,
			ParentOIDs:       []string{strings.Repeat("1", 40)},
			RefAtObservation: "test",
			ArtifactRoot:     artifact,
		},
		Coordinates:        coordinates,
		RequestedEffect:    "NONE",
		AuthorityProof:     AuthorityProof{Presence: "ABSENT"},
		PermitRef:          PermitReference{Presence: "ABSENT"},
		PolicyRoot:         nil,
		Decision:           "OBSERVE_ONLY",
		AuthorityEffect:    "NONE",
		ValidationReceipts: []HashRef{},
	}
}

func raw(t *testing.T, value any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func hashPointer(character string) *HashRef {
	return &HashRef{Alg: "sha256", Value: strings.Repeat(character, 64)}
}

func assertRejectedWith(t *testing.T, event EventEnvelope, fragments ...string) {
	t.Helper()
	err := ValidateEvent(event)
	if err == nil {
		t.Fatal("expected rejection")
	}
	for _, fragment := range fragments {
		if !strings.Contains(err.Error(), fragment) {
			t.Fatalf("rejection %q does not contain %q", err, fragment)
		}
	}
}
