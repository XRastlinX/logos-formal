# PROPOSED: Phase 0 Category Theory / SysML v2 Mapping Profile

**Profile ID:** `ct-sysmlv2-scalar-chain-0.1.0`  
**Artifact Status:** `PROPOSED_FOR_010_EVALUATION`  
**Governance State:** `010`  
**Validator Operator:** `000`  
**Governance Authority Effect:** `NONE`  
**SysML Basis:** OMG SysML v2.0, Language Specification `formal/2026-03-02`  
**Scope:** Illustrative scalar altitude → pressure → drag chain only  

---

## 1. Purpose

This profile defines one bounded mapping from selected SysML v2 constructs into the categories **Rel** and **Set**. Its purpose is to test whether categorical typing and composition can detect concrete modeling defects in a small, reproducible engineering example.

It is not:

- a replacement for KerML or the SysML v2 metamodel;
- a mapping of the complete SysML v2 language;
- an OMG conformance claim;
- an atmosphere or aerodynamic model;
- a proof of physical truth;
- a merge policy, deployment gate, or authority-bearing service.

## 2. Pilot Boundary

The pilot contains three typed value domains:

$$A = \{a \in \mathbb{R} \mid 0 \leq a \leq 5\} \text{ km}$$

$$B = \{b \in \mathbb{R} \mid 0.5b_0 \leq b \leq b_0\}$$

$$C = \{c \in \mathbb{R} \mid 0.5c_0 \leq c \leq c_0\}$$

where:

- $a$ is an altitude quantity;
- $b$ is an illustrative pressure quantity;
- $c$ is an illustrative drag-force quantity;
- $b_0 > 0$ is a declared reference pressure;
- $c_0 > 0$ is a declared reference drag force.

The profile defines:

$$f: A \rightarrow B, \quad f(a) = \left(1 - \frac{a}{10 \text{ km}}\right) b_0$$

and:

$$g: B \rightarrow C, \quad g(b) = \frac{c_0}{b_0} b$$

Their composition is:

$$(g \circ f)(a) = \left(1 - \frac{a}{10 \text{ km}}\right) c_0$$

For $a \in [0, 5]$ km, the direct image is:

$$(g \circ f)[A] = [0.5c_0, c_0]$$

These equations are normalized illustrative surrogates. They are not empirical atmosphere or aerodynamic laws.

## 3. Categorical Foundation

### 3.1 Relational Layer

The relational layer is a bounded subcategory of **Rel**:

- objects are the declared carrier sets $A$, $B$, and $C$;
- morphisms are binary relations between those sets;
- composition is relational composition;
- identities are equality relations on each carrier set.

A formalized requirement or constraint enters this layer as a relation or Boolean-valued predicate.

### 3.2 Functional Layer

The functional layer is a bounded subcategory of **Set**:

- objects are the same declared carrier sets;
- morphisms are total, single-valued functions;
- composition is ordinary function composition;
- identities are identity functions.

The profile uses the faithful, identity-on-objects embedding:

$$J: \mathbf{Set} \hookrightarrow \mathbf{Rel}$$

that maps a function to its graph.

No relation may be treated as a Set morphism until totality, single-valuedness, type compatibility, dimensional correctness, and domain coverage have been established.

### 3.3 Composition Rule

For $f: X \rightarrow Y$ and $g: Y' \rightarrow Z$, the composition $g \circ f$ is admissible only if:

1. $Y$ and $Y'$ are the same declared type, or an explicit approved conversion maps one into the other;
2. their physical dimensions agree;
3. the image of $f$ is contained in the accepted domain of $g$;
4. both functions are total over the composed operating envelope;
5. assumptions and uncertainty declarations do not conflict.

## 4. Categorical Mapping Profile

| SysML v2 Construct | Phase 0 Interpretation | Categorical Realization | Validation Rule |
|---|---|---|---|
| `package` and explicit import | Namespace and name-resolution boundary | No object or morphism | Imports must resolve; wildcard ambiguity is rejected |
| Attribute definition | Declared value type or quantity kind | Candidate carrier set | Type, dimension, unit policy, and value domain must be declared |
| Attribute usage | Variable or member of a carrier set | Element-valued feature; not automatically an object | Owning type and value domain must resolve |
| `calc def` with directed inputs and one result | Candidate deterministic calculation | Candidate Set morphism | Must be pure within the profile, total, single-valued, and dimensionally valid |
| Calculation usage | Application of a declared calculation | Morphism application | Arguments must match the declared domain in order and type |
| `constraint def` or constraint usage with Boolean result | Formal predicate | Rel morphism or subset of a product | Expression must be Boolean-valued and use supported operators |
| `requirement def` or requirement usage with formal `assume` and `require` constraints | Conditional predicate | Relation whose effective condition is assumptions ⇒ requirements | Text alone is documentation; only formal constraints enter Rel |
| Documentation comment | Human-readable explanation | No categorical realization | Preserved as provenance; never executed |
| Part definition or usage | Optional subject or scoping context | No Phase 0 object or morphism by default | Only subject binding is recognized; structural semantics are deferred |

