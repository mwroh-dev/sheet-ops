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
- Organism-level verifier specs now cover all 21 runtime draft planner
  organisms, so every required verifier spec id emitted by `runtime/templateclass`
  has an advisory acceptance contract.
- Operation planner contracts now cover all 21 runtime draft planner organisms,
  so every runtime-draft organism has an advisory classifier signal set,
  supported atom sequence, required verifier spec id, and fallback policy.
- Template-class harness scenarios now cover all 21 runtime draft planner
  organisms, so every organism has an advisory classify-plan-execute-verify-report
  scenario contract.
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
- `RunOrganismPlan` now includes standalone workbook semantic organism verifier
  slices for invoice line-item billing, expense reimbursement, and purchase
  order control, checking appended row evidence, extended formula evidence, and
  printable document evidence in the final workbook.
- `RunOrganismPlan` also includes a monthly budget semantic verifier slice that
  checks deterministic summary output, roll-forward cell output, and copied
  period formula evidence in the final workbook.
- `RunOrganismPlan` also includes a cash-flow monitor semantic verifier slice
  that checks deterministic inflow summary output, roll-forward opening balance
  output, and copied/extended closing formula evidence in the final workbook.
- `RunOrganismPlan` also includes an attendance register semantic verifier
  slice that checks next-period copied row evidence, attendance total formula
  evidence, status validation evidence, and protected formula sheet evidence in
  the final workbook.
- `RunOrganismPlan` also includes a timesheet hours log semantic verifier slice
  that checks next-period copied sheet evidence, appended employee-hour row
  evidence, pay formula extension evidence, work-code validation evidence, and
  protected formula sheet evidence in the final workbook.
- `RunOrganismPlan` also includes a project timeline tracker semantic verifier
  slice that checks next-period task evidence, progress formula extension
  evidence, timeline status summary evidence, status validation evidence, and
  protected formula sheet evidence in the final workbook.
- `RunOrganismPlan` also includes a shift roster planner semantic verifier
  slice that checks next-week roster row evidence, coverage formula evidence,
  shift-code validation evidence, and protected formula sheet evidence in the
  final workbook.
- `RunOrganismPlan` also includes a construction cost tracker semantic verifier
  slice that checks appended cost-row evidence, variance formula extension
  evidence, and protected formula sheet evidence in the final workbook.
- `RunOrganismPlan` also includes an inventory movement log semantic verifier
  slice that checks normalized movement headers, SKU lookup enrichment evidence,
  inventory reconciliation evidence, and protected movement formula sheet
  evidence in the final workbook.
- `RunOrganismPlan` also includes a procurement reconciliation semantic
  verifier slice that checks PO-to-invoice reconciliation output evidence,
  invoice review formula evidence, invoice status validation evidence, and
  protected review formula evidence in the final workbook.
- `RunOrganismPlan` also includes a warehouse reorder tracker semantic verifier
  slice that checks SKU lookup enrichment evidence, appended reorder-candidate
  evidence, reorder formula evidence, SKU validation evidence, and protected
  reorder formula evidence in the final workbook.
- `RunOrganismPlan` also includes a student gradebook semantic verifier slice
  that checks score summary evidence, weighted-score formula extension
  evidence, completion-status validation evidence, and protected calculated
  cell evidence in the final workbook.
- `RunOrganismPlan` also includes a training completion matrix semantic
  verifier slice that checks completion summary evidence, printable training
  report evidence, completion formula evidence, training-status validation
  evidence, and protected completion formula evidence in the final workbook.
- `RunOrganismPlan` also includes a service ticket queue semantic verifier
  slice that checks appended ticket evidence, SLA formula evidence, ticket
  summary evidence, ticket-status validation evidence, and protected SLA formula
  evidence in the final workbook.
- `RunOrganismPlan` also includes a sales pipeline tracker semantic verifier
  slice that checks appended deal evidence, weighted forecast formula evidence,
  pipeline stage summary evidence, stage validation evidence, and protected
  forecast formula evidence in the final workbook.
- `RunOrganismPlan` also includes a maintenance issue log semantic verifier
  slice that checks appended issue evidence, action-required formula evidence,
  maintenance action summary evidence, issue-status validation evidence, and
  protected action formula evidence in the final workbook.
- `RunOrganismPlan` also includes a compliance action register semantic
  verifier slice that checks appended action evidence, review-required formula
  evidence, compliance status summary evidence, printable register evidence,
  action-status validation evidence, and protected review formula evidence in
  the final workbook.
- `RunOrganismPlan` also includes a safety compliance register semantic
  verifier slice that checks appended safety-check evidence, action-required
  formula evidence, safety completion summary evidence, printable report
  evidence, safety-status validation evidence, and protected action formula
  evidence in the final workbook.
- `RunOrganismPlan` also includes a loan repayment calculator semantic
  verifier slice that checks interest, principal, and balance formula
  continuity on the next schedule row plus protected calculation-cell evidence
  in the final workbook.
- The open-layer request compiler and orchestrator decision surface now carry
  optional `template_class_plan` hints so organism id, atom sequence, verifier
  specs, and non-claims can survive the request-entry boundary.
- The open-layer orchestrator now has an explicit multi-step
  `OrganismExecutionRequest` bridge that converts concrete organism steps into
  `RunOrganismPlan` execution and preserves the organism verification report.
- The public typed envelope dispatcher now accepts
  `request.kind=organism_execution_request`, validates the request file, and
  routes it to the explicit organism execution bridge.
- Template-class draft planners now turn all five first productization classes
  into schema-valid explicit organism requests from workbook facts: invoice
  line-item, monthly budget, inventory movement, student gradebook, and loan
  repayment schedule mechanics.
- Draft planner coverage is now machine-recorded for all 21 roadmap organisms:
  all 21 have runtime draft planners.
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
- Aggregate organism-level verification exists, verifier specs cover all 21
  runtime draft planner organisms, and all 21 roadmap organisms now have
  standalone workbook semantic verifier slices.
- Operation planner and template class harness records are mostly advisory
  contracts, not broad autonomous runtime planner execution. The
  `runtime/templateclass` package plans and evaluates evidence, and the
  open-layer entry can preserve that plan hint, synthesize supported draft
  requests from workbook facts for all 21 roadmap organisms, and execute
  explicit organism steps.
- The organism harness currently covers representative productization classes
  and draft integrations now span all roadmap organisms; all 21 roadmap
  organisms have standalone final-workbook semantic verifier slices. This
  supports bounded 21-organism runtime evidence claims, not full template
  generation claims.
- Draft synthesis currently covers the first five productization classes,
  cash-flow continuity, weekly timesheet row logging, warehouse reorder
  tracking, expense reimbursement claims, purchase order line-item control,
  attendance register period-copy/status control, project timeline task-table
  status/progress control, shift roster period-copy/shift-code control,
  construction cost row/variance/overrun control, procurement
  PO-to-invoice reconciliation, training completion summary/reporting, service
  ticket queue append/SLA/status-summary control, sales pipeline
  deal/stage/forecast-formula control, and maintenance issue
  status/risk/action-formula control, compliance action
  status/overdue/printable-register control, and safety compliance
  risk/completion/printable-report control, and loan repayment
  schedule-formula/protection control. No round-009 roadmap organism remains
  explicit-request-only at the draft-planner coverage layer.
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
multi-operation organism harness coverage for all five productization classes,
and all five first productization classes have explicit draft-planner paths.
The remaining limits are explicitly documented as future runtime implementation
depth, especially for broad 21-organism standalone semantic verifier coverage,
domain-specific correctness, and visual/print QA, not untracked blockers in the
requested advisory-to-productization coverage layer.
