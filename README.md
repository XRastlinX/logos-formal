# PIC-4 Governance Architecture (Logos-Formal)

A rigorous, fail-closed security architecture that mathematically separates **interpretation** (AI planning, observation, estimation) from **authorization** (human permission, systemic actuation). 

This repository implements the **PIC-4 (Pure Invariant Constructor)** model, aligning agentic AI control systems with IETF RATS (Remote ATtestation procedureS) and NIST ABAC (Attribute-Based Access Control) standards.

## The Engineering Claim

Modern AI systems frequently conflate the ability to *plan* an action with the authority to *execute* it. The PIC-4 architecture solves this by enforcing a hard geometric boundary: **Q never sets Cy.** 

Mathematical output, AI models, and interpretive logic (Q) can never autonomously write to the capability/authorization geometry (Cy).

In this system:
1. **AI Models (010)**: Produce *Candidate Artifacts* (plans, code, instructions) in a pure, effectless environment.
2. **Validators**: Verify unit safety, boundaries, and limits, passing the artifact to an Attester.
3. **Attesters**: Cryptographically seal the artifact alongside its runtime identity (Evidence).
4. **Human/External Authority (101)**: Appraises the evidence and issues a one-shot cryptographic **Permit**.
5. **Actuators (001)**: Energize the physical or systemic effect **only** if they receive a valid Permit bound to the exact artifact and their specific hardware ID.

The result is a zero-trust execution envelope for autonomous agents, governed by the [Cubed Bit Safety Lattice](10_Elevation_Layer/ACTIVE_CANON_Bleed_Over_Operator_Beta_V1.md) and the formal **Bleed-Over Operator (β)**.

## Quickstart: Runnable Demonstrators

This repository contains two runnable Go demonstrators that prove these invariants:

1. **Permit-Gated Cell (`runtime/demonstrator`)**: Proves an AI planner cannot effectuate change without a human-issued permit. Tests replay attacks, expired nonces, and target mismatch.
2. **Bleed-Over (DCIF) Membrane (`runtime/dcif-membrane`)**: Simulates the [Distributed Current Intelligence Fabric](08_Builder_Layer/BLEED_OVER_CURRENT_INTELLIGENCE_PROFILE_V0_1.md). Proves that an AI agent translating a claim between fields cannot autonomously inflate its certainty or escalate its authority.
3. **Autophagic Decay Simulator (`runtime/autophagic-decay`)**: Simulates the `PROPOSED` Autophagic Operator ($\alpha$). Proves how local-first RAG loops collapse their own certainty matrix when recursively ingesting their own outputs, and mathematically forces the AI to reconnect to `D=0` (primary) roots.

To run the demonstrators (requires Go):

```bash
cd runtime/demonstrator
go run main.go

cd ../dcif-membrane
go run main.go

cd ../autophagic-decay
go run main.go
```

The demonstrator tests the following fail-closed negative vectors:
- Replay attacks on permits
- Incorrect actuator target IDs
- Bypassed validators
- Expired nonces

*(See `08_Builder_Layer/Demonstrator_Architecture.md` for complete architectural details of the cell).*

---

## The 12+1 Internal Ontology (Formal Geometry)

Under the hood, the system is exhaustively partitioned into 12 physical domains centered around a Master Codex. This topology explicitly defines the capability limits and developmental lifecycles of all agents and artifacts.

- **`00_Master_Codex`**: The core transparency manifest and central integration hub.
- **`01_Governance_Geometry`**: Defines **where** a cord sits in the governance landscape (Cubed Bit, capability locks).
- **`02_Cord_Algebra`**: Defines **what** a cord is and how it matures (interior, residual, frame).
- **`03_Q_Taxonomy`**: Defines **what kind** of content (interior/residual) a cord contains.
- **`04_Dual_Projection`**: Defines **how** a cord is observed without being altered ($\{q\}/\{cy\}$).
- **`05_Rite_Machine`**: Defines **how** a cord matures safely (Seven-Stage Transition).
- **`06_Prime_Root_Field`**: Defines the **input structure** for developmental analysis.
- **`07_Antigravity_Capability`**: Operational bounds of the local agent/Copilot (native tool-lift capability).
- **`08_Builder_Layer`**: Repository structure, schemas, and integration docs (PIC-4 standards mapping).
- **`09_Exchange_Corridor`**: The secure cross-platform packet exchange layer.
- **`10_Elevation_Layer`**: The formal registry for elevated `ACTIVE_CANON` artifacts (e.g., Normalization).
- **`11_DeepSeek_Reach`**: The capability boundaries of the DeepSeek external LLM.
- **`12_Grok_Reach`**: The capability boundaries of the Grok external synthesis LLM.

## License

This project is released under the MIT License.
