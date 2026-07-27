# PROPOSED: 13D Git Matrix Source Registry V0.1

**Record Type:** Typed Governance Coordinate Registry  
**Layer:** Builder and Persistence  
**Status:** PROPOSED  
**Authority Effect:** NONE  
**Physical Dimension Claim:** NONE  
**Spacetime Replacement Claim:** NONE

## Purpose

The 13D registry is the source model for describing governed repository events.
It preserves the coordinate partition authored in the earlier runnable
**13D Placetime & Epistemic Governance Studio**, then hardens its primitive
fields into typed records. Source archive and file hashes are recorded in
`registry/source_lineage.json`.

The normative machine-readable source is:

```text
registry/13D_coordinates.yaml
```

The repository binding is:

```text
registry/repository_profile.json
```

Generated and checked artifacts include:

```text
registry/event_envelope.schema.json
internal/placetime13d/registry_generated.go
```

## Coordinate Partition

| ID | Coordinate | Meaning |
|---|---|---|
| d01 | observed_time | Clock observation with source and uncertainty |
| d02 | latitude | Latitude under a declared geodetic reference |
| d03 | longitude | Longitude under the same reference |
| d04 | vertical_coordinate | Altitude or depth with unit, datum, and direction |
| d05 | semantic_axis_1 | First profile-defined semantic coordinate |
| d06 | semantic_axis_2 | Second profile-defined semantic coordinate |
| d07 | provenance_integrity_index | Method-bound provenance assessment, not truth |
| d08 | lineage_commitment | Cryptographic lineage commitment, never a Permit |
| d09 | observer_confidence | Method-bound observer assessment, not truth |
| d10 | instrumentation_class | Versioned sensor or agency classification |
| d11 | certifier_reference | Identity/key reference, never a Permit |
| d12 | claim_radius | Claim reach under a declared metric, unit, and scope |
| d13 | entity_binding | Entity identifier constraint, not authority |

All thirteen keys are present in every event vector. A coordinate may use its
declared null or absence representation; omission is not equivalent to absence.

## Governance Fence

Coordinates describe an observation or request. They do not authorize it.
Every event envelope carries the following fields outside the coordinate vector:

```text
requestedEffect
permitRef
policyRoot
decision
authorityEffect
validationReceipts
```

In particular:

```text
d08 lineage_commitment != Permit
d11 certifier_reference != Permit
d13 entity_binding != Permit
confidence or integrity score != truth
schema conformance != authorization
```

## Git Binding

The Git binding preserves distinct commitments:

```text
commit OID
tree OID
parent OIDs
external SHA-256 artifact root
```

A Git OID identifies a Git object under the repository's configured object
format. The external artifact root identifies the observed content under the
declared SHA-256 tree-hashing profile. Neither substitutes for the other.

Git provides content identity, ancestry, branching, and replay. Those Git and
governance controls are envelope fields; they do not replace the authored
thirteen-coordinate vector. External governance remains responsible for any
authorization or effect.

## Metric Boundary

The registry lists candidate derived metrics only. A metric becomes executable
only when its numerator, denominator, unit, scope, observation window,
uncertainty, and source are separately declared. No normative ratio is inferred
from coordinate names alone.

## Non-Claims

This profile does not:

- extend or replace physical spacetime;
- turn Git history into physical time;
- make a ledger physically irreversible;
- establish truth, legitimacy, or legal authority;
- authorize a transition;
- promote itself to ACTIVE_CANON.
