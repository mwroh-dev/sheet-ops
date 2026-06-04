#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUNDLE_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SKILL_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TARGET="$BUNDLE_ROOT/bin/sheet-ops-agent"

resolve_go_bin() {
  if [[ -n "${GO_BIN:-}" ]]; then
    printf '%s\n' "$GO_BIN"
    return 0
  fi
  printf '%s\n' "go"
}

if [[ ! -f "$SKILL_ROOT/go.mod" || ! -d "$SKILL_ROOT/cmd/sheet-ops-agent" ]]; then
  printf 'installed skill root is missing canonical agent build inputs: %s\n' "$SKILL_ROOT" >&2
  exit 1
fi

go_bin="$(resolve_go_bin)"
mkdir -p "$(dirname "$TARGET")"
(
  cd "$SKILL_ROOT"
  "$go_bin" build -o "$TARGET" ./cmd/sheet-ops-agent
)
chmod 755 "$TARGET"
printf 'built sheet-ops-agent at %s\n' "$TARGET"
