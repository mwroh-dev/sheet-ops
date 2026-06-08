# Advisory To Runtime Bridge

The template research ecosystem guides runtime expansion but does not execute
organisms directly.

## Bridge Flow

1. A user request is interpreted against workbook app primitives.
2. Organism and molecule records suggest likely atom builders.
3. The bridge splits atoms into supported, runtime primitive, and planned.
4. Only supported or runtime primitive atoms may enter deterministic execution.
5. Planned atoms remain backlog signals until capability contracts, runtime
   implementation, and operation-specific verifiers exist.

## Current Supported Path

The current runtime path remains:

```text
UseRequest / ValidatedExecutionRequest
-> TaskSpec
-> WorkbookOperationIR
-> operation plan where available
-> deterministic executor
-> operation-specific verifier
-> evidence artifacts
```

Atom builder records mirror this path for `group_summarize`,
`highlight_threshold`, `join_lookup`, `append_structured_rows`, and
`write_values`. They do not add new runtime support by themselves.

## Promotion Rule

An atom moves from planned to supported only when all of these exist:

- capability registry record
- schema-authorized task or IR fields
- deterministic executor
- operation-specific verifier
- fixture-backed harness coverage
- public claim review, when exposed as a public agent capability
