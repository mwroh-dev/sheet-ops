# Preview Status

The claim source of truth is `contracts/public_preview/claims.json`.

## Supported today

- experimental local skill harness for Codex/Claude workbook agents
- single human-facing `sheet-ops` skill entry
- currently implemented public path compositions:
  `append_structured_rows`, `extend_table_formulas`, `copy_period_sheet`,
  `add_data_validation`, `protect_formula_cells`, `normalize_headers`,
  `roll_forward_period`, `reconcile_tables`, `generate_printable_form`,
  `group_summarize`, `highlight_threshold`, and `join_lookup`
- schema-authorized `TaskSpec -> OperationIR -> Execute -> Verify -> Evidence`
  runtime path
- advisory template research corpus with atom/molecule/organism composition
  records and 21 roadmap organism coverage artifacts
- compiler decision `template_class_plan` evidence for template-like workbook
  requests before narrowing to supported atom execution
- explicit planning guidance for goal, known facts, selected path, verifier
  focus, and stop condition before runtime handoff
- explicit organism execution request path with draft planner coverage and
  final-workbook semantic verifier slices for 21 roadmap organisms
- deterministic fixture-backed harness smoke with checked-in frozen demo
  evidence
- render artifact emission when renderer tools exist, or explicit unavailable
  evidence when they do not

## Preview limitations

- live Codex delegation smoke is a non-blocking opt-in diagnostic, not a
  required public harness gate
- render artifact emission records file-open and preview artifact properties;
  it is not visual quality verification
- template research artifacts are advisory; they are not execution authority
  and do not imply full template generation
- Sheet Ops guides and records model orchestration; it does not claim that
  arbitrary Excel templates can be fully inferred or generated
- organism-level verifier slices prove bounded workbook evidence, not
  domain-specific correctness, financial advice, safety certification, or rich
  visual/print QA
- Sheet Ops is not hosted or multi-tenant ready
- `write_values` is not a public agent capability; it remains an internal
  runtime primitive until public orchestration and e2e tests exist

## Follow-up

- harden the optional live smoke with stronger timeout/process supervision
- add cross-platform visual quality assertions only after renderer support is
  proven
- close the assist orchestration backlog for decision artifacts, live guidance
  compliance, failure-to-knowledge promotion, and evidence-centered final
  answers
- add richer template generation, native pivot/matrix expansion, and
  domain-specific correctness layers after the current advisory/runtime bridge
- harden state/privacy boundaries before any hosted claim
- promote `write_values` only after public orchestration and e2e coverage exist

See [Assist Orchestration Backlog](orchestration-backlog.md) for the detailed
skill-harness gaps.
