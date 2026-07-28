# Tested Scope

The checked-in suite exercises:

- SQLite schema migration from version 2 to version 3 while preserving legacy
  receipt bytes as read-only, unbound custody records;
- exact graph binding for version 0.2 closure receipts;
- single-transaction nonce claim, mutation, receipt persistence, response
  persistence, and commit;
- authenticated, monotonically versioned registry updates and transactional
  key revocation;
- deterministic replay without duplicate mutation;
- immutable trace and legal operational-cost transitions;
- 50 independent SQLite stores under deterministic delay, duplication,
  reordering, partition, revocation, contention, and abrupt worker termination;
- exact equality between the independently calculated permitted affected set
  and the observed terminal set; and
- preservation of an orthogonal obligation during scoped closure propagation.

## Explicitly not established

- production-network liveness;
- arbitrary topology, scale, or workload behavior;
- Byzantine fault tolerance or malicious-node resistance;
- permanent-partition recovery;
- universal leak freedom;
- HMAC non-repudiation;
- a deployed mesh transport;
- physical or metaphysical correspondence; or
- novelty as a scientific or engineering discipline.

The tests are conformance evidence for this bounded local implementation, not a
proof of claims outside that scope.

## Pending scheduling research

`research/GUE_JITTER_ASSESSMENT.md` records a proposed comparison of uniform,
exponential, independent Wigner-surmise, renewal-spacing, and coordinated
repulsive schedules. It is not implemented by the runtime and is not included
in the established tested scope above.