The profile maps value domains — not arbitrary system components — to categorical objects. Calling every SysML element an "object" would be an unbounded and untested interpretation.

## 5. Supported SysML v2 Fragment

The following constructs are supported in Phase 0:

- packages and explicit imports required by the fixture;
- scalar attribute definitions and usages;
- `ScalarValues::Real`;
- quantity attributes specializing the applicable ISQ quantity kinds;
- SI quantity literals required by the fixture, including kilometre;
- calculation definitions and usages with directed input attributes, exactly one returned attribute, and a pure arithmetic result expression;
- constraint definitions and usages with Boolean result expressions;
- requirement definitions and usages containing formal `assume constraint` and `require constraint`;
- arithmetic operators: addition, subtraction, multiplication, division;
- comparison operators: equality, less than or equal, greater than or equal;
- Boolean conjunction;
- explicit parentheses;
- documentation comments.

Support means only that the Phase 0 mapper has declared semantics for the construct. It does not claim complete support for every legal variation of that SysML v2 construct.

## 6. Unsupported-Construct List

The following are outside Phase 0:

actions and control flow; states, transitions, events, and temporal semantics; ports, interfaces, connections, flows, and messages; allocations; items and material transfer; occurrence and snapshot semantics; verification, analysis, use-case, and trade-study execution; `assert constraint` as a model-consistency enforcement mechanism; stochastic, probabilistic, or fuzzy relations; differential equations and continuous-time simulation; vectors, matrices, tensors, and field-valued quantities; optimization objectives and solvers; multiplicities other than one scalar value per parameter; partial functions; nondeterministic calculations; recursive calculations; external-language calculation bodies; inheritance, redefinition, subsetting, and feature-chain semantics except where the fixture requires a simple quantity specialization; implicit unit conversion; uncertainty and correlation propagation; physical-interface compatibility; semantic interpretation of natural-language requirements; automatic elevation of a relation into a function.

Encountering an unsupported construct returns `NO_DECISION` with an `UNSUPPORTED_CONSTRUCT` finding. It does not produce `REJECT` unless the artifact falsely declares that the construct belongs to the supported profile.

## 7. Unit and Type Registry

| Registry ID | SysML-Oriented Type | Dimension | Canonical Unit | Allowed Domain | Notes |
|---|---|---|---|---|---|
| `P0-ALTITUDE` | Quantity specializing `ISQ::length` | Length | `km` | `[0,5] km` | Pilot input only |
| `P0-ALTITUDE-SCALE` | Quantity specializing `ISQ::length` | Length | `km` | Exactly `10 km` | Makes $a/(10\text{ km})$ dimensionless |
| `P0-PRESSURE` | Quantity specializing applicable ISQ pressure kind | Pressure | Profile-declared SI pressure unit | $[0.5b_0, b_0]$ | Illustrative surrogate |
| `P0-DRAG-FORCE` | Quantity specializing `ISQ::force` | Force | Profile-declared SI force unit | $[0.5c_0, c_0]$ | Illustrative surrogate |
| `P0-RATIO` | `ScalarValues::Real` | Dimensionless | `1` | `[0,1]` where applicable | No implicit percentage conversion |
| `P0-BOOLEAN` | Boolean | Dimensionless truth value | none | `{false, true}` | Constraint result |

### 7.1 Registry Rules

1. Addition and subtraction require identical dimensions.
2. Comparisons require compatible dimensions.
3. $a/(10 \text{ km})$ is dimensionless.
4. $(c_0/b_0)b$ has the dimension of force.
5. Unit conversion must be explicit and registered.
6. A bare literal `10` cannot replace `10 km` in the altitude-scale position.
7. Reference quantities $b_0$ and $c_0$ must be positive and supplied by the fixture.
8. Canonicalization changes representation, not physical meaning.

## 8. Claim Boundary

The Phase 0 implementation may claim:

> For the identified SysML v2 model bytes and mapping-profile version, the mapper recognized the supported scalar fragment, evaluated the declared typing and composition obligations, and emitted the recorded result.

It may not claim: complete SysML v2 conformance; OMG certification or endorsement; that Category Theory is the SysML v2 metamodel; functoriality beyond tested preservation obligations; atmospheric or aerodynamic validity; safety, optimality, or system correctness; objective semantic truth; canon elevation; authorization to merge, deploy, or actuate.

## 9. Validator and Authority Separation

The Phase 0 validator is a `000` observer.

It may: parse supported constructs; compute type and unit obligations; evaluate bounded fixtures; compute exact roots; emit `OBSERVE_ONLY`, `REJECT`, or `NO_DECISION`; exit nonzero when a declared invariant fails.

It may not: modify the SysML model; rewrite a relation as a function; block a commit or merge; approve a pull request; issue a Permit; elevate the profile; deploy or actuate an effect.

A separate authority-bearing policy layer may consume the receipt and decide whether to block a lifecycle transition. That layer must not claim operator state `000`.

