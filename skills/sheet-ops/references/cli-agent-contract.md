# CLI Agent Contract

The `sheet-ops` skill is the single human-facing workbook request entry.
`sheet-ops-codex` is an install, diagnostic, and agent-contract CLI surface,
not a second human-facing workbook entry.

Use these read-only contract surfaces for agent/CI discovery:

- `sheet-ops-codex agent-guide --json`
- `sheet-ops-codex capabilities --json`
- `sheet-ops-codex operation list --json`
- `sheet-ops-codex operation schema <operation> --json`
- `sheet-ops-codex operation example <operation> --json`
- `sheet-ops-codex schema command preflight --json`
- `sheet-ops-codex preflight --json`
- `sheet-ops-codex preview-request --json --intent-file <path-or-stdin> --input-file <path> --output-file <path>`
- `sheet-ops-codex evidence-summary --json --evidence-dir <path>`

`sheet-ops-codex run-validated` is an internal handoff surface. It is owned by
the installed skill handoff and must not be presented as the normal public
request route. Its stdout is an internal handoff envelope with runtime evidence
under `runtime`, not a public entry result. Validate it with
`contracts/cli/internal_handoff_result.schema.json`, not the public entry
schema `contracts/results/public_entry_result.schema.json`. Mutating workbook
commands currently report `dry_run_capable: false`; `preview-request` is impact
inspection only, not runtime dry-run evidence. Do not claim dry-run behavior
until a truthful runtime planning mode exists.

For large or generated normalized-intent payloads, prefer stdin over shell
escaping: `sheet-ops-codex preview-request --json --intent-file - ...` and
`sheet-ops-codex run-intent --intent-file - ...` read the intent JSON from
stdin. Treat the `stdin` planned-read path as an opaque source label and rely on
`fingerprints.normalized_intent_sha256` for identity.

`preview-request` also reports `planner:"requestcompiler_validate_intent"`,
`plan_confidence:"compiler_validated_boundary"`, `operation`, `would_mutate`,
`mutation_summary`, and input `fingerprints`. Treat these as trust metadata for
the inspected inputs, not proof of runtime execution success. After
`run-intent`, compare the execution result `fingerprints` to the preview
fingerprints, and when present use `output_workbook_sha256` to bind the
produced workbook bytes before trusting runtime artifacts. Cross-check it
against `runtime.verification.output_workbook_sha256`; verification artifacts
that name a non-empty `output_file` require that hash. Treat output workbooks
on `ok:false` results as failure evidence, not primary success.

After any mutating run, prefer
`sheet-ops-codex evidence-summary --json --evidence-dir <runtime evidence dir>`
before final reporting. Only report success when it returns
`status:"verified_success"`; otherwise follow `next_actions` and inspect
failure evidence or repair advice.
