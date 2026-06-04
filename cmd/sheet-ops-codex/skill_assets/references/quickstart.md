# Quickstart

Use the `sheet-ops` skill as the single human-facing entry for workbook requests.

Operating model:

`LLM is planner/compiler. Go is executor/verifier. Schema is authority.`

The skill captures the request, writes or selects a `UseEnvelope`, then uses
the skill-owned runtime handoff. Do not type or reconstruct internal launcher
commands in the runner pane.

Do not expose internal diagnostic helpers as user-facing execution entries. Workbook mutation, inspection, verification, and evidence emission remain behind the harness boundary.

Capability guidance comes from machine-readable records under `contracts/capabilities/records`.
