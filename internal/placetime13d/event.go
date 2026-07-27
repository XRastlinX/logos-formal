package placetime13d

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

func LoadEvent(path string) (EventEnvelope, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return EventEnvelope{}, err
	}
	schemaPath, err := findEventSchema(path)
	if err != nil {
		return EventEnvelope{}, err
	}
	if err := ValidateJSONFile(schemaPath, path); err != nil {
		return EventEnvelope{}, err
	}
	var event EventEnvelope
	if err := json.Unmarshal(data, &event); err != nil {
		return EventEnvelope{}, fmt.Errorf("parse event envelope: %w", err)
	}
	return event, nil
}

func ValidateEvent(event EventEnvelope) error {
	var findings []string
	if event.Schema != EventSchema || event.Version != "0.1.0" {
		findings = append(findings, "unsupported event schema or version")
	}
	if event.EventID == "" {
		findings = append(findings, "eventId is required")
	}
	if event.RegistryRoot != GeneratedRegistryRoot {
		findings = append(findings, "registryRoot does not match generated registry")
	}
	if event.BindingPhase != "PROPOSAL" && event.BindingPhase != "WITNESS" {
		findings = append(findings, "bindingPhase must be PROPOSAL or WITNESS")
	}
	findings = append(findings, validateHashRef("artifactRoot", event.ArtifactRoot, "sha256")...)
	findings = append(findings, validateGitBinding(event.BindingPhase, event.GitBinding)...)
	if event.ArtifactRoot != event.GitBinding.ArtifactRoot {
		findings = append(findings, "artifactRoot and gitBinding.artifactRoot must match")
	}

	if len(event.Coordinates) != 13 {
		findings = append(findings, fmt.Sprintf("coordinates contains %d entries, want 13", len(event.Coordinates)))
	}
	for _, dimension := range GeneratedDimensions {
		if _, exists := event.Coordinates[dimension.ID]; !exists {
			findings = append(findings, "coordinates is missing "+dimension.ID)
		}
	}
	for id := range event.Coordinates {
		if !isDimensionID(id) {
			findings = append(findings, "coordinates contains unknown key "+id)
		}
	}

	findings = append(findings, validateCoordinateRelations(event)...)
	findings = append(findings, validateGovernanceFence(event)...)
	for index, receipt := range event.ValidationReceipts {
		findings = append(findings, validateHashRef(fmt.Sprintf("validationReceipts[%d]", index), receipt, "")...)
	}
	if event.PolicyRoot != nil {
		findings = append(findings, validateHashRef("policyRoot", *event.PolicyRoot, "")...)
	}
	if len(findings) > 0 {
		sort.Strings(findings)
		return fmt.Errorf("event envelope rejected:\n- %s", strings.Join(findings, "\n- "))
	}
	return nil
}

func validateGitBinding(phase string, binding GitBinding) []string {
	var findings []string
	if binding.Repository == "" {
		findings = append(findings, "gitBinding.repository is required")
	}
	if binding.ObjectFormat != "sha1" && binding.ObjectFormat != "sha256" {
		findings = append(findings, "gitBinding.objectFormat must be sha1 or sha256")
	}
	if binding.RefAtObservation == "" {
		findings = append(findings, "gitBinding.refAtObservation is required")
	}
	if phase == "WITNESS" && (binding.CommitOID == nil || binding.TreeOID == nil) {
		findings = append(findings, "WITNESS binding requires commitOid and treeOid")
	}
	if phase == "PROPOSAL" && (binding.CommitOID != nil || binding.TreeOID != nil) {
		findings = append(findings, "PROPOSAL binding must not claim commitOid or treeOid before the commit exists")
	}
	if binding.CommitOID != nil && !validOID(*binding.CommitOID, binding.ObjectFormat) {
		findings = append(findings, "gitBinding.commitOid does not match objectFormat")
	}
	if binding.TreeOID != nil && !validOID(*binding.TreeOID, binding.ObjectFormat) {
		findings = append(findings, "gitBinding.treeOid does not match objectFormat")
	}
	for _, parent := range binding.ParentOIDs {
		if !validOID(parent, binding.ObjectFormat) {
			findings = append(findings, "gitBinding.parentOids contains an invalid OID")
		}
	}
	findings = append(findings, validateHashRef("gitBinding.artifactRoot", binding.ArtifactRoot, "sha256")...)
	return findings
}

