# PROPOSED: Cy-Egology Adoption Arc V1

**Record Type:** Evidence-Gated Adoption Strategy
**Layer:** Builder, Exposure, and Sustainability
**Status:** PROPOSED
**Authority Effect:** NONE
**Depends On:** Cyonic Codexica Sustainability Strategy V1

---

## 1. Strategic Position

The project seeks to become a small, technically credible reference for fail-closed separation of:

```
interpretation
authorization
effect
```

It is not presently established as a framework, standard, certification scheme, adopted dependency, or commercial product.

The phases below are evidence gates. They are not a guaranteed timeline, an inevitable growth sequence, or a basis for projecting revenue.

## 2. Phase Gates

### Phase 1 — First Contact

**Objective:** Make the public surface legible and runnable without prior project context.

**Required evidence:**

- three cold users independently locate the primary claim;
- median time from clone to passing demonstration is at most ten minutes;
- each user can accurately distinguish proposal, validation, authorization, and effect;
- at least one independently reported, reproducible cold-run limitation or
  confusion is recorded as verified friction;
- public claims match the demonstrators' actual test boundaries.

**Exit event:** `FIRST_CONTACT_VALIDATED`

Repository visits, stars, model praise, or internal walkthroughs do not satisfy
this gate. If the first three qualifying participants report no friction, the
cohort expands until one qualifying friction event is recorded or this
requirement is formally revised.

### Phase 2 — First Applied Friction

**Objective:** Observe a real external workflow, beyond the bounded cold-run
trial, and the limitation that prevents or complicates it.

**Required evidence:**

- an external actor identifies a concrete workflow beyond the First Contact
  setup and trial;
- the actor attempts to apply at least one project component;
- the encountered limitation is reproducible or documented with sufficient context;
- the limitation is classified as product, documentation, integration, governance, or conceptual friction;
- no internal proposal is counted as external friction.

**Exit event:** `FIRST_APPLIED_FRICTION_RECORDED`

Friction is evidence about use, not evidence that a requested feature is correct or should be built.

### Phase 3 — First Real Adoption

**Objective:** Establish recurring external reliance on a bounded component.

**Required evidence:**

- an identifiable external actor uses a named version or commit;
- use persists across at least two separate work cycles;
- the actor can describe the operational purpose served;
- a maintainer can reproduce or inspect the dependency path;
- limitations, support expectations, and status are explicit.

**Exit event:** `FIRST_ADOPTION_CONFIRMED`

A clone, fork, citation, unpaid trial, or letter of intent alone is not adoption.

### Phase 4 — Tooling Gravity

**Objective:** Productize the repeated component demanded by actual use.

**Required evidence:**

- at least two independent adopters require substantially the same component;
- the repeated requirement is more specific than a general request for "governance" or "AI safety";
- the component has a stable input, output, failure model, and support scope;
- implementation cost and ongoing maintenance are estimated;
- negative tests exist for unauthorized or contaminated paths.

**Candidate outputs:**

- permit-gate library
- policy/effect checker
- evidence-lineage monitor
- receipt and replay validator
- bounded starter kit

**Exit event:** `TOOLING_CANDIDATE_VALIDATED`

Formal elegance, document count, or a desired Cylang architecture cannot substitute for repeated external demand.

### Phase 5 — Institutional Touch

**Objective:** Enter a credible standards-adjacent, research, or organizational conversation.

**Required evidence:**

- an external institution initiates or accepts a substantive exchange;
- every named standards relationship has an explicit crosswalk;
- "aligned with," "informed by," and "compliant with" remain distinct;
- the work's empirical and formal limits are disclosed;
- no invitation, citation, grant inquiry, or workshop is described as endorsement unless the institution explicitly states that.

**Possible events:**

```
EXTERNAL_CITATION_RECORDED
WORKSHOP_INVITATION_RECORDED
GRANT_DISCUSSION_RECORDED
STANDARDS_CROSSWALK_REVIEWED
```

There is no automatic promotion from institutional attention to certification, standardization, safety approval, or commercial value.

## 3. Evidence Record

Every claimed transition should produce:

```json
{
  "eventId": "opaque-id",
  "eventType": "FIRST_CONTACT_VALIDATED",
  "observedAt": "RFC3339 civil timestamp",
  "actorRefs": ["scoped-external-actor"],
  "artifactRefs": ["commit-or-content-digest"],
  "evidenceRefs": ["test-log-or-interview-record"],
  "criteriaVersion": "CY_EGOLOGY_ADOPTION_ARC_V1",
  "limitations": [],
  "authorityEffect": "NONE"
}
```

The event records declared evidence. It does not prove honest intent, complete disclosure, market value, lawful ownership, or semantic truth.

## 4. Non-Inflation Rules

1. A later phase may not be claimed while an earlier phase lacks its required evidence.
2. Evidence from the project team is not counted as independent external evidence.
3. Multiple derivatives of one interaction count as one evidence family.
4. Attention is not adoption.
5. Adoption is not correctness.
6. Correctness under declared tests is not certification.
7. Institutional contact is not endorsement.
8. Payment is not technical authorization.
9. A requested feature receives no authority merely because a user requested it.
10. Negative or off-target feedback remains evidence and must not be silently discarded.

## 5. Observation Discipline

The project should record what outsiders actually reach for before selecting a flagship product.

Useful friction categories include:

| Category | Example signal |
|---|---|
| Legibility | User cannot state the core separation accurately |
| Tryability | Clone or first run fails |
| Integration | Permit boundary does not fit the user's architecture |
| Evidence | Lineage or receipt data cannot be supplied |
| Governance | Required decision owner or policy is absent |
| Product | Repeated manual step suggests a reusable tool |
| Scope mismatch | User wants a different problem solved |

The smallest recurring friction with a clear owner is a stronger product signal than the most elaborate internal roadmap.

## 6. Maintainer Obligations

Across all phases, maintainers must:

- correct public overclaims promptly and visibly;
- retain `PROPOSED`, `VALIDATED`, and `ACTIVE_CANON` distinctions;
- keep experimental operators from issuing authority;
- record failed tests and rejected interpretations;
- disclose when evidence comes from related or derivative sources;
- resist selecting a product before external use evidence exists;
- preserve a public path from claim to implementation and test.

## 7. Current Adjudication

The public repository establishes addressability and contains runnable reference material. That places the project inside Phase 1.

The following have not yet been demonstrated by the evidence presently bound to this strategy:

```
FIRST_CONTACT_VALIDATED
FIRST_APPLIED_FRICTION_RECORDED
FIRST_ADOPTION_CONFIRMED
TOOLING_CANDIDATE_VALIDATED
INSTITUTIONAL_TOUCH
```

The immediate work remains:

1. reconcile public claims with the actual demonstrator behavior;
2. complete three cold-clone tests;
3. record the first external reactions without inflating them;
4. treat the first reproducible external friction as the next design input.

## 8. Status

This record preserves the longer strategy while preventing phase inflation. It authorizes no outreach, contract, product build, standards claim, or public status change.
