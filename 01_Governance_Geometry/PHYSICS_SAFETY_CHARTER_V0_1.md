# PHYSICS SAFETY CHARTER V0.1

**Record Type:** Core Normative Governance Charter
**Layer:** Atomic Governance Invariants (Domain 1)
**Status:** ACTIVE_CANON
**Governing Rule:** Constitutio Alchemica

---

This charter establishes the normative, implementable invariants that govern the boundary between information processing (Q) and physical effectuation (Cy). It maps the abstract Governance Algebra into explicitly enforceable engineering rules, aligning with IETF/NIST standards (RATS, ABAC).

## The 8 Normative Clauses

### 1. Evidence is not authority
Evidence artifacts SHALL NOT by themselves authorize effectuation. The measurement of reality and the authorization to mutate reality remain structurally decoupled.

### 2. External permit required
Any transition that can change the physical or externally observable state of a system SHALL require a fresh external Permit bound to the specific artifact, actuator, nonce, and validity interval. (Enforces $A=1$).

### 3. Model scope required
Any physics-backed validation result SHALL identify the model, assumptions, and validity domain under which the result was computed or measured. (Bounds the Interpretation bit $I$).

### 4. Uncertainty required
Any measurement used in a safety or authorization decision SHALL include an uncertainty statement sufficient for downstream decision-making, in accordance with NIST metrology traceability.

### 5. Freshness required
Evidence and Permits SHALL include freshness material sufficient to prevent replay.

### 6. Traceability required
Measurements used for authorization gating SHALL be traceable to stated references through a documented chain contributing to uncertainty.

### 7. Safe failure
In the absence of a valid Permit or valid evidence, the actuator side SHALL default to no effect. The fail-closed condition represents the absolute baseline ($000$).

### 8. Post-effect witness
Every successful or attempted effectful transition SHALL emit a content-addressed receipt. This establishes the continuous lifecycle from stage through effect.
