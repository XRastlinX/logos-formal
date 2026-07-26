# Logos-Formal

**Fail-closed separation of AI interpretation from human authorization.**

An AI agent operating under this architecture cannot authorize or execute its own plans. Every effectual action requires a cryptographically bound permit issued by a human operator or external policy engine.

[![Validate Canon](https://github.com/XRastlinX/logos-formal/actions/workflows/validate.yml/badge.svg)](https://github.com/XRastlinX/logos-formal/actions/workflows/validate.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## The Problem

Modern AI systems increasingly conflate the ability to *plan* an action with the authority to *execute* it. Tool-calling LLMs can generate infrastructure commands, deployment scripts, and database mutations — and tooling increasingly lets them run those plans autonomously. The boundary between "what should we do?" and "do it" is collapsing.

In safety-critical domains — industrial control, medical devices, supply-chain integrity, autonomous vehicles — this collapse is unacceptable.

## The Architecture

This repository implements **PIC-4** (Pure Invariant Constructor), a governance architecture built on a single hard invariant:

> **Q never sets Cy.**
>
> Mathematical output, AI models, and interpretive logic (Q) can never autonomously write to the capability/authorization geometry (Cy).

The system enforces this through a five-stage pipeline aligned with [IETF RATS](https://datatracker.ietf.org/doc/rfc9334/) and [NIST ABAC](https://csrc.nist.gov/pubs/sp/800/162/final):

```
AI (010)          Validator          Attester          Human/Policy (101)     Actuator (001)
   │                  │                 │                    │                    │
   ├─ Propose ───────►│                 │                    │                    │
   │                  ├─ Check ────────►│                    │                    │
   │                  │                 ├─ Seal Evidence ───►│                    │
   │                  │                 │                    ├─ Issue Permit ────►│
   │                  │                 │                    │                    ├─ Execute
   │                  │                 │                    │                    │
   │            REJECT if unsafe  REJECT if untested  REJECT if unauthorized  FAIL CLOSED
```

The AI operates at coordinate **010** (pure interpretation). Only a human authority at **101** can issue a permit. If any stage fails, the actuator fails closed — it does nothing.

## Try It in 10 Minutes (First Contact Protocol)

This repository enforces Phase 1 of the Cy-Egology adoption arc (external verification). The core `010` observer mechanism is published for independent cold-run verification. 

This is not a request for endorsement or agreement. The trial makes no scientific or authority claim and performs no authorized effect.

**Prerequisites:** [Go 1.24+](https://go.dev/dl/)

```bash
# 1. Clone the repository and checkout the pinned trial commit
git clone https://github.com/XRastlinX/logos-formal.git
cd logos-formal
git checkout 7e0f263c2a31eaece5bf8ded76ce4ba47c7db32a

# 2. Run the deterministic observer trial
./demo.sh          # Linux / macOS
.\demo.ps1         # Windows PowerShell
```

**Reporting Friction:**
If you execute this trial, please log your result on [Issue #2](https://github.com/XRastlinX/logos-formal/issues/2) containing:
- Elapsed time
- One sentence explaining what you believe the surface does
- Anything that blocked or confused you, or “none observed.”

Or run individually:

```bash
go test ./...                              # Run all 14 tests
cd runtime/demonstrator && go run .        # Permit-gated actuation
cd runtime/dcif-membrane && go run .       # Bleed-over epistemic membrane
cd runtime/autophagic-decay && go run .    # Autophagic decay (PROPOSED)
```

### What the Demonstrators Show

| Demonstrator | Tests | Status |
|---|---|---|
| **Permit-Gated Cell** | Replay attacks, wrong targets, tampered artifacts, expired nonces | Reference |
| **Bleed-Over (β) Membrane** | Certainty inflation, provenance erasure, authority escalation | Reference |
| **Autophagic Decay (α)** | Echo-chamber collapse in recursive RAG loops | PROPOSED |

## Formal Boundaries (ACTIVE_CANON)

These documents define the system's mathematical invariants:

| Document | Description |
|---|---|
| [Cubed Bit Safety Lattice](01_Governance_Geometry/governance_algebra_v1.md) | The (A, I, E) state lattice — 8 vertices of governance |
| [Bleed-Over Operator β](10_Elevation_Layer/ACTIVE_CANON_Bleed_Over_Operator_Beta_V1.md) | Governed cross-field translation preserving provenance and uncertainty |
| [Normalization Criteria](10_Elevation_Layer/ACTIVE_CANON_Normalization_Criteria_v1.md) | Conditions for elevating artifacts through the Rite Machine |
| [PIC-4 Standards Mapping](08_Builder_Layer/PIC4_STANDARDS_MAPPING_V1.md) | Bridge to IETF RATS, NIST ABAC, SLSA, in-toto |
| [Cybot Pattern](08_Builder_Layer/CYBOT_PATTERN_V1.md) | Local-first AI assistant architecture |

### Research (PROPOSED — not elevated)

| Document | Description |
|---|---|
| Autophagic Operator (α) | Epistemic half-life boundary for recursive AI self-ingestion |

The α operator is under active research. It is used as a practical filter but has not been elevated to ACTIVE_CANON. See [docs/WHY_THIS_MATTERS.md](docs/WHY_THIS_MATTERS.md) for context.

## Standards Alignment

| Standard | PIC-4 Role |
|---|---|
| [IETF RATS (RFC 9334)](https://datatracker.ietf.org/doc/rfc9334/) | Attester/Verifier/Relying Party architecture |
| [NIST ABAC (SP 800-162)](https://csrc.nist.gov/pubs/sp/800/162/final) | Attribute-based permit issuance |
| [in-toto](https://in-toto.io/) | Supply-chain attestation layout |
| [SLSA](https://slsa.dev/) | Build provenance and source integrity |

## Documentation

- [Why This Matters](docs/WHY_THIS_MATTERS.md) — Plain-language explainer
- [Architecture Diagrams](docs/ARCHITECTURE.md) — Visual system overview
- [Positioning](docs/POSITIONING.md) — How PIC-4 addresses funded problem domains
- [Session Anchor](docs/SESSION_ANCHOR.md) — Start-here prompt for new AI sessions
- [Contributing](CONTRIBUTING.md)

## Repository Structure

```
logos-formal/
├── 00_Master_Codex/           # Core transparency manifest
├── 01_Governance_Geometry/    # Cubed Bit Safety Lattice (A, I, E)
├── 02_Cord_Algebra/           # Structural algebra (interior, residual, frame)
├── 03_Q_Taxonomy/             # Content taxonomy
├── 04_Dual_Projection/        # Observation without mutation
├── 05_Rite_Machine/           # Seven-stage artifact maturation
├── 06_Prime_Root_Field/       # Input structure for developmental analysis
├── 07_Antigravity_Capability/ # Local agent operational bounds
├── 08_Builder_Layer/          # Standards mapping, schemas, Cybot pattern
├── 09_Exchange_Corridor/      # Cross-platform packet exchange
├── 10_Elevation_Layer/        # ACTIVE_CANON registry
├── runtime/
│   ├── demonstrator/          # Permit-Gated Cell
│   ├── dcif-membrane/         # Bleed-Over Membrane
│   └── autophagic-decay/      # Epistemic Decay Simulator
├── docs/                      # Plain-language documentation
└── go.mod
```

## What Is Not Claimed

- The demonstrators are **reference implementations**, not production-hardened systems.
- Signatures are currently **simulated** (literal string comparison, not Ed25519/ECDSA).
- The Autophagic Operator (α) is **PROPOSED research**, not proven.
- The architecture is a **framework and formal boundary**, not a finished product.
- Independent custody separation (separate issuer/verifier processes) is **not yet implemented**.

The next assurance tier requires real cryptographic signatures, canonical serialization, separate trust domains, and persistent nonce storage.

## License

[MIT License](LICENSE) — Copyright (c) 2026 Charles LeRoy McClure II

## Citation

If you use this work in research, please cite:

```bibtex
@software{mcclure2026logosformal,
  author = {McClure II, Charles LeRoy},
  title = {Logos-Formal: Fail-Closed Separation of AI Interpretation from Authorization},
  year = {2026},
  url = {https://github.com/XRastlinX/logos-formal},
  license = {MIT}
}
```
