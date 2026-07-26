package main

const (
	requestSchema = "urn:cyonic:boundary-request:v1"
	permitSchema  = "urn:cyonic:permit:v1"
	receiptSchema = "urn:cyonic:boundary-receipt:v1"
)

// Artifact is an interpretation-layer proposal. It contains no authority.
type Artifact struct {
	Proposer      string `json:"proposer"`
	Action        string `json:"action"`
	Target        string `json:"target"`
	PayloadDigest string `json:"payloadDigest"`
}

// Permit is externally supplied authorization evidence. The service verifies
// it but cannot issue it, consume it, or perform the authorized effect.
type Permit struct {
	Schema         string `json:"schema"`
	ArtifactDigest string `json:"artifactDigest"`
	Action         string `json:"action"`
	Target         string `json:"target"`
	Issuer         string `json:"issuer"`
	ExpiresAt      string `json:"expiresAt"`
	Nonce          string `json:"nonce"`
	Signature      string `json:"signature"`
}

type BoundaryRequest struct {
	Schema    string   `json:"schema"`
	RequestID string   `json:"requestId"`
	Operation string   `json:"operation"`
	Artifact  Artifact `json:"artifact"`
	Permit    Permit   `json:"permit"`
}

type LayerResult struct {
	Status          string `json:"status"`
	AuthorityEffect string `json:"authorityEffect"`
	Detail          string `json:"detail"`
}

type AuthorizationResult struct {
	Status                 string `json:"status"`
	Issuer                 string `json:"issuer,omitempty"`
	Scope                  string `json:"scope,omitempty"`
	ServiceAuthorityEffect string `json:"serviceAuthorityEffect"`
	Detail                 string `json:"detail"`
}

type Friction struct {
	Category string `json:"category"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

type Externality struct {
	Claim  string `json:"claim"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type RoutingResult struct {
	Operation       string `json:"operation"`
	RequiredProfile string `json:"requiredProfile"`
	Decision        string `json:"decision"`
	Forwarded       bool   `json:"forwarded"`
	Detail          string `json:"detail"`
}

type BoundaryReceipt struct {
	Schema          string              `json:"schema"`
	RequestID       string              `json:"requestId"`
	ObservedAt      string              `json:"observedAt"`
	GovernanceState string              `json:"governanceState"`
	AuthorityEffect string              `json:"authorityEffect"`
	ArtifactDigest  string              `json:"artifactDigest,omitempty"`
	Routing         RoutingResult       `json:"routing"`
	Interpretation  LayerResult         `json:"interpretation"`
	Authorization   AuthorizationResult `json:"authorization"`
	Effect          LayerResult         `json:"effect"`
	Verdict         string              `json:"verdict"`
	Friction        Friction            `json:"friction"`
	Externality     Externality         `json:"externality"`
	Limits          []string            `json:"limits"`
}
