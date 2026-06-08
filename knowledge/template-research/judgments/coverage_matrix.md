# Template Research Coverage Matrix

Round: `round-009`

## By Status

| Status | Count | Notes |
| --- | ---: | --- |
| `roadmap_candidate` | 21 | The roadmap set stays stable in round-009; no promoted pattern regressed, and `attendance_register` remains firmly matrix-first. |
| `recurring` | 1 | `health_metrics_tracker` remains the only recurring pattern, but round-009 converts the last medium backlog into a stop decision: keep it recurring, exclude it from fixture kickoff, and stop further collection. |
| `observed` | 0 | No patterns remain parked at observed. |
| `rejected` | 14 | Rejections still include the retired `expense_policy_checker`, which round-009 did not reopen because no workbook-native policy gate appeared. |

## By Domain

| Domain | Judgments | Strongest status | Patterns |
| --- | ---: | --- | --- |
| analytics | 1 | `rejected` | `research_data_log` |
| compliance | 2 | `roadmap_candidate` | `safety_compliance_register`, `compliance_action_register` |
| construction | 1 | `roadmap_candidate` | `construction_cost_tracker` |
| education | 2 | `roadmap_candidate` | `attendance_register`, `student_gradebook` |
| facilities | 1 | `roadmap_candidate` | `maintenance_issue_log` |
| finance | 6 | `roadmap_candidate` | `monthly_budget_control`, `expense_reimbursement`, `loan_repayment_calculator`, `cash_flow_monitor`, `insurance_claim_tracker`, `expense_policy_checker` |
| health | 1 | `recurring` | `health_metrics_tracker` |
| hr | 2 | `roadmap_candidate` | `timesheet_hours_log`, `training_completion_matrix` |
| logistics | 2 | `roadmap_candidate` | `warehouse_reorder_tracker`, `fleet_dispatch_board` |
| manufacturing | 1 | `rejected` | `quality_inspection_log` |
| nonprofit | 1 | `rejected` | `donation_campaign_tracker` |
| operations | 6 | `roadmap_candidate` | `inventory_movement_log`, `asset_lifecycle_register`, `event_planning_runbook`, `shift_roster_planner`, `membership_renewal_monitor`, `subscription_renewal_ledger` |
| procurement | 3 | `roadmap_candidate` | `purchase_order_control`, `procurement_reconciliation`, `vendor_performance_scorecard` |
| project_management | 2 | `roadmap_candidate` | `project_timeline_tracker`, `project_status_dashboard` |
| sales | 2 | `roadmap_candidate` | `sales_pipeline_tracker`, `client_contact_master` |
| sales_admin | 1 | `roadmap_candidate` | `invoice_line_item_billing` |
| science | 1 | `rejected` | `lab_sample_tracker` |
| support | 1 | `roadmap_candidate` | `service_ticket_queue` |

## By Trigger Family

| Trigger family | Judgments touching family | Distinct trigger IDs used | Highest status reached | Notes |
| --- | ---: | ---: | --- | --- |
| `validation` | 34 | 11 | `roadmap_candidate` | Validation remains the main boundary-setting layer, and round-009 reinforces that the missing health step is not more threshold-like evidence in general but workbook-native control behavior strong enough for a fixture verifier. |
| `data` | 34 | 10 | `roadmap_candidate` | Data-shape mechanics stay central because the stable roadmap set still clusters around row growth, roll-forward, reconciliation, and summary preservation. |
| `visual` | 26 | 10 | `roadmap_candidate` | Visual structure still matters most where the workbook is a matrix, a printable control surface, or a dashboard fed by repeated period sheets, but round-009 confirms that visual range charts alone do not justify health promotion. |
| `workflow` | 18 | 10 | `roadmap_candidate` | Workflow signals are now mostly a scoping check: the only remaining non-roadmap family failed promotion because too many variants stay chart-like or review-light instead of proving a stronger control loop. |

## By Capability Candidate

| Capability candidate | Judgment count | Highest status | Representative patterns |
| --- | ---: | --- | --- |
| `add_data_validation` | 34 | `roadmap_candidate` | `invoice_line_item_billing`, `monthly_budget_control`, `inventory_movement_log` |
| `protect_formula_cells` | 25 | `roadmap_candidate` | `invoice_line_item_billing`, `attendance_register`, `health_metrics_tracker` |
| `group_summarize` | 22 | `roadmap_candidate` | `invoice_line_item_billing`, `monthly_budget_control`, `service_ticket_queue` |
| `create_pivot_summary` | 18 | `roadmap_candidate` | `monthly_budget_control`, `service_ticket_queue`, `health_metrics_tracker` |
| `normalize_headers` | 16 | `roadmap_candidate` | `attendance_register`, `expense_reimbursement`, `health_metrics_tracker` |
| `append_structured_rows` | 15 | `roadmap_candidate` | `invoice_line_item_billing`, `attendance_register`, `inventory_movement_log` |
| `generate_printable_form` | 14 | `roadmap_candidate` | `invoice_line_item_billing`, `project_timeline_tracker`, `expense_reimbursement` |
| `copy_period_sheet` | 12 | `roadmap_candidate` | `monthly_budget_control`, `project_timeline_tracker`, `attendance_register` |
| `extend_table_formulas` | 11 | `roadmap_candidate` | `invoice_line_item_billing`, `attendance_register`, `inventory_movement_log` |
| `flag_outliers` | 7 | `roadmap_candidate` | `inventory_movement_log`, `maintenance_issue_log`, `health_metrics_tracker` |

## Coverage Readout

- Round-009 is a closure round rather than a promotion round: the 21 roadmap candidates remain stable, `attendance_register` stays promoted, and `expense_policy_checker` remains retired.
- `health_metrics_tracker` gets materially broader evidence in round-009 across blood-count, glucose, weight-loss, growth, goal-sheet, and medical-progress variants, but the new material mostly reinforces a recurring threshold-aware family instead of proving workbook-native protected summaries or in-sheet control gates.
- The prior medium backlog is therefore resolved by judgment rather than promotion: no further evidence round is justified, and the health family stays outside the implementation fixture kickoff.
- The research phase is now complete enough to hand off to implementation fixture work on the supported roadmap set while keeping `health_metrics_tracker` as a documented non-roadmap boundary and `expense_policy_checker` as reopen-only.
