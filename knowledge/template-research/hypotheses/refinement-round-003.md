# Refinement Round 003

## What changed
- Sharpened `project_timeline_tracker`, `training_completion_matrix`, `shift_roster_planner`, `compliance_action_register`, `procurement_reconciliation`, `inventory_movement_log`, `warehouse_reorder_tracker`, `timesheet_hours_log`, and `cash_flow_monitor` so their verifier boundaries are more specific and easier to distinguish from nearby patterns.
- Separated `expense_reimbursement` from `expense_policy_checker` by making reimbursement about claim submission, receipts, approvers, and payout totals, while policy checking stays focused on limit and category enforcement before approval.
- Reduced trigger inflation on `sales_pipeline_tracker`, `service_ticket_queue`, `purchase_order_control`, and other weak workflow-heavy patterns by removing broader workflow handoff/repeating cues and leaning more on row, table, and snapshot mechanics.

## What remains ambiguous
- `project_timeline_tracker` still overlaps visually with dashboard-style project views, but its boundary is now anchored on dated milestone sequencing rather than general status display.
- `training_completion_matrix` still shares grid mechanics with other roster-style sheets, so its identity depends on the person-by-requirement cross-tab.
- `inventory_movement_log` and `procurement_reconciliation` can both mention reconciliation, but one is transaction-led while the other is document-match-led.
- `warehouse_reorder_tracker` is still vulnerable to collapse into a generic inventory list unless future evidence keeps showing reorder thresholds and snapshot mechanics.
- `sales_pipeline_tracker` and `service_ticket_queue` remain workflow-adjacent, but their trigger sets are now narrower and less inflated.

## Caveats
- No evidence was promoted; every `evidence_status` remains `hypothesis_only`.
- No new pattern ids were added, so the registry size is unchanged.