## 10. Test Fixtures

| Fixture ID | Condition | Expected Status | Expected Decision | Required Observation |
|---|---|---|---|---|
| `P0-PASS-001` | Valid $f: A \rightarrow B$ and $g: B \rightarrow C$ | `COMPLETE` | `OBSERVE_ONLY` | Composite is typed $A \rightarrow C$ |
| `P0-PASS-002` | Evaluate image of `[0,5] km` | `COMPLETE` | `OBSERVE_ONLY` | Output range is $[0.5c_0, c_0]$ |
| `P0-PASS-003` | Compose $f$ with $id_B$ | `COMPLETE` | `OBSERVE_ONLY` | $id_B \circ f = f$ |
| `P0-PASS-004` | Interpret formal requirement assumptions ⇒ requirements | `COMPLETE` | `OBSERVE_ONLY` | Predicate remains in Rel unless function obligations are separately proved |
| `P0-REJECT-001` | Feed pressure output into a temperature input | `COMPLETE` | `REJECT` | `TYPE_OR_DIMENSION_MISMATCH` |
| `P0-REJECT-002` | Use `1 - altitude / 10` with a bare denominator | `COMPLETE` | `REJECT` | `DIMENSIONAL_ERROR` |
| `P0-REJECT-003` | Evaluate at `6 km` | `COMPLETE` | `REJECT` | `DOMAIN_VIOLATION` |
| `P0-REJECT-004` | Promote a one-to-many relation into Set | `COMPLETE` | `REJECT` | `RELATION_NOT_SINGLE_VALUED` |
| `P0-REJECT-005` | Compose where image of $f$ exceeds domain of $g$ | `COMPLETE` | `REJECT` | `COMPOSITION_DOMAIN_GAP` |
| `P0-NONE-001` | Natural-language requirement without formal constraint | `INCOMPLETE` | `NO_DECISION` | `UNASSESSED_TEXTUAL_REQUIREMENT` |
| `P0-NONE-002` | Model contains a state transition in the requested target | `INCOMPLETE` | `NO_DECISION` | `UNSUPPORTED_CONSTRUCT` |
| `P0-NONE-003` | Required type or unit cannot be resolved | `INCOMPLETE` | `NO_DECISION` | `PROFILE_DEPENDENCY_UNAVAILABLE` |

### 10.1 Fixture Invariants

Every fixture must record: exact source bytes; SysML v2 specification and toolchain version; mapping-profile ID and root; supported and encountered construct lists; unit-registry root; target root before and after evaluation; ordered findings; decision; validator operator `000`; governance authority effect `NONE`.

### 10.2 Illustrative Source Sketch

The following sketch communicates intended structure. It is not yet an executable conformance fixture:

```sysml
package CTPhase0ScalarChain {
    calc def PressureFromAltitude {
        in attribute altitude :> ISQ::length;
        in attribute referencePressure :> ISQ::pressure;
        return :> ISQ::pressure =
            (1 - altitude / 10[km]) * referencePressure;
    }

    calc def DragFromPressure {
        in attribute pressure :> ISQ::pressure;
        in attribute referencePressure :> ISQ::pressure;
        in attribute referenceDrag :> ISQ::force;
        return :> ISQ::force =
            (referenceDrag / referencePressure) * pressure;
    }

    constraint def PilotAltitudeDomain {
        in attribute altitude :> ISQ::length;
        altitude >= 0[km] and altitude <= 5[km]
    }
}
```

Before promotion to an executable fixture, this source must parse in the chosen SysML v2 implementation and its qualified quantity names must be confirmed.

## 11. Phase 0 Acceptance Gate

Phase 0 passes only when:

- [ ] The profile is assigned an immutable version and content root.
- [ ] The selected SysML v2 toolchain and version are recorded.
- [ ] Every supported construct has one positive and one negative fixture.
- [ ] Every unsupported construct returns NO_DECISION.
- [ ] Set is represented through the faithful embedding into Rel.
- [ ] Relation-to-function promotion obligations are mechanically checked.
- [ ] Units, dimensions, and operating domains are explicit.
- [ ] The illustrative source parses without modification.
- [ ] The output range $[0.5c_0, c_0]$ is mechanically reproduced.
- [ ] Linux and Windows produce the same profile and target roots.
- [ ] Platform-specific runtime and decision roots are separately recorded.
- [ ] The validator performs no repository or governance mutation.

Passing Phase 0 establishes only that the bounded mapping profile is internally implemented and testable. It does not authorize Phase 1, establish empirical engineering benefit, or elevate the profile to canon.

## 12. References

- OMG SysML v2.0 specification: https://www.omg.org/spec/SysML/2.0
- OMG SysML v2.0 Language Specification: https://www.omg.org/spec/SysML/2.0/Language/PDF
- Fong and Spivak, *Seven Sketches in Compositionality*: https://arxiv.org/abs/1803.05316

---

**Status:** PROPOSED — This profile has not been implemented or tested. It defines the mapping and acceptance criteria for a bounded Phase 0 evaluation.
