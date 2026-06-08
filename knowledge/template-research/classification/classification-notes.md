# Classification Notes

Classification starts after the round-009 research gate. The goal is not to
group templates by business name, but to compress roadmap candidates into
Sheet Ops implementation families.

## Classification Principle

Roadmap candidates are grouped by the mechanics Sheet Ops must implement and
verify:

- row growth and formula preservation
- period copy and roll-forward
- keyed reconciliation and threshold flags
- grid and schedule propagation
- matrix scoring and protected summaries
- status pipelines and review outputs

Domain labels such as invoice, budget, timesheet, or gradebook remain useful
examples, but they are secondary. The implementation family decides which
fixture and verifier should be built first.

## Primary Groups

- `form_document_output`: form-style workbooks with line items, totals, and
  printable outputs.
- `period_roll_forward`: recurring period workbooks with locked summaries and
  copy-forward behavior.
- `transaction_reconciliation`: operational tables that need matching,
  exceptions, thresholds, and lookup enrichment.
- `schedule_grid_propagation`: dated rows and grids that must grow without
  losing formulas or summaries.
- `matrix_scored_summary`: dense matrices with weighted scores, completion
  state, and protected totals.
- `status_review_workflow`: status-driven queues, risk/action registers, and
  review dashboards.
- `calculation_schedule`: specialized formula-generated schedules.

## Fixture Direction

Start with P0 fixtures because they exercise mechanics reused by later groups:

1. Form/document output.
2. Period roll-forward.
3. Reconciliation and schedule groups.
4. Matrix and workflow groups.
5. Specialized formula schedule.

This keeps fixture work tied to verifiable Sheet Ops operations rather than
building one-off replicas of template catalogs.

## Sweep Findings

Before adding subtypes, the current primary groups were swept for ambiguity.
The sweep found that several patterns are not "obvious" domain buckets:

- `attendance_register` could look like a matrix pattern, but its current
  verifier is period roll-forward with protected roster totals.
- `purchase_order_control` could drift into procurement matching, but the
  current boundary stops at PO document/control and handoff checkpoints.
- `maintenance_issue_log` has append-row mechanics, but its verifier is due
  state and review workflow.
- `construction_cost_tracker` must stay active-cost-control only; estimates,
  bid comparisons, and generic budgets are excluded.
- `warehouse_reorder_tracker` is threshold-snapshot inventory, while
  `inventory_movement_log` owns transaction-led movement.

The machine-readable sweep is in `classification-sweep.json`. It records each
ambiguous pattern's current group, alternatives, boundary rule, and trigger for
future reclassification.
