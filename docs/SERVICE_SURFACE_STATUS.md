# Cyonic Service Surface Status

**Status:** PROPOSED
**Authority effect:** NONE
**Surface:** `runtime/cyonic-service`

## Verified implementation evidence

| Requirement | Current evidence | Result |
|---|---|---|
| Single entry point | `go run ./runtime/cyonic-service demo` | Implemented |
| Standalone HTTP entry point | `go run ./runtime/cyonic-service serve-demo` | Implemented |
| Cold-call network probe | `go run ./runtime/cyonic-service probe-http` | Implemented |
| One-command HTTP cold path | `go run ./runtime/cyonic-service smoke-http` | Implemented |
| Stable validation aliases | `/api/service/cyonic-validate` and `/api/cyonic/validate` | Implemented |
| Governance headers on every HTTP response | `010`, `NONE`, `NOT_PERFORMED`, `forwarded=false`, bounded router decision | Implemented |
| Apply compatibility boundary | HTTP 405 before body or permit parsing | Implemented |
| External CLI path | `fixture` then `evaluate` commands | Implemented |
| Minimal client | PowerShell and shell client examples | Implemented |
| Interpretation separated | Receipt has an `interpretation` record | Implemented |
| Authorization evidence separated | Receipt has `authorization.status = EXTERNAL_PERMIT_VERIFIED` only after Ed25519 verification | Implemented |
| Effect separated | Receipt always reports `effect.status = NOT_PERFORMED` | Implemented |
| 010 lock explicit | Receipt reports `governanceState = 010` and `authorityEffect = NONE` | Implemented |
| Effectful routes rejected | Only `VERIFY_PERMIT_EVIDENCE` is routable; `APPLY` is a typed rejection | Implemented |
| No proposed alpha authority | No depth or uncertainty fields are emitted | Implemented |
| Interaction record | Every evaluation prints a receipt; optional JSONL append | Implemented |
| Internal/external distinction | External is recorded only as `CLAIMED_EXTERNAL` pending separate adjudication | Implemented |
| Friction classification | Rejections carry category, code, and message | Implemented |
| First Contact capture | Interactive trial records setup time, articulation, friction, source ref, and pending adjudication | Implemented |
| Participant evidence immutability | Trial reports are create-only and existing report paths fail before prompting | Implemented |
| Additive adjudication | Exact report-byte hash + explicit reviewer findings + mechanical receipt/time checks | Implemented |
| Evidence threshold summary | Deduplicates participant and friction evidence; cannot issue authority | Implemented |
| Portable Node Setup | Version 0.1.1 archive pins the tested service commit and carries a detached receipt | Implemented |
| Fail-closed tests | Signature, issuer, expiry, scope, digest, malformed input, and effect-route tests | Implemented |

## Local verification

Observed on 2026-07-26:

```text
go test ./...: PASS
go vet ./...: PASS
standalone HTTP live probe: PASS
  validate aliases: 2/2
  governance state: 010
  authority effect: NONE
  effect: NOT_PERFORMED
  forwarded: false
  Apply compatibility status: 405
  Apply rejection: EFFECT_ROUTE_FORBIDDEN
  externality status: NOT_ADJUDICATED
PowerShell client example: PASS
First Contact trial rehearsal: PASS (internal evidence only)
participant report create-only guard: PASS
  existing report path: rejected before prompting
  existing report bytes: unchanged
adjudication + summary rehearsal: PASS
  qualification: NOT_QUALIFYING
  qualifying external participants: 0/3
  verified external friction events: 0/1
  status: EVIDENCE_INCOMPLETE
clean-clone demo + full tests: PASS
clean-clone elapsed time: 3.36 seconds
valid permit receipt:
  routing.decision: OBSERVE_ONLY
  routing.forwarded: false
  effect.status: NOT_PERFORMED
  authorityEffect: NONE
guarded builder clone:
  remote count: 0
  authorityEffect: NONE
  securityBoundary: GUARDRAIL_ONLY
portable Node Setup v0.1.1:
  archive sha256: 2b539eea8ee392bb5f1318b3661b1d04bc090bffb19e7482d9813453a9001c12
  pinned service commit: 7e0f263c2a31eaece5bf8ded76ce4ba47c7db32a
  archive unsafe paths: 0
  Windows cold setup: PASS (16.55 seconds)
  repository validation: 5/5
  service probe: COLD_CALL_PASSED
  externality: NOT_ADJUDICATED
```

These are internal verification events. They do not count as external First
Contact or First Friction evidence.

The exact implementation is published on the public pull-request branch:

```text
codex/cyonic-service-surface-v1
3a5b0801bc243a1716273dfe98acc5265e6c6882
```

GitHub Actions run 16 passed independent validator-conformance jobs on Windows
and Ubuntu, followed by the public Windows and Unix validation scripts. This is
implementation evidence, not independent First Contact evidence.

## Adoption-arc audit

| Goal requirement | Status |
|---|---|
| Outsider can call in under ten minutes | Standalone HTTP path and native probe pass internally; not externally verified |
| Outsider can articulate the separation | No qualifying external evidence |
| Three cold users complete First Contact | Not achieved |
| At least one external friction event recorded | Not achieved |
| Service is public primary discovery entry | Public PR candidate and Issue #2 trial entry; not merged into `main` |
| Phase 3+ claims withheld | Satisfied |

## Remaining gates

1. Review and merge the candidate through the protected repository path.
2. Run three cold external trials and record their elapsed time and
   interpretation.
3. Record the first independently verified external friction event.

No internal run may be relabeled as external evidence.
