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
| `normalize_headers` | supported | explicit header-row mapping scope: capability record, TaskSpec/IR, executor, verifier, fixtures |
| `roll_forward_period` | supported | explicit closing-to-opening cell mapping scope: capability record, TaskSpec/IR, executor, verifier, fixtures |
| `reconcile_tables` | supported | explicit key and compare mapping scope: capability record, TaskSpec/IR, executor, verifier, fixtures |
| `generate_printable_form` | supported | explicit field/table binding scope: capability record, TaskSpec/IR, executor, verifier, fixtures |
| `write_values` | runtime primitive | primitive only, not public agent capability |
| `create_pivot_summary` | planned | p2+ summary depth blocker |

## P0 Organism Gap

| Organism | Runtime-Covered Atoms | Missing Runtime Atoms | Immediate Strategy |
| --- | --- | --- | --- |
| `invoice_line_item_billing` | `join_lookup`, `group_summarize`, `append_structured_rows`, `extend_table_formulas`, `add_data_validation`, `protect_formula_cells`, `normalize_headers`, `generate_printable_form` | none in current p0 atom list | broaden organism preview before claiming full template generation |
| `monthly_budget_control` | `group_summarize`, `highlight_threshold`, `extend_table_formulas`, `copy_period_sheet`, `protect_formula_cells`, `roll_forward_period` | none in current p0 atom list | broaden organism preview before claiming full template generation |
| `cash_flow_monitor` | `group_summarize`, `extend_table_formulas`, `copy_period_sheet`, `protect_formula_cells`, `roll_forward_period` | none in current p0 atom list | broaden organism preview before claiming full template generation |
| `attendance_register` | `group_summarize`, `copy_period_sheet`, `add_data_validation`, `protect_formula_cells`, `normalize_headers` | none in current p0 atom list | broaden organism preview before claiming full template generation |
| `timesheet_hours_log` | `group_summarize`, `copy_period_sheet`, `append_structured_rows`, `extend_table_formulas`, `add_data_validation`, `protect_formula_cells` | none in current p0 atom list | broaden organism preview before claiming full template generation |
| `inventory_movement_log` | `join_lookup`, `group_summarize`, `append_structured_rows`, `protect_formula_cells`, `normalize_headers`, `reconcile_tables` | none in current p0 atom list | broaden organism preview before claiming full template generation |

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
- Remaining template needs at that checkpoint: period copy/roll-forward was
  the largest p0 cross-cutting blocker; `normalize_headers`,
  `reconcile_tables`, and `generate_printable_form` still blocked specific
  organism families. Phase 8 resolves the explicit-mapping
  `normalize_headers` portion.
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
- Remaining template needs at that checkpoint: `roll_forward_period` was the
  main blocker for budget and cash-flow continuity; `normalize_headers`
  blocked invoice, attendance, and inventory paths; `reconcile_tables` and
  `generate_printable_form` remained family-specific blockers. Phase 8 resolves
  the explicit-mapping `normalize_headers` portion.
- Verifier strength: the verifier directly compares copied sheet values,
  formulas, styles, target sheet existence, and source hash preservation.
- Advisory claim risk: period copy must not be described as a full
  roll-forward. Label rewriting, input clearing, and carry-forward mapping
  remain outside supported scope.

## Phase 8 Result

Promoted `normalize_headers` with a deliberately narrow first supported scope:

- same-workbook output copy
- explicit source sheet
- explicit header row
- explicit `from` to `to` header mappings
- only mapped header-row cells are changed
- non-header values and formulas are preserved
- preserve original workbook

Non-goals for this phase:

- fuzzy alias inference
- duplicate header resolution
- multi-row header detection
- semantic canonicalization without explicit mappings
- table restructuring beyond header-cell rewrites

## Phase 8 Self-Retro

- P0 gaps reduced: `invoice_line_item_billing`, `attendance_register`, and
  `inventory_movement_log` no longer depend on header normalization as an
  unsupported atom for explicit mapping cases.
