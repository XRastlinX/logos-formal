# Operational Closure Through Receipts and Status Projections

```text
status:             PROPOSED INTERPRETIVE PROFILE
engineering effect: NONE
ontological claim:  NONE
```

This document applies Luhmann's vocabulary as an interpretive model for the
repository boundary. It does not claim that a Git repository is literally a
social system, that cryptography proves meaning, or that an engineering test
settles the ontic status of a human or artificial observer.

## Operational code

The repository recognizes a communication only through its own typed forms:
schemas, content digests, receipt proofs, policy roots, and externally
certified state transitions. Raw agent output can be retained as evidence, but
it is not thereby an operation of the governed surface.

The operational code is deliberately not a single `authorized / proposed`
boolean. A single bit would allow proof verification, permit verification,
lifecycle status, and effect status to masquerade as one another.

The actual membrane is the product of independent codes:

```text
artifact:             PROPOSED
permit evidence:      NOT_PRESENT | NOT_VERIFIED | VERIFIED
receipt proof:        NO_DECISION | VERIFIED | REJECTED
router:               NONE | OBSERVE_ONLY | REJECT
effect:               NOT_PERFORMED
governance authority: NONE
```

Recognition therefore does not equal admission:

```text
well-formed communication != authorized transition
valid signature           != semantic truth
verified receipt          != Permit
status label              != effect
```

## Deparadoxication

The boundary creates a self-reference problem: the system can describe a
change to itself, but under A4 it cannot treat that description as the
authority to make the change.

The receipt deparadoxicates this in a bounded engineering sense:

1. An agent emits a proposal.
2. The observer classifies it without changing the governed target.
3. The observer binds the proposal, evidence, policy, operator, runtime, and
   non-effect result into an immutable receipt core.
4. The receipt makes that second-order observation available to an external
   reviewer or policy gate.
5. A separate authority may later issue a Permit or a certified rejection.
6. Any lifecycle change is recorded as a separate transition, never as a
   rewrite of the proposal or observer receipt.

This is a temporalization technique: a present self-reference is converted
into a stable object that can be adjudicated in a later operation. It does not
literally stop time, resolve the system's authority from within, or make the
receipt an authority source.

`observedAt` records a clock reading only. Ordering comes from
content-addressed predecessor links and an external ledger/policy; even those
links do not by themselves prove that the presented chain is complete.

## Status labels as projections

Labels are projections from evidence, not imperative verbs:

| Label | Permitted meaning | Forbidden inference |
|---|---|---|
| `PROPOSED` | An inert candidate exists | It may be applied |
| `PENDING_REVIEW` | An external workflow has queued review | Approval is likely or implied |
| `ILLUSTRATIVE_DEMO` | The artifact is explanatory | It is production or canonical |
| `ACTIVE_CANON` | A valid external transition selected the artifact under a pinned policy | The artifact authorized itself |
| `DISPUTED` / `REVOKED` | A later certified transition changes the projected claim state | The immutable original was rewritten |

The projection layer must fail closed when required receipts, policy roots,
transition links, signer roles, or trust roots are absent.

## Structural coupling without identity transfer

Humans, Codex, local models, CI, and other tools couple to the repository
through the same socket. They remain external vantages unless a separately
defined role and permit grants an operation.

The architectural rule is normative:

```text
vantage identity does not create repository authority
```

Luhmann's distinction between consciousness and communication helps describe
why system communications do not themselves supply human judgment. The
repository does not use that sociological distinction as a technical proof
about consciousness. Its enforceable claim is narrower: no agent output,
self-description, or receipt is accepted as self-authorization.

## Reliability weights and hallucination

Weighting coefficients do not prevent hallucinations. They only determine how
strongly measured consistency, semantic correctness, and execution success
contribute to a halt signal:

```text
R = omega_C*C + omega_S*S + omega_E*E
```

A consistently repeated hallucination can score high on `C`. A successfully
executed wrong tool call can score high on `E`. For knowledge- or
constraint-heavy work, increasing `omega_S` makes the aggregate more
sensitive to semantic mismatch, but only if `S` is measured by independent
tests, source checks, schemas, or reviewers rather than by the same model's
self-rating.

Illustrative offline starting points:

| Task class | omega_C | omega_S | omega_E | Required non-compensating floor |
|---|---:|---:|---:|---|
| Constraint-heavy reasoning | 0.20 | 0.65 | 0.15 | `S >= S_min` |
| Deterministic extraction | 0.40 | 0.45 | 0.15 | schema/source match |
| API orchestration | 0.15 | 0.30 | 0.55 | `S >= S_min` and typed tool contract |

These are examples, not calibrated production values. Weights and thresholds
must be selected offline on held-out tasks and frozen for a declared profile.
Critical component floors prevent a high `C` or `E` from compensating for a
semantic failure.

When a score falls below threshold:

```text
score may halt
score may propose repair
score may not issue Permit
score may not expand credentials or authority
```

Cryptographic receipts make a hallucinated claim tamper-evident; they do not
make the claim true. Semantic checks and the external A4 gate remain required.
