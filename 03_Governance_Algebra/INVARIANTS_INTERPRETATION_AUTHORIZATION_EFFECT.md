# Invariants: Interpretation ≠ Authorization ≠ Effect

**Record Type:** Epistemic Canon  
**Layer:** Governance Algebra  
**Status:** ACTIVE_CANON  

## 1. The Core Constitutional Invariant

The fundamental law of the `logos-formal` architecture is the strict segregation of semantic interpretation from cryptographic authorization and physical effect. 

**Interpretation does not imply Authorization. Authorization does not guarantee Effect.**

### 1.1 The Cubed Bit Model
Every assertion must map to the Cubed Bit matrix, differentiating the nature of the claim:
*   **The Witness (`000`)**: Records exact bytes and provenance without interpretation.
*   **The Interpreter (`010`)**: Parses syntax and evaluates logical models. Yields `OBSERVE_ONLY`, `REJECT`, or `NO_DECISION`.
*   **The Actuator (`101`)**: Performs state-changing physical operations, gated by strict cryptographic authorization.

### 1.2 The β Boundary
The `β` boundary separates the Physical Domain ($A$, the empirical world) from the Model Domain ($B$, abstract representation).
*   Models cannot dictate physical truth.
*   The `β` function maps empirical measurements into the Model Domain. 
*   Statements made in $B$ must remain explicitly identified as models or surrogates unless bridged by concrete `β` receipts.

### 1.3 The Permit Doctrine
*   **Sovereign Apply:** No operational change, deployment, or state mutation can occur via implicit assumption or system default.
*   **Explicit Cryptography:** Every `101` effect requires a formally structured, verifiable Permit (`cyonic.permit/v1`) signed by an authorized principal.
*   **Fail-Closed:** Without a verified Permit, the system defaults to `REJECT`.

### 1.4 The Apply Boundary
Agents, Oracles, and Interpreters may discover, stage, and propose configurations. However, they lack the inherent authority to cross the Apply Boundary. The Sovereign (human operator or mathematically designated authority) retains sole promotive authority.
