#!/usr/bin/env bash
set -euo pipefail

IDENTITY="${1:-friend-node-01}"
PATHWAY="${2:-codex}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CANON_DIR="$ROOT/canon/logos-formal"
REPOSITORY="@@SOURCE_REPOSITORY@@"
SOURCE_BRANCH="@@SOURCE_BRANCH@@"
SOURCE_COMMIT="@@SOURCE_COMMIT@@"
SOURCE_TREE="@@SOURCE_TREE@@"

[[ "$IDENTITY" =~ ^[A-Za-z0-9._-]{1,64}$ ]] || {
  echo "Invalid node identity." >&2
  exit 2
}
case "$PATHWAY" in
  codex|antigravity|manual) ;;
  *) echo "Pathway must be codex, antigravity, or manual." >&2; exit 2 ;;
esac

command -v git >/dev/null || { echo "Git is required." >&2; exit 2; }
command -v go >/dev/null || { echo "Go 1.24 or newer is required." >&2; exit 2; }

GO_VERSION="$(go env GOVERSION)"
if [[ ! "$GO_VERSION" =~ ^go([0-9]+)\.([0-9]+) ]]; then
  echo "Could not parse Go version: $GO_VERSION" >&2
  exit 2
fi
GO_MAJOR="${BASH_REMATCH[1]}"
GO_MINOR="${BASH_REMATCH[2]}"
if (( GO_MAJOR < 1 || (GO_MAJOR == 1 && GO_MINOR < 24) )); then
  echo "Go 1.24 or newer is required. Found: $GO_VERSION" >&2
  exit 2
fi

if [[ -d "$CANON_DIR/.git" ]]; then
  [[ -z "$(git -C "$CANON_DIR" status --porcelain)" ]] || {
    echo "Existing canon checkout has local changes; refusing to discard them." >&2
    exit 2
  }
elif [[ -e "$CANON_DIR" ]]; then
  echo "Canon path exists but is not a Git checkout: $CANON_DIR" >&2
  exit 2
else
  git clone --filter=blob:none --no-checkout "$REPOSITORY" "$CANON_DIR"
fi

git -C "$CANON_DIR" fetch origin "$SOURCE_BRANCH"
git -C "$CANON_DIR" switch --detach "$SOURCE_COMMIT"
[[ "$(git -C "$CANON_DIR" rev-parse HEAD)" == "$SOURCE_COMMIT" ]]
[[ "$(git -C "$CANON_DIR" rev-parse 'HEAD^{tree}')" == "$SOURCE_TREE" ]] || {
  echo "Canon checkout does not match the pinned source tree." >&2
  exit 2
}

if [[ ! -f "$ROOT/config/node.yaml" ]]; then
  sed -e "s/NODE_ID/$IDENTITY/g" -e "s/PATHWAY/$PATHWAY/g" \
    "$ROOT/config/node.example.yaml" > "$ROOT/config/node.yaml"
else
  EXISTING_IDENTITY="$(sed -n 's/^node_id: //p' "$ROOT/config/node.yaml")"
  EXISTING_PATHWAY="$(sed -n 's/^preferred_pathway: //p' "$ROOT/config/node.yaml")"
  [[ "$EXISTING_IDENTITY" == "$IDENTITY" && "$EXISTING_PATHWAY" == "$PATHWAY" ]] || {
    echo "Existing node configuration does not match the requested identity/pathway. Refusing to report unapplied values." >&2
    exit 2
  }
fi

"$ROOT/scripts/validate.sh"

printf '\nNODE_SETUP_COMPLETE\n'
printf 'Node: %s\nPathway: %s\nPinned commit: %s\nPinned tree: %s\n' \
  "$IDENTITY" "$PATHWAY" "$SOURCE_COMMIT" "$SOURCE_TREE"
printf 'Local sessions remain 010 with authority_effect NONE.\n'
