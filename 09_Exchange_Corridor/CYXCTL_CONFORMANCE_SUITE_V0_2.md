# PROPOSED: cyxctl Conformance Suite V0.2

**Record Type:** Conformance Harness  
**Layer:** Exchange Corridor  
**Status:** PROPOSED  

## 1. Overview

The `cyxctl` tool serves as both a manual testing tool and an automated conformance harness for the CyExchange Execution Substrate. It validates the constraints enforced by the `cyxd` daemon.

## 2. Command Reference

### `send`
Transmits a `CyExchangeEnvelope` over the active transport layer (loopback HTTP for V0.2).

### `inspect-node`
Leverages the mDNS registry to detect active NodeCards and output their reported `capability_scope`.

### `verify-receipt`
Audits the cryptographic signature and monotonic timestamps of a `DeliveryReceipt` or `EvaluationReceipt`.

### `conformance`
Automates the injection of malformed envelopes, expired Permits, and illegal `111` required states. The suite strictly requires that the `cyxd` daemon returns terminal `REJECT` for all invalid injections, mathematically verifying the safety of the `010` epistemic lock before Sovereign Apply is ever considered.
