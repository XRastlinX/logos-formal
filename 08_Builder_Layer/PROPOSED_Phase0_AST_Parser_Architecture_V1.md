# PROPOSED: Phase 0 Mechanical AST Parser Architecture

**Record Type:** Implementation Blueprint  
**Layer:** Builder  
**Status:** PROPOSED  
**Authority Effect:** NONE  
**Branch Scope:** `track-1-ast-parser`

---

## 1. 000 and 010 Non-Interference

The separation works through immutable input binding, distinct capabilities, and a one-way handoff:

```text
SysML source bytes
   ↓ hash + provenance only
000 witness receipt
   ↓ exact artifactRoot binding
010 evaluator
   ↓ semantic findings only
OBSERVE_ONLY | REJECT | NO_DECISION
```

### 1.1 Controls
- **Exact-byte binding:** The `010` evaluator recomputes the source hash and requires it to equal the `artifactRoot` in the `000` receipt.
- **Separate responsibilities:** `000` does not parse; `010` does not attest original acquisition.
- **Immutable input:** Read-only source bytes.
- **Restricted output vocabulary:** Success is `OBSERVE_ONLY`. Never `ALLOW`, `PERMIT`, or `APPLY`.
- **Fail-closed classification:** Invalid supported material $\to$ `REJECT`; unsupported material $\to$ `NO_DECISION`.

---

## 2. Bounded Go Evaluator Implementation Steps

Because we are building a deliberately bounded parser rather than full OMG XMI conformance, we explicitly pin the language basis and tokenize only the supported fragment.

### Step 1: Pin the Language Basis
```text
SysML specification: formal/26-03-02
Phase 0 profile root
Unit-registry root
```

### Step 2: Verify the Witness Handoff
- Read exact SysML bytes.
- Compute SHA-256 and compare with the `000` receipt.
- Reject on mismatch.

### Step 3: Tokenize the Supported Fragment
Recognize only: `package`, imports, `calc def`, `constraint def`, directed input attributes, single returns, scalar types, arithmetic, comparisons, and calculation calls.

### Step 4: Deliberately Small AST
The abstract syntax tree restricts itself strictly to the Phase 0 scalar chain domain.
```go
type Package struct {
    Imports     []Import
    Calcs       []CalcDef
    Constraints []ConstraintDef
}
```

### Step 5: Resolve Names and Imports
Construct a symbol table. Unresolved dependencies yield `NO_DECISION`.

### Step 6: Type and Dimension Checking
Assign every expression a value type, physical dimension, canonical unit, and declared domain. Track dimensions mathematically (e.g. $a/10\text{ km}$ is dimensionless).

### Step 7: Extract Carrier Domains
Convert `constraint def`s into `Interval` objects spanning lower and upper closed/open bounds.

### Step 8: Classify Functions and Relations
A `calc def` enters `Set` as a function only after proving totality, single-valuedness, output dimension correctness, and output-image containment. Otherwise it is a relation in `Rel`.

### Step 9: Evaluate Obligations
Execute the 12 golden fixtures. Exact rational arithmetic is preferred for coefficients.

### Step 10 & 11: Classify and Emit Receipt
Emit deterministic JSON receipts carrying `cubedBit: 010` and `authorityEffect: NONE`.

---

## 3. Rel and SysML Mapping

In `Rel`:
- Objects are sets.
- A morphism $R \subseteq X \times Y$ is a binary relation.

**SysML to Phase 0 Mapping:**
- Scalar value domain $\to$ Object/Set
- Formal `constraint def` $\to$ Relation or predicate
- Valid total, single-valued `calc def` $\to$ Function in `Set`

A `calc def` is first evaluated as a relation. The evaluator rejects attempted relation-to-function promotion (e.g., one-to-many outputs, domain gaps), preserving the strict mathematical distinction between `Rel` and `Set`.
