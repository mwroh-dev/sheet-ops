# Template Research Taxonomy

## Source Classes

- `free_template_site`: free or public template catalog used only for why/how
  observation.
- `open_source_repo`: repository with explicit license metadata.
- `research_corpus`: public or academic workbook corpus used as analysis seed.

## Domains

- `finance`: budgets, statements, loan calculators, cash flow, variance.
- `sales_admin`: invoices, quotes, receipts, statements, customer records.
- `operations`: inventory, orders, stock counts, maintenance, work orders.
- `hr`: timesheets, payroll, attendance, leave, reimbursement.
- `project`: Gantt charts, task trackers, WBS, risks, status reports.
- `education`: gradebooks, attendance, schedules, lesson planning.
- `health`: logs, trackers, meal plans, measurements.
- `household`: chores, personal lists, events, family records.

## Pattern Dimensions

Visual/frontend labels describe workbook presentation: `printable_form`,
`dashboard`, `timeline_grid`, `calendar_grid`, `roster_grid`,
`weekly_matrix`, `input_area`, `protected_formula_area`, `sectioned_form`,
`status_color`, `monthly_matrix`, `line_item_table`, `summary_table`,
`total_box`, `chart_view`.

Data/backend labels describe mechanics: `entity_table`, `lookup`,
`append_rows`, `aggregate`, `formula_propagation`, `reconciliation`,
`period_rollup`, `period_rollover`, `category_mapping`,
`date_range_expansion`.

Validation labels describe safety checks: `required_fields`, `dropdown_values`,
`numeric_bounds`, `date_windows`, `referential_integrity`,
`duplicate_detection`, `valid_date_range`, `valid_time_range`,
`known_status`, `known_category`.

Workflow labels describe business state: `draft_sent_paid`, `planned_actual`,
`stock_in_out`, `todo_in_progress_done`, `submitted_approved`,
`period_close`, `variance_review`, `coverage_review`, `overdue_review`,
`schedule_update`, `scenario_review`, `review_publish`.

## Capability Mapping

Existing Sheet Ops capabilities are evidence targets, not limits:

- `group_summarize`: summaries, rollups, category totals, attendance totals.
- `highlight_threshold`: low stock, over-budget, overdue, high risk.
- `join_lookup`: customer, SKU, employee, rate, and category enrichment.
- `write_values`: internal primitive only.

Opportunity records track missing capabilities until separate implementation and
verification promote them into the supported registry.
