# PROPOSED: Cyxd Multi-Node Demonstrator V0.2

**Record Type:** Service Surface Expansion  
**Layer:** Exposure Layer  
**Status:** PROPOSED  

## 1. Two-Node Loopback Exchange

The `cyxd` daemon V0.2 introduces a runnable multi-node demonstrator simulating two physical entities communicating across the CyExchange corridor.

*   **Alpha Node (`config/alpha.yaml`):** Listens on port 8081.
*   **Beta Node (`config/beta.yaml`):** Listens on port 8082.

## 2. Constraints Enforced

When running the loopback transport, `cyxd` strictly enforces:
1.  **Transport Level:** HTTP payloads are blindly accepted into the routing handler, but isolated immediately.
2.  **Envelope Canonicalization:** The Envelope payload is hashed and checked against Ed25519 signatures. A single mutated byte causes a `400 Bad Request` and a terminal `REJECT` state in the Inbox adapter.
3.  **No Execution:** Even if the payload schema requests an action, the V0.2 node enforces `010` constraints. It outputs an `EvaluationOutcome` indicating `OUTCOME_OBSERVE_ONLY` or halts.

## 3. Running the Simulation
*(Requires compilation of `cmd/cyxd`)*

Terminal A:
```bash
cyxd serve 8081
```

Terminal B:
```bash
cyxctl send --dest localhost:8081 --payload envelope_example.json
```
