# Node validation

`cyonic-validate` is the repository's witness-only validation entrypoint. It
binds a result to an exact tree root, runs an allowlisted set of checks, and
rejects the run if the target changes while those checks execute.

It does not issue a Permit, elevate canon, prove a claim true, or authorize an
effect. A successful result is `OBSERVE_ONLY` under Cubed Bit `010` with
`authorityEffect: NONE`.

## One command

On Linux, macOS, or another Bash environment:

```bash
./scripts/validate.sh
```

On Windows PowerShell:

```powershell
.\scripts\validate.ps1
```

If local execution policy blocks direct `.ps1` invocation, use the
process-scoped command wrapper:

```bat
scripts\validate.cmd
```

For JSON:

```bash
./scripts/validate.sh --format json
```

```powershell
.\scripts\validate.ps1 -Format json
```

The Bash and PowerShell wrappers build the validator into an operating-system temporary directory,
run it from the repository root, preserve its exit code, and remove the
temporary binary.

## Checks

The manifest at `cyonic.validation.json` names check identifiers, never shell
commands. Version 0.1 accepts exactly these checks:

| Identifier | Predicate |
| --- | --- |
| `CV-TARGET-BOUNDARY` | The resolved directory is inside this repository. |
| `CV-REQUIRED-PATHS` | Every declared structural path exists and remains inside the repository. |
| `CV-GO-VET` | `go vet ./...` succeeds from the repository root. |
| `CV-GO-TEST-FRESH` | `go test -count=1 ./...` succeeds without cached test results. |
| `CV-TARGET-IMMUTABILITY` | The target's deterministic SHA-256 tree root is unchanged after all checks. |

The validator never executes a command supplied by JSON. Its two process
invocations are compiled into the approved operator registry in the Go code.

## Result semantics

Successful text output includes:

```text
X-Cubed-Bit: 010
X-Validator-Operator: 000
X-Authority-Effect: NONE
X-Target-Root: sha256:<tree-root>
X-Checks-Passed: 5/5
X-Router-Decision: OBSERVE_ONLY
X-Rejected-Invariant: none
X-Decision-Root: sha256:<decision-root>
```

`Target-Root` is a domain-separated SHA-256 digest over repository-relative
paths, file kinds, symlink targets, lengths, and bytes. `.git` metadata
is excluded. Symlinks that resolve outside the repository are rejected.

`Decision-Root` binds the target root, declared profile and scope, fixed
governance fields, final decision, and ordered check statuses. Human-readable
details and civil timestamps are deliberately excluded so they cannot create
cross-platform identity drift.

`VALIDATED` means only that the declared checks passed over the identified
bytes. It does not establish semantic truth, lawful ownership, GTS/GMEOW
compatibility, canon status, external witness independence, or Apply authority.

The before/after immutability check is not an operating-system snapshot or file
lock. A concurrent writer that changes bytes and restores them exactly between
the two roots is outside this reference implementation's detection guarantee.
Production use requires a read-only snapshot or equivalent filesystem
isolation.

## Exit codes

| Code | Meaning |
| ---: | --- |
| `0` | Checks passed; result is `VALIDATED` / `OBSERVE_ONLY`. |
| `10` | A declared validation check rejected the target. |
| `11` | Target is missing, invalid, or outside the repository. |
| `12` | Manifest is missing or invalid. |
| `13` | Required Go toolchain is unavailable. |
| `14` | Internal validator or invocation failure. |

## External storage boundary

GTS/GMEOW is an external graph transport project. It is not imported by this
validator and supplies no Cubed Bit, Permit, authority, or validation semantics.
Any future GTS-backed storage adapter must remain separately `PROPOSED` and pass
its own conformance tests.

## Interpretive boundary

This is a structured interpretive tool that re-reads alchemical symbols as
operators in a transformation calculus. It is not laboratory science,
religious doctrine, or final metaphysical truth. Its value is measured only by
whether it produces clearer thinking about reactive processes.
