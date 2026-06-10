# Sheet Ops CLI Agent Contract Final Review

Branch: `codex/cli-agent-contract`

Final status: pass after hardening follow-up.

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
  `capabilities`, `schema command prepare-use`, `schema command preflight`, and
  `preflight`.
- Public docs and embedded skill docs share the CLI agent-contract language.

## Hardening Evidence Matrix

| Area | Status | Authoritative Evidence | Success Claim |
| --- | --- | --- | --- |
| Installed contract smoke | pass | `TestInstallSkillBundledCLIExposesAgentContract` executes installed `.codex/skills/sheet-ops/bin/sheet-ops-codex` for `capabilities --json`, `schema command prepare-use --json`, `schema command preflight --json`, and `preflight --json`. | Installed CLI exposes the agent contract after materialized install. |
| Source CLI capabilities | pass | `go run ./cmd/sheet-ops-codex capabilities --json` returns `root_command.name:"sheet-ops-codex"`, `read_only_commands` without the root command, and `agent_contract` commands including `capabilities`, `preflight`, `prepare-use`, and `schema`. | Agents can discover safe command boundaries without parsing help prose. |
| Error taxonomy | pass | `capabilities --json` returns `error_contract.codes[]` with non-empty `status` and `producer`; emitted codes are `invalid_usage`, `missing_required_option`, `unknown_command`, and `internal_error`; runtime/future categories are `reserved`. | Agents can distinguish currently emitted errors from reserved taxonomy. |
| Prepare-use schema | pass | `go run ./cmd/sheet-ops-codex schema command prepare-use --json` returns `classification:"agent_contract"`, `mutating:true`, `read_only:false`, `dry_run_capable:false`, and `--envelope-file`. | The typed handoff surface is machine-readable. |
| Preflight | pass | `go run ./cmd/sheet-ops-codex preflight --json` returns `ok:true`, `read_only:true`, `project_directory:pass`, `go_runtime:pass`, `package_manifest:pass`, and skips optional file checks when no file args are provided. | Preflight is a read-only readiness diagnostic, not a fake dry-run. |
| Full test suite | pass | `go test ./... -count=1` passed. | No known repo test regression from the CLI contract hardening. |
| Real dry-run | deferred | `dry_run_capable:false` remains explicit for mutating workbook commands. | No dry-run success is claimed. |
| Full installed workbook execution smoke | deferred | No fast stable fixture selected. | No workbook execution success is claimed from meta-contract smoke alone. |

## Unsupported Assumptions

- Full workbook execution from an installed skill is not proven by this branch.
- Runtime-level reserved error categories are not proven emitted until typed
  producers and focused tests exist.
- Preflight proves readiness signals only; it does not prove a future workbook
  mutation would succeed.
- `dry_run_capable:false` is intentional until the runtime has a truthful
  planning mode.

## Allowed Strategy

- Add focused producer tests before moving any error code from `reserved` to
  `emitted`.
- Add a fast installed workbook fixture later if it can assert result JSON,
  output workbook existence, and evidence artifacts without excessive runtime
  cost.
- Keep `sheet-ops` as the single human-facing workbook request entry while
  expanding machine-readable `sheet-ops-codex` discovery.

## Rejected Strategy

- Do not call preflight a dry-run or infer mutation success from readiness.
- Do not treat `sheet-ops-agent` or bundled internal launchers as a public human
  workbook CLI.
- Do not report installed workbook execution as covered by `capabilities`,
  `schema`, or `preflight` smoke.
- Do not collapse reserved runtime errors into emitted errors without live
  producer evidence.

## Verification

- `go test ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts -count=1`: pass.
- `go run ./cmd/sheet-ops-codex --help`: pass.
- `go run ./cmd/sheet-ops-codex capabilities --json`: pass.
- `go run ./cmd/sheet-ops-codex schema command prepare-use --json`: pass.
- `go run ./cmd/sheet-ops-codex schema command preflight --json`: pass.
- `go run ./cmd/sheet-ops-codex preflight --json`: pass.
- `go test ./... -count=1`: pass.

## Backlog

- Full installed workbook execution smoke remains optional until a fast,
  stable fixture is selected.
- Error taxonomy includes `reserved` categories that should move to `emitted`
  only when typed producers and focused tests exist.
- Real dry-run remains deferred until the runtime has a truthful planning mode.
- Preflight checks output parent ancestor accessibility without creating probe
  files; deeper permission probing can be added later if required.
