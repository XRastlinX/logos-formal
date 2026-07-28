# PROPOSED: cyxctl Reliability Suite V0.3

**Record Type:** Conformance Harness
**Layer:** Exchange Corridor
**Status:** PROPOSED
**authority_effect:** NONE

## 1. Overview
The `cyxctl reliability` command acts as a stress-tester for the V0.3 Storage and Discovery layers.

## 2. Bounded Local Probes
The command creates an isolated temporary SQLite store and reports `PASS` only after each corresponding check returns the expected result.

*   **Crash Recovery:** Verifies that a persisted inbox record can be reconstructed into the in-memory projection.
*   **Duplicate Injection:** Fires the identical `EnvelopeID` into the SQLite backend twice, demanding an `ErrReplay` outcome.
*   **Lineage Forgery:** Attempts to register a child envelope referencing a parent `EnvelopeID` that was never ingested. Expects an `ErrLineage` rejection.
*   **Readiness Probe:** Confirms `/healthz` returns unavailable before an explicit readiness transition and succeeds afterward.

The suite records conformance for these fixtures and this build only. It does not prove universal crash safety, global replay immunity, deployment readiness, or authority.
