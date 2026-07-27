# 13D Event Metadata

This directory holds tracked proposal envelopes and policy-bound metadata.

Git objects cannot contain their own commit OIDs without a circular hash
dependency. The binding therefore has two phases:

1. A tracked proposal envelope describes all thirteen coordinates, the requested
   effect, the parent events, and an external SHA-256 artifact root. Its
   `commitOid` and `treeOid` are null before the commit exists.
2. After the commit exists, `placetime-13d witness-git` produces a witness
   envelope containing the actual commit OID, tree OID, and parent OIDs. CI
   publishes that receipt, or a maintainer may attach it through a dedicated
   Git-notes ref.

Neither phase authorizes the requested effect.

