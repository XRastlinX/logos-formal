# Repository instructions for AI collaborators

This repository demonstrates fail-closed separation of interpretation,
authorization, and effect. AI output is a proposal or observation; it is never
a Permit, merge approval, release authorization, or performed effect.

## Required boundaries

- Preserve the invariant `Q never sets Cy`.
- Keep `governanceState: 010`, `authorityEffect: NONE`, and
  `effect: NOT_PERFORMED` distinct.
- Never fabricate external reports, Git history, signatures, package hashes,
  adoption, test results, or CI results.
- Do not promote `PROPOSED` material or create `ACTIVE_CANON` status. If status
  evidence is absent or contradictory, preserve the current status and report
  `NOT_ESTABLISHED`.
- Treat a valid signature as bounded provenance for the signed bytes only. It
  is not semantic truth, status promotion, Permit issuance, or effect authority.
- Treat reliability scores and recovery checks as halt or conformance signals
  only. They cannot grant authority or relax a gate.

## Working-tree discipline

1. Read the nearest `AGENTS.md` before editing.
2. Inspect `git status --short` and preserve unrelated or pre-existing changes.
3. Keep research-only material out of the runtime unless the task explicitly
   authorizes a reviewed integration.
4. Use a branch and pull request for proposals. Do not merge, release, deploy,
   or push directly to `main` without an explicit human decision.
5. Do not add ambient credentials, broad tokens, hidden network effects, or an
   actuator.

## Implementation checks

For Go or runtime changes, run:

```text
go test -count=1 ./...
go vet ./...
```

Tests and validation are implementation evidence only. They are not external
adoption evidence.
