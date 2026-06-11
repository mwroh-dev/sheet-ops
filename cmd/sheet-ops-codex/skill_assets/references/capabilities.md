# Capabilities

Capability records live under `contracts/capabilities/records`. This reference is guidance over that registry, not a second authority.

Operating model:

`LLM is planner/compiler. Go is executor/verifier. Schema is authority.`

The skill routes execution through the skill-owned runtime handoff, not through
a runner-pane command.

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
request route. Its stdout is an internal handoff envelope with runtime evidence
under `runtime`, not a public entry result. Validate it with
`contracts/cli/internal_handoff_result.schema.json`, not the public entry
schema. Mutating workbook commands currently report `dry_run_capable: false`;
`preview-request` is impact inspection only, not runtime dry-run evidence. Do
not claim dry-run behavior until a truthful runtime planning mode exists.

`preview-request` includes `planner:"requestcompiler_validate_intent"` and
`plan_confidence:"compiler_validated_boundary"`. Treat its `operation`, `would_mutate`,
`mutation_summary`, and `fingerprints` fields as planning/trust metadata only;
they do not prove that runtime execution succeeded. After `run-intent`, compare
the execution result `fingerprints` to the preview fingerprints, and use
`output_workbook_sha256` when present to bind the produced workbook bytes before
trusting runtime artifacts. Cross-check it against
`runtime.verification.output_workbook_sha256`; verification artifacts that name
a non-empty `output_file` require that hash. Treat output workbooks on
`ok:false` results as failure evidence, not primary success.

## Public agent capabilities

- `group_summarize`: use when the request asks to summarize rows by keys, such as "summarize revenue by region".
- `highlight_threshold`: use when the request asks to highlight rows or cells over, under, or equal to a threshold.
- `join_lookup`: use when the request asks to merge or append lookup-sheet data by key.
- `append_structured_rows`: use when the request asks to append complete new rows into an existing table while preserving the source workbook and table headers.
- `extend_table_formulas`: use when the request asks to copy existing row formulas into specified target rows while preserving relative row references.
- `copy_period_sheet`: use when the request asks to copy an existing period sheet into a new target sheet while preserving values, formulas, and styles.
- `add_data_validation`: use when the request asks to add explicit list/dropdown validation rules to declared ranges while preserving the source workbook.
- `protect_formula_cells`: use when the request asks to lock explicit formula ranges, leave declared input ranges editable, and protect the sheet while preserving the source workbook.
- `normalize_headers`: use when the request provides explicit source-to-target header mappings for a declared header row.
- `roll_forward_period`: use when the request provides explicit closing-to-opening cell mappings between existing period sheets.
- `reconcile_tables`: use when the request provides explicit source/lookup sheets, left/right keys, and compare mappings to create a reconciliation result sheet.
- `generate_printable_form`: use when the request provides explicit field bindings, table binding, target sheet, form title, and print area for a fixed printable workbook region.

These records have `status: supported` and `exposure: public_agent_capability`.
Before selecting a capability, read the matching machine-readable record in
`contracts/capabilities/records`.

## Internal runtime primitives

- `write_values`: `status: supported`, `exposure: runtime_primitive`. This is
  a deterministic runtime primitive for tests and lower-level execution paths.
  It is not selected directly by the public agent flow.

## Agent decision table

| Signal | Classification | Agent action |
| --- | --- | --- |
| User asks to summarize, group, aggregate, or create a summary sheet | Supported public capability | Select `group_summarize` after reading the registry record. |
| User asks to highlight values over/under/equal to a threshold | Supported public capability | Select `highlight_threshold` after reading the registry record. |
| User asks to merge lookup data by key | Supported public capability | Select `join_lookup` after reading the registry record. |
| User asks to add complete new table rows with provided field values | Supported public capability | Select `append_structured_rows` after reading the registry record. |
| User asks to extend row formulas into specific target rows | Supported public capability | Select `extend_table_formulas` after reading the registry record. |
| User asks to copy an existing period sheet to a named new sheet | Supported public capability | Select `copy_period_sheet` after reading the registry record. |
| User asks to add dropdown/list validation to specific ranges with allowed values | Supported public capability | Select `add_data_validation` after reading the registry record. |
| User asks to protect formula cells and keep input cells editable with explicit ranges | Supported public capability | Select `protect_formula_cells` after reading the registry record. |
| User asks to rename known headers with explicit from/to mappings on a declared header row | Supported public capability | Select `normalize_headers` after reading the registry record. |
| User asks to carry a closing value into a next-period opening cell with explicit mappings | Supported public capability | Select `roll_forward_period` after reading the registry record. |
| User asks to reconcile two tables by explicit keys and compare fields | Supported public capability | Select `reconcile_tables` after reading the registry record. |
| User asks to create a printable/reviewable form from explicit field and table bindings | Supported public capability | Select `generate_printable_form` after reading the registry record. |
| Runtime needs to write literal cells inside a verified lower-level path | Internal primitive | Use `write_values` only through runtime-owned flows, not as a public request capability. |
| Deterministic smoke uses fixture-backed specialist decisions | Deterministic smoke | Treat as the required public harness gate, not live LLM delegation. |
| Live Codex is available and the caller opts in | Live smoke | Run only as a non-blocking diagnostic. |
| Fuzzy header inference, duplicate header resolution, fuzzy reconciliation, numeric tolerance policies, inferred printable layouts, visual quality assertions, PDF pagination, full period generation, label rewriting, input clearing, hosted deployment, or public `write_values` is requested | Preview limitation | Do not claim support; report the limitation or require additional hardening/tests. |

Install contract reminder:

```bash
command -v go && go version
/path/to/sheet-ops/install-skill.sh --project /path/to/target/workspace
```

Go 1.25 or newer is required. If Go is installed but not visible in the current
app or shell `PATH`, find the actual executable first and pass
`--go-bin /absolute/path/to/go`. Do not guess the Go path or install a second
copy before checking the existing installation. If Go is ready, install-skill
does not prompt. If Go is missing or outdated, it asks before continuing. The
install flow does not use a zip file, tarball, or prebuilt platform binary.
