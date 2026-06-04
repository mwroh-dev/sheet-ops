# Usage

This runtime is bundled inside the installed Sheet Ops Codex skill:

- `./.codex/skills/sheet-ops/agent-system`

The installed executable is intentionally not a human-facing runner command.

The internal proof-gated execution boundary is identified by capability. Do not
type, paste, or reconstruct the installed executable path in the runner pane.

Humans should reach this boundary through the `sheet-ops` Codex skill, which is
the single human-facing public entry for workbook requests. This bundled
runtime enforces the internal launcher contract after the skill has captured a
`UseEnvelope`.
