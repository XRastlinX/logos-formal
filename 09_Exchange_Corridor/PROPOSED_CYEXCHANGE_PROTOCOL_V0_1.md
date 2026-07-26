# PROPOSED: CyExchange Protocol V0.1

**Record Type:** Protocol Specification  
**Layer:** Exchange Corridor  
**Status:** PROPOSED  

## 1. Overview

The CyExchange Protocol is the operational organ of the Agency application. It provides the secure, verifiable routing of state transitions and artifacts between isolated agents, operators, and environments.

## 2. Core Mechanics

### 2.1 Envelopes and Node Cards
*   **Envelopes:** All payloads traversing the corridor are cryptographically wrapped in CyExchange Envelopes. Envelopes carry explicit metadata detailing provenance, destination, and the required interpretive schema.
*   **Node Cards:** Participants (agents, oracles, human operators) are identified by Node Cards, which securely bind their public keys and capability scopes.

### 2.2 Security Primitives
*   **mTLS:** Network transport is strictly secured via mutually authenticated TLS.
*   **Ed25519:** All message signing and Permit generation utilize pure Ed25519 signatures, adhering to RFC 8032 and RFC 8785 for canonical JSON representations.

### 2.3 Durability and Delivery
*   **Inbox / Outbox:** Messages are persisted to durable inbox/outbox queues prior to network transmission, ensuring that network failures result in recoverable, atomic retries.
*   **Delivery Semantics:** The corridor guarantees at-least-once delivery with idempotent receiver processing.
*   **Receipt Types:** Senders receive cryptographically signed receipts upon successful message ingestion and separate receipts upon successful application of operational effects.

## 3. Cubed Bit Routing and Invariants

The CyExchange Corridor inherently understands the Cubed Bit matrix and enforces the Permit separation doctrine.

### 3.1 "Send ≠ Apply"
Transmitting an Envelope through the corridor strictly constitutes a `Send` operation. It places data in a receiver's inbox. It *never* implies an implicit `Apply`. 

### 3.2 Invariant Enforcement
If an Envelope contains an actuator command targeting the `101` (Physical Domain), the corridor's conformance tests mandate that the Envelope must contain a valid Sovereign Apply Permit (`cyonic.permit/v1`). If the permit is absent, invalid, or expired, the routing mechanism halts execution and returns a `REJECT` receipt.
