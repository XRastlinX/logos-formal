# Tested Scope

The checked-in suite exercises:

- SQLite schema migration from versions 2 and 3 to version 4 while preserving
  legacy v0.1/v0.2 receipt payload bytes and retaining `NULL` result digests;
- exact graph binding for legacy version 0.2 closure receipts;
- mandatory SHA-256 result binding for newly emitted version 0.3 closure
  receipts across canonical identity, HMAC input, JSON payload, and SQLite row;
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
