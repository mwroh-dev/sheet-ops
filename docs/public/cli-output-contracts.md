# CLI Output Contracts

Sheet Ops publishes machine-readable CLI output contracts for agent and
automation callers. JSON output contracts use JSON Schema 2020-12 and keep the
root `$schema` and `$id` fields explicit.

## Published Schemas

- `contracts/cli/capabilities.schema.json` validates `sheet-ops-codex capabilities --json`.
- `contracts/cli/command_schema.schema.json` validates `sheet-ops-codex schema --json` and `sheet-ops-codex schema command <name> --json`.
- `contracts/cli/error_envelope.schema.json` validates JSON error envelopes.
- `contracts/cli/preflight_result.schema.json` validates `sheet-ops-codex preflight --json`.
- `contracts/results/public_entry_result.schema.json` validates public execution results from `run-prompt`, `run-text`, and `run-intent`.

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
