# Sheet Ops CLI Agent Contract Final Review

Branch: `codex/cli-agent-contract`

Final status: pass.

## Implemented Contract

- `sheet-ops` remains the single human-facing workbook request entry.
- `sheet-ops-codex` now exposes classified install, agent-contract, internal
  handoff, maintainer diagnostic, compatibility, and support surfaces.
- `capabilities --json` exposes machine-readable command groups, read-only
  boundaries, mutating flags, hidden flags, and error taxonomy.
- `schema --json` and `schema command <name> --json` expose command metadata,
  artifacts, state behavior, safety notes, related commands, mutation flags,
  read-only flags, and `dry_run_capable`.
- `preflight --json` is a read-only diagnostic surface for project, Go,
  embedded manifest, input readability, output parent ancestry, and state-root
  checks.
- Installed `.codex/skills/sheet-ops/bin/sheet-ops-codex` is smoke-tested for
  `capabilities`, `schema`, and `preflight`.
- Public docs and embedded skill docs share the CLI agent-contract language.

## Verification

- `go test ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts -count=1`: pass.
- `go run ./cmd/sheet-ops-codex --help`: pass.
- `go run ./cmd/sheet-ops-codex capabilities --json`: pass.
- `go run ./cmd/sheet-ops-codex schema command preflight --json`: pass.
- `go run ./cmd/sheet-ops-codex preflight --json`: pass.
- `go test ./... -count=1`: pass.

## Backlog

- Full installed workbook execution smoke remains optional until a fast,
  stable fixture is selected.
- Error taxonomy includes reserved categories that should be wired to typed
  producers as runtime command error handling is hardened.
- Real dry-run remains deferred until the runtime has a truthful planning mode.
- Preflight checks output parent ancestor accessibility without creating probe
  files; deeper permission probing can be added later if required.
