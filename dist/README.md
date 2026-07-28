# First Contact package

This directory publishes the portable setup for the proposed first Cyonic
Service Surface.

Current package:

- [`cyonic-node-setup-v0.1.2-portable-7e0f263.zip`](cyonic-node-setup-v0.1.2-portable-7e0f263.zip)
- [`cyonic-node-setup-v0.1.2-portable-7e0f263.receipt.json`](cyonic-node-setup-v0.1.2-portable-7e0f263.receipt.json)

The archive pins source commit:

```text
7e0f263c2a31eaece5bf8ded76ce4ba47c7db32a
```

Its recorded SHA-256 is:

```text
8d1263b1dff0fcd84c42ac8911f132d270e8b172d50d06982d73d227499315a7
```

Unpack it, then follow its `README.md`. On Windows:

```powershell
# Prefer a short extraction root such as C:\cyonic-trial\
.\setup.cmd friend-node-01 codex
.\trial.cmd cyonic-first-contact-report-01.json
```

On Linux or macOS:

```bash
chmod +x scripts/*.sh
./scripts/setup.sh friend-node-01 codex
./scripts/trial.sh cyonic-first-contact-report-01.json
```

The trial wrapper produces the participant report required by the First
Contact evidence contract. It is interactive, create-only, and preserves
`PENDING_EXTERNAL_REVIEW`.

The archive and receipt have passed an internal Windows cold setup and trial.
That run is implementation evidence only:

```text
externalityStatus: NOT_ADJUDICATED
authorityEffect: NONE
```

It does not count toward the independent First Contact threshold.
