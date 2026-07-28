# Ark Codex Intake: Mesh Routing and Pseudometric Cost Algebra

> **Status:** `PROPOSED_REFERENCE_IMPLEMENTATION`
> **Production readiness:** `NOT_ESTABLISHED`
> **Deployment effect:** `NOT_PERFORMED`
> **Authority effect:** `NONE`

See [GOVERNANCE.md](GOVERNANCE.md) for lifecycle and control boundaries and
[TESTED_SCOPE.md](TESTED_SCOPE.md) for the exact claim boundary.

This repository implements a locally simulated, deterministic pseudophysical space for decentralized obligation mediation. It relies on a dual-measure runtime—combining spatial adjacency (topology) and operational cost (pseudometrics)—to enforce fail-closed dependency resolution across distributed agent swarms.

## 1. Data Integrity Under Hostile Conditions

Within the tested local simulation, integrity is maintained through layered controls:

*   **SQLite transactions:** `BEGIN IMMEDIATE … COMMIT` makes the nonce claim, mutation, receipt, response, and final commitment one atomic operation. A failure rolls the entire transaction back.
*   **WAL recovery:** SQLite’s write-ahead log supports recovery after abrupt worker termination. Tests confirmed that uncommitted ledger writes and `CLAIMED` nonces disappeared after restart.
*   **Exact identity binding:** Obligations include the frozen graph digest in their SHA-256 identity.
*   **Authenticated receipts:** HMAC-SHA256 detects modification by parties lacking the shared secret. It authenticates custody; it does not prove scientific truth or provide public non-repudiation.
*   **Idempotent nonces:** Duplicate delivery retrieves the prior committed response without repeating the mutation.
*   **Database constraints and triggers:** These reject graph mismatches, unauthorized transitions, malformed costs, trace deletion, and direct trust-registry modification.
*   **Governed revocation:** Authenticated, monotonically versioned registry updates revoke keys transactionally.
*   **Independent oracle:** The test calculates the graph-permitted affected set separately and requires exact equality with observed terminal nodes. Orthogonal obligations must remain unchanged.

*"Hostile" here means the exercised deterministic faults: delay, loss with fair retry, duplication, reordering, temporary partition, revocation, contention, and local process termination.*

---

## 2. Engineering Constraints and Limitations

The important boundaries are explicitly defined to separate proven mechanics from theoretical networking behavior.

### ESTABLISHED_FOR_TESTED_SCOPE:
*   Local SQLite atomicity and rollback
*   Graph-bound state transitions
*   Deterministic replay
*   Immutable audit traces
*   Bounded routing locality
*   Replicated revocation behavior

### NOT_ESTABLISHED:
*   Production-network liveness
*   Arbitrary topology or workload behavior
*   Byzantine fault tolerance
*   Malicious-node resistance
*   Internet-scale performance
*   Permanent partitions
*   HMAC non-repudiation
*   Universal leak freedom
*   Physical or metaphysical correspondence

*Additional constraints include shared-secret custody, SQLite’s single-writer model, bounded hop counts, locally replicated trust registries, and dependence on declared fair-delivery conditions. The routing algorithms currently live in a deterministic test harness, not a deployed network transport.*

---

## 3. Operational Costs and State Transitions

Operational cost represents whether an obligation remains unresolved or has lawfully crossed its boundary:

*   `SUSPENDED_STATE`
    *   **opCost** = `INFINITY`
    *   **meaning:** mediation remains unresolved.
*   `CLEARED_BY_RECEIPT`
    *   **opCost** = `0`
    *   **meaning:** administrative dependency cleared by a valid bound receipt.
*   `RESOLVED_BY_WITNESS`
    *   **opCost** = finite nonnegative integer
    *   **meaning:** a verified witness produced a bounded resolution.

The permitted transition is mathematically bounded:

$$(\text{SUSPENDED},\infty) \xrightarrow{\text{authenticated, graph-bound receipt}} (\text{TERMINAL},c)$$

A transition is accepted **only** when the obligation is still suspended, its graph digest matches, the receipt key remains locally valid, the cost is legal for the requested terminal state, and the operation occurs inside the protected transaction.

Cost changing to `0` or finite $c$ does not erase history. The receipt and resolved obligation remain immutable, so operational clearance and evidentiary custody stay separate.
