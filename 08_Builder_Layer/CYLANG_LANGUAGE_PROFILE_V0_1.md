# CYLANG LANGUAGE PROFILE V0.1

**Record Type:** Language Specification / Intermediate Representation
**Status:** PROPOSED
**authority_effect:** NONE
**target_runtime:** Go

---

**Definition:** Cylang is an AI-assisted governance DSL and intermediate representation targeting Go. It is not an AI; it is a governed specification language through which AI can safely produce deterministic, capability-gated Go code.

## 1. Lexical Grammar and AST
The Cylang grammar extends a C-like syntax with explicit capability and effect keywords.

- **Keywords:** `pure`, `measure`, `stage`, `witness`, `apply`, `fn`, `struct`, `ensures`
- **AST Nodes:** The AST natively isolates Effect Declarations from Capability Values.
- **Functions:** Every function must declare an effect. The absence of an effect keyword is a compilation error.

## 2. Static Types
Cylang enforces strict structural typing. Core governance types are unforgeable and cannot be cast or mocked.

- `Evidence`: Opaque struct representing raw, unmodified observational data.
- `Result`: The output of a `pure` computation.
- `Candidate`: A structured proposal for state mutation.
- `Attestation`: A cryptographically verifiable claim regarding a Candidate.
- `Permit<T>`: A parameterized capability value locking authorization to a specific cryptographic digest (`T`).
- `ApplyReceipt`: A terminal record confirming effectuation.

## 3. Effect System
The effect system separates computation, observation, and mutation. Crucially, **Permit is a capability value, not an effect.**

* `pure`: Computation without effects.
* `measure`: Evidence acquisition. Carries no authorization.
* `stage`: Candidate artifact creation. Carries no authorization.
* `witness`: Attestation without authority. Enforces `authority_effect == NONE`.
* `apply`: Explicitly authorized effect, strictly requiring a `Permit`.

**Syntax Example:**
```cylang
pure fn derive(e: Evidence) -> Result

measure fn observe(source: Source) -> Evidence

stage fn prepare(x: Result) -> Candidate

witness fn attest(c: Candidate) -> Attestation
    ensures authority_effect == NONE

apply fn commit(
    c: Candidate,
    p: Permit<HashOf<c>>
) -> ApplyReceipt
```

## 4. Canonical IR
The Intermediate Representation (IR) strips syntactic sugar and resolves all types into a strict, serialized graph of governance dependencies. The IR explicitly maps every function call against the Seven-Stage Transition axioms to ensure no capability lift occurs outside of $\sigma_6$.

## 5. Deterministic Go Lowering
Cylang lowers into Go by mapping effects to interfaces and capability values to strict generic structs. The Go generator makes illegal states structurally difficult or impossible to construct.

**Go Lowering Example:**
```go
func Apply(
    ctx context.Context,
    candidate SealedCandidate,
    permit Permit[CandidateDigest],
) (ApplyReceipt, error)
```
*Note: The `Permit[CandidateDigest]` uses Go generics to bind the authorization explicitly to the candidate's hash, preventing permit replay or reassignment.*

## 6. Conformance and Rejection Tests
Promotion eligibility to `ACTIVE_CANON` requires demonstrating fail-closed behavior across the compiler pipeline. 

The following criteria must be met:
- Grammar frozen for the declared version.
- Deterministic AST and IR serialization.
- Effect-checking rules mathematically proven.
- Go-generation replay equivalence (idempotent transpilation).
- **Negative Tests:** Explicit test suites proving that `measure`, `stage`, and `witness` paths absolutely cannot produce a `Permit`.
- Delivery of a reference compiler or verifier capable of asserting these invariants.
