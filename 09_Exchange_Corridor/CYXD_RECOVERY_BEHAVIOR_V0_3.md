# PROPOSED: Cyxd Recovery Behavior V0.3

**Record Type:** Service Surface Expansion
**Layer:** Exposure Layer
**Status:** PROPOSED
**authority_effect:** NONE

## 1. Daemon Crash Geometry
In the event of physical power loss or process termination, the `cyxd` daemon must reconstruct the local epistemic state precisely as it was prior to the failure.

## 2. The Replay Sequence (`recover.go`)
1.  On `cyxd serve`, the serve path opens the configured `SQLiteStore`.
2.  `recoverInbox()` loads and validates the append-only queue histories.
3.  The recovery package restores the in-memory projection through the dedicated `Restore` entrypoint; handlers are not invoked.
4.  Expired active records become `REJECTED`, and an interrupted `EVALUATING` record becomes `OUTCOME_UNKNOWN`.
5.  **Critical Constraint:** `EFFECT_REQUESTED` and `APPLIED` are rejected by the `010` recovery path rather than restored.

## 3. Liveness Probes (`/healthz`)
Until `recoverInbox()` completes successfully, the readiness flag remains false. The HTTP listener is created only after recovery, and the flag is marked ready immediately before listening. A standalone prober returns `503 Service Unavailable` until that explicit transition.
