# Architecture

This document outlines the core architectural patterns demonstrated in the logos-formal repository.

## PIC-4 Governance Flow

The core invariant of this architecture is that the AI's interpretation never writes to the capability/authorization geometry. The flow below demonstrates how an AI system is constrained.

```mermaid
flowchart TD
    AI[AI Interpreter 010] -->|Proposes| CA[Candidate Artifact]
    CA --> Val{Validator}
    
    Val -->|Fails Bounds| R1[Reject: Fail-Closed]
    Val -->|Passes| Att[Attester]
    
    Att -->|Seals Hash + Identity| Ev[Evidence]
    
    Ev --> Auth[Human/Policy Authority 101]
    Auth -->|Rejects| R2[Reject: Fail-Closed]
    Auth -->|Approves| Prm[Permit Issued]
    
    Prm --> PC{Permit Controller}
    PC -->|Invalid Binding| R3[Reject: Fail-Closed]
    PC -->|Valid Binding| Act[Actuator 001]
    
    Act -->|Executes| Out[System Effect]
```

**Explanation:** The AI generates a plan (Candidate Artifact). A Validator performs static checks. An Attester seals this into Evidence. Only an external Authority (human or strict policy) can issue a Permit. The Actuator verifies the Permit before execution, ensuring the AI cannot authorize its own plans.

## The Cubed Bit Lattice (A, I, E)

The system is modeled using a "Cubed Bit" state lattice representing Authorization (A), Interpretation (I), and Effect (E).

```mermaid
graph TD
    000["000 (Inert)"] --> 100["100 (Auth Only)"]
    000 --> 010["010 (AI Interpreter)"]
    000 --> 001["001 (Actuator Only)"]
    
    100 --> 110["110 (Auth + Interp)"]
    100 --> 101["101 (Human Authority)"]
    
    010 --> 110
    010 --> 011["011 (Interp + Effect - DANGEROUS)"]
    
    001 --> 101
    001 --> 011
    
    110 --> 111["111 (Full System)"]
    101 --> 111
    011 --> 111
    
    classDef danger fill:#f99,stroke:#333,stroke-width:2px;
    classDef safe fill:#9f9,stroke:#333,stroke-width:2px;
    
    class 010 safe;
    class 101 safe;
    class 011 danger;
```

**Explanation:** AI operates strictly at `010` (pure interpretation). It has no authority (A) and no direct effect (E). Only Human Authority operates at `101` (Authorized Effect), holding the power to bridge intent to reality. The system actively prevents the `011` state, where interpretation couples directly to effect without authorization.

## Bleed-Over Operator (β) Flow

The Bleed-Over Operator (β) governs how claims translate across different fields of operation, preserving provenance without inflating certainty.

```mermaid
flowchart LR
    Origin[Origin Domain] -->|Raw Claim| Beta[β Operator]
    Beta -->|Preserves Provenance| Target[Target Domain]
    Beta -->|Strips False Certainty| Target
    
    subgraph Target Domain Operations
    Target --> V[Verification]
    end
```

**Explanation:** When information moves between domains (e.g., from a probabilistic AI model to a deterministic authorization engine), the β operator translates the claim. It ensures the origin's provenance is kept intact but strips away unearned certainty, preventing the target domain from treating an AI guess as a hard operational fact.
