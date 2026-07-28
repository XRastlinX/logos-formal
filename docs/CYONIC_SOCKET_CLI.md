# `/cyonic` Socket Command Contract

```text
slash adapter:       PROPOSED / NOT_IMPLEMENTED
backend commands:    PARTIALLY_IMPLEMENTED
effect command:      ABSENT
governance state:    010
authority_effect:    NONE
```

`/cyonic` is a vantage-agnostic adapter vocabulary. It gives Codex, a local
model, CI, or a human-operated client the same typed socket without attaching
repository authority or privileged credentials to the model.

The slash syntax is not yet an installed shell executable or chat adapter.
The table below defines its mapping to the current repository commands.

| Socket command | Current backend | State | Result |
|---|---|---|---|
| `/cyonic help` | `cyonic-service help` plus this contract | Implemented backend | Read-only command discovery |
| `/cyonic observe` | `cyonic-service evaluate` | Implemented backend | Verifies permit evidence; no effect |
| `/cyonic receipt verify` | `cyonic-service receipt verify` | Implemented backend | Verifies a signed observer receipt |
| `/cyonic probe` | `cyonic-service probe-http` | Implemented backend | Probes the HTTP observer boundary |
| `/cyonic fixture` | `cyonic-service fixture` | Implemented backend | Creates a short-lived evaluation fixture |
| `/cyonic trial` | `cyonic-service trial` | Implemented backend | Records a participant report without promotion |
| `/cyonic repo validate` | `scripts/validate.*` / `cmd/cyonic-validate` | Implemented backend | Validates the governed tree without mutating it |
| `/cyonic status` | Aggregate the preceding read-only checks | Proposed | Reports evidence, missing controls, and `NONE` authority |
| `/cyonic apply` | No backend | Forbidden | Typed `EFFECT_ROUTE_FORBIDDEN` rejection |

## Proposed adapter grammar

```text
/cyonic observe
  --request <request.json>
  --trust-profile <operator-selected-profile>
  [--evidence-log <append-only.jsonl>]

/cyonic receipt verify
  --receipt <decision-receipt.json>
  --trust-profile <operator-selected-profile>
  [--expected-signer sha256:<digest>]

/cyonic probe
  --base-url <loopback-or-approved-service-url>
  --source-ref <caller-selected-source-ref>
  [--expected-instance <out-of-band process identity>]

/cyonic repo validate
  [--target <repository-root>]
  [--format text|json]

/cyonic status
  [--format text|json]
```

The adapter resolves `--trust-profile` outside the model-supplied request. The
profile contains public verification material and policy references only. A
raw private key, bearer token, cloud credential, or arbitrary tool URL MUST
NOT be accepted as an inline model argument.

## Invariants for every command

Every response must include or project:

```text
routerDecision:   OBSERVE_ONLY | REJECT | NONE
effect:           NOT_PERFORMED
forwarded:        false
governanceState:  010
authorityEffect:  NONE
```

The adapter must:

- parse a typed allowlisted command before reading permit or credential data;
- reject unknown commands, flags, schemas, and operations;
- bind the action, target, payload digest, evidence identifiers, policy root,
  operator root, and runtime root in any signed receipt;
- keep trust profiles, signer selection, and connector credentials outside the
  model request;
- treat missing control files, trust roots, schemas, or policy roots as
  `REJECT`, never as a default approval;
- preserve the original proposal and rejection receipt as inert evidence;
- never translate success into `Apply`, forwarding, merge, deploy, send, or
  another external effect.

## Why this narrows confused-deputy risk

A confused deputy appears when a less-privileged requester persuades a more
privileged component to misuse authority that the requester could not exercise
directly.

The current socket removes the deputy's usable authority:

1. The model can select only observation commands.
2. The request cannot supply the trust anchor used to validate itself.
3. The receipt binds the exact action, target, payload, evidence, policy,
   operator, and runtime.
4. The verifier owns no actuator or private signing credential.
5. All successful routes remain `forwarded: false` and
   `effect: NOT_PERFORMED`.

This is prevention by capability absence for the current observer service, not
a proof about every future integration. If an effect gateway is later added,
it becomes a new security boundary and must independently implement
least-privilege credentials, audience/resource restrictions, action/target
binding, expiration, one-time nonce consumption, replay storage, revocation,
policy versioning, and externally governed approval.

Model Context Protocol or another transport does not supply these properties
by itself. The transport may carry a typed proposal to this socket; it may not
turn the model into the credential holder or permit issuer.

Reference:

- [NSA, Security Design Considerations for AI-Driven Automation Leveraging MCP](https://www.nsa.gov/Portals/75/documents/Cybersecurity/CSI_MCP_SECURITY.pdf)
