# Foreign-Clone Experience Report

**Date:** 2026-07-26  
**Tester:** Local agent simulating outsider clone  
**Environment:** Windows 10, Go 1.26.2, PowerShell  

---

## Procedure

Cloned the public repository to a fresh temporary directory with no prior state:

```powershell
git clone https://github.com/XRastlinX/logos-formal.git $tempDir
cd $tempDir
go test ./...
```

## Results

### Clone
- **Time:** ~2 seconds
- **Size:** Small (no large binaries, no vendored dependencies)
- **Friction:** None

### `go test ./...`
- **Time:** 0.572s (first run, uncached)
- **Result:** All 14 tests pass across 3 packages
- **Dependencies:** Zero external dependencies (standard library only)
- **Friction:** None

### `go run .` (individual demonstrator)
- **Time:** <1 second
- **Result:** Clear, self-explanatory output showing the permit-gated boundary
- **Friction:** None

### `.\demo.ps1` (single-command runner)
- **Time:** ~5 seconds total
- **Result:** All three demos run sequentially with clear visual framing, followed by test suite
- **Friction:** None

## What an Outsider Sees (First 30 Seconds)

1. README opens with "Fail-closed separation of AI interpretation from human authorization"
2. ASCII pipeline diagram shows the five-stage flow
3. "Run It (< 2 minutes)" section is immediately visible
4. Clone + test takes under 2 minutes including download

## What an Outsider Sees (First 5 Minutes)

1. Permit-gated demonstrator shows: valid permit accepted, then replay/target/tamper attacks all rejected
2. Bleed-over membrane shows: valid translation passes, certainty inflation rejected, provenance erasure rejected
3. Autophagic decay shows: certainty collapse table as generational depth increases
4. All 14 tests pass

## Friction Points Identified

| Issue | Severity | Status |
|---|---|---|
| No friction found in clone-to-run path | — | Clean |
| `demo.sh` not tested on actual Linux/macOS | Low | Needs external verification |
| README links to docs/ that are only on main (not in the v1.1.0-canon tag) | Low | Will resolve with next tag |
| Internal ontology section (12+1 domains) may confuse outsiders who scroll past the fold | Low | Acceptable — it's below the fold |

## Assessment

The first-run experience is clean. An outsider with Go installed can go from zero to running demonstrators in under 2 minutes. The output is self-explanatory. The README answers "what is this?" and "how do I run it?" without requiring prior context.

The surface is ready for first contact.
