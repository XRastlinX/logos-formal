# PIC-4 SCHEMAS V1

**Record Type:** Concrete Implementation Schemas
**Layer:** Builder Layer (Domain 8)
**Status:** PROPOSED
**target_runtime:** JSON / YAML / CBOR

---

These schemas provide the concrete implementation types for the PIC-4 governance structure.

## 1. PIC4Descriptor (YAML)

Provides the explicit identity, invariance, and determinism claims for the constructor.

```yaml
descriptorType: "urn:pic4:descriptor:v1"
operatorId: "kappa-cord"
operatorVersion: "1.0.0"
inputSchema: "urn:pic4:schema:kappa-input:v1"
outputSchema: "urn:pic4:schema:kappa-cord:v1"
invariantSchema: "urn:pic4:invariant:cord-balance:v1"
effectClass: "none"
determinism: "deterministic"
canonicalization: "JCS_RFC8785"
operatorRoot:
  alg: "sha3-512"
  value: "..."
testVectorRoot:
  alg: "sha3-512"
  value: "..."
policyProfile: "urn:psc:profile:charter:v0.1"
measurementProfile: "urn:psc:profile:metrology:v0.1"
cubedBitPolicy:
  allowEffectOnlyWhen:
    authorization: 1
  forbid:
    - { authorization: 0, effectuation: 1 }
```

## 2. Root Manifest (JSON)

Binds the dependencies of an attestation package. Should use JCS (RFC 8785) prior to hashing.

```json
{
  "schema": "urn:pic4:root-bundle:v1",
  "profile": "urn:psc:metrology-attestation:v0.1",
  "hashAlgorithm": "sha3-512",
  "policyRoot": { "alg": "sha3-512", "value": "..." },
  "operatorRoot": { "alg": "sha3-512", "value": "..." },
  "runtimeRoot": { "alg": "sha3-512", "value": "..." },
  "artifactRoot": { "alg": "sha3-512", "value": "..." }
}
```

## 3. Sealed Artifact in-toto Predicate (JSON)

An in-toto Statement wrapper for provenance.

```json
{
  "_type": "https://in-toto.io/Statement/v1",
  "subject": [
    {
      "name": "pic4-artifact",
      "digest": {
        "sha3-512": "..."
      }
    }
  ],
  "predicateType": "urn:pic4:sealed-artifact:v1",
  "predicate": {
    "descriptorRoot": { "alg": "sha3-512", "value": "..." },
    "policyRoot": { "alg": "sha3-512", "value": "..." },
    "runtimeRoot": { "alg": "sha3-512", "value": "..." },
    "governanceBits": { "A": 0, "I": 0, "E": 0 },
    "lifecycleState": "sealed",
    "measurementProfile": "urn:psc:metrology:v0.1",
    "modelRoot": { "alg": "sha3-512", "value": "..." },
    "uncertaintyStatementRoot": { "alg": "sha3-512", "value": "..." }
  }
}
```

## 4. Physics Evidence Schema (JSON)

Complies with NIST traceability and uncertainty rules.

```json
{
  "schema": "urn:psc:physics-evidence:v0.1",
  "measurand": "heater.energy",
  "value": 82.4,
  "unit": "mJ",
  "combinedStandardUncertainty": 1.7,
  "coverage": 0.95,
  "rawDataRoot": { "alg": "sha3-256", "value": "..." },
  "instrumentRoot": { "alg": "sha3-256", "value": "..." },
  "firmwareRoot": { "alg": "sha3-256", "value": "..." },
  "calibrationChainRoot": { "alg": "sha3-256", "value": "..." },
  "modelRoot": { "alg": "sha3-256", "value": "..." },
  "validityWindow": {
    "notBefore": "2026-07-26T15:00:00Z",
    "notAfter": "2026-07-26T15:05:00Z"
  },
  "nonce": "base64url..."
}
```

## 5. Permit Token (JSON / COSE target)

The final ABAC authority decision explicitly bound to an actuator and an artifact.

```json
{
  "profile": "urn:psc:permit:v1",
  "iss": "permit-authority.example",
  "sub": "actuator:heater-cell-01",
  "aud": "permit-controller:cell-01",
  "exp": 1785100000,
  "nonce": "base64url...",
  "artifactRoot": { "alg": "sha3-256", "value": "..." },
  "policyRoot": { "alg": "sha3-256", "value": "..." },
  "runtimeRoot": { "alg": "sha3-256", "value": "..." },
  "limits": {
    "maxCurrent_mA": 500,
    "maxEnergy_mJ": 100,
    "maxTemp_C": 60
  },
  "governanceBits": { "A": 1, "I": 0, "E": 1 }
}
```
