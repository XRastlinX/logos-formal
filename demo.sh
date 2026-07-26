#!/usr/bin/env bash
set -euo pipefail

# Logos-Formal: Run all demonstrators
# Requires: Go 1.24+
# Usage: ./demo.sh

GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ""
echo -e "${BOLD}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║  Logos-Formal: Fail-Closed AI Governance Demonstrators      ║${NC}"
echo -e "${BOLD}║  https://github.com/XRastlinX/logos-formal                  ║${NC}"
echo -e "${BOLD}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check Go is available
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed. Get it at https://go.dev/dl/${NC}"
    exit 1
fi

echo -e "${CYAN}Go version:${NC} $(go version)"
echo ""

# ── Demo 1: Permit-Gated Cell ───────────────────────────────────────
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  DEMO 1: Permit-Gated Actuation${NC}"
echo -e "  Shows: AI proposes → Validator checks → Attester seals →"
echo -e "         Human permits → Actuator executes. Then attacks it."
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
(cd "$SCRIPT_DIR/runtime/demonstrator" && go run .)
echo ""

# ── Demo 2: Bleed-Over (β) Membrane ─────────────────────────────────
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  DEMO 2: Bleed-Over (β) Epistemic Membrane${NC}"
echo -e "  Shows: Cross-field translation preserving provenance."
echo -e "         Rejects certainty inflation and provenance erasure."
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
(cd "$SCRIPT_DIR/runtime/dcif-membrane" && go run .)
echo ""

# ── Demo 3: Autophagic Decay (α) ────────────────────────────────────
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  DEMO 3: Autophagic Decay (α) — PROPOSED RESEARCH${NC}"
echo -e "  Shows: Certainty collapse when AI recursively ingests"
echo -e "         its own outputs. Forces reconnection to D=0 roots."
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
(cd "$SCRIPT_DIR/runtime/autophagic-decay" && go run .)
echo ""

# ── Tests ───────────────────────────────────────────────────────────
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  RUNNING TEST SUITE${NC}"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
(cd "$SCRIPT_DIR" && go test ./...)

echo ""
echo -e "${GREEN}${BOLD}All demonstrators complete. All tests passed.${NC}"
echo ""
echo -e "Read more:"
echo -e "  docs/WHY_THIS_MATTERS.md   — Plain-language explainer"
echo -e "  docs/ARCHITECTURE.md       — Visual system diagrams"
echo -e "  docs/POSITIONING.md        — Standards alignment (RATS, ABAC, SLSA)"
echo ""
