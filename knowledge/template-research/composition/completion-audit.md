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
- All 21 roadmap organisms are at `preview_fixture` tier.
- The knowledge harness validates roadmap coverage, fixture priority alignment,
  advisory authority, supported atom references, absence of planned blockers,
  and preview evidence presence.
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
- Native Excel pivot artifacts are not supported.
- Matrix auto-growth, rich visual layout QA, full printable pagination QA, and
  domain-specific schedule semantics remain future runtime work.
- `loan_repayment_calculator` proves formula extension and protected
  calculation mechanics, not amortization correctness or financial advice.

## Completion Judgment

The advisory ecosystem, runtime bridge, coverage ladder, and representative
preview fixture coverage are complete for the current 21 roadmap organisms.
The remaining limits are explicitly documented as future runtime depth, not
untracked blockers in the requested advisory-to-preview coverage layer.
