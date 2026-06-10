---
name: sheet-ops
description: Single human-facing public entry for Sheet Ops workbook requests.
argument-hint: <workbook-task>
allowed-tools: Read Grep Glob Bash
---

# Sheet Ops

Use this installed skill as the single human-facing public entry for repeated
workbook work that should become a local process artifact, not only a one-off
chat answer. Accept the user's natural-language workbook request, separate the
task facts needed for execution, write or select a structured request reference,
and hand a typed `UseEnvelopeV2` to the internal proof-gated execution boundary.

This file is the installed skill documentation surface generated from the
canonical public package. The canonical package contract remains the authority
in repository architecture, and this copy is self-contained for
installed/bundled consumers.

The internal proof-gated execution boundary is a skill-owned runtime handoff,
not a runner-pane command.

Do not present that launcher as a second human-facing public entry. The user
calls this skill; this skill creates or selects the envelope and uses the
runtime handoff internally.

runner-pane bypass guard: Do not type, paste, or run the skill-local launcher path directly.
In a loop-station or consumer runner, the user-facing entry is this `sheet-ops`
skill; the launcher remains internal to skill execution.
Do not search for, inspect, or construct internal command lines in the runner
pane. Treat this skill document as the public entry instructions.

This skill is installed from the repository-local `install-skill.sh` command. It
does not ship a zip file, tarball, or prebuilt platform binary. It builds the
skill-local agent runtime during install. The runtime handoff remains internal
to this skill and to the installed wrapper.

## Why not just Codex/Claude?

Modern LLMs can already help with workbook tasks. Sheet Ops starts from that
assumption and focuses on what becomes useful after repetition: keeping the
process behind a workbook edit as local files that can be reviewed, tested,
adapted, and reused later.

In this skill, the LLM only plans and compiles. Go runtime executes and verifies
workbook work; schema contracts decide what is valid, and evidence records what happened for review and repair.

This skill is a guideline and orchestration harness for a strong model, not a
claim that the model lacks spreadsheet knowledge. Its value is to reduce
repeated reasoning cost, keep decisions tied to local contracts, and leave
evidence that can improve later similar workbook work.

## Public Preview Claims

Supported today:

- local source-install Sheet Ops skill harness entry for workbook requests
- public agent capabilities: `group_summarize`, `highlight_threshold`,
  `join_lookup`, `append_structured_rows`, `extend_table_formulas`,
  `copy_period_sheet`, `add_data_validation`, `protect_formula_cells`,
  `normalize_headers`, `roll_forward_period`, `reconcile_tables`, and
  `generate_printable_form`
- deterministic fixture-backed harness smoke with the public preview claim
  contract
- schema-authorized `TaskSpec -> OperationIR -> Execute -> Verify -> Evidence`
  runtime path
- advisory `template_class_plan` evidence for template-like workbook requests
  before narrowing to supported atom execution
- explicit decision guidance for goal, known facts, selected path, verifier
  focus, and stop condition before runtime handoff
- render artifact emission as preview artifact evidence

Preview limitations:

- live Codex delegation smoke is a non-blocking opt-in diagnostic, not a
  required public harness gate
- render artifact emission is not visual quality verification
- Sheet Ops is not hosted or multi-tenant ready
- `write_values` is not a public agent capability

Follow-up:

- harden live delegation smoke
- add visual quality checks after renderer support is reliable
- harden hosted state/privacy boundaries before hosted claims
- add public orchestration and e2e tests before promoting `write_values`

## Primary Workflow

1. Capture the user's workbook request in natural language.
2. Separate the task facts needed to identify the scenario, request text or request file, input workbook, and output workbook.
3. Record the goal, known facts, selected public atom capability or
   template-class hint, verifier focus, and stop condition before execution.
   For template-like requests, use the advisory atom/molecule/organism
   ecosystem as decision scaffolding, not runtime authority.
