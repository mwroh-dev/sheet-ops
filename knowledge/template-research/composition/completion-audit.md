# Template Research Runtime Coverage Completion Audit

This audit checks the current state against the advisory-to-runtime coverage
goal. It does not claim full workbook generation or broad Excel automation.

## Completed Evidence

- Research/advisory layers exist for raw observations, hypotheses, judgments,
  classification, molecule composition, organism composition, atom-builder
  planning, runtime bridge guidance, and executable coverage.
- Supported capability status remains owned by
  `contracts/capabilities/capability.schema.json`, runtime contracts,
  deterministic executors, and operation-specific verifiers.
- `create_pivot_summary` remains outside the supported capability registry.
  P2 organisms use deterministic `group_summarize` previews instead of
  claiming native pivot support.
- `executable-organism-coverage.json` covers all 21 round-009 roadmap
  organisms exactly once.
- All 21 roadmap organisms remain preview-backed by workbookcase evidence.
- P2 and P3 depth contracts are recorded at `organism_verified_class` tier with
  explicit acceptance criteria, non-claims, and promotion gates.
- Runtime productization contracts now cover organism-level verifier specs,
  operation planner sequences, template class harness scenarios, and advanced
  atom promotion decisions.
- `runtime/templateclass` provides a deterministic classifier/planner/evidence
  evaluator for the first productization slice.
- `runtime/workbookcase.RunOrganismPlan` provides the first multi-operation
  organism harness slice for invoice line item billing.
- The knowledge harness validates roadmap coverage, fixture priority alignment,
  advisory authority, supported atom references, absence of planned blockers,
  preview evidence presence, verified-class depth contract references, and
  runtime productization artifact boundaries.
- Workbookcase preview fixtures cover p0, p1, p2, and p3 organism priorities.
- Each phase records a self-retro in `executable-coverage-matrix.md`.

## Verified Commands

- `jq empty` over checked-in JSON artifacts under `contracts`, `agents`,
  `knowledge`, and `examples`
- `go test ./runtime/workbookcase ./runtime/knowledge`
- `go test ./...`
- `git diff --check`

## Remaining Limits

- Preview fixtures are representative, not full template generators.
- Organism-level verifiers do not exist yet; previews rely on
  operation-specific verifiers.
- Operation planner and template class harness records are advisory contracts,
  not autonomous runtime planner execution. The `runtime/templateclass` package
  plans and evaluates evidence, but it does not execute workbook operations.
- The organism harness currently covers the invoice class only. Other organism
  classes still need equivalent multi-operation fixtures before broad
  organism-runtime claims are appropriate.
- Native Excel pivot artifacts are not supported.
- Matrix auto-growth, rich visual layout QA, full printable pagination QA, and
  domain-specific schedule semantics remain future runtime work.
- `loan_repayment_calculator` proves formula extension and protected
  calculation mechanics, not amortization correctness or financial advice.

## Completion Judgment

The advisory ecosystem, runtime bridge, coverage ladder, representative preview
fixture coverage, p2/p3 verified-class depth contracts, and runtime
productization contracts are complete for the current 21 roadmap organisms.
The first deterministic template-class planner/evaluator slice and invoice
multi-operation organism harness are also present. The remaining limits are
explicitly documented as future runtime implementation depth, not untracked
blockers in the requested advisory-to-productization coverage layer.
