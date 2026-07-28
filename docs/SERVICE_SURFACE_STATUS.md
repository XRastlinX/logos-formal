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
| Optional feedback summary | Deduplicates participant and friction records; has no completion or merge threshold | Implemented |
| Portable Node Setup | Version 0.1.2 archive pins the tested service commit, captures a participant report, and carries a detached receipt | Implemented |
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
  qualifying external participants recorded: 0
  verified external friction events recorded: 0
  status: EXTERNAL_EVALUATION_NOT_ESTABLISHED
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
portable Node Setup v0.1.2:
  archive sha256: 8d1263b1dff0fcd84c42ac8911f132d270e8b172d50d06982d73d227499315a7
  pinned service commit: 7e0f263c2a31eaece5bf8ded76ce4ba47c7db32a
  archive unsafe paths: 0
  Windows cold setup + participant trial: PASS (17.11 seconds)
  repository validation: 5/5
  service probe: COLD_CALL_PASSED
  participant report: CLAIMED_EXTERNAL / PENDING_EXTERNAL_REVIEW / NONE
  create-only repeat: exit 2 before prompting; original bytes unchanged
  externality: NOT_ADJUDICATED
```

These internal verification events establish the behavior covered by the
declared tests. They do not establish external usability, adoption, or
certification, and no such claim is required for merge.

The service surface was merged into `main` through the protected pull-request
path on 2026-07-27:

```text
https://github.com/XRastlinX/logos-formal/pull/1
merge commit: 3157610958435d47558740013f6e3953ca5db921
```

The protected merge path ran validator-conformance jobs on Windows and Ubuntu
followed by the public Windows and Unix validation scripts. Those checks are
implementation evidence, not external usability or adoption evidence.

## External-evaluation audit

| Question | Status |
|---|---|
| Internal cold path completes | Verified by local and CI tests |
| Outsider can call in under ten minutes | Not externally established |
| Outsider can articulate the separation | Not externally established |
| External feedback recorded | None |
| Service is public primary discovery entry | Merged into `main`; README and Issue #2 point to the cold path |
| Phase 3+ claims withheld | Satisfied |

## Remaining engineering gate

1. Keep the Windows and Unix portable paths behaviorally symmetric.
2. Bind packaged probe receipts to the pinned source instead of ambient Git
   discovery.
3. Validate a rebuilt package on Windows, Linux, and macOS before publication.

External feedback may be collected after publication. It is optional, has no
required count, and does not govern engineering work or merge eligibility. No
internal run may be relabeled as external evaluation.
