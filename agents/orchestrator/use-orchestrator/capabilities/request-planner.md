# Request Planner Capability

Agent-owned capability for lowering normalized request fields into
harness-facing request fields.

Responsibilities:

- accept normalized semantic request fields
- produce orchestrator-facing request fields and assumptions
- keep execution-shape lowering deterministic

Boundaries:

- this is not a root-level discoverable skill
- routing and checkpoint authority stay with `agents/orchestrator/use-orchestrator/`
- executable workbook authority stays in `contracts/` and `runtime/`
