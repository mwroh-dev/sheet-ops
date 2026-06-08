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
| `copy_period_sheet` | supported | explicit sheet-copy scope: capability record, TaskSpec/IR, executor, verifier, fixtures |
| `add_data_validation` | supported | explicit list/dropdown scope: capability record, TaskSpec/IR, executor, verifier, fixtures |
| `protect_formula_cells` | supported | explicit formula/input range scope: capability record, TaskSpec/IR, executor, verifier, fixtures |
| `write_values` | runtime primitive | primitive only, not public agent capability |
| `roll_forward_period` | planned | budget/cash continuity blocker |
| `normalize_headers` | planned | invoice/attendance/inventory ambiguity blocker |
| `reconcile_tables` | planned | inventory reconciliation blocker |
| `create_pivot_summary` | planned | p2+ summary depth blocker |
| `generate_printable_form` | planned | invoice document-output blocker |

## P0 Organism Gap

| Organism | Runtime-Covered Atoms | Missing Runtime Atoms | Immediate Strategy |
| --- | --- | --- | --- |
| `invoice_line_item_billing` | `join_lookup`, `group_summarize`, `append_structured_rows`, `extend_table_formulas`, `add_data_validation`, `protect_formula_cells` | `generate_printable_form`, `normalize_headers` | printable form or normalize headers next |
| `monthly_budget_control` | `group_summarize`, `highlight_threshold`, `extend_table_formulas`, `copy_period_sheet`, `protect_formula_cells` | `roll_forward_period` | roll-forward continuity next |
| `cash_flow_monitor` | `group_summarize`, `extend_table_formulas`, `copy_period_sheet`, `protect_formula_cells` | `roll_forward_period` | reuse budget period fixture for carry-forward |
| `attendance_register` | `group_summarize`, `copy_period_sheet`, `add_data_validation`, `protect_formula_cells` | `normalize_headers` | header normalization next |
| `timesheet_hours_log` | `group_summarize`, `copy_period_sheet`, `append_structured_rows`, `extend_table_formulas`, `add_data_validation`, `protect_formula_cells` | none in current p0 atom list | broaden organism preview before claiming full template generation |
| `inventory_movement_log` | `join_lookup`, `group_summarize`, `append_structured_rows`, `protect_formula_cells` | `reconcile_tables`, `normalize_headers` | normalize headers before reconciliation |

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

## Phase 6 Result

Promoted `protect_formula_cells` with a deliberately narrow first supported
scope:

- same-workbook output copy
- explicit source sheet
- explicit formula ranges
- optional explicit input ranges
- sheet protection enabled
- preserve original workbook
- verifier checks formula cells are locked, input cells are unlocked, sheet
  protection options are present, and source hash is preserved

Non-goals for this phase:

- automatic formula discovery
- complex named/protected ranges
- role-based or collaborative permissions
- workbook-level protection
- visual proof of Excel protection UI

## Phase 6 Self-Retro

- P0 gaps reduced: all p0 organisms no longer depend on formula protection as
  an unsupported atom for explicit-range cases.
- Remaining template needs: period copy/roll-forward is now the largest p0
  cross-cutting blocker; `normalize_headers`, `reconcile_tables`, and
  `generate_printable_form` still block specific organism families.
- Verifier strength: the verifier directly inspects cell protection styles,
  sheet protection options, formulas in protected cells, and source hash
  preservation. It does not prove user-facing Excel edit rejection behavior.
- Advisory claim risk: automatic discovery of formula regions must not be
  implied. Callers must supply formula/input ranges until a discovery verifier
  exists.

## Phase 7 Result

Promoted `copy_period_sheet` with a deliberately narrow first supported scope:

- same-workbook output copy
- explicit source sheet
- explicit target sheet
- target sheet must not already exist
- values, formulas, and styles are preserved
- preserve original workbook

Non-goals for this phase:

- period label rewriting
- input clearing
- carry-forward opening/closing balance continuity
- block-level copy inside a sheet
- tables, charts, drawings, and pictures beyond what the underlying workbook
  copy API preserves

## Phase 7 Self-Retro

- P0 gaps reduced: `monthly_budget_control`, `cash_flow_monitor`,
  `attendance_register`, and `timesheet_hours_log` no longer depend on period
  copying as an unsupported atom for explicit sheet-copy cases.
- Remaining template needs: `roll_forward_period` is now the main blocker for
  budget and cash-flow continuity; `normalize_headers` blocks invoice,
  attendance, and inventory paths; `reconcile_tables` and
  `generate_printable_form` remain family-specific blockers.
- Verifier strength: the verifier directly compares copied sheet values,
  formulas, styles, target sheet existence, and source hash preservation.
- Advisory claim risk: period copy must not be described as a full
  roll-forward. Label rewriting, input clearing, and carry-forward mapping
  remain outside supported scope.
