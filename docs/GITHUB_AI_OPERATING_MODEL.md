# GitHub + AI operating model

**Reviewed:** 2026-07-27
**Scope:** public repository coordination and evidence
**Authority effect:** NONE

GitHub can coordinate human and AI development without allowing an AI proposal
to become an effect by implication. The separation is not automatic: it
depends on repository instructions, permissions, protected branches or
rulesets, required checks, review, and explicit release controls.

## Supported GitHub capabilities

Current GitHub documentation establishes that:

- assigning an issue to Copilot can produce a pull request for human review;
- third-party coding agents can work alongside Copilot and create or update
  pull requests;
- repository-wide instructions, path-specific instructions, and `AGENTS.md`
  files can provide durable project context;
- custom agent profiles can define specialized development roles;
- Copilot code review can analyze pull requests, but its output must still be
  validated;
- protected branches and rulesets can require pull requests, reviews, status
  checks, signed commits, and other merge conditions;
- GitHub Actions can generate verifiable build-provenance attestations for
  release artifacts.

Official references:

- [Assigning an issue to Copilot](https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/assigning-issues-and-pull-requests-to-other-github-users)
- [Third-party coding agents](https://docs.github.com/en/copilot/concepts/agents/about-third-party-coding-agents)
- [Repository and agent instructions](https://docs.github.com/en/copilot/how-tos/copilot-on-github/customize-copilot/add-custom-instructions/add-repository-instructions)
- [Custom agents](https://docs.github.com/en/copilot/how-tos/copilot-on-github/customize-copilot/customize-cloud-agent/create-custom-agents)
- [Copilot code review](https://docs.github.com/en/copilot/concepts/agents/code-review)
- [Protected branches](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches)
- [Artifact attestations](https://docs.github.com/en/actions/how-tos/secure-your-work/use-artifact-attestations/use-artifact-attestations)

Availability varies by plan, policy, repository settings, and product surface.
This repository does not claim that every listed capability is enabled.

## Project operating loop

```text
issue or declared task
  → agent inspects repository instructions
  → agent proposes a branch and pull request
  → deterministic tests and boundary checks
  → human or external policy review
  → merge decision
  → separately authorized packaging or release
  → content-bound receipt or attestation
```

The pull request is a proposal surface. CI reports are observations about the
tested commit. Neither is a Permit to deploy, publish, spend, or mutate an
external system.

```text
agent proposal ≠ merge approval
green check ≠ correctness or authority
merge ≠ release
receipt or attestation ≠ truth
```

This is the engineering pattern called bounded or proposal-only adaptation in
this project. It does not imply consciousness, self-sovereignty, or autonomous
governance.

## Capability vocabulary

The operating model keeps visibility, executable capability, invocation, and
authorization as different types:

| Object | Canonical name | Boundary |
|---|---|---|
| Python process speaking MCP | MCP Server | Hosts one or more tools; connection alone is not invocation |
| Schema-declared callable operation | MCP Tool | Advertised capability; discovery is not authorization |
| `.py` supplied only as context | Reference Module | Readable artifact with no ambient execution right |
| Governed procedural package | Skill Package | Instructions plus declared references, assets, or scripts |
| Workflow step behind authorization | Permit-Gated Tool Node | Invocation occurs only after independent permit verification |

```text
Attach(file)      != Register(server)
Register(server)  != Advertise(tool)
Advertise(tool)   != Invoke(tool)
Invoke(tool)      != AuthorizedEffect
```

An agent may read a Reference Module or follow a Skill Package inside its
declared scope. It must not reinterpret visibility as registration, discovery
as invocation, or invocation as authorization. A Permit-Gated Tool Node fails
closed before invocation when external permit evidence is missing or invalid.

## Repository memory

The repository supplies durable context through:

- `AGENTS.md` for the cross-agent operating contract;
- `.github/copilot-instructions.md` for GitHub Copilot repository guidance;
- specifications and architecture decisions;
- tests and fixtures;
- committed receipts and release manifests;
- the parent-linked Git commit graph.

Instruction files guide model behavior; they are not security boundaries.
Enforcement comes from permissions and independently configured repository
rules. Tests and receipts then provide evidence about the bounded behavior
they actually cover.

## Archive and provenance limits

Git objects are content-addressed and commits bind parent history, but Git
history is not inherently immutable or permanent. References can be moved,
branches can be force-pushed or deleted when policy allows, hosting can be
removed, and unreachable objects can be pruned.

Durable custody therefore needs an explicit combination of:

1. protected branches or rulesets with bypass constrained;
2. maintained mirrors or backups;
3. pinned commit and artifact digests;
4. signed commits or tags where identity binding is required;
5. independently verified release attestations where build provenance matters.

A verified signature associates signed bytes with a key or identity under a
verification policy. It does not prove the code is correct, safe, authorized,
or canonical. Likewise, an artifact attestation is useful only when its subject
digest and provenance are actually verified.

## CI and review boundaries

For the public Cyonic surface, required checks should cover:

- Go tests and vetting;
- the repository validator on Windows, Linux, and macOS;
- real HTTP refusal tests;
- deterministic package construction;
- source-bound cold paths on all three operating systems.

Required status checks should be bound to the expected GitHub App where
possible. GitHub documents that actors and integrations with write permission
can otherwise set commit statuses, so a check name alone is not a complete
trust policy.

Copilot or another AI reviewer may provide useful findings, but it must not be
the only approval for a capability, governance, packaging, or release change.

## Current repository posture

| Layer | Current repository artifact | Claim boundary |
|---|---|---|
| Cross-agent guidance | `AGENTS.md` | Procedure, not enforcement |
| Copilot guidance | `.github/copilot-instructions.md` | Procedure, not enforcement |
| Ownership routing | `.github/CODEOWNERS` | Requires branch/ruleset configuration to become a merge gate |
| CI | `.github/workflows/validate.yml` | Evidence from the commit actually run |
| Runtime boundary | `runtime/cyonic-service` | Verification and refusal; no actuator |
| Portable package | `dist/` plus packaging source | Candidate or release status must be explicit |
| External reports | First Contact / Issue #2 | Claims awaiting declared adjudication |

The proposed release-custody boundary is documented separately in
[`ARTIFACT_ATTESTATION_POLICY.md`](ARTIFACT_ATTESTATION_POLICY.md). It does not
enable attestations or grant workflow permissions.

Repository settings, agent enablement, branch-protection bypass lists, and
required-check sources live outside Git. They must be audited on GitHub before
the repository can claim those controls are active.

## Missing-reference fail-closed mechanism

The public validation profile has a validator-compiled minimum control-file
floor. `cyonic.validation.json` may add required paths, but it cannot remove
the following controls while retaining the
`logos-formal-node-setup-010` profile:

- the root and GitHub agent instruction files;
- CODEOWNERS routing;
- the Q classification status source;
- the A4 matrix and artifact-attestation policy;
- the Codex and GitHub operating profiles;
- the validation manifest itself.

The mechanism has three layers:

1. `LoadManifest` rejects a public-profile manifest that omits a minimum
   control path.
2. `CV-REQUIRED-PATHS` rejects when a declared control is absent, unreadable,
   outside the repository, or reached through a symlink/reparse point.
3. CI exits nonzero when the governed validator returns `REJECT`.

This does not make CI self-authorizing. The check blocks merge only if an
external GitHub branch rule requires the exact check source and restricts
bypass. If that configuration is absent or cannot be inspected, the merge-gate
status remains `NOT_ESTABLISHED`.

Action-specific authorization records are deliberately not reconstructed from
repository text. If a promotion, release, deployment, or Apply operation names
a required Permit, adjudication, source digest, or decision record and that
reference is missing or cannot be verified, the operation remains
`NO_DECISION` or `REJECT`. The artifact stays `PROPOSED`; an agent must not
create the missing authority for its own proposal.

## Explicit non-goals

- No agent may promote its own work to canon.
- No instruction file is treated as a Permit.
- No failed CI loop is permitted to merge its own repair.
- No AI review substitutes for accountable human or policy review.
- No Git or GitHub feature establishes external adoption.
- No build attestation is generated merely to decorate an unreleased artifact.
