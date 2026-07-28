#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CANON_DIR="$ROOT/canon/logos-formal"
CACHE_DIR="$ROOT/.cache"
EVIDENCE_DIR="$ROOT/evidence"
PORT="${1:-18787}"
BASE_URL="http://127.0.0.1:$PORT"
EXECUTABLE="$CACHE_DIR/cyonic-service"
EXPECTED_COMMIT="@@SOURCE_COMMIT@@"
EXPECTED_TREE="@@SOURCE_TREE@@"
INSTANCE_ID="cyonic-validation-$$-$(date -u +%s)"

[[ -d "$CANON_DIR/.git" ]] || {
  echo "Canon checkout is missing. Run scripts/setup.sh first." >&2
  exit 2
}

mkdir -p "$CACHE_DIR" "$EVIDENCE_DIR"

ACTUAL_COMMIT="$(git -C "$CANON_DIR" rev-parse HEAD)"
[[ "$ACTUAL_COMMIT" == "$EXPECTED_COMMIT" ]] || {
  echo "Canon checkout does not match the pinned source commit." >&2
  exit 2
}
[[ "$(git -C "$CANON_DIR" rev-parse 'HEAD^{tree}')" == "$EXPECTED_TREE" ]] || {
  echo "Canon checkout does not match the pinned source tree." >&2
  exit 2
}
[[ -z "$(git -C "$CANON_DIR" status --porcelain)" ]] || {
  echo "Canon checkout has local changes; validation requires the pinned clean tree." >&2
  exit 2
}

REPOSITORY_VALIDATION="$("$CANON_DIR/scripts/validate.sh" --format json)"
for required in \
  '"validationStatus": "COMPLETE"' \
  '"status": "VALIDATED"' \
  '"cubedBit": "010"' \
  '"validatorOperator": "000"' \
  '"governanceAuthorityEffect": "NONE"' \
  '"executionContainment": "HOST"' \
  '"routerDecision": "OBSERVE_ONLY"'
do
  grep -Fq "$required" <<<"$REPOSITORY_VALIDATION" || {
    echo "Repository validation output violated the node boundary: missing $required" >&2
    exit 2
  }
done

(
  cd "$CANON_DIR"
  go build -o "$EXECUTABLE" ./runtime/cyonic-service
)

"$EXECUTABLE" serve-demo -listen "127.0.0.1:$PORT" -instance-id "$INSTANCE_ID" \
  >"$CACHE_DIR/server.stdout.log" 2>"$CACHE_DIR/server.stderr.log" &
SERVER_PID=$!
cleanup() {
  kill "$SERVER_PID" 2>/dev/null || true
  wait "$SERVER_PID" 2>/dev/null || true
}
trap cleanup EXIT

PROBE_OUTPUT=""
for _ in $(seq 1 40); do
  if PROBE_OUTPUT="$("$EXECUTABLE" probe-http \
      -base-url "$BASE_URL" \
      -source-ref "$EXPECTED_COMMIT" \
      -expected-instance "$INSTANCE_ID" 2>/dev/null)"; then
    break
  fi
  kill -0 "$SERVER_PID" 2>/dev/null || {
    echo "Service exited before readiness." >&2
    cat "$CACHE_DIR/server.stderr.log" >&2
    exit 2
  }
  sleep 0.25
done

[[ -n "$PROBE_OUTPUT" ]] || {
  echo "Cold-call probe did not pass." >&2
  exit 2
}
grep -Fq "\"sourceRef\": \"$EXPECTED_COMMIT\"" <<<"$PROBE_OUTPUT" || {
  echo "Probe receipt did not bind the pinned source commit." >&2
  exit 2
}
grep -Fq "\"serverInstance\": \"$INSTANCE_ID\"" <<<"$PROBE_OUTPUT" || {
  echo "Probe receipt did not bind the started server instance." >&2
  exit 2
}

STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
REPOSITORY_EVIDENCE_PATH="$EVIDENCE_DIR/repository-validation-$STAMP.json"
SERVICE_EVIDENCE_PATH="$EVIDENCE_DIR/service-probe-$STAMP.json"
printf '%s\n' "$REPOSITORY_VALIDATION" | tee "$REPOSITORY_EVIDENCE_PATH"
printf '%s\n' "$PROBE_OUTPUT" | tee "$SERVICE_EVIDENCE_PATH"
printf 'Repository evidence: %s\n' "$REPOSITORY_EVIDENCE_PATH"
printf 'Service evidence: %s\n' "$SERVICE_EVIDENCE_PATH"
