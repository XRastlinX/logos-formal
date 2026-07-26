# PIC-4 STANDARDS MAPPING V1

**Record Type:** Standards Bridge / Engineering Substrate
**Layer:** Builder Layer (Domain 8)
**Status:** ACTIVE_CANON
**Governing Rule:** Constitutio Alchemica

---

This artifact bridges the formal theoretical concepts of the PIC-4 Governance Algebra into deployable IETF/NIST engineering standards. The architecture is framed as a **manifest-level safety lattice** that utilizes existing protocols.

## Conceptual Mapping

| PIC-4 Concept | Standards-facing realization |
|---|---|
| **Pure Constructor** | Deterministic function plus separately attested execution provenance (SLSA) |
| **Cubed-Bit (A, I, E)** | Manifest field or policy bitset used by the verifier and policy engine |
| **policyRoot** | Policy bundle digest; ABAC policy or permit profile digest |
| **operatorRoot** | Hash of constructor spec, tests, and interfaces; correlates to EAT or CoRIM reference values |
| **runtimeRoot** | Attested boot/runtime state from TPM/PSA/TF-M evidence |
| **artifactRoot** | Digest of the canonical payload |
| **Physics Safety Charter Evidence** | EAT/COSE or DSSE/in-toto predicate carrying measurement metadata and uncertainty |
| **Permit** | ABAC decision output, encoded as COSE/CBOR or JWT/JWS and cryptographically bound |
| **receiptRoot** | Post-effect digest of execution receipt |

## Architectural Roles

The PIC-4 pure constructor isolates code computation. The external lifecycle explicitly maps to RATS (Remote ATtestation procedureS) roles:

- **Attester:** Generates the runtime identity, boot state, and evidence roots.
- **Verifier (Validator):** Appraises the evidence (units, safety envelopes, uncertainty) without generating authorization.
- **Relying Party (Permit Controller):** Evaluates ABAC policy to issue the `Permit`, and securely controls actuation.

## Encoding Disciplines

1. **Device-native / Bandwidth-constrained:** Deterministic CBOR (RFC 8949) + COSE (RFC 9052) + EAT structure.
2. **Human-readable / Supply-chain JSON:** JCS (RFC 8785) canonicalization wrapped in DSSE / in-toto envelopes.
