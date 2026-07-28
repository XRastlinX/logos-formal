#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CANON_DIR="$ROOT/canon/logos-formal"
EXPECTED="@@SOURCE_COMMIT@@"

printf 'Node Setup: @@PACKAGE_VERSION@@\n'
printf 'Expected commit: %s\n' "$EXPECTED"
printf 'Governance state: 010\nAuthority effect: NONE\n'

if [[ -f "$ROOT/config/node.yaml" ]]; then
  printf '\n'
  cat "$ROOT/config/node.yaml"
fi

if [[ -d "$CANON_DIR/.git" ]]; then
  ACTUAL="$(git -C "$CANON_DIR" rev-parse HEAD)"
  printf '\nActual commit: %s\n' "$ACTUAL"
  [[ "$ACTUAL" == "$EXPECTED" ]] && printf 'Pinned commit match: true\n' || printf 'Pinned commit match: false\n'
  [[ -z "$(git -C "$CANON_DIR" status --porcelain)" ]] && printf 'Working tree clean: true\n' || printf 'Working tree clean: false\n'
else
  printf 'Canon checkout: MISSING\n'
fi
