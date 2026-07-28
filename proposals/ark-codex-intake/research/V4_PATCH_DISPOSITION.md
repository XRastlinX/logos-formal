# Antigravity V4 Patch Disposition

```text
classification:       CODE_REVIEW_RECORD
source:               LOCAL_ANTIGRAVITY_WORKSPACE
source_status:        FAILED_LOCAL_SUITE
adapted_status:       PROPOSED_V0.4.0
authority_effect:     NONE
```

## Direct assay

The local Antigravity v4 draft was run before import:

```text
tests: 31
passed: 4
failed: 27
dominant failure: SQLITE_FOUNDATION_INVARIANT_FAILURE
```

The draft declared a v4 closure schema but did not provide an operational v0.3
writer: `applyClosureReceipt` still emitted `CLOSURE_RECEIPT_v0.2` and did not
insert `result_digest`. No v3-to-v4 migration assay, missing-result assay, or
result-tamper assay was present.

## Adapted implementation

The GitHub proposal was advanced independently from its passing v3 baseline:

- schema version is `4`;
- v2 and v3 databases migrate transactionally to v4;
- all triggers that can reference the renamed receipt table are dropped and
  recreated during migration;
- v0.1 and v0.2 payload bytes are preserved with `result_digest = NULL`;
- newly emitted closures use `CLOSURE_RECEIPT_v0.3`;
- `resultDigest` is mandatory lowercase SHA-256;
- the digest participates in canonical receipt identity and a distinct
  `ARK_CLOSURE_RECEIPT_V3` HMAC domain;
- the JSON payload and relational column must agree;
- SQLite rejects missing result bindings; and
- verification rejects result-digest tampering.

The adapted implementation passed 37/37 tests in three consecutive complete
runs. This is tested-scope conformance, not production certification.
