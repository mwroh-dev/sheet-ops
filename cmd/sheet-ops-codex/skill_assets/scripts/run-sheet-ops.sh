#!/usr/bin/env bash
set -euo pipefail

PACKAGE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_PATH="__SHEET_OPS_CODEX_BIN__"
STATE_ROOT="${SHEET_OPS_STATE_ROOT:-$PWD/.sheet-ops-state}"

export SHEET_OPS_PACKAGE_ROOT="${SHEET_OPS_PACKAGE_ROOT:-$PACKAGE_ROOT}"
export SHEET_OPS_ARTIFACT_ROOT="${SHEET_OPS_ARTIFACT_ROOT:-$STATE_ROOT/artifacts}"
export SHEET_OPS_KNOWLEDGE_ROOT="${SHEET_OPS_KNOWLEDGE_ROOT:-$STATE_ROOT/knowledge}"
export SHEET_OPS_RETENTION_MODE="${SHEET_OPS_RETENTION_MODE:-redacted}"
export SHEET_OPS_RENDER_MODE="${SHEET_OPS_RENDER_MODE:-never}"
export SHEET_OPS_CLI_NAME="${SHEET_OPS_CLI_NAME:-sheet-ops-runtime}"

if [[ ! -x "$BIN_PATH" ]]; then
  printf 'sheet-ops repository entry missing or not executable: %s\n' "$BIN_PATH" >&2
  printf 'reinstall with: /path/to/sheet-ops/install-skill.sh --project /path/to/target/workspace\n' >&2
  exit 1
fi

exec "$BIN_PATH" "$@"
