#!/usr/bin/env python3
"""Small fail-closed repository invariant verifier."""

from __future__ import annotations

import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]

REQUIRED_FILES = (
    ".github/copilot-instructions.md",
    "AGENTS.md",
    "DOMAIN_MAP_V1.md",
    "README.md",
    "01_Governance_Capability_and_Effect/cubed-bit-governance-capability-lattice-v1.md",
    "02_Cord_Algebra/CORD_ALGEBRA_V1.md",
    "03_Q_Content_Classification/Q_CONTENT_CLASSIFICATION_STATUS_V1.md",
    "04_Safe_Paired_Projection/q-cy-dual-projection-adjudication-v1.md",
    "05_Analytical_Lifecycle_and_Rite/rite-machine-v1-adjudication.md",
    "05_Analytical_Lifecycle_and_Rite/rite-machine-v1.md",
    "05_Analytical_Lifecycle_and_Rite/seven-stage-transition-and-product-category-adjudication-v1.md",
    "06_Prime_Root_Field/prime-root-field-v1-completion-audit.md",
    "07_Collaborator_and_Connector_Capabilities/agency-of-agents-cylang-schema-v1-adjudication.md",
    "07_Collaborator_and_Connector_Capabilities/COPILOT_CAPABILITY_PROFILE_V1.md",
    "08_Builder_and_Repository/CONTRIBUTING.md",
    "08_Builder_and_Repository/SECURITY.md",
    "09_Exchange_and_Review_Corridor/tools/grok-bridge/bridge.py",
    "09_Exchange_and_Review_Corridor/tools/grok-bridge/test_bridge.py",
    "09_Exchange_and_Review_Corridor/grok-rite-normalization-drive-mint-v1.json",
    "10_Status_Promotion_and_Canon_Ledger/CANON_STATUS.md",
    "10_Status_Promotion_and_Canon_Ledger/NORMALIZATION_CRITERIA_V1.md",
    "11_DeepSeek_External_Reviewer_Reach/DEEPSEEK_CAPABILITY_PROFILE_V1.md",
    "11_DeepSeek_External_Reviewer_Reach/deepseek-remote-text-reviewer-v1.json",
)

REQUIRED_ASSERTIONS = {
    "README.md": (
        "Q never sets Cy. Cy never rewrites Q.",
        "capability_vector: 010",
        "authority_effect: NONE",
        "does not mutate the analyzer into\n`101`",
    ),
    "10_Status_Promotion_and_Canon_Ledger/CANON_STATUS.md": (
        "HOLD - REQUIRES FORMAL REVISION",
        "PROPOSED - CORRECTED",
        "do not alter an artifact's status",
    ),
    ".github/copilot-instructions.md": (
        "The Prime-Root Field analyzer remains `010`",
        "No `sigma` transition",
        "`101` is not shorthand for successful application",
    ),
}

FORBIDDEN_ACTIVE_CLAIMS = (
    "sigma_6(i',1,f',010) = (i',1,f',101)",
    "σ₆(i',1,f',010) = (i',1,f',101)",
    "all artifacts as authoritative and immutable",
)


def main() -> int:
    errors: list[str] = []
    for relative in REQUIRED_FILES:
        if not (ROOT / relative).is_file():
            errors.append(f"MISSING_REQUIRED_FILE:{relative}")

    for relative, assertions in REQUIRED_ASSERTIONS.items():
        path = ROOT / relative
        if not path.is_file():
            continue
        text = path.read_text(encoding="utf-8")
        for assertion in assertions:
            if assertion not in text:
                errors.append(f"MISSING_ASSERTION:{relative}:{assertion}")

    active_surfaces = (
        ROOT / "README.md",
        ROOT / ".github/copilot-instructions.md",
        ROOT / "AGENTS.md",
    )
    for path in active_surfaces:
        if not path.is_file():
            continue
        text = path.read_text(encoding="utf-8")
        for forbidden in FORBIDDEN_ACTIVE_CLAIMS:
            if forbidden in text:
                errors.append(f"FORBIDDEN_ACTIVE_CLAIM:{path.relative_to(ROOT)}:{forbidden}")

    if errors:
        for error in errors:
            print(error, file=sys.stderr)
        print(f"FORMAL_CORE_REJECTED:{len(errors)}", file=sys.stderr)
        return 1

    print(f"FORMAL_CORE_ACCEPTED:{len(REQUIRED_FILES)} required files")
    print("GOVERNANCE_STATE:010")
    print("AUTHORITY_EFFECT:NONE")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
