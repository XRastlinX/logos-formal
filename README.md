# logos-formal

**Fail-closed separation of interpretation from authorization.**

An AI system may propose. It must not authorize or execute its own proposals
without an external permit. If any required check fails, the effect does not
happen.

The reference Cyonic Boundary Service verifies Ed25519 permit evidence bound to
an exact artifact proposal. It cannot issue, consume, revoke, or Apply a permit,
and it has no actuator capability.

[![Validate Canon](https://github.com/XRastlinX/logos-formal/actions/workflows/validate.yml/badge.svg)](https://github.com/XRastlinX/logos-formal/actions/workflows/validate.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## A Codex-first, cross-agent repository

This repository is designed to be opened and worked in with Codex. Its
repo-level [`AGENTS.md`](AGENTS.md) gives Codex durable operating constraints,
while tests and validators define what completion means. The same public
documents and schemas remain readable by other agents without giving any agent
ambient merge, release, or effect authority.

```text
agent reads repository guidance
  -> prepares a bounded proposal
  -> runs declared checks
  -> reports evidence and remaining uncertainty
  -> external review decides any merge, release, or effect
```

Start with the [Codex repository profile](docs/CODEX_REPOSITORY_PROFILE.md).

## Quick check (about 5–10 minutes)

Requires [Go 1.24+](https://go.dev/dl/).

```bash
git clone https://github.com/XRastlinX/logos-formal.git
cd logos-formal
git checkout --detach 7e0f263c2a31eaece5bf8ded76ce4ba47c7db32a
go run ./runtime/cyonic-service smoke-http
```

You should see:

```text
governanceState: 010
authorityEffect: NONE
effect: NOT_PERFORMED
applyProbeStatus: 405
applyProbeReason: EFFECT_ROUTE_FORBIDDEN
```

The trial validates a boundary and records that the Apply compatibility route
rejected during that run. It does not issue a permit or perform an effect.

Independent cold-run reports belong in
[Issue #2](https://github.com/XRastlinX/logos-formal/issues/2), which records
the portable package SHA-256, qualification rules, and reporting instructions.

## Cold-reader map

| Question | Public answer |
|---|---|
| What runs? | The Cyonic Boundary Service and repository validator |
| What is archived? | Committed specifications, source, tests, manifests, receipts, and their Git history |
| What is refused? | Permit issuance, self-authorization, forwarding, and every effect route |
| How do I verify it? | Run the quick check above or the pinned portable package |
| Where do I report friction? | Use the copy-ready fields in the First Contact guide and Issue #2 |

Start with the [service contract](docs/SERVICE_SURFACE.md), the
[validation contract](docs/NODE_VALIDATION.md), or the
[five-to-ten-minute trial](docs/FIRST_CONTACT_TRIAL.md). No prior project
vocabulary is required to run the refusal path.

## What this is

- A reference governance boundary for agent-style systems
- A runnable service for checking externally supplied authorization evidence
- A formal separation between proposing, validating, authorizing, and applying
- A versioned custody surface for specifications, tests, receipts, and release
  manifests that have actually been committed
- A research architecture mapped in spirit to separation patterns in RATS,
  ABAC, in-toto, and SLSA

## Two roles, one repository

`logos-formal` is deliberately both an executable reference backend and a Git
archive:

| Role | What Git holds | What an outsider can verify |
|---|---|---|
| Executable reference | Service, validator, demonstrators, tests, and package builder | A selected commit builds, validates, and refuses every effect route |
| Versioned custody | Committed specifications, fixtures, manifests, receipts, and their change history | Which exact bytes and rules existed at a selected commit |

The roles reinforce one another: the backend emits bounded evidence, while Git
preserves the committed source and history needed to inspect or replay that
evidence. They do not collapse:

```text
runtime result ≠ repository authority
commit history ≠ independent immutable custody
receipt ≠ Permit
uncommitted or private work ≠ public archive
```

Git makes committed history content-addressable and widely reproducible. It
does not guarantee permanent retention by itself; durable archival custody
also requires maintained mirrors or backups and an explicit retention policy.

## What this is not

- Not a full agent framework or actuator
- Not a permit issuer or source of ambient authority
- Not production-hardened cryptographic infrastructure
- Not a claim of scientific proof, standards conformance, certification, or
  completed external adoption

**External evaluation status:** Not yet established. Independent reports are
welcome but optional; they are not prerequisites for engineering work, merge,
or publication. Internal verification establishes only the behavior covered by
the declared tests.

## Status labels

| Label | Meaning |
|---|---|
| `ACTIVE_CANON` | Binding within this repository's declared process |
| `PROPOSED` | Research or design material; no authority |
| `authority_effect: NONE` | The operation does not authorize or effectuate |
| `010` | Interpretive or validation state |
| `101` | Effect-capable state requiring an external permit |

---

## Service boundary

The network service separates three questions:

```text
interpretation: was the proposal structurally validated?
authorization: was external permit evidence verified?
effect: NOT_PERFORMED
```

Read the [Cyonic Service Surface](docs/SERVICE_SURFACE.md) for the client path,
receipt contract, limitations, and non-claims. The
[First Contact Trial](docs/FIRST_CONTACT_TRIAL.md) defines the evidence intake,
and [dist/README.md](dist/README.md) documents the portable package.

## Validate this node

The repository validator runs an allowlisted check manifest against an exact
tree root and rejects observed in-run mutation:

```bash
./scripts/validate.sh        # Linux / macOS
scripts\validate.cmd         # Windows
```

Success is `010`, `governanceAuthorityEffect: NONE`, and `OBSERVE_ONLY`—never a
permit, canon elevation, or truth claim. The checks run with the host process's
ordinary OS permissions; this reference validator is not a sandbox. See
[Node Validation](docs/NODE_VALIDATION.md) for the result contract, exit codes,
and known limits.

## Architecture

```text
AI proposal (010)
    |
    v
validation / witness (authority_effect: NONE)
    |
    v
external human or policy decision
    |
    v
permit-gated Apply (101) or fail-closed rejection
```

The core invariant is:

> **Q never sets Cy.**
>
> Interpretive output cannot autonomously write to authorization or
> effectuation state.

The current service demonstrates the observer, evidence-verification, and
refusal boundary. It does not establish a production permit issuer or actuator.

## Demonstrators

```bash
go test ./...
go run ./runtime/cyonic-service demo
go run ./runtime/cyonic-service smoke-http
go run ./runtime/demonstrator
go run ./runtime/dcif-membrane
go run ./runtime/autophagic-decay
```

| Demonstrator | Boundary exercised | Status |
|---|---|---|
| Cyonic Boundary Service | Ed25519 binding, expiry, issuer, scope, refusal of effect | Reference |
| Permit-Gated Cell | Replay, wrong target, tampering, expiry | Reference |
| Bleed-Over Membrane | Provenance, uncertainty, authority escalation | Reference |
| Autophagic Decay | Recursive self-ingestion and evidence depth | `PROPOSED` |

## Formal core

| Document | Purpose |
|---|---|
| [Cubed Bit Safety Lattice](01_Governance_Geometry/governance_algebra_v1.md) | Defines the `(A, I, E)` governance states |
| [Bleed-Over Operator β](10_Elevation_Layer/ACTIVE_CANON_Bleed_Over_Operator_Beta_V1.md) | Preserves provenance and uncertainty across translation |
| [Normalization Criteria](10_Elevation_Layer/ACTIVE_CANON_Normalization_Criteria_v1.md) | Defines post-lifecycle validation conditions |
| [PIC-4 Standards Mapping](08_Builder_Layer/PIC4_STANDARDS_MAPPING_V1.md) | Relates the reference model to established assurance vocabulary |
| [Cybot Pattern](08_Builder_Layer/CYBOT_PATTERN_V1.md) | Describes a local-first, proposal-only assistant pattern |

The standards mapping is explanatory. It is not a claim of conformance,
certification, or endorsement by the named standards bodies.

## Start here

- [Why This Matters](docs/WHY_THIS_MATTERS.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Codex Repository Profile](docs/CODEX_REPOSITORY_PROFILE.md)
- [GitHub + AI Operating Model](docs/GITHUB_AI_OPERATING_MODEL.md)
- [A4 Boundary Test Matrix](docs/A4_BOUNDARY_TEST_MATRIX.md)
- [Artifact Attestation Policy](docs/ARTIFACT_ATTESTATION_POLICY.md)
- [Cyonic Decision Receipt v1](docs/DECISION_RECEIPT_V1.md)
- [`/cyonic` Socket Command Contract](docs/CYONIC_SOCKET_CLI.md)
- [Operational Closure Through Receipts](docs/OPERATIONAL_CLOSURE_RECEIPTS.md)
- [Cyonic Service Surface](docs/SERVICE_SURFACE.md)
- [First Contact Trial](docs/FIRST_CONTACT_TRIAL.md)
- [Guarded Builder Session](docs/BUILDER_SESSION.md)
- [Contributing](CONTRIBUTING.md)
- [Session Anchor](docs/SESSION_ANCHOR.md)

## Repository map

```text
01_Governance_Geometry/  capability and authorization states
02_Cord_Algebra/         structural algebra
03_Q_Taxonomy/           content classification
04_Dual_Projection/      observation without cross-modification
05_Rite_Machine/         seven-stage construction lifecycle
06_Prime_Root_Field/     bounded input profile
08_Builder_Layer/        schemas, mappings, and builder guidance
09_Exchange_Corridor/    proposed exchange and transport work
10_Elevation_Layer/      ACTIVE_CANON records
runtime/                 runnable reference implementations
docs/                    public explanations and trial guidance
dist/                    pinned portable trial package
```

## Non-claims and current limits

- The demonstrators are reference implementations, not production-hardened
  systems.
- The Cyonic Boundary Service uses Ed25519. Older actuation demonstrators still
  use simulated signatures and must not be represented as cryptographic
  enforcement.
- The service verifies permit evidence but does not issue, consume, revoke, or
  Apply permits.
- The Autophagic Operator is `PROPOSED` research, not proven.
- Independent custody separation and independent security review are not yet
  established.
- External adoption evidence remains incomplete.

The next assurance tier requires canonical serialization, separate trust
domains, persistent nonce consumption at an external actuator, and independent
security review.

## License

[MIT License](LICENSE) — Copyright (c) 2026 Charles LeRoy McClure II

## Citation

See [CITATION.cff](CITATION.cff), or use:

```bibtex
@software{mcclure2026logosformal,
  author = {McClure II, Charles LeRoy},
  title = {Logos-Formal: Fail-Closed Separation of AI Interpretation from Authorization},
  year = {2026},
  url = {https://github.com/XRastlinX/logos-formal},
  license = {MIT}
}
```
