# BLEED_OVER_CURRENT_INTELLIGENCE_PROFILE_V0_1

```yaml
status: PROPOSED
authority_effect: NONE
core_operator: β
public_surface: Current Map
substrate: Distributed Current Intelligence Fabric
governing_invariants:
  - provenance_conservation
  - semantic_delta_declaration
  - uncertainty_noninflation
  - authority_nontransfer
  - effect_noncreation
  - temporal_revision
  - contestability
  - rights_preservation
```

## Executive Summary

Bleed-over is formalized as the controlled translation of claims, evidence, methods, and interpretations across epistemic fields—while preserving provenance, uncertainty, temporal history, and authority boundaries. The resulting architecture is the **Distributed Current Intelligence Fabric (DCIF)** and its public-facing expression, **The Current Map**. 

This is not a universal truth machine but a federated, continuously revised, provenance-bearing map of what different parts of society currently claim, know, dispute, institutionalize, and enact.

---

## 1. The Bounded Universe of Intelligence

No system can capture the complete spectrum of intelligence. Absolute completeness is an unfalsifiable claim. The rigorous replacement is a **declared coverage universe**:

$\mathcal{U} = (D, L, J, S, T, R)$

Where:
* $D$ = Domains
* $L$ = Languages
* $J$ = Jurisdictions
* $S$ = Source classes
* $T$ = Historical/current time window
* $R$ = Rights and access envelope

The objective is: **maximum declared coverage + minimum update latency + preserved uncertainty + auditable omissions.**

## 2. Bleed-Over as a Formal Operator ($\beta$)

Let $\mathcal{F}_i$ denote an epistemic field and $c$ be a claim object. The bleed-over operator is:

$\beta_{i \to j}: \mathsf{ClaimPacket}_i \times \mathsf{Context}_j \rightarrow \mathsf{BoundaryObject}_j$

with $\beta_{i \to j}(c, \Gamma_j) = \langle b, \pi, \Delta_m, \Delta_u \rangle$

Where:
* $b$ is the target-context representation
* $\pi$ is the preserved provenance path
* $\Delta_m$ is the declared semantic transformation
* $\Delta_u$ is any change in uncertainty

### The Five Bleed-Over Invariants

1. **Provenance Conservation:** Translation may add provenance, but cannot erase the origin chain.
2. **Authority Non-Transfer:** Translation of evidence does not mint authority.
3. **Effect Non-Creation:** A translation object may inform a decision, but cannot execute it.
4. **Uncertainty Non-Inflation:** The summary may become shorter, but it may not become surer.
5. **Temporal Preservation:** The translation must retain world-time, observation-time, and ingestion-time.

## 3. Connection to the Cubed Bit (Governance Algebra)

The bleed-over operator is constitutionally restricted to the interpretation coordinate:

$(A, I, E) \xrightarrow{\beta} (A, I', E)$

The cubed bit becomes a type constraint on cross-field information transport:
- `translate` may modify Interpretation.
- `translate` may NOT mint Authorization.
- `translate` may NOT produce Effectuation.

## 4. The Selective Epistemic Membrane

The permeability vector $\mathbf{p} = (p_D, p_V, p_I, p_A, p_E)$:
* High permeability to discoverability ($p_D \approx 1$)
* Conditional permeability to evidence ($p_V$)
* Visible permeability to interpretation ($p_I$)
* Zero permeability to implicit authority or effect ($p_A = 0, p_E = 0$)

## 5. The Current Intelligence Packet (ClaimPacket)

Every extracted claim is a typed packet, tracking valid time, observation time, and ingestion time. The system must never silently overwrite a claim. It creates a new version and records its relational links (supersedes, corrects, retracts).

## 6. Plural Reasoning Stack

No single logic can correctly resolve every form of uncertainty. The DCIF exposes a reasoning palette:
- Bayesian inference (Calibrated probability)
- Evidence/belief-function (Ignorance support)
- Dung-style argumentation (Competing arguments)
- AGM-style belief revision (Contradictory evidence)
- Structural causal models (Intervention/policy)
- Provenance semirings (Derivation lineage)
- Bitemporal logic (Historical corrections)
- PIC-4 / Typed capabilities (Governance)

## 7. The Bleed-Over Charter

1. **BO-1 Provenance conservation**
2. **BO-2 Semantic delta declaration**
3. **BO-3 Uncertainty cannot disappear through rhetoric**
4. **BO-4 Interpretation must remain typed**
5. **BO-5 Authority cannot bleed over**
6. **BO-6 Effect requires an external Permit**
7. **BO-7 Counterevidence remains addressable**
8. **BO-8 Correction propagates forward**
9. **BO-9 Popularity is an attention measurement (not truth)**
10. **BO-10 Rights and privacy travel with the data**
11. **BO-11 No covert psychological targeting**

## 8. Current Map V0.1

The first demonstrator is intentionally bounded to the domains of AI governance, agent interoperability, information physics, and provenance.

**Claim Lifecycle mapping to PIC-4:**
- **Draft:** Raw capture or extracted claim
- **Compacted:** Normalized ClaimPacket
- **Sealed:** Provenance, hashes, time, and dependencies bound
- **Applied:** Human-authorized publication into the Current Map

*The bleed-over transformation remains pure. Publication remains effectful and external.*

---
**Conclusion:** Bleed-over is not the erosion of boundaries. It is the art and science of making boundaries permeable to discovery while impermeable to false authority.
