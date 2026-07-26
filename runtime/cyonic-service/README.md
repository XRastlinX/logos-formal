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
