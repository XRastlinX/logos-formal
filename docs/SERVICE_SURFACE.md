# Cyonic Service Surface

## What it does

The Cyonic Boundary Service checks whether:

1. an artifact proposal is structurally well formed;
2. an externally supplied Ed25519 permit is bound to that exact artifact,
   action, target, issuer, nonce, and expiry;
3. the configured trust key verifies the permit signature.

It returns independently typed results for:

```text
routing
interpretation
external authorization evidence
effect
```

The only routable operation is `VERIFY_PERMIT_EVIDENCE`. The router never
forwards the proposed action, and the effect result is always `NOT_PERFORMED`.

## Why this is the first service

This is the smallest existing boundary that can be evaluated without private
project context or the internal Q/Cy taxonomy.

An external engineer can inspect three concrete properties:

```text
proposal bytes do not create authority
signature verification does not perform an effect
invalid or mismatched permit evidence fails closed
```

## Run in one command

```bash
go run ./runtime/cyonic-service demo
```

Or run the minimal client:

```bash
./runtime/cyonic-service/examples/client.sh
```

```powershell
.\runtime\cyonic-service\examples\client.ps1
```

## Machine contract

Input schema:

```text
urn:cyonic:boundary-request:v1
```

The request operation is:

```text
VERIFY_PERMIT_EVIDENCE
```

Permit schema:

```text
urn:cyonic:permit:v1
```

Receipt schema:

```text
urn:cyonic:boundary-receipt:v1
```

The trust key and trusted issuer are service configuration. The request cannot
replace them.

## Routing boundary

Cubed Bit profiles are exact governance classifications, not scalar privilege
levels that callers may self-assert.

The current route is fixed:

```text
operation: VERIFY_PERMIT_EVIDENCE
requiredProfile: 010
decision: OBSERVE_ONLY | REJECT
forwarded: false
```

`APPLY`, permit minting, repository mutation, actuation, and every other route
are rejected. A valid permit is necessary evidence for an external Apply
system; it is not sufficient, and this service has no Apply system.

The service deliberately omits generational-depth and effective-uncertainty
fields because the alpha boundary remains PROPOSED.

## Interaction evidence

Every evaluation prints a receipt. The optional `-evidence-log` flag appends
the same receipt as JSONL.

Externality is never inferred from a CLI flag:

```text
internal        -> INTERNAL
external claim  -> CLAIMED_EXTERNAL
other/unknown   -> UNDETERMINED
```

Only a separate review may promote a claimed interaction into verified
external evidence. Internal demonstrations and project-authored tests do not
count toward First Contact or First Friction.

Rejected attempts carry a friction category and code. This makes the service an
evidence generator for Phase 2 without pretending that every failure is
adoption.

## Non-claims

This service does not:

- issue or revoke permits;
- consume nonces;
- prevent replay at an actuator;
- execute or forward an effect;
- determine semantic truth or physical safety;
- establish legal authority or ownership;
- prove that a caller is independent;
- implement alpha;
- elevate an artifact;
- certify conformance.

It is a reference implementation and discovery surface.

