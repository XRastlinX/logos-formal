# PROPOSED: CyExchange Error Taxonomy V0.2

**Record Type:** Protocol Specification  
**Layer:** Exchange Corridor  
**Status:** PROPOSED  

## 1. Overview

To strictly enforce the invariant geometry of the CyExchange Protocol, all edge cases and failures must be mechanically quantifiable. The CyExchange Error Taxonomy classifies corridor failures to ensure monotonic handling. 

## 2. Taxonomy Classes

### 2.1 Transport Errors (`ErrTransport`)
Failures in the mTLS mesh, connectivity, or node discovery (mDNS).
*   *Action:* Retry subject to exponential backoff.

### 2.2 Validation Errors (`ErrValidation`)
The Envelope violates structural schemas (e.g., missing required fields, illegal Cubed Bit `required_state` transitions like `111`).
*   *Action:* Terminal `REJECT`. Return receipt.

### 2.3 Signature Errors (`ErrSignature`)
The cryptographic Ed25519 signature over the payload or envelope metadata is invalid.
*   *Action:* Terminal `REJECT`. Immediate receipt generation to log tampering attempts.

### 2.4 Permit Errors (`ErrPermit`)
The Sovereign Apply Permit attached to a `101` Envelope request is missing, structurally invalid, or fails cryptographic verification by the stated issuer.
*   *Action:* Terminal `REJECT`. Immediate halt of the `101` pipeline.

### 2.5 Replay & Expiration Errors (`ErrReplay`, `ErrExpiration`)
The envelope or permit contains a reused nonce, or the monotonic TTL (Time-to-Live) bounds have expired.
*   *Action:* Terminal `REJECT`. Prevents delayed action execution.

### 2.6 Lineage Errors (`ErrLineage`)
The payload claims a `010` observer outcome, but the hashed lineage chain back to the `000` byte witness is broken or fabricated.
*   *Action:* Terminal `REJECT`. Epistemic contamination halted.

## 3. Enforcement

Any of the above errors transitioning an envelope into the `FAILED` or `REJECTED` state immediately terminates execution. The envelope cannot be "rescued" or modified. A new envelope must be issued.
