# First Contact package

This directory publishes the portable setup for the proposed first Cyonic
Service Surface.

Current package:

- [`cyonic-node-setup-v0.1.1-portable-7e0f263.zip`](cyonic-node-setup-v0.1.1-portable-7e0f263.zip)
- [`cyonic-node-setup-v0.1.1-portable-7e0f263.receipt.json`](cyonic-node-setup-v0.1.1-portable-7e0f263.receipt.json)

The archive pins source commit:

```text
7e0f263c2a31eaece5bf8ded76ce4ba47c7db32a
```

Its recorded SHA-256 is:

```text
2b539eea8ee392bb5f1318b3661b1d04bc090bffb19e7482d9813453a9001c12
```

Unpack it, then follow its `README.md`. On Windows:

```powershell
.\setup.cmd friend-node-01 codex
```

On Linux or macOS:

```bash
chmod +x scripts/*.sh
./scripts/setup.sh friend-node-01 codex
```

The archive and receipt have passed an internal Windows cold run. That run is
implementation evidence only:

```text
externalityStatus: NOT_ADJUDICATED
authorityEffect: NONE
```

It does not count toward the independent First Contact threshold.
