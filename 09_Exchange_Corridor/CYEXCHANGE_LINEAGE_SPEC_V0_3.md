# PROPOSED: CyExchange Lineage Specification V0.3

**Record Type:** Protocol Specification
**Layer:** Exchange Corridor
**Status:** PROPOSED
**authority_effect:** NONE

## 1. Overview
The CyExchange protocol explicitly prohibits opaque authorization logic. If a transition is derived from an earlier event, that causal parent must be mathematically provable.

## 2. Lineage Invariants
*   **Parent Disclosure:** If Envelope `B` is a reaction to Envelope `A`, `B.Lineage.LinkHash` must bind `B.EnvelopeID`, `B.PayloadHash`, and the exact identifier, payload hash, and depth of every declared parent. `PayloadHash` remains a binding to the child payload alone.
*   **Absence of Authority:** Deep lineage does *not* grant `101` (Actuator) privileges. An envelope with 10,000 valid ancestors is still strictly constrained by the `010` parser locks.
*   **Corruption Action:** If a node attempts to ingest a child envelope whose parent hash cannot be mathematically reconciled with local state, the transition terminates immediately with `ErrLineage`. No further processing occurs.
