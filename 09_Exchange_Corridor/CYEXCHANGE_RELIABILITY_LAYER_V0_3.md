# PROPOSED: CyExchange Reliability Layer V0.3

**Record Type:** Architecture Specification
**Layer:** Exchange Corridor
**Status:** PROPOSED
**authority_effect:** NONE

## 1. The Need for Reliability
A system governed by rigorous, immutable logic is vulnerable if its underlying transport and storage are ephemeral. The V0.3 Reliability Layer solidifies the epistemic structures built in V0.2 into crash-safe, deterministic reality.

## 2. Durable Monotonic Transitions
Using a pure Go SQLite backend (`modernc.org/sqlite`), `queue_messages` stores each exact envelope while `queue_events` appends every accepted state transition. Existing events are not rewritten.

## 3. Replay Immunity
The database primary key binds queue kind and `envelope_id`. Duplicate inserts return `ErrReplay`; the HTTP adapter maps a confirmed duplicate to a conflict response. This is local replay resistance for the tested store, not a claim of global replay prevention.

## 4. Zero-Authority Constraints
V0.3 remains a `PROPOSED` local implementation inside `010`. Recovery reconstructs validated queue projections and converts uncertain in-flight states to `OUTCOME_UNKNOWN`; it contains no `101` Apply path.
