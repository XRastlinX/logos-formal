# Positioning

This document frames the logos-formal system against several funded problem domains and established engineering paradigms.

## High-Assurance Actuation
*Aerospace, medical devices, industrial control systems*

This architecture addresses the integration of AI into high-assurance environments by physically and cryptographically separating the AI's planning from the execution of physical systems. It aligns with existing safety-critical standards by ensuring an AI cannot unilaterally actuate hardware, requiring a discrete, auditable permit for every system effect.

## Attested Toolchains
*SLSA, in-toto, software supply chain integrity*

The system aligns with software supply chain integrity models by treating AI-generated actions as artifacts requiring rigorous attestation. It provides a framework for generating and sealing evidence about the AI's runtime state and the resulting plan, requiring verifiable signatures before any deployment or execution occurs.

## AI Control Interfaces
*The alignment/control problem from an engineering angle*

Rather than relying on behavioral alignment or prompt engineering, this architecture addresses AI control through formal capability restriction. It provides a framework for structural containment, ensuring that regardless of the AI's internal state or intent, it mathematically lacks the authorization tokens required to affect the external environment. 

## Fail-Closed Capability Separation
*Capability-based security, least privilege*

The architecture implements a strict fail-closed capability model mapped to a "Cubed Bit" state lattice. It aligns with the principle of least privilege by locking the AI in a pure interpretation state (010), demonstrating how to prevent capability escalation by requiring an external authority (101) to grant execution rights.

## RATS / ABAC / Zero-Trust
*Direct standards mapping*

The governance flow directly maps to IETF RATS (Remote ATtestation procedureS) and NIST ABAC (Attribute-Based Access Control). It addresses zero-trust requirements by treating every AI proposal as untrusted, requiring continuous, attribute-based validation and attestation before granting execution permits.
