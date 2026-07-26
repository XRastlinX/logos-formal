# PROPOSED: CyExchange State Machine V0.1

**Record Type:** State Machine Architecture  
**Layer:** Exchange Corridor  
**Status:** PROPOSED  

## 1. Inbox/Outbox Transitions

The CyExchange State Machine manages the lifecycle of an Envelope as it transverses the protocol. To decouple network volatility from operational integrity, the state machine enforces strict transitional logic.

### 1.1 Outbox States
1.  **QUEUED:** Envelope committed to local durable storage.
2.  **TRANSMITTING:** Envelope currently traversing the mTLS boundary.
3.  **DELIVERED:** Cryptographic receipt received from the destination's Inbox.
4.  **FAILED:** Exhausted retry logic or explicit cryptographic rejection from the destination.

### 1.2 Inbox States
1.  **RECEIVED:** Envelope durably persisted locally; receipt dispatched.
2.  **EVALUATING:** Envelope contents passed to the `010` parser/evaluator for syntactic and semantic validation.
3.  **OBSERVE_ONLY:** Evaluation complete. No further action required.
4.  **EFFECT_REQUESTED:** Envelope contains a valid Sovereign Apply Permit requesting a `101` state change.
5.  **APPLIED:** Operational effect successfully executed and empirically verified.
6.  **REJECTED:** Validation failed, signature invalid, or constraints violated.

## 2. Conformance and Re-Entrancy

Transitions between states must be strictly monotonic. An Envelope cannot revert from `EFFECT_REQUESTED` to `EVALUATING`. If an `APPLIED` state transition fails midway, it must halt at `OUTCOME_UNKNOWN` until an operator mechanically reconciles the system. Automated re-entrancy of physical effects is prohibited.