func validateCoordinateRelations(event EventEnvelope) []string {
	var findings []string
	var certifier struct {
		Presence   string `json:"presence"`
		Identifier string `json:"identifier"`
	}
	if err := json.Unmarshal(event.Coordinates["d11"], &certifier); err != nil {
		findings = append(findings, "d11 certifier_reference is invalid")
	} else if event.AuthorityProof.Presence == "PRESENT" {
		if certifier.Presence != "PRESENT" || certifier.Identifier == "" || certifier.Identifier != event.AuthorityProof.Subject {
			findings = append(findings, "present authorityProof must match a present d11 certifier reference")
		}
	}

	var lineage struct {
		Presence string   `json:"presence"`
		Root     *HashRef `json:"root"`
	}
	if err := json.Unmarshal(event.Coordinates["d08"], &lineage); err != nil {
		findings = append(findings, "d08 lineage_commitment is invalid")
	} else if lineage.Presence == "PRESENT" && lineage.Root == nil {
		findings = append(findings, "present d08 lineage commitment requires a root")
	}

	var entity struct {
		Presence   string `json:"presence"`
		Identifier string `json:"identifier"`
		Namespace  string `json:"namespace"`
	}
	if err := json.Unmarshal(event.Coordinates["d13"], &entity); err != nil {
		findings = append(findings, "d13 entity_binding is invalid")
	} else if entity.Presence == "PRESENT" && (entity.Identifier == "" || entity.Namespace == "") {
		findings = append(findings, "present d13 entity binding requires identifier and namespace")
	}
	return findings
}

func validateGovernanceFence(event EventEnvelope) []string {
	var findings []string
	validEffects := map[string]bool{"NONE": true, "READ": true, "STAGE": true, "APPLY": true}
	if !validEffects[event.RequestedEffect] {
		findings = append(findings, "requestedEffect is invalid")
	}
	validDecisions := map[string]bool{"NONE": true, "OBSERVE_ONLY": true, "REJECT": true, "PERMIT_REQUIRED": true, "PERMITTED": true}
	if !validDecisions[event.Decision] {
		findings = append(findings, "decision is invalid")
	}
	if event.RequestedEffect == "APPLY" {
		if event.PermitRef.Presence != "PRESENT" {
			findings = append(findings, "APPLY requires permitRef presence PRESENT")
		}
		if event.Decision != "PERMITTED" || event.AuthorityEffect != "PERMITTED" {
			findings = append(findings, "APPLY requires PERMITTED decision and authorityEffect")
		}
	} else if event.AuthorityEffect != "NONE" {
		findings = append(findings, "non-APPLY event must have authorityEffect NONE")
	}
	if event.PermitRef.Presence == "PRESENT" && (event.PermitRef.ID == "" || event.PermitRef.Root == nil) {
		findings = append(findings, "present permitRef requires id and root")
	}
	if event.AuthorityProof.Presence == "PRESENT" && (event.AuthorityProof.Subject == "" || event.AuthorityProof.Root == nil) {
		findings = append(findings, "present authorityProof requires subject and root")
	}
	return findings
}

func validateHashRef(field string, value HashRef, requiredAlgorithm string) []string {
	var findings []string
	if value.Alg == "" || value.Value == "" {
		return []string{field + " requires alg and value"}
	}
	if requiredAlgorithm != "" && value.Alg != requiredAlgorithm {
		findings = append(findings, field+" must use "+requiredAlgorithm)
	}
	decoded, err := hex.DecodeString(value.Value)
	if err != nil {
		findings = append(findings, field+" value must be lowercase hexadecimal")
		return findings
	}
	expectedBytes := map[string]int{"sha256": 32, "sha3-256": 32, "sha3-512": 64}
	size, supported := expectedBytes[value.Alg]
	if !supported || len(decoded) != size || value.Value != strings.ToLower(value.Value) {
		findings = append(findings, field+" length or algorithm is invalid")
	}
	return findings
}

func validOID(value, objectFormat string) bool {
	expected := 40
	if objectFormat == "sha256" {
		expected = 64
	}
	if len(value) != expected || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func isDimensionID(value string) bool {
	for _, dimension := range GeneratedDimensions {
		if dimension.ID == value {
			return true
		}
	}
	return false
}

func isNull(value json.RawMessage) bool {
	return len(value) == 0 || string(value) == "null"
}
