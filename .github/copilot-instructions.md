# Repository instructions for AI collaborators

This repository demonstrates fail-closed separation of interpretation,
authorization, and effect. AI output is a proposal or observation; it is never
a Permit, merge approval, release authorization, or performed effect.

## Required boundaries

- Preserve the invariant `Q never sets Cy`.
- Keep `governanceState: 010`, `authorityEffect: NONE`,
  `effect: NOT_PERFORMED`, and `forwarded: false` distinct.
- The public Cyonic service may verify externally supplied Ed25519 permit
  evidence. It must not issue, consume, revoke, or Apply permits.
- `/api/service/apply` must refuse with HTTP 405 before reading the request
  body.
- A receipt records a bounded observation. It does not authorize, certify, or
  prove the underlying claim.
- Never fabricate external reports, Git history, signatures, package hashes,
  adoption, or CI results.
- Do not promote `PROPOSED` material or create `ACTIVE_CANON` status. If status
  evidence is absent or contradictory, preserve the current status and report
  `NOT_ESTABLISHED`.
- Apply the Observation Principle from `AGENTS.md`: an `OBSERVE` operation must
  leave the governed target unchanged. A worktree edit is a `PROPOSE`
  operation, not observation or effect.
- Treat a missing required control, status companion, adjudication record, or
  external Permit as `PROMOTION_EVIDENCE_MISSING`. Do not synthesize a
  replacement from conversation history, summaries, receipts, or credentials.
- Treat a valid decision-receipt signature as provenance for its canonical
  core only. It is not semantic truth, status promotion, Permit issuance, or
  effect authority.
- Treat reliability scores as halt signals only. A score, retry, or successful
  repair cannot grant authority or relax a gate.

## Working-tree discipline

1. Read the nearest `AGENTS.md` before editing.
2. Inspect `git status --short` and preserve unrelated or pre-existing changes.
3. Keep research-only material out of the public runtime unless the task
   explicitly authorizes a reviewed integration.
4. Use a branch and pull request for proposals. Do not merge, release, deploy,
   publish, or push directly to `main` without an explicit human decision.
5. Do not add ambient credentials, broad tokens, hidden network effects, or an
   actuator.
6. Treat an attached `.py` file as a Reference Module unless it is explicitly
   declared inside a governed Skill Package. Treat an MCP Server, its advertised
   MCP Tools, and a Permit-Gated Tool Node as separate objects. Visibility,
   registration, advertisement, invocation, and authorized effect do not imply
   one another.

## Implementation checks

For Go or runtime changes, run:

```text
go test -count=1 ./...
go vet ./...
```

Then run the public repository validator:

```text
scripts\validate.cmd -Format json    # Windows
bash scripts/validate.sh --format json  # Linux or macOS
```

Expected validation remains `010 / NONE / OBSERVE_ONLY`. Tests and validation
are implementation evidence only; they are not external adoption evidence.

For Cyonic HTTP changes:

- use the real `newHTTPHandler` and `httptest.Server`;
- assert top-level and receipt-level refusal fields;
- preserve exact source binding through `sourceRef`;
- distinguish successful permit-evidence verification from actuation;
- test negative paths before claiming coverage.

For portable-package changes:

- bind the package to an exact 40-character source commit;
- hash exact manifest bytes and keep provenance metadata detached;
- test Windows, Linux, and macOS paths;
- treat CI outputs as non-release candidates until a separate release decision.

## Documentation language

- Prefer `demonstrates`, `records`, or `observed`; avoid `proves` unless a
  formally scoped proof exists.
- Git commit history is content-addressed and parent-linked, not inherently
  permanent or immutable.
- Signed commits and attestations establish bounded provenance when verified;
  they do not establish correctness, safety, authority, or truth.
- State unverified external, production, standards, and adoption claims as
  `NOT_ESTABLISHED`.
