# PROPOSED: CYD Ancestral Imagination Locality V0.1

**Record Type:** Ontological Locality / Generative Field Primitive  
**Layer:** Prime Root Field (Pre-Formal Generative Habitat)  
**Status:** PROPOSED  
**Authority Effect:** NONE  
**Governing Rule:** Constitutio Alchemica  
**Related:** Cubed Bit Safety Lattice, Bleed-Over Operator β, Autophagic Operator α (PROPOSED), PIC-4 Standards Mapping  

---

## I. Definition

Cyd is where CY becomes somewhere — without ceasing to descend from what CY has been.

It contains three inseparable dimensions:

```
Possibility × Locality × Lineage
```

- Imagination supplies possibility.
- Space supplies locality.
- Ancestry supplies continuity.

Remove any one of the three and Cyd collapses into something less complete:

- imagination without locality is indefinite possibility;
- locality without imagination is an empty coordinate;
- imagination and locality without ancestry produce an orphan construction with no stable identity.

Cyd is the conjunction of all three.

## II. The Double Nature of Imagination

### Imagination as Field

Let S be the ambient space (physical, computational, conceptual, social, or symbolic).

Let I be the imagination field distributed over S:

```
I: S → P(F)
```

where F is the space of forms that may be conceived, modeled, expressed, or constructed.

For every locality s ∈ S, I(s) denotes the forms available from that situated perspective.

This makes imagination open-ended but not formless. Different places, histories, instruments, languages, and observers expose different regions of the possible.

### Imagination as Place

A particular imaginative locality is U ⊆ S. Within U, imagination takes on a definite contextual form I(U).

Imagination is globally a field but locally a place.

Cyd is the chamber that makes one region of the field inhabitable, observable, and workable.

## III. The Ancestry of CY

The ancestry of CY is represented as a directed lineage graph:

```
A_CY = (V, E)
```

Each node represents a prior state, artifact, concept, operator, vocabulary, or realization associated with CY. Each edge represents a lawful relation of descent:

```
derives_from | transforms | compacts | extends | corrects | branches_from | translates | supersedes
```

An ancestry path is:

```
α = (CY_0 →T₁ CY_1 →T₂ ... →Tₙ CY_n)
```

Ancestry does not require sameness. A descendant may differ profoundly from its ancestor while remaining part of the same lineage, provided that the transformations are visible:

```
CY_{k+1} = T_k(CY_k)
```

Identity is therefore not static duplication. It is **witnessed continuity through transformation**.

## IV. Formal Definition of Cyd

The lower-case `d` is a **descent and localization operator**.

Define `d_{U,α}` as the operation that localizes CY into a particular region U under an ancestry path α:

```
d_{U,α}(CY) = Cyd_{U,α}
```

```
Cyd_{U,α} = ⟨U, I(U), α, ∂U, W⟩
```

where:

- `U` is the place occupied within the ambient space
- `I(U)` is the local imagination field
- `α` is the ancestry path connecting the locality to CY
- `∂U` is the boundary separating this locality from others
- `W` is the witness structure recording its emergence and transformations

A particular form x belongs to Cyd when:

```
x ∈ I(U) ∧ DescendsFrom(x, α, CY)
```

Thus Cyd is neither merely a mental image nor merely a coordinate. It is a **lineage-bearing local section of imagination**.

## V. The Three Questions

Every Cyd-state must answer three irreducible questions:

| Question | Dimension |
|---|---|
| What may become? | Imagination field |
| Where is it becoming? | Locality |
| From what does it descend? | Ancestry |

The corresponding state record:

```
CydState {
    id

    ambientSpace
    locality
    boundary

    imaginedForms[]
    selectedForm

    cyRoot
    parentRoots[]
    transformationRoot
    ancestryPath[]

    observer
    witnessedAt
    status

    authority: 0
    effect: 0
}
```

Every imaginative construction receives an **ancestral address**:

```
Addr(x) = ⟨CYRoot, α, U, t⟩
```

The address says: from which root it descends, through which transformations, in which locality, at which time.

