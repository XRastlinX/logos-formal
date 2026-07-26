# ACTIVE_CANON: Cybot Pattern Specification V1

**Record Type:** Local-First System Architecture  
**Layer:** Builder Layer / Implementations  
**Status:** ACTIVE_CANON  
**Governing Rule:** Constitutio Alchemica  
**Related:** Bleed-Over Operator β V1, PIC-4 Standards  

---

## 1. What is a Cybot?

A **Cybot** is a local, persona-aware operator assistant bound to a BuilderLMS. It is a governed epistemic surface—an AI integration that synchronizes analytical outputs (such as lesson materials, thesis notes, and operator calculus transformations) directly into a local Learning Management System (LMS) or Knowledge Graph. 

A Cybot **is not** an autonomous agent that acts on external human structures. It cannot grade students, permanently alter public records, or issue credentials. A Cybot translates and synthesizes information under strict epistemic constraints (Bleed-Over rules), outputting structured knowledge while keeping `Authority = 0` and `Effect = 0`.

## 2. Required Local Components

Standing up a Cybot requires four primary components:

1. **Oracle Surface:** The UI and interaction layer where the operator inputs dilemmas or queries (e.g., the Alchemical Operator Oracle web app).
2. **Persona Set:** The behavioral profiles the Cybot adopts to perform specific semantic transformations (e.g., strict formalist, dialectic, empirical, or specific historical personas like Carl Jung).
3. **LMS Adapter:** A local integration bridge (REST API, Webhook, or file sync) that pushes structured outputs (NotebookLM packages, lesson modules) into platforms like Canvas, Moodle, Blackboard, or custom endpoints.
4. **State Vector Matrix:** A real-time tracker of the operator's learning or exploration state, mapping their trajectory within the formal field.

## 3. Epistemic Constraints

A compliant Cybot operates strictly within the formal geometry of the [Cubed Bit Safety Lattice](../10_Elevation_Layer/ACTIVE_CANON_Bleed_Over_Operator_Beta_V1.md) and the DCIF:
- **Rule 8 / Uncertainty Non-Inflation:** The Cybot may compress, summarize, and adapt text to its persona, but it may **never** inflate certainty. It cannot present a hypothesis as a proven fact, nor erase the provenance of its sources.
- **Authority Effect = NONE:** The Cybot operates strictly in the `010` (pure interpretation) coordinate. Only an external operator (the human) can issue a permit (`101`) to finalize the material into a formal syllabus or publication.

## 4. Initialization Prompt (Start Here)

To build your own Cybot using an LLM session, paste the following prompt into your preferred AI builder:

```text
Act as a Cybot Builder. We are constructing a local, persona-aware operator assistant bound to a local LMS integration.

**Architecture Requirements:**
1. Design an Oracle Surface (UI) for interacting with the persona.
2. Create a standard JSON/REST adapter to push synthesized lesson plans and briefings to my local LMS.
3. Enforce strict epistemic governance: You may translate and synthesize, but you may not inflate certainty or assume administrative authority. All outputs must maintain provenance.

Please begin by scaffolding the initial React/Vite web application and defining the LMS adapter schema.
```

## 5. Deployment

Once built, the Cybot is deployed locally. It acts as an integration loop between the operator's chaotic exploration and the structured, permanent ledger of their LMS or knowledge base.
