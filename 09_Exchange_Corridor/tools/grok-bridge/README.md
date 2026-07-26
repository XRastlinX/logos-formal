# Grok Collaboration Bridge V1

This bridge moves a bounded set of formal artifacts between this local Codex
workspace and a remote Grok sandbox that can only read uploaded files from
`/home/workdir/artifacts`.

It is a file exchange, not a tunnel into `C:\`, and not a credential bridge.

## Authority boundary

Every exported packet and imported response carries:

```text
governance_state: 010
authority_effect: NONE
mode: REVIEW_ONLY
```

Export does not transfer Owner, Seal, Apply, lease, or state-commit authority.
Return material is always `QUARANTINED_UNTRUSTED` and is never auto-applied.

## Create a packet

```powershell
python .\bridge.py pack `
  --task .\tasks\dual_projection_review.md `
  --include ..\..\..\04_Safe_Paired_Projection\q-cy-dual-projection-adjudication-v1.md `
  --include ..\..\..\05_Analytical_Lifecycle_and_Rite\seven-stage-transition-and-product-category-adjudication-v1.md `
  --output ..\..\..\dist\grok-dual-projection-review-v1.zip
```

Only files named by individual `--include` arguments are exported. Directories
are never enumerated implicitly, and absolute local paths are not written into
the packet.

## Verify before upload

```powershell
python .\bridge.py verify-export ..\..\..\dist\grok-dual-projection-review-v1.zip
```

Upload the verified ZIP to Grok. Its included `GROK_INSTRUCTIONS.md` explains
the bounded task and the required response shape.

## Ingest Grok's return

Grok must return:

```text
INPUT_PACKET_ID.txt
response.md
artifacts/...       # optional text/data artifacts
```

Then run:

```powershell
python .\bridge.py ingest-response `
  --response C:\path\to\downloaded-response.zip `
  --expected-packet-id sha256:... `
  --stage-dir .\quarantine
```

The command validates the packet binding, path and type constraints, size
limits, and high-confidence secret patterns. Only then does it write the
response and a content-addressed intake receipt under quarantine.

## What the receipt proves

It proves the exact locally received bytes, their hashes, their declared input
packet binding, and their quarantine disposition.

It does not prove Grok's identity, semantic truth, completeness, lawful
ownership, or independence. It grants no promotion or Apply authority.

## Exit semantics

- `0`: packet created, verified, or response staged
- `1`: fail-closed typed rejection
- `2`: CLI usage error