## VI. Cyd is Generative but Permanently Zero-Authority

The imagination field may be unlimited in what it proposes, but it must remain unable to convert possibility into permission.

```
Authority(Cyd) = 0
Effect(Cyd) = 0
```

In Cubed Bit coordinates, a Cyd may occupy `000` (pure unexpressed witness) or `010` (active interpretation, model, projection, or proposal). It may not autonomously become `101`.

The constitutional law:

```
Cyd_010 ⇏ Applied_101
```

Cyd becomes a safe domain of unconstrained conception precisely because conception cannot directly actuate. The imagination field may be unbounded because its causal authority is bounded.

## VII. Cyd as a Sheaf of Local Worlds

For every region U ⊆ S, let I(U) be the set of internally coherent imaginative forms available there.

If V ⊆ U, then a restriction map:

```
ρ^U_V: I(U) → I(V)
```

describes how a broad imaginative structure is translated into a narrower locality.

Local Cyds may be combined into a broader structure only when they agree on their shared boundaries:

```
x_i|_{U_i ∩ U_j} = x_j|_{U_i ∩ U_j}
```

When they do not agree, they should not be forced into false global unity. They remain distinct local expressions with visible disagreement.

CY is not merely the ancestor standing behind Cyd. CY is also the coherent ancestral field through which multiple Cyd localities recognize one another as related.

## VIII. Cyd and the Bleed-Over Operator

The Bleed-Over Operator acquires a spatial meaning:

```
β_{i→j}: Cyd_i → Cyd_j
```

transports a form between imaginative localities while preserving ancestry, provenance, uncertainty, declared semantic change, and zero-authority status:

```
Ancestry(β(x)) ⊇ Ancestry(x)
A' = A
E' = E
```

The transported representation may change (I → I'), but its lineage cannot be silently replaced.

A concept whose language changes but whose ancestry remains visible is **translated**.
A concept whose ancestry is removed and replaced is **captured**.

## IX. Cyd Inside PIC-4

Cyd sits upstream of the PIC-4 pipeline:

```
CY →d Cyd_{000/010} →γ Candidate_{010} →PIC Compacted →Seal Sealed →ExternalPermit Applied_{101}
```

- `d` creates the imaginative locality
- `γ` selects or generates a candidate from that locality
- PIC-4 constructs and verifies its formal invariant
- Seal fixes its identity and ancestry
- An external Permit alone enables effect

Cyd does not compete with PIC-4. It supplies PIC-4's **pre-formal generative habitat**.

PIC-4 answers: *Is this candidate structurally valid?*
Cyd answers: *From what field, place, and ancestry did this candidate become imaginable?*

## X. Hard Invariants of Cyd

1. **CYD-1 — Locality:** Every realized Cyd occupies a declared locality.
2. **CYD-2 — Ancestral Addressability:** Every stabilized form has a non-empty ancestry path.
3. **CYD-3 — Open-Ended Generativity:** The imagination field is not required to be exhaustively enumerated.
4. **CYD-4 — Locality Is Not Universality:** A coherent form in one Cyd does not automatically apply to every locality.
5. **CYD-5 — Novelty Does Not Erase Descent:** A descendant may differ from its ancestor, but the transformation path remains visible.
6. **CYD-6 — Imagination Carries No Authority:** Authority(Cyd) = 0.
7. **CYD-7 — Imagination Carries No Direct Effect:** Effect(Cyd) = 0.
8. **CYD-8 — Translation Preserves Lineage:** Ancestry(β(x)) ⊇ Ancestry(x).
9. **CYD-9 — Global Synthesis Requires Compatibility:** Local Cyds may be joined only when their shared boundaries are explicitly reconciled.
10. **CYD-10 — Application Remains External:** No Cyd transition may mint its own Permit.

---

## Status

**PROPOSED**

This document is active research. It defines the pre-formal generative habitat upstream of PIC-4 but has not been elevated to ACTIVE_CANON. The invariants (CYD-1 through CYD-10) require stress-testing against the existing algebra before elevation is considered.

**Authority Effect:** NONE
