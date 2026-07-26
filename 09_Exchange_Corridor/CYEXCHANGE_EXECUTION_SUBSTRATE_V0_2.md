# PROPOSED: CyExchange Execution Substrate V0.2

**Record Type:** Architecture Specification  
**Layer:** Exchange Corridor  
**Status:** PROPOSED  

## 1. Substrate Reality

The V0.2 substrate transitions the CyExchange protocol from theoretical scaffolding to a runnable, testable implementation. By pairing the `cyxd` daemon with the `cyxctl` conformance harness, the multi-node reality of the network can be simulated safely under strict `010` observer constraints.

## 2. Invariants Enforced in V0.2 Code

1.  **Strict Monotonicity:** Handled by `pkg/cyexchange/inbox` and `pkg/cyexchange/outbox`. A state can never regress backwards.
2.  **Epistemic Separation:** The daemon possesses no knowledge of application logic; it strictly executes cryptographic schema verification.
3.  **Discovery Without Authority:** The `discovery` package allows nodes to locate each other via mDNS. This provides network proximity but zero authorization to cross the $\beta$ boundary.

## 3. The 101 Terminal Block

Most critically, the V0.2 execution substrate proves that an Envelope requesting an `EFFECT_REQUESTED` state is immediately halted and parked until the operator physically bridges the gap using Sovereign Apply. The automated substrate cannot, structurally, trigger an external actuation.
