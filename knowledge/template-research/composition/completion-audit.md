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
- `runtime/workbookcase.RunOrganismPlan` provides multi-operation organism
  harness slices for the five first productization classes: invoice line item
  billing, monthly budget control, inventory movement log, student gradebook,
  and loan repayment schedule mechanics.
- `RunOrganismPlan` now emits an organism-level verification result artifact
  that records expected atoms, executed atoms, passed step counts, verifier spec
  id, output workbook, and failure reasons before `templateclass` accepts the
  runtime claim.
- The open-layer request compiler and orchestrator decision surface now carry
  optional `template_class_plan` hints so organism id, atom sequence, verifier
  specs, and non-claims can survive the request-entry boundary.
- The open-layer orchestrator now has an explicit multi-step
  `OrganismExecutionRequest` bridge that converts concrete organism steps into
  `RunOrganismPlan` execution and preserves the organism verification report.
- The public typed envelope dispatcher now accepts
  `request.kind=organism_execution_request`, validates the request file, and
  routes it to the explicit organism execution bridge.
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
- Aggregate organism-level verification exists for the first productization
  slice, but domain-specific organism verifiers do not exist yet. Runtime
  evidence still relies on operation-specific verifier depth for workbook
  semantics.
- Operation planner and template class harness records are advisory contracts,
  not autonomous runtime planner execution. The `runtime/templateclass` package
  plans and evaluates evidence, and the open-layer entry can preserve that plan
  hint and execute explicit organism steps, but it does not synthesize missing
  operation parameters.
- The organism harness currently covers the first five productization classes.
  The remaining roadmap organisms still need equivalent multi-operation
  fixtures before broad 21-organism runtime claims are appropriate.
- Native Excel pivot artifacts are not supported.
- Matrix auto-growth, rich visual layout QA, full printable pagination QA, and
  domain-specific schedule semantics remain future runtime work.
- `loan_repayment_calculator` proves formula extension and protected
  calculation mechanics, not amortization correctness or financial advice.

## Completion Judgment

The advisory ecosystem, runtime bridge, coverage ladder, representative preview
fixture coverage, p2/p3 verified-class depth contracts, and runtime
productization contracts are complete for the current 21 roadmap organisms.
The first deterministic template-class planner/evaluator slice now has
multi-operation organism harness coverage for all five productization classes.
The remaining limits are explicitly documented as future runtime implementation
depth, especially for broad 21-organism harness coverage and true
domain-specific organism verifiers, not untracked blockers in the requested
advisory-to-productization coverage layer.
