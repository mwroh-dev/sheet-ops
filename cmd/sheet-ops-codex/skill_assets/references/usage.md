# Usage

This `codex-skill` surface is the single human-facing public entry for Sheet Ops
workbook requests in the source-based local harness. The skill accepts natural
language from the user, creates or selects a `UseEnvelope`, then uses the
internal proof-gated execution boundary as a skill-owned runtime handoff. Do
not type or reconstruct internal launcher commands in the runner pane.

This file is installed skill documentation generated from the canonical public
package. The canonical package contract remains the authority in repository
architecture, and this copy is self-contained for installed/bundled consumers.

Do not present the internal dispatcher as a second public entry for humans. It is
the internal launcher used after the skill has captured the request boundary.

The skill install flow is source based. It does not use a zip file, tarball,
prebuilt platform binary, or Homebrew dependency. It builds the skill-local
agent binary during install.

## Install

Go 1.25 or newer is required. Check whether Go is already visible in the
current shell:

```bash
command -v go && go version
```

If Go is missing or older than 1.25, install or upgrade Go from the official Go
guide:

- https://go.dev/doc/install

Then run the repository-local install command. This single install command also
places the bundled agent runtime under `.codex/skills/sheet-ops/agent-system`:

```bash
/path/to/sheet-ops/install-skill.sh --project /path/to/target/workspace
```

If Go is installed but not visible in the current app or shell `PATH`, find the
actual Go executable first and pass it explicitly:

```bash
/path/to/sheet-ops/install-skill.sh --project /path/to/target/workspace --go-bin /absolute/path/to/go
```

Do not guess the Go path or install a second copy before checking the existing
installation.

If Go is already 1.25 or newer, `install-skill` does not prompt. If Go is
missing, it asks before directing the user to the official Go installation
guide. If Go is present but older than 1.25, it asks before directing the user
to upgrade.

## CLI agent contract

The `sheet-ops` skill is the single human-facing workbook request entry.
`sheet-ops-codex` is an install, diagnostic, and agent-contract CLI surface,
not a second human-facing workbook entry.

Use these read-only contract surfaces for agent/CI discovery:

- `sheet-ops-codex capabilities --json`
- `sheet-ops-codex schema command preflight --json`
- `sheet-ops-codex preflight --json`
- `sheet-ops-codex preview-request --json --intent-file <path> --input-file <path> --output-file <path>`

`sheet-ops-codex run-validated` is an internal handoff surface. It is owned by
the installed skill handoff and must not be presented as the normal public
request route. Mutating workbook commands currently report
`dry_run_capable: false`; `preview-request` is impact inspection only, not
runtime dry-run evidence. Do not claim dry-run behavior until a truthful runtime
planning mode exists.

`preview-request` also reports `planner:"requestcompiler_validate_intent"`,
`plan_confidence:"compiler_validated_boundary"`, `operation`, `would_mutate`,
`mutation_summary`, and input `fingerprints`. Treat these as trust metadata for
the inspected inputs, not proof of runtime execution success. After
`run-intent`, compare the execution result `fingerprints` to the preview
fingerprints, and use `output_workbook_sha256` to bind the produced workbook
bytes before trusting runtime artifacts.

## Public split surface

The public-side path is `sheet-ops-codex prepare-use`, which captures the user
request into a typed `UseEnvelopeV2` before any execution boundary is entered.

```bash
sheet-ops-codex prepare-use --scenario-id example-scenario --request-kind structured_use_request --request-ref /abs/path/request.json --input-file /abs/path/input.xlsx --output-file /abs/path/output.xlsx --envelope-file /abs/path/use-envelope.json
sheet-ops-codex prepare-use --scenario-id prompt-scenario --request-kind prompt_text --request-ref /abs/path/prompt.md --input-file /abs/path/input.xlsx --output-file /abs/path/output.xlsx --envelope-file /abs/path/use-envelope.json
sheet-ops-codex prepare-use --scenario-id organism-scenario --request-kind organism_execution_request --request-ref /abs/path/organism-request.json --input-file /abs/path/input.xlsx --output-file /abs/path/output.xlsx --envelope-file /abs/path/use-envelope.json
```

The compatibility boundary remains `UseEnvelope -> internal dispatcher`.
Runtime subroutes such as `use-open`, `use-structured`, explicit organism
execution, and `run-validated` stay behind that skill-owned handoff.

`prepare-use` writes the typed envelope that the compatibility boundary routes.
It does not expose `write_values` as a public agent capability.

## Diagnostics / regression helpers

This maintainer-only diagnostic surface is not the primary public entry. It
exists to bypass the `UseEnvelope -> internal dispatcher` public path for
case-local state while still keeping the split public surface and typed
envelope flow intact.

- It is a maintainer-only diagnostic surface, not a public capability escape hatch.
- The retained state lives under case-local state in `.sheet-ops-state`.
- Public entry runs reject `SHEET_OPS_STATE_ROOT` unless it equals the
  workbook case root's `.sheet-ops-state` path.
- Public installs default to `SHEET_OPS_RETENTION_MODE=redacted`.
- Public installs default to `SHEET_OPS_RENDER_MODE=never`.
- Full forensic retention overrides are for maintainer local testing mode only
  and are not deployment-safe yet.

When the open request-compiler lane is exercised through the public
`prepare-use -> internal dispatcher -> use-open` path, generated
request-compiler artifacts are stored under:

```text
<effective-state-root>/artifacts/work/<work-unit-id>/request-compiler/
```

If this diagnostic path returns `blocked`, stop and inspect the corresponding
request-compiler artifacts before attempting any follow-up run.

If a compatibility helper reports `needs_human_checkpoint`, keep the same rule:
stop and inspect the corresponding request-compiler artifacts before attempting
any follow-up run.

This is case-local state. Full forensic retention is opt-in and should be used
only for bounded maintainer diagnostics.
