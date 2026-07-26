# Why This Matters

## The Problem

Modern AI systems increasingly conflate planning with execution. Large Language Models (LLMs) routinely generate code, infrastructure commands, and deployment scripts. As these systems are integrated with execution tooling, the historical separation between planning an action and executing it is collapsing.

When an AI system operates without strict separation between interpretation and authorization, any hallucination, prompt injection, or logic error can translate directly into unauthorized action. We lose the fundamental governance boundary that ensures human intent governs system effect.

This collapse of boundaries matters urgently for safety-critical systems, software supply chains, and any domain where unauthorized actuation carries significant consequences. We need a way to mathematically guarantee that an AI system cannot authorize its own plans.

## The Claim

This repository demonstrates a formal, fail-closed architecture that separates interpretation from authorization. Under this model, an AI agent operates strictly as an interpreter—it proposes plans but literally cannot authorize or execute them. The core invariant is that the mathematical or AI output never autonomously writes to the authorization geometry. A human, or a strictly defined external policy engine, must issue a cryptographically bound permit for every effectual action.

## How It Works

The architecture follows a strict, five-step governance flow:

*   **AI proposes:** The AI generates a Candidate Artifact (a plan, script, or command).
*   **Validator checks:** A static validator checks the artifact against safety bounds and structure.
*   **Attester seals:** The system seals the validated artifact into Evidence, binding a cryptographic hash and runtime identity.
*   **Human/Policy authorizes:** An external human or policy engine reviews the Evidence and issues a cryptographic Permit.
*   **Actuator executes:** The Actuator verifies the Permit's binding to the Evidence and executes the action. If the permit is missing or invalid, the system fails closed.

## What You Can Run Today

The repository includes three runnable Go demonstrators:

1.  **Permit-Gated Cell:** Demonstrates the core fail-closed execution loop requiring external authorization.
2.  **Bleed-Over Membrane:** Demonstrates the Bleed-Over Operator (β) governing cross-field translation of claims while preventing certainty inflation.
3.  **Autophagic Decay Simulator:** Demonstrates the PROPOSED Autophagic Operator (α) addressing the echo-chamber problem when AI systems ingest their own outputs.

## Standards Alignment

This architecture aligns with established industry standards:

*   **IETF RATS** (Remote ATtestation procedureS)
*   **NIST ABAC** (Attribute-Based Access Control)
*   **in-toto** (Software Supply Chain Security)
*   **SLSA** (Supply-chain Levels for Software Artifacts)

## What Is Not Claimed

*   **These are demonstrators, not production software.** They are reference implementations meant to illustrate the architecture.
*   **Signatures are simulated.** The current Go code simulates cryptographic signing for ease of demonstration.
*   **The Autophagic Operator (α) is PROPOSED.** It is active research, not established active canon.
*   **This is a framework, not a final product.** It demonstrates a security pattern rather than providing a drop-in enterprise solution.
*   **This demonstrates, rather than proves.** We rely on rigorous engineering patterns, avoiding hyperbolic claims of absolute mathematical proof.
