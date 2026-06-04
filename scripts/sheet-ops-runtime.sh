#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

export SHEET_OPS_PACKAGE_ROOT="${SHEET_OPS_PACKAGE_ROOT:-$REPO_ROOT}"

SHEET_OPS_CLI_NAME=sheet-ops-runtime exec "${GO_BIN:-go}" -C "$REPO_ROOT" run ./cmd/sheet-ops-codex "$@"
