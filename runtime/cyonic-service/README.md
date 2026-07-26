# Cyonic Boundary Service

A read-only CLI that verifies whether an externally supplied Ed25519 permit is
bound to an exact artifact proposal.

It makes three results explicit:

```text
interpretation: structurally validated or rejected
authorization: external permit evidence verified or rejected
effect: always NOT_PERFORMED
```

The only accepted request operation is:

```text
VERIFY_PERMIT_EVIDENCE
```

Every other operation, including `APPLY`, is rejected before permit
verification. The router never forwards the artifact action to another
runtime.

The service runs at governance state `010` with `authorityEffect: NONE`. It has
no actuator, cannot issue a permit, and cannot Apply anything.

## One-command first contact

From the repository root:

```bash
go run ./runtime/cyonic-service demo
```

The command creates an ephemeral test principal, signs one artifact-bound
permit, verifies it, prints a receipt, and discards the private key. The receipt
still reports:

```json
"effect": {
  "status": "NOT_PERFORMED",
  "authorityEffect": "NONE"
}
```

The demo is an internal interaction. It must not be counted as external
adoption evidence.

## Standalone HTTP surface

Run the complete HTTP boundary once, including both aliases and the fail-closed
Apply probe, with one command:

```bash
go run ./runtime/cyonic-service smoke-http
```

This opens an ephemeral loopback listener, probes it, shuts it down, and prints
the structured result. It does not establish externality or adoption.

To keep the service running for interactive calls, start the isolated
demonstration server:

Start the isolated demonstration server from the repository root:

```bash
go run ./runtime/cyonic-service serve-demo
```

It listens on `127.0.0.1:8787` by default and exposes:

```text
GET  /health
GET  /api/demo/request
POST /api/service/cyonic-validate
POST /api/cyonic/validate
POST /api/service/apply              always rejects before body parsing
```

Every response carries:

```text
X-Cubed-Bit: 010
X-Authority-Effect: NONE
X-Router-Decision: OBSERVE_ONLY | REJECT
X-Effect: NOT_PERFORMED
X-Forwarded: false
```

The demo request contains a signed, ephemeral permit-evidence fixture. The
private key is discarded before the server begins listening. The fixture can
exercise the verifier but cannot cause an effect.

In another terminal, run the dependency-free cold-call probe:

```bash
go run ./runtime/cyonic-service probe-http
```

It exercises both validate aliases and the fail-closed Apply compatibility
path. Its output remains `externalityStatus: NOT_ADJUDICATED`; an internal
probe is not external adoption evidence.

You can also use `curl`:

```bash
curl -s http://127.0.0.1:8787/api/demo/request |
  curl -sS \
    -H 'Content-Type: application/json' \
    --data-binary @- \
    http://127.0.0.1:8787/api/service/cyonic-validate
```

For a caller-supplied trust anchor instead of the ephemeral demo fixture:

```bash
go run ./runtime/cyonic-service serve \
  -trust-key ./trusted-public.key \
  -issuer example-principal
```

The HTTP process is intentionally not an Apply gateway. Its compatibility
`/api/service/apply` path always returns `405 EFFECT_ROUTE_FORBIDDEN` before
reading the body, permit material, or credentials.

## First Contact trial

An independent tester can run:

```bash
go run ./runtime/cyonic-service trial \
  -origin external \
  -report cyonic-first-contact-report-01.json
```

The command displays the boundary result and records the tester's own
explanation, setup time, and friction. It never scores or promotes the report.

Every report remains:

```text
externalityStatus: CLAIMED_EXTERNAL
adjudicationStatus: PENDING_EXTERNAL_REVIEW
authorityEffect: NONE
```

See [`docs/FIRST_CONTACT_TRIAL.md`](../../docs/FIRST_CONTACT_TRIAL.md).

An independent reviewer can bind an additive adjudication to the participant
report's exact bytes:

```bash
go run ./runtime/cyonic-service adjudicate \
  -report cyonic-first-contact-report-01.json \
  -out cyonic-first-contact-adjudication-01.json \
  -reviewer independent-reviewer-01 \
  -externality VERIFIED_EXTERNAL \
  -source VERIFIED \
  -comprehension ACCEPTED \
  -friction ACCEPTED
```

The adjudication is still `authorityEffect: NONE`. It records reviewer
attestations; it does not make them self-proving facts or modify the original
participant report.

## External CLI evaluation

Create a short-lived fixture:

```bash
go run ./runtime/cyonic-service fixture -dir .cyonic-fixture
```

Evaluate it:

```bash
go run ./runtime/cyonic-service evaluate \
  -request .cyonic-fixture/request.json \
  -trust-key .cyonic-fixture/trusted-public.key \
  -issuer example-principal \
  -origin external \
  -evidence-log .cyonic-fixture/interactions.jsonl
```

PowerShell:

```powershell
go run ./runtime/cyonic-service evaluate `
  -request .cyonic-fixture/request.json `
  -trust-key .cyonic-fixture/trusted-public.key `
  -issuer example-principal `
  -origin external `
  -evidence-log .cyonic-fixture/interactions.jsonl
```

`-origin external` is only a claim. The receipt records it as
`CLAIMED_EXTERNAL`; a separate reviewer must establish that the caller is
actually independent of the project.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Permit evidence is valid; no effect was performed |
| `1` | Typed boundary rejection |
| `2` | Malformed input or service configuration |

Every result is a structured receipt. When `-evidence-log` is supplied, the
same receipt is appended as one JSONL interaction record.

## Friction categories

Rejected attempts are classified using the adoption-arc categories:

- `TRYABILITY`
- `EVIDENCE`
- `GOVERNANCE`
- `INTEGRATION`
- `SCOPE_MISMATCH`

The current implementation emits the first three. The other categories remain
available for externally adjudicated integration observations; the service
does not invent them.

## Signing contract

Artifact and permit fields use domain-separated, length-prefixed UTF-8 bytes.
The service hashes the artifact with SHA-256 and verifies the permit with
Ed25519.

The trust key and trusted issuer are operator configuration. They are not
accepted from the request being evaluated.

## Limits and non-claims

- This is a reference service surface, not production-hardened infrastructure.
- It verifies one signature against one configured trust key.
- It does not persist nonce consumption or prevent replay at an actuator.
- It does not issue, revoke, or consume permits.
- It does not perform an effect.
- It does not route or forward an effect.
- It does not evaluate semantic truth, physical safety, lawful authority,
  ownership, or complete provenance.
- It does not certify a system as Cyonic-conformant.
- It does not implement the PROPOSED alpha operator.
