package placetime13d

import "encoding/json"

const (
	RegistrySchema = "urn:logos-formal:placetime-13d-registry:v0.1"
	EventSchema    = "urn:logos-formal:placetime-13d-event:v0.1"
	ProfileSchema  = "urn:logos-formal:placetime-13d-repository-profile:v0.1"
)

type DimensionDescriptor struct {
	ID              string
	Name            string
	CoordinateClass string
}

type HashRef struct {
	Alg   string `json:"alg"`
	Value string `json:"value"`
}

type GitBinding struct {
	Repository       string   `json:"repository"`
	ObjectFormat     string   `json:"objectFormat"`
	CommitOID        *string  `json:"commitOid"`
	TreeOID          *string  `json:"treeOid"`
	ParentOIDs       []string `json:"parentOids"`
	RefAtObservation string   `json:"refAtObservation"`
	ArtifactRoot     HashRef  `json:"artifactRoot"`
}

type AuthorityProof struct {
	Presence string   `json:"presence"`
	Subject  string   `json:"subject,omitempty"`
	Root     *HashRef `json:"root,omitempty"`
}

type PermitReference struct {
	Presence string   `json:"presence"`
	ID       string   `json:"id,omitempty"`
	Root     *HashRef `json:"root,omitempty"`
}

type EventEnvelope struct {
	Schema             string                     `json:"schema"`
	Version            string                     `json:"version"`
	EventID            string                     `json:"eventId"`
	RegistryRoot       string                     `json:"registryRoot"`
	BindingPhase       string                     `json:"bindingPhase"`
	ArtifactRoot       HashRef                    `json:"artifactRoot"`
	ParentEvents       []string                   `json:"parentEvents"`
	GitBinding         GitBinding                 `json:"gitBinding"`
	Coordinates        map[string]json.RawMessage `json:"coordinates"`
	RequestedEffect    string                     `json:"requestedEffect"`
	AuthorityProof     AuthorityProof             `json:"authorityProof"`
	PermitRef          PermitReference            `json:"permitRef"`
	PolicyRoot         *HashRef                   `json:"policyRoot"`
	Decision           string                     `json:"decision"`
	AuthorityEffect    string                     `json:"authorityEffect"`
	ValidationReceipts []HashRef                  `json:"validationReceipts"`
}
