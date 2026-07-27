# Node validation

`cyonic-validate` is the repository's witness-only validation entrypoint. It
binds a result to an exact tree root, runs an allowlisted set of checks, and
rejects the run if the target changes while those checks execute.

It does not issue a Permit, elevate canon, prove a claim true, or authorize an
effect. A successful result is `OBSERVE_ONLY` under Cubed Bit `010` with
`governanceAuthorityEffect: NONE`.

`go vet` and `go test` execute repository code with the host process's ordinary
filesystem, process, and network capabilities. `executionContainment: HOST`
makes that explicit. Governance authority `NONE` is not a claim of
side-effect-free execution.

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
X-Governance-Authority-Effect: NONE
X-Execution-Containment: HOST
X-Validation-Status: COMPLETE
X-Target-Root: sha256:<tree-root>
X-Checks-Passed: 5/5
X-Router-Decision: OBSERVE_ONLY
X-Rejected-Invariant: none
X-Decision-Root: sha256:<decision-root>
```

`Target-Root` is a domain-separated SHA-256 digest over repository-relative
paths, file kinds, lengths, and bytes:

- `.git` metadata is excluded;
- canonical paths use `/` and valid UTF-8;
- entries are sorted by canonical UTF-8 byte order before hashing;
- directories and regular files are covered;
- executable and other mode bits are not covered;
- every symlink or reparse point is rejected in v0.1;
- case-colliding paths are rejected for cross-platform portability;
- tracked and untracked entries are covered;
- `.gitattributes` fixes ordinary text line endings across checkouts.

`Profile-Root` binds the declarative manifest without executing manifest
content. `Validator-Runtime-Root` binds the local validator executable.
`Decision-Root` binds those roots, the target root, declared profile and scope,
fixed governance fields, runtime coordinates, final decision, and ordered
check statuses. Human-readable details and civil timestamps are deliberately
excluded so they cannot create cross-platform identity drift.

These SHA-256 values are local integrity checksums. They are not signatures or
external witness certificates.

`VALIDATED` means only that the declared checks passed over the identified
bytes. It does not establish semantic truth, lawful ownership, GTS/GMEOW
compatibility, canon status, external witness independence, or Apply authority.

The before/after check proves only that target roots were identical at
validation entry and exit. It is not an operating-system snapshot or file lock.
A concurrent writer that changes bytes and restores them exactly between the
two roots is outside this reference implementation's detection guarantee.
Production use requires a disposable read-only snapshot, constrained network,
isolated temporary directories, or equivalent containment.

## Exit codes

| Code | Meaning |
| ---: | --- |
| Code | Validation status | Router decision | Meaning |
| ---: | --- | --- | --- |
| `0` | `COMPLETE` | `OBSERVE_ONLY` | Checks passed over the exact target. |
| `10` | `COMPLETE` | `REJECT` | A declared validation check rejected the target. |
| `11` | `INPUT_ERROR` | `NONE` | Target is missing, invalid, or outside the repository. |
| `12` | `PROFILE_ERROR` | `NONE` | Manifest is missing or invalid. |
| `13` | `ENVIRONMENT_ERROR` | `NONE` | Required Go toolchain is unavailable. |
| `14` | `INTERNAL_ERROR` | `NONE` | Internal validator or invocation failure. |

Only exit `10` is an artifact rejection. Other nonzero exits fail closed
without claiming that the artifact violated a declared invariant.

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
