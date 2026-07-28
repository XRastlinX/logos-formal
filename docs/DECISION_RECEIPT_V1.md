# Cyonic Decision Receipt v1

```text
wire profile:       PROPOSED
implementation:     VERIFIER_ONLY
receipt signer:     NOT_IMPLEMENTED
effect gateway:     ABSENT
governance state:   010
authority_effect:   NONE
```

This profile records a bounded observation of a proposal. It binds the
proposal, the evidence identifiers examined by an observer, the observer's
decision, and the exact policy/operator/runtime roots used for that decision.

A valid receipt proves only that the canonical receipt core was signed by the
trusted receipt key. It does not prove that the proposal is true, safe, lawful,
authorized for execution, or already effective.

The normative transport schemas are:

- [`schemas/cyonic-decision-receipt.schema.json`](../schemas/cyonic-decision-receipt.schema.json)
- [`schemas/cyonic-decision-receipt-verification.schema.json`](../schemas/cyonic-decision-receipt-verification.schema.json)

## Do not collapse the state dimensions

`authorized / proposed` is useful as a sociological shorthand but is too
coarse to be an executable repository code. The wire profile keeps these
questions independent:

| Dimension | v1 values | Meaning |
|---|---|---|
| Artifact lifecycle | `PROPOSED` (verifier projection) | The artifact has not been promoted |
| Permit evidence | `NOT_PRESENT`, `NOT_VERIFIED`, `VERIFIED` | What the receipt signer records about separate permit evidence |
| Receipt proof | `NO_DECISION`, `VERIFIED`, `REJECTED` | Whether this verifier accepted the receipt signature and structure |
| Router | `NONE`, `OBSERVE_ONLY`, `REJECT` | What the verifier itself did |
| Effect | `NOT_PERFORMED` | No actuator ran |
| Governance | `010 / NONE` | No authority was created |

In particular:

```text
receipt proof VERIFIED
  != permit evidence VERIFIED
  != proposal authorized
  != effect performed
```

`authorizationBasis.verificationStatus: VERIFIED` is a signed statement by the
receipt witness that it checked external permit evidence under the bound
policy root. The current receipt verifier does not resolve that evidence and
must not restate the field as its own authorization decision.

## Immutable core and proof

The receipt has two top-level members:

```json
{
  "core": {
    "schema": "urn:cyonic:decision-receipt:v1",
    "version": "1.0",
    "canonicalization": "RFC8785",
    "observedAt": "2026-07-27T20:00:00Z",
    "proposal": {
      "requestDigest": "sha256:<64 lowercase hex>",
      "artifactDigest": "sha256:<64 lowercase hex>",
      "action": "repository.propose",
      "target": "refs/heads/example",
      "payloadDigest": "sha256:<64 lowercase hex>"
    },
    "authorizationBasis": {
      "kind": "EXTERNAL_PERMIT_EVIDENCE",
      "permitDigest": "sha256:<64 lowercase hex>",
      "verificationStatus": "VERIFIED",
      "issuer": "external-issuer-reference",
      "keyId": "sha256:<64 lowercase hex>"
    },
    "decision": {
      "operation": "VERIFY_PERMIT_EVIDENCE",
      "outcome": "OBSERVE_ONLY",
      "reasonCode": "NONE",
      "forwarded": false
    },
    "effect": {
      "status": "NOT_PERFORMED",
      "authorityEffect": "NONE"
    },
    "governance": {
      "state": "010",
      "authorityEffect": "NONE"
    },
    "policyRoot": "sha256:<64 lowercase hex>",
    "operatorRoot": "sha256:<64 lowercase hex>",
    "runtimeRoot": "sha256:<64 lowercase hex>",
    "evidenceIds": [
      "sha256:<permit digest>"
    ]
  },
  "proof": {
    "algorithm": "Ed25519",
    "keyId": "sha256:<receipt public-key digest>",
    "signature": "<base64 Ed25519 signature>"
  }
}
```

`evidenceIds` MUST be sorted lexicographically and MUST NOT contain
duplicates. A recorded `permitDigest` MUST also occur in `evidenceIds`.
`observedAt` is an observation, not identity, ordering authority, or freshness
proof.

The proof lives outside `core`. No signature, receipt identifier, or mutable
status is included in the material from which that same receipt identifier is
computed.

## Canonicalization, identity, and signature

Let:

```text
domain       = UTF8("cyonic-decision-receipt-core.v1") || 0x00
core_bytes   = RFC8785(JSON(core))
signing_bytes = domain || core_bytes

receiptDigest = "sha256:" || HEX(SHA256(signing_bytes))
proof.keyId   = "sha256:" || HEX(SHA256(raw_32_byte_Ed25519_public_key))
proof.signature = BASE64(Ed25519_SIGN(private_key, signing_bytes))
```

The verifier rejects:

- invalid UTF-8 or duplicate JSON member names;
- unknown JSON members;
- a non-RFC-8785 profile;
- unsorted, duplicate, malformed, or unbound evidence identifiers;
- a signer key mismatch or invalid Ed25519 signature;
- any `forwarded: true`, effect other than `NOT_PERFORMED`, governance state
  other than `010`, or authority effect other than `NONE`.

RFC 8785 defines the portable canonical JSON representation used by the
cryptographic operations. The implementation pins
`github.com/gowebpki/jcs v1.0.1`.

## Verification command

```bash
go run ./runtime/cyonic-service receipt verify \
  -receipt ./decision-receipt.json \
  -trust-key ./receipt-public.key \
  -expected-signer sha256:<expected-public-key-digest>
```

The command loads only a public verification key. It has no receipt-signing
command, private-key loader, permit issuer, actuator, or Apply route.

Exit codes:

| Code | Meaning |
|---|---|
| `0` | Receipt structure and signature verified; observation only |
| `1` | Typed receipt rejection |
| `2` | Malformed input or verifier configuration |

## Status labels are projections

`PROPOSED`, `PENDING_REVIEW`, `ACTIVE_CANON`, `DISPUTED`, and `REVOKED` are not
mutable flags inside the receipt core.

- The verifier projects `PROPOSED` for a parsed receipt input.
- A review system may project `PENDING_REVIEW` from its external queue.
- Promotion, dispute, or revocation requires a separately content-addressed,
  externally certified transition record.
- A transition never rewrites this receipt.

`previousReceiptDigest` may bind one claimed predecessor, but it does not prove
ledger completeness or select between forks. Those remain external
ledger/policy responsibilities.

## Security boundary and non-claims

This profile narrows a confused-deputy path because the proposed action,
target, payload, permit evidence, policy root, observer operator, and runtime
are all bound before verification. Trust keys are operator configuration, not
model-supplied request fields, and the verifier cannot perform an effect.

It does not complete a future effect gateway. Such a gateway would still need
an independently authorized principal, audience/resource/action/target
restriction, nonce consumption and replay prevention, expiration, revocation,
key rotation, policy-version checks, and least-privilege connector
credentials. The decision receipt cannot supply those authorities.

References:

- [RFC 8785 — JSON Canonicalization Scheme](https://www.rfc-editor.org/rfc/rfc8785.html)
- [RFC 9728 — OAuth 2.0 Protected Resource Metadata](https://www.rfc-editor.org/rfc/rfc9728.html)
