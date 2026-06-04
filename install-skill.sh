#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GO_INSTALL_URL="https://go.dev/doc/install"

resolve_go_bin() {
  local go_bin="${GO_BIN:-go}"
  local args=("$@")
  local index=0
  while [[ $index -lt ${#args[@]} ]]; do
    case "${args[$index]}" in
      --go-bin)
        if [[ $((index + 1)) -lt ${#args[@]} ]]; then
          go_bin="${args[$((index + 1))]}"
        fi
        break
        ;;
      --go-bin=*)
        go_bin="${args[$index]#--go-bin=}"
        break
        ;;
    esac
    index=$((index + 1))
  done
  printf '%s\n' "$go_bin"
}

open_go_install_guide() {
  if command -v open >/dev/null 2>&1; then
    open "$GO_INSTALL_URL" >/dev/null 2>&1 || true
  fi
  printf 'install Go 1.25 or newer from %s, then rerun install-skill\n' "$GO_INSTALL_URL" >&2
}

confirm_open_guide() {
  printf '%s [y/N] ' "$1" >&2
  local answer=""
  if ! IFS= read -r answer; then
    return 1
  fi
  answer="${answer,,}"
  [[ "$answer" == "y" || "$answer" == "yes" ]]
}

ensure_go_ready() {
  local go_bin="$1"
  local version_output=""
  if ! version_output="$("$go_bin" version 2>/dev/null)"; then
    printf 'Go was not found. Sheet Ops requires Go 1.25 or newer.\n' >&2
    if confirm_open_guide 'Open the official Go installation guide before continuing?'; then
      open_go_install_guide
    fi
    exit 1
  fi

  if [[ "$version_output" =~ go([0-9]+)\.([0-9]+) ]]; then
    local major="${BASH_REMATCH[1]}"
    local minor="${BASH_REMATCH[2]}"
    if (( major < 1 || (major == 1 && minor < 25) )); then
      printf 'Go 1.25 or newer is required. Current version: go%s.%s.\n' "$major" "$minor" >&2
      if confirm_open_guide 'Open the official Go upgrade guide before continuing?'; then
        open_go_install_guide
      fi
      exit 1
    fi
    return 0
  fi

  printf 'Go was not found. Sheet Ops requires Go 1.25 or newer.\n' >&2
  if confirm_open_guide 'Open the official Go installation guide before continuing?'; then
    open_go_install_guide
  fi
  exit 1
}

go_bin="$(resolve_go_bin "$@")"
ensure_go_ready "$go_bin"

exec "$go_bin" -C "$SCRIPT_DIR" run ./cmd/sheet-ops-codex install-skill \
  "$@"
