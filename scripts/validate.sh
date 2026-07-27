#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
repo_root="$(cd -- "$script_dir/.." && pwd -P)"
temp_dir="$(mktemp -d)"

cleanup() {
  rm -rf -- "$temp_dir"
}
trap cleanup EXIT

binary="$temp_dir/cyonic-validate"
(
  cd -- "$repo_root"
  GOFLAGS= go build -o "$binary" ./cmd/cyonic-validate
)

set +e
(
  cd -- "$repo_root"
  "$binary" "$@"
)
status=$?
set -e
exit "$status"
