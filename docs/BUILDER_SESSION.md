# Guarded Builder Session

## Outcome

Use `tools/New-CyonicBuilderSession.ps1` to create a separate local clone with:

- an exact recorded base commit;
- a dedicated `session/*` branch;
- no configured Git remote;
- the clone-local credential helper disabled;
- a session record declaring `authorityEffect: NONE`.

Example:

```powershell
.\tools\New-CyonicBuilderSession.ps1 `
  -SessionDirectory C:\Users\CHARL\Documents\builder-sessions\logos-service-01
```

The script prints the environment settings for entering the session without
inheriting ordinary Git credential helpers.

## Why this uses a clone rather than a worktree

A Git worktree shares the source repository's object database and Git
configuration. It is useful for concurrent development, but it is not a
credential or filesystem isolation boundary.

The guarded clone is created through a temporary Git bundle, after which its
remote is removed. This prevents accidental push through the source
repository's normal remote configuration.

## What this does not secure

This is an operational guardrail, not a security sandbox.

It does not restrict:

- network access;
- access to parent directories;
- access to Windows Credential Manager through other processes;
- process creation;
- commands that reconfigure Git;
- credentials supplied through other applications or files.

For an actual sandbox, place the clone inside Windows Sandbox, a container, a
dedicated restricted OS account, or another process boundary with explicit
filesystem, network, and credential controls.

## Promotion boundary

The builder session may produce commits and test results. It should export a
candidate for Principal review rather than configuring a GitHub remote itself.

No local commit, test pass, or builder report changes canon status or grants
push authority.

