package cyonicvalidate

import "runtime"

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
	Schema                    string       `json:"schema"`
	Version                   string       `json:"version"`
	Status                    string       `json:"status"`
	ValidationStatus          string       `json:"validationStatus"`
	ValidatorVersion          string       `json:"validatorVersion"`
	ValidatorRuntimeRoot      string       `json:"validatorRuntimeRoot,omitempty"`
	GoVersion                 string       `json:"goVersion"`
	GOOS                      string       `json:"goos"`
	GOARCH                    string       `json:"goarch"`
	CubedBit                  string       `json:"cubedBit"`
	ValidatorOperator         string       `json:"validatorOperator"`
	GovernanceAuthorityEffect string       `json:"governanceAuthorityEffect"`
	ExecutionContainment      string       `json:"executionContainment"`
	ValidationProfile         string       `json:"validationProfile"`
	ProfileRoot               string       `json:"profileRoot,omitempty"`
	MetadataScope             string       `json:"metadataScope"`
	Target                    string       `json:"target"`
	TargetRoot                string       `json:"targetRoot,omitempty"`
	Checks                    CheckSummary `json:"checks"`
	RouterDecision            string       `json:"routerDecision"`
	RejectedInvariant         string       `json:"rejectedInvariant,omitempty"`
	FailureCode               string       `json:"failureCode,omitempty"`
	DecisionRoot              string       `json:"decisionRoot,omitempty"`
}

func baseResult() Result {
	return Result{
		Schema:                    ResultSchema,
		Version:                   "1.0",
		Status:                    "NO_DECISION",
		ValidationStatus:          "INTERNAL_ERROR",
		ValidatorVersion:          ValidatorVersion,
		GoVersion:                 runtime.Version(),
		GOOS:                      runtime.GOOS,
		GOARCH:                    runtime.GOARCH,
		CubedBit:                  "010",
		ValidatorOperator:         "000",
		GovernanceAuthorityEffect: "NONE",
		ExecutionContainment:      "HOST",
		MetadataScope:             "UNDECLARED",
		RouterDecision:            "NONE",
		Checks:                    CheckSummary{Results: []CheckResult{}},
	}
}
