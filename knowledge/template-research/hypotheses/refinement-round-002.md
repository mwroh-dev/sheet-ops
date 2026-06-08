# Refinement Round 002

## What changed
- Split inventory movement/reconciliation from static stock lists by making `inventory_movement_log` explicitly transactional and `warehouse_reorder_tracker` explicitly snapshot-based.
- Tightened `monthly_budget_control` toward plan-vs-actual variance control and `cash_flow_monitor` toward opening-to-closing balance continuity.
- Reframed `maintenance_issue_log` as a scheduled service log instead of an incident-and-repair tracker.
- Clarified `attendance_register` as a roster-by-date attendance sheet and `training_completion_matrix` as a person-by-requirement completion matrix.
- Reduced broad workflow triggers on weaker patterns such as `service_ticket_queue`, `sales_pipeline_tracker`, and `project_status_dashboard`.
- Kept every `evidence_status` at `hypothesis_only`.

## What stays ambiguous
- `attendance_register` and `training_completion_matrix` still share a lot of grid mechanics, so the boundary is conceptual rather than proven.
- `sales_pipeline_tracker`, `service_ticket_queue`, and `project_status_dashboard` still read like adjacent dashboard/queue variants, not settled promotion candidates.
- `warehouse_reorder_tracker` may still collapse into a generic inventory list if future evidence does not show a true snapshot/reorder workflow split.

## What Agent C should judge
- Whether the inventory pair is now distinct enough to keep both patterns.
- Whether the budget and cash-flow pair are separated by behavior instead of just by wording.
- Whether the attendance and training patterns are materially different capabilities or just different row/column layouts.
- Whether the weaker workflow-heavy patterns should stay parked until direct evidence arrives.
