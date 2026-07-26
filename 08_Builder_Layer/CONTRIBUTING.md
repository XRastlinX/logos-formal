# @# D08 — Contributing

## Before changing a formal artifact

1. Read the root `DOMAIN_MAP_V1.md`.
2. Read `10_Status_Promotion_and_Canon_Ledger/CANON_STATUS.md`.
3. Read the relevant adjudication.
4. Preserve the artifact's status and authority effect unless a separate,
   evidence-bearing promotion record is in scope.

## Required separation

```text
content classification
!= lifecycle phase
!= component capability
!= permit state
!= external effect
!= artifact status
```

## Verification

From the repository root:

```powershell
python 08_Builder_and_Repository\tools\verify_formal_core.py
python 09_Exchange_and_Review_Corridor\tools\grok-bridge\test_bridge.py -v
```

Pull requests must state the affected domain, preserved invariant, exact test
results, and remaining proof obligations.
