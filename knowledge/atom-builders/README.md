# Atom Builders

This namespace records the first runtime-adjacent layer above existing Sheet Ops
operation execution. It is intentionally advisory unless an atom maps to an
existing supported capability record and runtime verifier.

- `workbook-app-primitives.json`: LLM-friendly workbook domain shapes.
- `builders.json`: supported capability mirrors and planned atom-builder
  contracts.
- `builder-plan-shapes.json`: reusable planning shapes for future atom builders.

Supported runtime status remains owned by `contracts/capabilities` and
`runtime`. Planned atom builders must stay backed by opportunity records until
separate implementation, executor, verifier, and fixture work promotes them.
