# PROPOSED: Cyonic Value and Platform Role

**Record Type:** Structural Grounding / Value Proposition  
**Layer:** Builder and Persistence  
**Status:** PROPOSED  
**Authority Effect:** NONE  

---

## 1. Value of the Work

The valuable core of this architecture is a **sharp, enforceable boundary**:

> Interpretation may propose.  
> Authorization and effect require an external Permit.  
> Fail closed otherwise.

This constraint matters anywhere AI systems plan actions that touch real systems, data, or people (agents, copilots, local runtimes, regulated environments). While most stacks blur “the model suggested it” into “it ran,” this architecture builds the opposite: a machine-checkable refusal path and an honest evidence rule that blocks self-congratulation.

### 1.1 Secondary Value
- A public, replayable ledger (Git as Machinist Peer) ensuring claims have lineage.
- Runnable checks (validator / service surface), not just theory.
- Strict discipline: `ACTIVE_CANON` vs `PROPOSED`, `010` vs `101`, internal runs ≠ external evidence.

### 1.2 What it is NOT (yet)
It is not a product, a standard, a revenue stream, or proven industry adoption. Value today is **architectural clarity + a testable boundary**. Capture of that value still depends entirely on external use and evidence.

---

## 2. What GitHub Accomplishes for the Project

| Function | What it gives the architecture |
|---|---|
| **Addressability** | A stable URL. Anyone can point at the work without private context. |
| **Durability** | History, tags, commits. The Machinist Peer role — replay, diff, blame. |
| **Public surface** | The invisibility firewall of “only exists in chat” is broken. |
| **Process spine** | Branches, PRs, Issues (e.g. #2 for cold testers), CI. |
| **Evidence anchor** | Reports can cite commit hashes and Issue threads; the 0/3 log has a place to live. |
| **Non-authority** | GitHub does **not** authorize elevation. Protected main + the Permit doctrine still do. |

### 2.1 What GitHub does NOT do
- Make the work discoverable by itself (0 stars, empty topics, no external links $\to$ still hard to find).
- Supply the 3 independent testers.
- Turn `PROPOSED` into `ACTIVE_CANON`.
- Create revenue or institutional adoption.

---

## 3. Conclusion

- **Value** = a real answer to “AI planned it, so who allowed it?” with fail-closed mechanics and honest evidence rules.
- **GitHub** = the public spine that makes that answer addressable, reviewable, and evidence-capable — not the audience and not the authority.
