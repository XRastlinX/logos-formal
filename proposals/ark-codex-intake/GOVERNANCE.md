# Ark Codex Intake — Governance Record

```text
artifact:              ark-codex-intake
classification:        PROPOSED_REFERENCE_IMPLEMENTATION
secondary_class:       ENGINEERING_TEST_HARNESS
lifecycle:              PROPOSED
promotion_evidence:    MISSING
production_readiness:  NOT_ESTABLISHED
deployment_effect:     NOT_PERFORMED
authority_effect:      NONE
```

## Purpose

This proposal encodes a deterministic local test surface for graph-bound
obligations, authenticated closure receipts, transactional key revocation,
idempotent nonce handling, and adversarial routing assays.

## Boundaries

- The implementation may be read, built, and tested during review.
- The implementation is not active canon and is not wired into the repository's
  production runtime.
- HMAC values authenticate possession of a configured shared secret. They do
  not establish scientific truth, public non-repudiation, or authority.
- Passing tests establishes conformance only for the exercised fixtures and
  deterministic local simulation.
- Physical, metaphysical, universal-scale, Byzantine-fault-tolerance, and
  production-network claims remain `NOT_ESTABLISHED`.
- No file, label, receipt, model response, or successful execution grants
  permission to merge, deploy, or elevate this proposal.

## Control-file gap

The root `AGENTS.md` references:

```text
.github/copilot-instructions.md
10_Status_Promotion_and_Canon_Ledger/CANON_STATUS.md
```

Those files are not present on the base branch at proposal time. This proposal
therefore records `PROMOTION_EVIDENCE_MISSING` and makes no promotion request.
It does not synthesize replacements for the missing repository controls.

## Custody

The files in this directory were imported from the owner's local
`ark-codex-intake` working implementation for review in a dedicated branch.
Git commit and tree identities provide transport lineage after commit; they do
not change this lifecycle or authority classification.
