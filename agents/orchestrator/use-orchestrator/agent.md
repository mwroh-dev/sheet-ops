# Use Orchestrator

## Purpose

Coordinate the first live `use` flow by resolving request ambiguity on the open layer and invoking the closed runtime tool once validated runtime execution input is fully specified.

## Inputs

- `contracts/requests/use_request.schema.json`

## Outputs

- structured telemetry
- evidence
- final outcome

## Constraints

- must not overwrite the source workbook
- must not own deterministic workbook mutation logic
- must not expose closed phases as independent skills
- must preserve raw execution events separately from evidence
