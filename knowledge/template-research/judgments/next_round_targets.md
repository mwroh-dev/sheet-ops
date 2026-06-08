# Next Stage Targets

Stage anchor: `post-round-009`

## Research Gate

No further collection round is justified. Round-009 resolves the last medium backlog by keeping `health_metrics_tracker` at `recurring` and ending research in favor of implementation fixture work on the supported roadmap set.

## Fixture Work Targets

1. Form-first row growth and printable output:
   `invoice_line_item_billing`, `expense_reimbursement`, `purchase_order_control`
2. Period roll-forward with locked summaries:
   `monthly_budget_control`, `cash_flow_monitor`, `attendance_register`, `timesheet_hours_log`
3. Reconciliation and transaction preservation:
   `inventory_movement_log`, `procurement_reconciliation`, `warehouse_reorder_tracker`
4. Schedule and grid propagation:
   `project_timeline_tracker`, `shift_roster_planner`, `construction_cost_tracker`
5. Dense matrix and scored-summary control:
   `student_gradebook`, `training_completion_matrix`, `vendor_performance_scorecard`
6. Status-pipeline and review workflows:
   `service_ticket_queue`, `sales_pipeline_tracker`, `compliance_action_register`, `safety_compliance_register`

## Fixture Exclusions

1. Keep `health_metrics_tracker` out of fixture kickoff unless later evidence shows workbook-native protected summaries, locked threshold logic, or another direct in-sheet control gate.
2. Keep `expense_policy_checker` retired unless a sheet itself blocks or flags non-compliant claims with formulas, validations, or locked rule cells.
3. Keep `attendance_register` matrix-first; do not widen it back into append-only sign-in logs or dashboard-only attendance summaries during fixture design.

## Judge Carry-Forward Notes

1. The roadmap set is stable enough to drive fixture selection without a round-010 evidence pass.
2. Any future reopening should be evidence-triggered, not quota-driven:
   `health_metrics_tracker` only for stronger workbook-native control mechanics,
   `expense_policy_checker` only for a true in-sheet rule gate.
