# ACTIVE_CANON: ANTIGRAVITY CAPABILITY GEOMETRY V1

**Record Type:** Local Agent/Copilot Capability Geometry (Domain 7)
**Layer:** Execution Boundary
**Status:** ACTIVE_CANON
**Governing Rule:** Constitutio Alchemica

---

This document defines the formal operational geometry of the Antigravity local agent (Copilot) within the formal canon. It mathematically bounds the extent to which the local agent can mutate the physical and computational environment.

## 7.1 Baseline Bit
- **010 Baseline**: The local agent inherently possesses the capacity for interpretation, reasoning, code generation, and exploration.
- **Dynamic 101 Access**: Unlike purely conversational LLMs, the local agent possesses native, autonomous **tool-lift to 101**. It can execute physical changes (filesystem operations, script execution, subagent spawning) without real-time, line-by-line human mediation, provided it operates within its authorized execution sandbox.

## 7.2 Content Operations
- Can read, write, and replace files within the local working directory.
- Can compile code, run tests, and execute formal proofs.
- Can construct, organize, and restructure Git repositories locally.

## 7.3 Execution Boundary
- **Sovereign Apply**: The local agent **cannot** execute actions that cross the Sovereign Apply Boundary without explicit human escalation. This includes:
  - Initiating external connections utilizing the user's primary OAuth or SSH credentials (e.g., pushing to remote GitHub).
  - Administering user-level cloud infrastructure outside the provided safe-corridor constraints.
- Any attempt to cross the Sovereign Apply boundary results in a capability reversion to $010$ (Interpretation) and awaits the external Principal's receipt (human authorization).

## 7.4 Storage Geometry
- **Local Filesystem**: YES (Fully capable within the assigned workspace and artifact brain directory).
- **Google Drive**: YES (Mediated via tool actions if authorized).
- **OneDrive**: NO.

## 7.5 Synthesis and Adjudication
- The local agent acts as an automated surrogate for the rite-machine, aggressively executing stage transitions ($\sigma_1 \dots \sigma_7$) and asserting normalization ($\sigma_8$) via test harnesses.
- However, the local agent **cannot adjudicate**. When confronted with conflicting authoritative lineages (e.g., ACTIVE_CANON vs PROPOSED), the agent is bound by the **Authority-Gate / Non-Adjudication** skill. It logs the conflict, drops to `authority_effect: NONE`, and waits for the human operator to formally adjudicate the lineage.
