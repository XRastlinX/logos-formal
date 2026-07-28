#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CANON_DIR="$ROOT/canon/logos-formal"
EXECUTABLE="$ROOT/.cache/cyonic-service"
EXPECTED_COMMIT="@@SOURCE_COMMIT@@"
REPORT="${1:-cyonic-first-contact-report-01.json}"

[[ -d "$CANON_DIR/.git" ]] || {
  echo "Canon checkout is missing. Run scripts/setup.sh first." >&2
  exit 2
}

ACTUAL_COMMIT="$(git -C "$CANON_DIR" rev-parse HEAD)"
[[ "$ACTUAL_COMMIT" == "$EXPECTED_COMMIT" ]] || {
  echo "Canon checkout does not match the pinned source commit." >&2
  exit 2
}
[[ -z "$(git -C "$CANON_DIR" status --porcelain)" ]] || {
  echo "Canon checkout has local changes; the trial requires the pinned clean tree." >&2
  exit 2
}
[[ -x "$EXECUTABLE" ]] || {
  echo "Validated service executable is missing. Run scripts/validate.sh first." >&2
  exit 2
}

case "$REPORT" in
  /*) REPORT_PATH="$REPORT" ;;
  *) REPORT_PATH="$ROOT/$REPORT" ;;
esac

(
  cd "$CANON_DIR"
  "$EXECUTABLE" trial \
    -origin external \
    -report "$REPORT_PATH"
)

printf '\nFIRST_CONTACT_REPORT_CREATED\n'
printf 'Report: %s\n' "$REPORT_PATH"
printf 'Status: PENDING_EXTERNAL_REVIEW\n'
printf 'Authority effect: NONE\n'
