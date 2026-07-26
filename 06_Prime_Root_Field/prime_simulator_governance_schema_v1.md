# Prime Simulator Governance Schema V1

**Record Type:** Formal JSON Schema
**Layer:** Simulator Output Bound
**Status:** ACTIVE_CANON
**Governing Rule:** Constitutio Alchemica

---

This schema strictly enforces the epistemic boundary of the prime gap simulator. By embedding the Governance Algebra directly into the data contract, the schema guarantees that any downstream consumer reads the output structurally as **pure interpretation (`010`)**—permanently stripped of any authorization or effectuation claims.

It also explicitly codifies the mathematical null-case for the first prime gap (\(3 - 2 = 1\)).

## JSON Schema Definition

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://antigravity.local/schemas/prime_simulator_governance_v1",
  "title": "Prime Gap Simulator Record",
  "description": "A mathematically rich, governance-null, interpretation-only record of a prime gap transform.",
  "type": "object",
  "required": [
    "prime_1",
    "prime_2",
    "gap",
    "A",
    "B",
    "C",
    "D",
    "governance_state",
    "authorization",
    "interpretation",
    "effectuation",
    "binding"
  ],
  "properties": {
    "prime_1": {
      "type": "integer",
      "minimum": 2
    },
    "prime_2": {
      "type": "integer",
      "minimum": 3
    },
    "gap": {
      "type": "integer",
      "minimum": 1
    },
    "A": {
      "type": ["array", "null"],
      "items": { "type": "integer" }
    },
    "B": {
      "type": "array",
      "items": { "type": "integer" }
    },
    "C": {
      "type": ["array", "null"],
      "items": { "type": "integer" }
    },
    "D": {
      "type": ["array", "null"],
      "items": { "type": "integer" }
    },
    
    "governance_state": {
      "type": "string",
      "const": "010",
      "description": "Hardcoded 010 state indicating Interpretation only."
    },
    "authorization": {
      "type": "integer",
      "const": 0,
      "description": "Must be 0. This record cannot grant permission."
    },
    "interpretation": {
      "type": "integer",
      "const": 1,
      "description": "Must be 1. This record contains interpretive modeling (prime factors)."
    },
    "effectuation": {
      "type": "integer",
      "const": 0,
      "description": "Must be 0. This record cannot actuate a physical state."
    },
    "binding": {
      "type": "string",
      "const": "non-authoritative",
      "description": "Declares that any patterns found herein are descriptive, never prescriptive or normative."
    }
  },
  
  "allOf": [
    {
      "description": "Mathematical Invariant: No prime-root is defined for gap = 1.",
      "if": {
        "properties": { "gap": { "const": 1 } }
      },
      "then": {
        "properties": {
          "A": { "type": "null" },
          "B": { "type": "array", "maxItems": 0 },
          "C": { "type": "null" },
          "D": { "type": "null" }
        }
      }
    }
  ]
}
```

## Structural Guarantees

### 1. `010` is a First-Class Machine Constraint
Because `governance_state`, `authorization`, `interpretation`, and `effectuation` are typed using JSON Schema `const`, any simulator emitting this format will fail validation if it attempts to assert `authorization: 1` or `effectuation: 1`. The epistemic boundary is now a machine-checkable compilation error.

### 2. The Null-Prime Rule
The `allOf` block enforces the condition:
\[g_i = 1 \Rightarrow A_i = null,\; B_i = \emptyset,\; C_i = null,\; D_i = null\]
This acts as an algorithmic lock against downstream agents attempting to extract meaning or "prime structure" out of the singular gap between 2 and 3.

### 3. The Non-Binding Flag
The `binding: "non-authoritative"` tag ensures that even if this simulator identifies a profound cycle, cluster, or distribution, that pattern legally remains a description. It cannot be used to authorize an entitlement force (Fact $\to$ Veritas Mandatum) because it is structurally decoupled from authority at the schema layer.
