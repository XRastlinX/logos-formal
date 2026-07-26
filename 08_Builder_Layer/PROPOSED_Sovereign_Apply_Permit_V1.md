# PROPOSED: Sovereign Apply Permit Schema and Verification

**Record Type:** Cryptographic Design Blueprint  
**Layer:** Core Governance  
**Status:** PROPOSED  
**Authority Effect:** NONE  

## 1. Required Verification Corrections

The Permit architecture enforces non-interference between interpretive modeling and physical actuation. To ensure the implementation remains consistent with the Cy-Egology governance model, the following core strictures are mandatory:

1. **Key Resolution Must Remain Pure:** 
   The verifier should not fetch keys or revocation data during verification. It must be given a pinned trust-bundle snapshot:
   ```text
   Verify(permit, trustBundle, revocationSnapshot, trustedTime) → VALID | REJECT
   ```
   *Note:* The bundle and revocation snapshot each need content roots and freshness policy. A `did:key` contains public-key material, so it does not solve trust by itself; the issuer must still be externally pinned.

2. **Nonce Consumption State Machine:**
   Nonce consumption cannot generally be atomic with a GitHub merge. A local database and GitHub do not share one transaction. A durable state machine is required:
   ```text
   ISSUED → RESERVED → EFFECT_REQUESTED → APPLIED | FAILED | OUTCOME_UNKNOWN
   ```
   *Note:* If the network fails after the merge request, record `OUTCOME_UNKNOWN` and reconcile before retrying. Never assume that a missing receipt means no effect occurred.

3. **Distinguish Git Object IDs from Artifact Hashes:**
   Git repositories use 40-character SHA-1 or 64-character SHA-256 object IDs. Do not blanket label every commit `sha256:`. Use explicit `object_format` matching GitHub's preconditions:
   ```json
   "expected_pr_head": {
     "object_format": "sha1",
     "oid": "<40 lowercase hex>"
   },
   "artifact_root": "sha256:<64 lowercase hex>"
   ```

4. **No Cryptographic Null Permit:**
   `SHA256("")` identifies empty bytes. It does not mean absent, revoked, or unauthorized. Represent those states explicitly:
   ```text
   permit: absent
   permitStatus: REVOKED
   verificationDecision: REJECT
   ```

---

## 2. Corrected Signed Core (YAML Schema)

```yaml
core:
  schema: cyonic.permit/v1
  version: "1.0"
  permit_id: prm_...
  alg: Ed25519
  issuer: principal:...
  kid: 2026-key-01
  subject: actor:...
  audience: pep:github-merge
  policy_root: sha256:<64 hex>
  scope:
    repository: XRastlinX/logos-formal
    pull_number: 1
    base_ref: refs/heads/main
    expected_pr_head:
      object_format: sha1
      oid: <40 hex>
    expected_base_head:
      object_format: sha1
      oid: <40 hex>
    action: MERGE_PULL_REQUEST
  artifact_root: sha256:<64 hex>
  nonce: <canonical base64url, 32 decoded bytes>
  issued_at: 1234567890
  not_before: 1234567890
  expires_at: 1234571490
  constraints:
    max_uses: 1
signature: <canonical base64url Ed25519 signature>
```

### 2.1 Cryptographic Notes
- The maximum clock skew belongs to the verifier policy, not the Permit, because a Permit must not loosen its own validation boundary.
- Signature Construction:
  ```text
  Ed25519(
    "cyonic.permit.v1\0" || JCS(core)
  )
  ```
- Pin plain Ed25519 explicitly (RFC 8032) rather than leaving Ed25519/Ed25519ph ambiguous. RFC 8785 defines the canonical JSON representation (JCS).

---

## 3. Current Adjudication

```text
Permit specification: PROPOSED
Crypto implementation: NOT_IMPLEMENTED
Atomic Apply guarantee: NOT_ESTABLISHED
Current Permit path: SIMULATED
authority_effect: NONE
8208 relationship: symbolic only
External evidence: unchanged
```
*(No repository changes or elevation were performed.)*
