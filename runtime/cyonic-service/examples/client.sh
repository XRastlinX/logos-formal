#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)"
FIXTURE_DIR="$(mktemp -d)"
trap 'rm -rf "$FIXTURE_DIR"' EXIT

cd "$REPO_DIR"

go run ./runtime/cyonic-service fixture -dir "$FIXTURE_DIR"
go run ./runtime/cyonic-service evaluate \
  -request "$FIXTURE_DIR/request.json" \
  -trust-key "$FIXTURE_DIR/trusted-public.key" \
  -issuer example-principal \
  -origin external \
  -evidence-log "$FIXTURE_DIR/interactions.jsonl"

printf '\nInteraction receipt:\n'
cat "$FIXTURE_DIR/interactions.jsonl"
printf '\nExternality remains CLAIMED_EXTERNAL until independently adjudicated.\n'

