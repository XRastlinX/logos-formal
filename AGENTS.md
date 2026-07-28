# Agent operating contract

All automated collaborators must follow `.github/copilot-instructions.md`.

Agents may inspect, analyze, implement, test, and stage proposals. They may not
self-promote an artifact or claim Owner, Seal, Apply, or state-commit authority.

## Observation Principle

```text
Observe(O, R) -> (O', receipt)
R' = R
```

An observation may change the observer's working context and may produce a
detached receipt. It must not change the governed target. Reading, hashing,
testing, comparing, and reporting are observations only while the inspected
repository state remains unchanged.

Before acting, classify the operation:

- `OBSERVE`: make no persistent repository or external-system change;
- `PROPOSE`: edit only the declared branch or worktree and leave a reviewable
  diff; the diff is not canon, authorization, merge approval, or release;
- `EFFECT`: do not proceed without an independently verifiable external Permit
  covering the exact action, target, and content.

If an observation needs generated files, place them in an excluded temporary
location or use an isolated worktree. If a command unexpectedly changes the
target, stop, preserve the evidence, and report the mutation. Do not relabel
the changed run as observation.

```text
Table(x) != Invoke(x)
visibility != invocation
proposal != authorization
receipt != Permit
green check != merge or release approval
```

## Missing-reference fail-closed rule

The minimum control references for this profile are:

- `AGENTS.md`;
- `.github/copilot-instructions.md`;
- `.github/CODEOWNERS`;
- `03_Q_Taxonomy/Q_CONTENT_CLASSIFICATION_STATUS_V1.md`;
- `docs/A4_BOUNDARY_TEST_MATRIX.md`;
- `docs/ARTIFACT_ATTESTATION_POLICY.md`;
- `docs/CODEX_REPOSITORY_PROFILE.md`;
- `docs/CYONIC_SOCKET_CLI.md`;
- `docs/DECISION_RECEIPT_V1.md`;
- `docs/GITHUB_AI_OPERATING_MODEL.md`;
- `docs/OPERATIONAL_CLOSURE_RECEIPTS.md`;
- `schemas/cyonic-decision-receipt.schema.json`;
- `schemas/cyonic-decision-receipt-verification.schema.json`;
- `cyonic.validation.json`.

If a required control, status source, adjudication record, Permit, or
content-binding reference is missing, unreadable, ambiguous, stale, or
contradictory:

1. do not infer or recreate the missing authority;
2. do not promote, merge, release, deploy, Apply, or resume a prior effect;
3. preserve the artifact as `PROPOSED` with `authority_effect: NONE`;
4. report `PROMOTION_EVIDENCE_MISSING` or the validator's exact rejection;
5. continue only with bounded observation or a repair proposal.

The repository validator mechanically rejects a missing minimum control file
under `CV-REQUIRED-PATHS`. This becomes a merge block only when GitHub rules
require the exact validation check and constrain bypass; instruction text and
CI output do not create that external configuration.

Before editing a formal artifact:

1. read `03_Q_Taxonomy/Q_CONTENT_CLASSIFICATION_STATUS_V1.md`;
2. read the target artifact's declared status and any adjudication or status
   companion it names;
3. if the status source is missing, ambiguous, or contradictory, make no
   promotion and keep the change `PROPOSED` with `authority_effect: NONE`;
4. preserve explicit claim class, status, and authority effect;
5. keep Q/content, Cy/capability, lifecycle, permit, and effect carriers
   separate;
6. run the relevant tests and report exact results.

## Reliability-triggered recovery

A reliability score is an observational halt signal, never a source of
authority. Use one only when a versioned task profile defines the metrics,
window, non-negative weights, fixed offline threshold, and critical roles.
Weights must sum to one. Do not tune weights or thresholds online to make a
blocked action pass.

A workflow average cannot override a failed critical role. When a local or
workflow score crosses its declared halt threshold, contain the failure and
prepare a bounded recovery proposal. Do not acquire credentials, broaden
permissions, change policy, or resume an effect unless a separate external
gate covers that exact transition.

A valid decision-receipt signature proves only the exact canonical core and
trusted signer profile. It does not prove semantic truth, verify an underlying
Permit by itself, promote a status label, or authorize an effect.

External responses belong in quarantine until verified and deliberately
admitted.
