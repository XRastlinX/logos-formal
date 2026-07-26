# Independent review request: Rite Machine and normalization

Review the supplied Rite Machine profile, its adjudication, and the corrected
Seven-Stage formal specification.

## Questions

1. Is the developmental interpretation faithful to the formal lifecycle
   without turning interpretation into authorization?
2. Should normalization remain a post-lifecycle pure verifier, or is there a
   rigorous reason to make it an eighth lifecycle transition?
3. Propose a typed `NormalizationPolicy` containing exactly six independently
   testable criteria. For each criterion provide:
   - identifier;
   - input type;
   - deterministic predicate;
   - threshold or exact acceptance condition;
   - typed failure code;
   - evidence required for replay.
4. Define the result type:

   ```text
   Valid(normalizationReceipt) | Reject(reasonCodes)
   ```

5. Demonstrate that normalization:
   - preserves the analyzer capability vector `010`;
   - creates no permit;
   - commits no external effect;
   - does not automatically authorize retention or export.
6. Identify any remaining category errors involving:
   - content state;
   - lifecycle phase;
   - capability vector;
   - permit state;
   - human developmental language.

## Required response

- findings classified as `BLOCKING`, `MATERIAL`, or `EDITORIAL`;
- explicit agreement and disagreement;
- proposed formal signatures;
- six normalization test vectors, including at least two rejection cases;
- `governance_state: 010`;
- `authority_effect: NONE`;
- no Owner, Seal, Apply, or state-commit claim.
