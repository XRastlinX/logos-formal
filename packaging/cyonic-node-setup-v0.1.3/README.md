# Cyonic Node Setup v0.1.3 packaging source

**Status:** local packaging input
**Authority effect:** NONE

This directory contains the auditable overlays used to derive the next
portable node package from the pinned v0.1.2 archive.

The builder:

- requires an exact 40-character source commit;
- starts from the tracked v0.1.2 archive;
- replaces the version, README, and platform scripts with these sources;
- generates `MANIFEST.sha256` from the exact package bytes;
- emits a deterministic ZIP with fixed entry timestamps;
- refuses to overwrite an existing package.

It does not commit, push, publish, sign, or release anything.

Build only after the source changes have a real commit:

```powershell
.\scripts\build-cyonic-node-package.ps1 `
  -SourceCommit <40-character-commit>
```

`-SourceBranch` and `-SourceRepository` may point a CI candidate at its exact
review branch. A publishable package must be rebuilt against the accepted
`main` commit.

The resulting package remains a local candidate until its Windows, Linux, and
macOS cold paths have been run and recorded.
