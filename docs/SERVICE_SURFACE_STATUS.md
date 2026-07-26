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
| Additive adjudication | Exact report-byte hash + explicit reviewer findings + mechanical receipt/time checks | Implemented |
| Evidence threshold summary | Deduplicates participant and friction evidence; cannot issue authority | Implemented |
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
```

These are internal verification events. They do not count as external First
Contact or First Friction evidence.

The exact implementation is committed on the local candidate branch:

```text
codex/cyonic-service-surface-v1
```

## Adoption-arc audit

| Goal requirement | Status |
|---|---|
| Outsider can call in under ten minutes | Standalone HTTP path and native probe pass internally; not externally verified |
| Outsider can articulate the separation | No qualifying external evidence |
| Three cold users complete First Contact | Not achieved |
| At least one external friction event recorded | Not achieved |
| Service is public primary discovery entry | Local candidate only; not published |
| Phase 3+ claims withheld | Satisfied |

## Remaining gates

1. Review the candidate branch against its exact bytes.
2. Publish through an external Principal-controlled repository transition.
3. Run three cold external trials and record their elapsed time and
   interpretation.
4. Record the first independently verified external friction event.

No internal run may be relabeled as external evidence.
