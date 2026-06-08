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
`highlight_threshold`, `join_lookup`, `append_structured_rows`,
`extend_table_formulas`, `copy_period_sheet`, `add_data_validation`,
`protect_formula_cells`, `normalize_headers`, and `write_values`. The supported `add_data_validation` scope is explicit
list/dropdown rules on declared ranges; richer validation forms remain
opportunity scope. The supported `protect_formula_cells` scope is explicit
formula/input ranges with sheet protection enabled; automatic formula discovery
and collaborative permission policies remain opportunity scope. Atom builder
records do not add new runtime support by themselves. The supported
`copy_period_sheet` scope copies an explicit source sheet to a target sheet;
label updates, input clearing, and carry-forward continuity remain
`roll_forward_period` opportunity scope. The supported `normalize_headers`
scope rewrites only explicitly mapped header-row cells; alias inference,
duplicate resolution, and multi-row header detection remain opportunity scope.

## Promotion Rule

An atom moves from planned to supported only when all of these exist:

- capability registry record
- schema-authorized task or IR fields
- deterministic executor
- operation-specific verifier
- fixture-backed harness coverage
- public claim review, when exposed as a public agent capability
