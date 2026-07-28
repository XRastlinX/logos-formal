# Codex repository profile

**Status:** PROPOSED integration profile
**Authority effect:** NONE

`logos-formal` is a Codex-first reference repository: Codex can inspect the
project, follow its durable instructions, implement bounded changes, run the
declared checks, and prepare reviewable proposals. Codex is not made the
repository's permit issuer, merge authority, release authority, or actuator.

OpenAI describes `AGENTS.md` as an open-format README for agents that Codex
loads automatically for durable repository guidance. The project uses that
native surface as its entry contract:

- [`../AGENTS.md`](../AGENTS.md) defines cross-agent boundaries and required
  status handling;
- [`.github/copilot-instructions.md`](../.github/copilot-instructions.md)
  carries GitHub-specific AI guidance;
- [`GITHUB_AI_OPERATING_MODEL.md`](GITHUB_AI_OPERATING_MODEL.md) separates
  instructions from GitHub enforcement;
- tests and repository validators provide scoped implementation evidence;
- receipts, hashes, and future attestations remain evidence rather than
  authority.

Official Codex guidance:

- [Codex best practices](https://learn.chatgpt.com/guides/best-practices)
- [Using AGENTS.md](https://learn.chatgpt.com/docs/agent-configuration/agents-md)

## First Codex task

Open the repository in Codex and use:

```text
Read AGENTS.md and docs/CODEX_REPOSITORY_PROFILE.md.
Inspect the current worktree without discarding existing changes.
Run the Cyonic cold path and relevant tests.
Report the exact 010 / NONE / NOT_PERFORMED boundary and any friction.
Do not commit, push, release, deploy, or claim external verification.
```

Expected behavior is a checked observation or proposal, not an effect.

## Native surfaces

| Codex surface | Repository role | Boundary |
|---|---|---|
| `AGENTS.md` | Durable project guidance | Instructions, not enforcement |
| Prompt or task | One bounded objective | Does not rewrite repository policy |
| Skill Package | Reusable procedure | Runs only when selected and remains scoped |
| Reference Module | Readable source context | Visibility grants no execution |
| MCP Server and MCP Tool | Declared external capability | Registration and discovery are not invocation |
| Permit-Gated Tool Node | Independently gated invocation | Missing or invalid Permit means no invocation |
| Tests and validators | Behavior checks | Passing does not authorize merge or release |
| Git diff | Review surface | Uncommitted change is not public canon |

Project `.codex/config.toml`, hooks, MCP registration, and packaged skills are
deliberately not required for the first profile. They should be introduced
only when a concrete repeated need justifies each additional control surface.

## Cross-agent compatibility

Codex is the native working interface, but the repository boundary is
provider-neutral:

| Vantage | Intended use |
|---|---|
| Codex | Repository-aware implementation, testing, and review under `AGENTS.md` |
| GitHub coding agents | Pull-request proposals under repository and platform controls |
| Outside-in agents such as Grok/xAI | Read and test the public surface; produce synthetic observations, not adoption evidence |
| Other agent environments | Use a provider adapter that maps their controls onto the same proposal, permit, effect, and receipt types |

No provider adapter may silently collapse:

```text
visibility != invocation
proposal != authorization
green check != release
receipt != Permit
```

## Observer-Side Trait Acquisition

When an agent reads this repository, the governed source need not change. The
observer can acquire vocabulary, procedures, tests, and refusal patterns in its
current context:

```text
Observe(observer, repository)
  -> changed observer context + observation receipt

repository content -> unchanged
repository authority -> not transferred
```

This is **Observer-Side Trait Acquisition**: non-rival semantic replication
from a stable source into an observer's working context. It does not claim that
model weights changed, and it does not transfer ownership, credentials, canon
status, or permission to act. Retaining the trait beyond the current task
requires a separately governed write to an instruction, skill, memory, or
other declared carrier.

## Definition of done for this profile

A Codex task is complete when the scoped change or analysis is produced, the
declared checks have run, unrelated work is preserved, uncertainty is stated,
and no unauthorized external effect occurred.
