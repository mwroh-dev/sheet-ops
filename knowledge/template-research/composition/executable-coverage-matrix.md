# Executable Coverage Matrix

This matrix tracks the gap between advisory template research and executable
runtime coverage. It is intentionally narrower than the full template corpus:
the first execution target is p0 organism coverage.

## Current Atom Status

| Atom | Status | Runtime Evidence |
| --- | --- | --- |
| `group_summarize` | supported | capability record, TaskSpec/IR, executor, verifier, fixtures |
| `highlight_threshold` | supported | capability record, TaskSpec/IR, executor, verifier, fixtures |
| `join_lookup` | supported | capability record, TaskSpec/IR, executor, verifier, fixtures |
| `append_structured_rows` | supported | capability record, TaskSpec/IR, executor, verifier, fixtures |
| `extend_table_formulas` | supported | capability record, TaskSpec/IR, executor, verifier, fixtures |
| `add_data_validation` | supported | explicit list/dropdown scope: capability record, TaskSpec/IR, executor, verifier, fixtures |
| `write_values` | runtime primitive | primitive only, not public agent capability |
| `protect_formula_cells` | planned | broad p0 blocker, needs scoped protection semantics |
| `copy_period_sheet` | planned | budget/cash/timesheet blocker |
| `roll_forward_period` | planned | budget/cash continuity blocker |
| `normalize_headers` | planned | invoice/attendance/inventory ambiguity blocker |
| `reconcile_tables` | planned | inventory reconciliation blocker |
| `create_pivot_summary` | planned | p2+ summary depth blocker |
| `generate_printable_form` | planned | invoice document-output blocker |

## P0 Organism Gap

| Organism | Runtime-Covered Atoms | Missing Runtime Atoms | Immediate Strategy |
| --- | --- | --- | --- |
| `invoice_line_item_billing` | `join_lookup`, `group_summarize`, `append_structured_rows`, `extend_table_formulas`, `add_data_validation` | `generate_printable_form`, `protect_formula_cells`, `normalize_headers` | protect formula cells next, then printable form |
| `monthly_budget_control` | `group_summarize`, `highlight_threshold`, `extend_table_formulas` | `copy_period_sheet`, `roll_forward_period`, `protect_formula_cells` | period copy before roll-forward |
| `cash_flow_monitor` | `group_summarize`, `extend_table_formulas` | `copy_period_sheet`, `roll_forward_period`, `protect_formula_cells` | reuse budget period fixture after period copy exists |
| `attendance_register` | `group_summarize`, `add_data_validation` | `copy_period_sheet`, `protect_formula_cells`, `normalize_headers` | header normalization first, then period copy |
| `timesheet_hours_log` | `group_summarize`, `append_structured_rows`, `extend_table_formulas`, `add_data_validation` | `copy_period_sheet`, `protect_formula_cells` | reuse invoice row-growth mechanics, then period copy/protection |
| `inventory_movement_log` | `join_lookup`, `group_summarize`, `append_structured_rows` | `reconcile_tables`, `normalize_headers`, `protect_formula_cells` | normalize headers before reconciliation |

## Phase Checklist

Each runtime atom promotion must satisfy all checks before it can move from
planned to supported:

- capability record exists
- capability schema enum includes the atom
- task schema and request schema describe the public shape
- TaskSpec builder exists
- Operation IR compiler exists
- deterministic executor exists
- operation-specific verifier exists
- workbookcase path emits plan, execution, verification artifacts
- request compiler/use orchestrator can carry structured requests
- focused fixture test exists
- at least one p0 organism preview uses the atom, when applicable
- advisory records remove the atom from planned gaps
- self-retro records remaining scope limits

## Phase 5 Result

Promoted `add_data_validation` with a deliberately narrow first supported scope:

- same-workbook output copy
- target sheet range list
- list validation only
- explicit allowed values
- preserve original workbook
- verifier checks validation rule presence and allowed values

Non-goals for this phase:

- date windows
- numeric bounds
- cross-sheet reference lists
- invalid user entry simulation
- UI dropdown rendering proof

## Phase 5 Self-Retro

- P0 gaps reduced: `invoice_line_item_billing`, `attendance_register`, and
  `timesheet_hours_log` no longer depend on validation as an unsupported atom
  for explicit list/dropdown cases.
- Remaining template needs: formula-cell protection, printable document
  generation, period copy/roll-forward, header normalization, and
  reconciliation still block broader organism execution.
- Verifier strength: the verifier directly inspects workbook data validation
  rules and source preservation evidence. It does not simulate UI entry or
  prove Excel dropdown rendering quality.
- Advisory claim risk: numeric bounds, date windows, formula-backed lists, and
  referential integrity remain excluded from supported scope and must stay in
  opportunity/backlog language.
