package cyonicvalidate

const (
	ValidatorVersion = "0.1.0"

	ManifestSchema = "urn:cyonic:validation-manifest:v1"
	ResultSchema   = "urn:cyonic:validation-result:v1"

	CheckTargetBoundary     = "CV-TARGET-BOUNDARY"
	CheckRequiredPaths      = "CV-REQUIRED-PATHS"
	CheckGoVet              = "CV-GO-VET"
	CheckGoTestFresh        = "CV-GO-TEST-FRESH"
	CheckTargetImmutability = "CV-TARGET-IMMUTABILITY"
)

const (
	ExitValidated         = 0
	ExitRejected          = 10
	ExitInvalidTarget     = 11
	ExitInvalidManifest   = 12
	ExitMissingDependency = 13
	ExitInternalFailure   = 14
)

var requiredCheckOrder = []string{
	CheckTargetBoundary,
	CheckRequiredPaths,
	CheckGoVet,
	CheckGoTestFresh,
	CheckTargetImmutability,
}

type Manifest struct {
	SchemaRef     string   `json:"$schema,omitempty"`
	Schema        string   `json:"schema"`
	Version       string   `json:"version"`
	Profile       string   `json:"profile"`
	Target        string   `json:"target"`
	RequiredPaths []string `json:"requiredPaths"`
	Checks        []string `json:"checks"`
}

type CheckResult struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type CheckSummary struct {
	Passed  int           `json:"passed"`
	Total   int           `json:"total"`
	Results []CheckResult `json:"results"`
}

type Result struct {
	Schema            string       `json:"schema"`
	Version           string       `json:"version"`
	Status            string       `json:"status"`
	ValidatorVersion  string       `json:"validatorVersion"`
	CubedBit          string       `json:"cubedBit"`
	ValidatorOperator string       `json:"validatorOperator"`
	AuthorityEffect   string       `json:"authorityEffect"`
	ValidationProfile string       `json:"validationProfile"`
	MetadataScope     string       `json:"metadataScope"`
	Target            string       `json:"target"`
	TargetRoot        string       `json:"targetRoot,omitempty"`
	Checks            CheckSummary `json:"checks"`
	RouterDecision    string       `json:"routerDecision"`
	RejectedInvariant string       `json:"rejectedInvariant,omitempty"`
	DecisionRoot      string       `json:"decisionRoot,omitempty"`
}

func baseResult() Result {
	return Result{
		Schema:            ResultSchema,
		Version:           "1.0",
		Status:            "REJECTED",
		ValidatorVersion:  ValidatorVersion,
		CubedBit:          "010",
		ValidatorOperator: "000",
		AuthorityEffect:   "NONE",
		MetadataScope:     "UNDECLARED",
		RouterDecision:    "REJECT",
	}
}
