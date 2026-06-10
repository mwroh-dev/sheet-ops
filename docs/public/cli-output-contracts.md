# CLI Output Contracts

Sheet Ops publishes machine-readable CLI output contracts for agent and
automation callers. JSON output contracts use JSON Schema 2020-12 and keep the
root `$schema` and `$id` fields explicit.

## Published Schemas

- `contracts/cli/capabilities.schema.json` validates `sheet-ops-codex capabilities --json`.
- `contracts/cli/command_schema.schema.json` validates `sheet-ops-codex schema --json` and `sheet-ops-codex schema command <name> --json`.
- `contracts/cli/error_envelope.schema.json` validates JSON error envelopes.
- `contracts/cli/preflight_result.schema.json` validates `sheet-ops-codex preflight --json`.
- `contracts/cli/preview_request_result.schema.json` validates `sheet-ops-codex preview-request --json`.
- `contracts/results/public_entry_result.schema.json` validates public execution results from `run-prompt`, `run-text`, and `run-intent`.

## Preview Contract

`preview-request` is read-only impact inspection, not runtime dry-run. Agents
must treat `dry_run:false` and `plan_confidence:"compiler_validated_boundary"` as the
authoritative limit of the claim.

The preview result includes:

- `planner:"requestcompiler_validate_intent"`: the command ran non-persisting
  request-compiler validation over the normalized intent and workbook boundary,
  but did not run runtime execution or persist compiler artifacts.
- `operation`: the operation selected by request-compiler validation,
  `unresolved` for checkpoint/blocked decisions, or `unknown` when no operation
  is selected.
- `would_mutate`: true only when request-compiler validation selected a
  compiled operation. It is false for unresolved/checkpoint/blocked previews.
- `mutation_summary`: machine-readable categories for planned workbook,
  artifact, and state-root mutation; unresolved previews use `none`.
- `planned_writes` and `planned_artifacts`: populated only for compiled preview
  decisions. Unresolved/checkpoint/blocked previews use empty arrays.
- `fingerprints`: SHA-256 hashes for the normalized intent file and input
  workbook read by preview.

Executed public results include the same input fingerprint fields. When a run
materializes an output workbook, they also include `output_workbook_sha256`.
Agents can compare `preview-request.fingerprints` with the later
`run-intent.fingerprints` to prove the run used the same normalized intent and
input workbook, then bind any reported output workbook to a concrete file hash
before trusting runtime artifacts. The same output hash is also carried in
`runtime.verification.output_workbook_sha256` and the persisted verification
artifact. Any verification artifact that names a non-empty `output_file`
requires this hash. Treat `output_workbook` as the primary success artifact
only when the public result has `ok:true`; on `ok:false`, a materialized output
workbook is failure evidence.

## Versioning Policy

The current schema version is `sheet-ops-cli/v1`.

Breaking output changes require a new `schema_version`. Breaking changes include
renaming a field, removing a field, changing a field type, changing a command's
success/failure meaning, or weakening a required success artifact.

Additive optional fields do not require a new `schema_version` when existing
fields keep their meaning. Producers should update the matching JSON Schema in
the same change that adds the field.

Agent consumers should accept unknown fields and rely on the documented required
fields for control flow. For public execution results, agents should decide
success from `ok`, `status`, `recoverable`, `artifacts`, and
`runtime.verification`, not from incidental stdout text.