- Remaining template needs: `roll_forward_period` still blocks budget and
  cash-flow continuity, `reconcile_tables` blocks inventory movement proof, and
  `generate_printable_form` blocks document-style invoice output.
- Verifier strength: the verifier checks source hash preservation, mapped
  output headers, non-header row values, and non-header formulas.
- Advisory claim risk: the runtime must not imply automatic header discovery or
  alias matching. Callers must provide the mapping until a separate verifier
  exists for inference and ambiguity handling.

## Phase 9 Result

Promoted `roll_forward_period` with a deliberately narrow first supported
scope:

- same-workbook output copy
- explicit source sheet
- explicit existing target sheet
- explicit closing-to-opening cell mappings
- formula-backed closing values are calculated before carry-forward
- non-mapped cells and formulas are preserved
- preserve original workbook

Non-goals for this phase:

- creating the next period sheet
- period label rewriting
- input clearing
- automatic opening/closing field inference
- multi-sheet workbook calendar generation

## Phase 9 Self-Retro

- P0 gaps reduced: `monthly_budget_control` and `cash_flow_monitor` no longer
  depend on roll-forward as an unsupported atom for explicit mapping cases.
- Remaining p0 template needs at the end of Phase 9: `generate_printable_form`
  blocked invoice-style document output, and `reconcile_tables` blocked
  inventory movement proof.
- Verifier strength: the verifier checks source hash preservation, carried
  closing-to-opening values, non-mapped cell values, and non-mapped formulas.
- Advisory claim risk: this is carry-forward continuity, not full period
  generation. Callers must create/copy the target period sheet separately until
  a richer verifier exists.

## Phase 10 Result

Promoted `reconcile_tables` with a deliberately narrow first supported scope:

- explicit source and lookup sheets
- explicit left/right key columns
- explicit compare mappings
- new reconciliation result sheet
- deterministic buckets: `matched`, `left_only`, `right_only`,
  `value_mismatch`
- preserve original workbook and unchanged source/lookup sheets

Non-goals for this phase:

- fuzzy key matching
- duplicate-key resolution
- numeric tolerances or rounding policies
- composite keys
- many-to-many reconciliation
- external file reconciliation

## Phase 10 Self-Retro

- P0 gaps reduced: `inventory_movement_log` no longer depends on reconciliation
  as an unsupported atom for explicit key/value comparison cases.
- Remaining p0 template need: `generate_printable_form` still blocks
  invoice-style document output.
- Verifier strength: the verifier recomputes expected reconciliation rows from
  the input workbook, checks source hash preservation, checks source/lookup
  sheet preservation, and compares matched/missing/mismatch rows.
- Advisory claim risk: this is row-level reconciliation, not general audit
  automation. Ambiguous keys, duplicate keys, tolerances, and fuzzy matching
  must stay outside supported scope until separate verifiers exist.

## Phase 11 Result

Promoted `generate_printable_form` with a deliberately narrow first supported
scope:

- explicit target printable sheet
- explicit form title and print area
- explicit field bindings from source cells to label/value cells
- explicit table binding from declared source columns to a fixed output region
- source workbook preserved

Non-goals for this phase:

- automatic layout inference
- styling quality or visual design assertions
- merged-cell layout
- logos/images
- PDF rendering or pagination QA
- dynamic multi-page forms

## Phase 11 Self-Retro

- P0 gaps reduced: `invoice_line_item_billing` no longer depends on printable
  form generation as an unsupported atom for explicit field/table binding
  cases.
- Remaining p0 atom gaps: none in the current p0 atom list. This does not mean
  full template generation is complete; organism preview breadth and richer
  layout verification still need follow-up.
- Verifier strength: the verifier checks source hash preservation, source
  sheet preservation, populated field cells, populated table cells, and the
  workbook print-area defined name.
- Advisory claim risk: this is fixed-region materialization, not a general
  document designer. Visual quality, pagination, merged cells, images, and
  inferred layouts must stay outside supported scope until render/verifier
  coverage exists.
