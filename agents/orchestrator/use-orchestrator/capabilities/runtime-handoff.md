# Runtime Handoff Capability

Agent-owned capability for invoking the closed deterministic workbook runtime.

Responsibilities:

- accept validated runtime execution input from the orchestrator
- invoke the closed runtime pipeline
- return telemetry, evidence, and verification artifacts

Boundaries:

- this is not a root-level discoverable skill
- request interpretation and routing stay in `agents/`
- execution authority stays in `contracts/` and `runtime/`