4. If the user already supplied the source sheet, output boundary, and operation shape clearly enough, write a structured `UseRequest` JSON request file and pack it with `--request-kind structured_use_request`. For explicit multi-step template-class organism execution, write an `OrganismExecutionRequest` JSON request file and pack it with `--request-kind organism_execution_request`. Do not pre-read the workbook through Python, `openpyxl`, ZIP/XML scraping, or other ad hoc inspection just to confirm headers.
5. If additional workbook facts are still needed, use a plain text or markdown request file and pack it with `--request-kind prompt_text` so the internal request-compiler inspects workbook facts in Go.
6. Derived request, envelope, output, report, and evidence files must be written outside immutable case input folders. In loop-station or any harness that provides an attempt output directory, write the request reference and typed `UseEnvelopeV2` under that attempt output directory. Do not create `use-envelope.json` inside the case input folder.
7. Hand the envelope to the internal compatibility dispatcher through the
   skill-owned runtime handoff. Do not type or reconstruct an internal command
   in the runner pane. The dispatcher routes typed structured request JSON to
   `use-structured`, explicit organism execution requests to the organism
   execution bridge, and text or markdown requests to `use-open`. Closed
   validated execution requests are for `run-validated` only.
8. Read the JSON result on stdout.
9. Inspect the evidence path and output workbook only when the result succeeds.
10. Return the final user-facing answer from this parent skill context.

## Tool Authority

Consumer restrictions live in `agents/profiles/*.toml`. This public skill
surface does not widen runner-pane tool permissions; profile-rendered consumer
surfaces remain the code-reviewed source of truth for allowed tools, commands,
carriers, and writable scope.

## Routing Guardrail

Do not route a live Sheet Ops request through consumer-local demo or
compatibility surfaces such as `.codex/skills/example-*` or
`.codex/skills/shared-workbook-case-executor`. Those are harness fixtures, not
the single human-facing public entry, and they may exercise legacy adapters or
non-authoritative execution paths.

## Installation Contract

The user environment must have Go 1.25 or newer before installing this skill.

Official install flow:

1. Check whether Go is already visible in the current shell:
   `command -v go && go version`
2. If Go is missing or older than 1.25, install or upgrade Go from the official
   Go guide: https://go.dev/doc/install
3. Run the repository-local installer:

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

If Go is missing or older than 1.25, `install-skill` asks before sending the user to the official Go installation guide. If Go 1.25 or newer is already available, it does not prompt and installs without asking.

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

## Diagnostic Workflow

Public split workflow:

1. Write a typed `UseEnvelopeV2` JSON file with the deterministic packing step under the attempt output directory or another writable derived-output location. Do not create `use-envelope.json` inside the case input folder.
2. Use the internal compatibility launcher through the skill-owned runtime
   handoff; do not type or reconstruct its command line.
3. Let the launcher dispatch the request:
   - `request.kind=prompt_text` goes through `use-open`
   - `request.kind=structured_use_request` goes through `use-structured`
   - `request.kind=organism_execution_request` goes through explicit organism execution
   - closed validated execution requests belong to `run-validated`
4. Read the JSON result on stdout.
5. If the result status is `failed` and `terminal_state` is `BLOCKED_AT_REQUEST_COMPILER`, treat it as a blocked compiler outcome, stop, and inspect the request-compiler artifacts under `.sheet-ops-state/artifacts/work/<work-unit-id>/request-compiler/`. Those artifacts may contain compiler decision details such as `needs_human_checkpoint`.
6. Inspect the output workbook only when the result status is `succeeded`.

External normalized-intent flow remains an advanced maintainer diagnostic
surface. It is not the normal human-facing Sheet Ops route and must not be used
by runner panes as a substitute for invoking this public skill.

Legacy public compiler entries are not part of the Task 4 split. Do not route workbook requests through direct prompt/text compilation from the public codex CLI.

## Runtime State

The wrapper script resolves `SHEET_OPS_STATE_ROOT` as `$PWD/.sheet-ops-state` only when that environment variable is unset, then derives:

- artifacts: `.sheet-ops-state/artifacts`
- knowledge: `.sheet-ops-state/knowledge`

For public open-layer requests, there is always one effective state root. The runtime workspace is derived from the input workbook directory:

- workspace: `<input-workbook-dir>`
- default state root when `SHEET_OPS_STATE_ROOT` is unset: `<workspace>/.sheet-ops-state`
- only accepted explicit state root: `<workspace>/.sheet-ops-state`
- artifacts root: `<effective-state-root>/artifacts`
- knowledge root: `<effective-state-root>/knowledge`

If neither supported state-root condition is true, the command is rejected.

Installed public runs default to:

- `SHEET_OPS_RETENTION_MODE=redacted`
- `SHEET_OPS_RENDER_MODE=never`

Use explicit environment overrides only when a maintainer intentionally needs
full forensic retention or render evidence. That override path is for local testing mode only and is not deployment-safe yet.
