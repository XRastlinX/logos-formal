# PROPOSED: Phase 0 `010` Mechanical Verification Steps

**Record Type:** Verification Procedure  
**Governance State:** `010`  
**Validator Operator:** `010` (Mechanical Evaluator)  
**Witness Operator:** `000` (Raw Attester)  
**Authority Effect:** NONE  

---

## 1. The Epistemic Handoff

The integrity of this evaluation relies on a strict separation of powers defined by the Cubed-Bit capability lattice:

1. **The `000` Witness:** This component secures the exact bytes of `CTPhase0ScalarChain.sysml` and provides a cryptographic hash commitment. It acts as the "Raw attester" or "Witness without judgment." It asserts nothing about the correctness of the model.
2. **The `010` Validator:** This component mechanically evaluates the fixtures against the parsed SysML v2 Abstract Syntax Tree (AST). It acts as the "Interpreter" but holds `authority_effect: NONE`. It cannot write a status change to the repository.

## 2. Mechanical Execution Steps for the `010` Validator

To enforce the structural obligations of the Phase 0 mapping profile, the `010` Validator must perform the following mechanical steps against the `CTPhase0ScalarChain.sysml` file:

### Step 1: Parse Phase
Parse `CTPhase0ScalarChain.sysml` into a SysML v2 Abstract Syntax Tree (AST). 
*   **Gate:** If the syntax is invalid or fails to parse, emit `REJECT`.

### Step 2: Domain Extraction
Extract the asserted constraints for the carrier sets from the `constraint def` blocks:
*   `PilotAltitudeDomain`: `[0[km], 5[km]]`
*   `PilotPressureDomain`: `[0.5 * referencePressure, referencePressure]`
*   `PilotDragDomain`: `[0.5 * referenceDrag, referenceDrag]`

### Step 3: Morphism Typing Check
For the composite calculation `AltitudeToDrag`:
*   Verify that the `return` type of `AltitudeToPressure` (`ISQ::pressure`) exactly matches the `in` type parameter `pressure` of `PressureToDrag` (`ISQ::pressure`).
*   Verify that dimensions and units resolve structurally across the boundary.

### Step 4: Fixture Evaluation
*   Inject the input values from the 12 Phase 0 test fixtures (defined in the profile matrix) into the AST representation of `AltitudeToDrag`.
*   Mechanically evaluate the mathematical results against the defined domain constraints (`PilotAltitudeDomain`, `PilotPressureDomain`, `PilotDragDomain`).
*   Verify whether the output mathematically matches the expected status outcome (`PASS`, `REJECT`, or `NO_DECISION`) defined in the Phase 0 mapping matrix.

### Step 5: Receipt Emission
Output a deterministic JSON receipt of the evaluation.
*   **Gate:** The receipt must explicitly carry `authority_effect: NONE`. The validation evaluates internal consistency; it does not issue a Permit or merge a pull request.
