# PROPOSED: Cyonic Service Surface — Long-Term Architectural Vision

**Record Type:** Aspirational Architecture / Strategic Vision
**Layer:** Builder, Exchange, and Sustainability
**Status:** PROPOSED (Vision)
**Authority Effect:** NONE
**Governing Rule:** Constitutio Alchemica

---

> [!IMPORTANT]
> This document describes aspirational long-term architecture. It does not describe current capabilities. The distinction between what exists (the reference demonstrators) and what is envisioned (a production service surface) must remain explicit at all times.

## 1. Computational and Protocol Substrate (The Exchange Layer)

The long-term foundation relies on escaping monolithic, cloud-hosted AI paradigms in favor of a decentralized, locally executing network of agents.

- **CyExchange Protocol:** The surface should be built upon the `09_Exchange_Corridor`, functioning as a secure, local-to-local cross-platform packet exchange layer.
- **Fail-Closed Daemon:** The core runtime should be a hardened `cyxd` daemon relying on a pure `DecisionKernel` that never performs I/O, utilizing SQLite in WAL mode for persistent, concurrent node-local inbox/outbox state.
- **Secure Transport:** Inter-agent communication should eventually transition from local mDNS discovery to hardened gRPC over mTLS transport, ensuring that unauthenticated peers cannot exchange payloads.

**Current status:** The `runtime/cyonic-service` provides a read-only Ed25519 boundary service. It verifies permit evidence but does not issue, consume, or Apply permits. The daemon, mTLS transport, and exchange protocol do not yet exist.

## 2. Epistemic and Mathematical Governance (The Law Layer)

The service surface should be governed by formally specified invariants rather
than relying on heuristic prompt engineering alone.

- **The κ (kappa) Cord Operator:** Where a transition lies inside its declared
  dimensional domain, κ may be evaluated within a Cylang computational
  substrate under an explicit calibration contract. κ is not a universal
  constructor for all agent state transitions.
- **Four-Root Locality-Evidence Model:** The execution context should be cryptographically sealed using four domain-separated roots: `policyRoot` (human authorization), `operatorRoot` (physical implementation/code), `runtimeRoot` (execution environment), and `artifactRoot` (payload commitment).
- **Epistemic Tracking:** Candidate self-modeling metrics may measure declared
  forms of semantic change and deviation. No metric may trigger an automatic
  halt until its observable, calibration set, error model, false-positive cost,
  recovery path, and external authority have been specified and tested.

**Current status:** The Cubed Bit lattice and β operator are ACTIVE_CANON. The α operator is PROPOSED. The κ Cord Operator and Cylang substrate are architectural goals, not implementations. The four-root model aligns with the existing validator receipt structure but is not yet cryptographically enforced across separate trust domains.

## 3. Hierarchical Execution and Safety (The Workflow Layer)

To prevent decision-space explosion, the surface should utilize a rigid, topological execution model.

- **Hierarchical Skill-Based Architecture:** Capabilities should be organized into a formal tree, strictly segregating Decision-Making Skills (internal orchestrators that plan but cannot physically act) from Executable Skills (terminal leaf nodes that actuate deterministically but cannot reason).
- **Pushdown Automaton Memory:** Agent memory should be constrained by a LIFO Skill Path Stack, safely compartmentalizing active context and preventing cross-branch memory contamination.
- **PIC-4 Governance Algebra:** The reference implementations demonstrate
  rejection behavior for declared `010` paths under the tested software model.
  The Cubed Bit classification does not by itself prove non-interference or
  physical safety. Any `010 -> 101` transition requires separately supplied,
  externally issued authorization and a bounded effect handler.

**Current status:** PIC-4 is implemented in the reference demonstrators. The hierarchical skill architecture and pushdown automaton memory are design patterns, not running code.

## 4. Commercial Viability (The Enterprise Layer)

For the service surface to survive long-term, it must demonstrate a clear pathway to economic sustainability.

- **Strategic Repository Governance:** Organization ownership may be adopted
  when teams, organization-scoped rules, delegated bypass administration, or
  other named controls are actually required. It is not a prerequisite for the
  present public reference surface.
- **Dual-Licensing Model:** Base specifications and validation engines under permissive open-source licenses (MIT); optional proprietary commercial tier with specific institutional SLAs for regulated industries.
- **Continuous CI/CD Integration:** Agentic capabilities should be natively compiled from declarative specifications into deterministic lock files, executing within standard CI workflows with strict telemetry for computational cost control.

**Current status:** The repository is under a personal account with MIT license. Dual licensing is a Cy-Egology PROPOSED strategy. CI runs `go vet` and `go test ./...` on push/PR. No organizational migration, SLA framework, or enterprise tier exists.

## 5. External Stabilization (The Social Layer)

As the network scales, it must resist institutional degeneracy and capture.

- **Axiom of Consent and Friction Dynamics:** Consent and coordination-friction
  measures remain proposed research. They may not affect authority until their
  observables, scales, sampling procedure, uncertainty, and decision boundary
  are operationally defined and validated.
- **Reality Contact and External Verification:** The service surface must maintain independent verification channels entirely decoupled from the core execution thread, preventing the system from suppressing its own compliance signals.

**Current status:** The First Contact Trial infrastructure provides a structured external verification channel. The friction dynamics and consent axioms are theoretical constructs, not implementations.

---

## Epistemic Guard

This document describes five architectural pillars for a long-term vision. The language throughout uses "should" and "must eventually" rather than "does" or "has been proven." The following distinctions are mandatory:

| Claim Level | Current Evidence |
|---|---|
| PIC-4 governance flow | Tested rejection behavior in the reference implementation; physical safety not established |
| Ed25519 authorization-evidence verification | Demonstrated for the bounded Cyonic Boundary Service fixture; issuance and Apply are absent |
| β operator invariants | **ACTIVE_CANON** with test coverage |
| α operator | **PROPOSED** research |
| κ Cord Operator / Cylang | **Aspirational** — no implementation exists |
| cyxd daemon / mTLS / mDNS | **Aspirational** — no implementation exists |
| Hierarchical skill tree / pushdown automaton | **Design pattern** — not running code |
| Dual licensing / enterprise SLAs | **PROPOSED** strategy — no commercial activity |
| Institutional standards engagement | **Phase 5** — no external institution has engaged |

No aspirational element may be described as demonstrated, proven, or established until it passes the evidence gates defined in the Cy-Egology Adoption Arc V1.

---

**Status:** PROPOSED vision document.
It does not authorize any implementation, commercial activity, or standards claim.

```text
authority_effect: NONE
scientific_claim: NONE
physical_guarantee: NONE
historical_identity_claim: NONE
```
