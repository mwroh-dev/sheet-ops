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
| `invoice_line_item_billing` | `join_lookup`, `group_summarize`, `append_structured_rows`, `extend_table_formulas`, `add_data_validation`, `protect_formula_cells`, `normalize_headers`, `generate_printable_form` | none in current p0 atom list | fixture-backed preview exists; do not claim full template generation |
| `monthly_budget_control` | `group_summarize`, `highlight_threshold`, `extend_table_formulas`, `copy_period_sheet`, `protect_formula_cells`, `roll_forward_period` | none in current p0 atom list | fixture-backed preview exists; do not claim full template generation |
| `cash_flow_monitor` | `group_summarize`, `extend_table_formulas`, `copy_period_sheet`, `protect_formula_cells`, `roll_forward_period` | none in current p0 atom list | fixture-backed preview exists; do not claim full template generation |
| `attendance_register` | `group_summarize`, `copy_period_sheet`, `add_data_validation`, `protect_formula_cells`, `normalize_headers` | none in current p0 atom list | fixture-backed preview exists; do not claim full template generation |
| `timesheet_hours_log` | `group_summarize`, `copy_period_sheet`, `append_structured_rows`, `extend_table_formulas`, `add_data_validation`, `protect_formula_cells` | none in current p0 atom list | fixture-backed preview exists; do not claim full template generation |
| `inventory_movement_log` | `join_lookup`, `group_summarize`, `append_structured_rows`, `protect_formula_cells`, `normalize_headers`, `reconcile_tables` | none in current p0 atom list | fixture-backed preview exists; do not claim full template generation |

## P0 Organism Preview Coverage

These previews prove representative supported-atom composition for each p0
organism. They are intentionally not full workbook generators.

| Organism | Fixture-Backed Preview Evidence | Coverage Limit |
| --- | --- | --- |
| `invoice_line_item_billing` | append structured rows, extend formulas, add validation, protect formulas, generate printable form | does not prove full invoice layout/design inference |
| `monthly_budget_control` | group summary and threshold highlight | does not prove period copy, protected summary, or full budget workbook generation in one flow |
| `cash_flow_monitor` | roll-forward period carry-forward | does not prove full cash-flow dashboard generation |
| `attendance_register` | copy period sheet, add validation, protect formula cells | does not prove matrix auto-growth or calendar projection |
| `timesheet_hours_log` | copy period sheet | does not prove full row append, formula extension, and protection in one organism flow |
| `inventory_movement_log` | normalize headers and reconcile tables | does not prove full stock ledger generation |

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

## Phase 12 Result

Broadened p0 organism preview coverage so every p0 organism has fixture-backed
evidence that at least one representative supported-atom composition executes
through the workbookcase path:

- `invoice_line_item_billing`: append, formula extension, validation,
  protection, and printable form generation
- `monthly_budget_control`: group summary and threshold exception highlighting
- `cash_flow_monitor`: period roll-forward
- `attendance_register`: period copy, dropdown validation, and formula-cell
  protection
- `timesheet_hours_log`: period copy
- `inventory_movement_log`: header normalization and table reconciliation

Non-goals for this phase:

- full template generation for all p0 organisms
- all atom combinations inside each organism
- all 21 roadmap organisms
- p2+ `create_pivot_summary`
- visual/rendered workbook QA beyond operation-specific verifier evidence

## Phase 12 Self-Retro

- P0 coverage improved: there are no remaining p0 atom gaps in the current
  matrix, and every p0 organism now has at least one executable preview backed
  by workbook fixtures.
- Remaining template needs: this is still below L5. The preview layer proves
  representative composition only; it does not prove full template assembly,
  all 21 organism coverage, pivot-style analytical summaries, or rich visual
  document verification.
- Verifier strength: each preview depends on existing operation-specific
  verifiers rather than a new organism-level verifier. That keeps runtime
  claims narrow but leaves organism-level acceptance criteria as future work.
- Advisory claim risk: organism names must not be interpreted as supported
  public runtime capabilities. Supported status still comes only from the
  capability registry, runtime contracts, and verifier-backed operation paths.

## Phase 13 Result

Added `executable-organism-coverage.json` and a knowledge harness that maps all
21 round-009 roadmap organisms to one of four claim tiers:

- `preview_fixture`: representative workbookcase fixture exists
- `supported_atom_plan`: supported atom sequence can be planned, but no
  organism preview fixture exists yet
- `blocked_by_planned_atom`: at least one planned atom, currently
  `create_pivot_summary`, blocks stronger runtime claims
- `advisory_only_gap`: core workbook behavior lacks a supported atom sequence

The harness validates that every roadmap organism appears exactly once, fixture
priority matches the organism catalog, executable sequence atoms are supported
capabilities, planned blockers are opportunity records, and each entry states a
claim, remaining gaps, and next promotion gate.

## Phase 13 Self-Retro

- Connection improved: the 21-organism set is no longer only an advisory list.
  Each organism now has a verified coverage-tier record tying it to supported
  atoms, preview evidence, or explicit blockers.
- Remaining template needs: p1 and p2 organisms still lack broad organism
  preview fixtures. The ladder makes that visible but does not itself execute
  those organisms.
- Verifier strength: the new harness checks catalog/roadmap/atom consistency.
  It does not prove workbook behavior beyond existing operation-specific
  previews.
- Advisory claim risk: `supported_atom_plan` must be read as a planning
  decomposition only. It is weaker than `preview_fixture` and much weaker than
  full template generation.

## Phase 14 Result

Broadened p1 organism preview coverage. The following p1 roadmap organisms now
have representative workbookcase preview fixtures:

- `expense_reimbursement`: append expense rows, extend reimbursable formula,
  generate printable claim
- `purchase_order_control`: enrich supplier metadata, validate PO status
- `project_timeline_tracker`: copy timeline period, validate task status
- `shift_roster_planner`: copy roster period, validate shift codes
- `construction_cost_tracker`: append cost row, highlight overrun variance
- `procurement_reconciliation`: reconcile PO and invoice amount tables
- `warehouse_reorder_tracker`: enrich SKU metadata, highlight low-stock rows

The executable coverage ladder now marks p0 and p1 roadmap organisms as
`preview_fixture`.

## Phase 14 Self-Retro

- Coverage improved: p1 no longer stops at supported-atom planning. Each p1
  organism has at least one fixture-backed execution path.
- Remaining template needs: these previews are intentionally representative.
  They do not prove every molecule in each organism, organism-level acceptance
  criteria, or full template generation.
- Verifier strength: each preview still relies on operation-specific verifiers.
  Cross-operation organism verifiers remain future work.
- Advisory claim risk: p1 preview evidence should not be read as broad domain
  automation. It proves that a narrow supported-atom composition executes for
  the organism family.

## Phase 15 Result

Broadened p2 organism preview coverage without promoting
`create_pivot_summary` into the supported capability registry. The phase uses
the opportunity record's own first-step guidance: deterministic summary-sheet
verification before native Excel pivot artifacts.

The following p2 roadmap organisms now have representative workbookcase preview
fixtures:

- `student_gradebook`: score summary and status validation
- `training_completion_matrix`: completion summary and printable report
- `service_ticket_queue`: ticket append and SLA threshold flag
- `sales_pipeline_tracker`: stage summary and risk threshold flag
- `maintenance_issue_log`: issue append and overdue threshold flag
- `compliance_action_register`: status summary and printable report
- `safety_compliance_register`: risk threshold flag and printable report

The executable coverage ladder now marks p0, p1, and p2 roadmap organisms as
`preview_fixture`. `create_pivot_summary` remains an opportunity, not a
supported atom.

## Phase 15 Self-Retro

- Coverage improved: p2 no longer depends on a planned pivot atom before any
  executable preview can run. Deterministic summary previews cover the first
  useful runtime layer.
- Remaining template needs: native pivot tables, matrix auto-growth,
  score/completion dashboard semantics, and organism-level acceptance criteria
  remain outside supported scope.
- Verifier strength: `group_summarize`, `highlight_threshold`,
  `append_structured_rows`, `add_data_validation`, and
  `generate_printable_form` verifiers provide the evidence. There is still no
  native pivot verifier.
- Advisory claim risk: this phase intentionally avoids pretending that
  deterministic summary sheets are full Excel pivot tables. Any future native
  pivot support needs a separate capability, verifier, and fixture path.

## Phase 16 Result

Added the p3 `loan_repayment_calculator` preview fixture:

- extend repeated payment-row formulas across interest, principal, and balance
  columns
- protect calculated schedule cells while leaving payment input cells editable

The executable coverage ladder now marks all 21 round-009 roadmap organisms as
`preview_fixture`.

## Phase 16 Self-Retro

- Coverage improved: there are no remaining roadmap organisms stuck at
  `supported_atom_plan` or `blocked_by_planned_atom`.
- Remaining template needs: the loan preview does not prove amortization
  correctness, payoff termination, cumulative balance semantics, or a dedicated
  `calculation_schedule_block` verifier.
- Verifier strength: the evidence is formula extension plus formula-cell
  protection, not a domain-specific loan calculator verifier.
- Advisory claim risk: this should be described as executable schedule
  mechanics, not as a supported financial calculator or financial advice
  capability.

## Phase 17 Result

Added `verified-organism-classes.json` and a knowledge harness for p2/p3 depth
contracts. This adds a higher advisory tier, `organism_verified_class`, for the
organisms where representative previews are backed by explicit acceptance
criteria, non-claims, runtime atom references, and promotion gates:

- p2 deterministic summary depth: gradebook, sales pipeline, compliance action,
  safety compliance
- p2 matrix/status depth: training completion, service tickets, maintenance
  issues
- p3 calculation schedule depth: loan repayment schedule mechanics

The coverage ladder now distinguishes:

- `preview_fixture`: representative workbookcase preview exists
- `organism_verified_class`: representative preview plus an advisory
  verified-class depth contract exists

## Phase 17 Self-Retro

- Coverage improved: p2/p3 are no longer only breadth previews. Their summary,
  matrix/status, and calculation schedule limits are now explicit and
  harness-validated.
- Remaining template needs: native pivot tables, arbitrary matrix growth,
  organism-level runtime verifiers, and amortization correctness remain future
  work.
- Verifier strength: the new harness validates the depth contracts and their
  references, while workbook behavior is still proven by operation-specific
  preview tests.
- Advisory claim risk: `organism_verified_class` is not a public capability
  and must not be described as full template generation.

## Phase 18 Result

Added the first runtime productization contract layer under
`knowledge/template-research/runtime/`:

- `organism-verifier-specs.json`: organism-level acceptance checks that
  aggregate existing atom verifier evidence for invoice, budget, inventory,
  gradebook, and loan schedule classes
- `operation-planner.json`: advisory organism-to-supported-atom planning
  sequences and fallback policies
- `template-class-harness.json`: request-to-classify-to-plan-to-execute-to-
  verify-to-report harness scenarios
- `advanced-atom-promotion.json`: promotion decisions for native pivots,
  matrix growth, timeline projection, calculation schedule verification, and
  printable render QA

The knowledge harness validates these records against schemas and checks that
planner/runtime atom references stay inside the supported capability registry,
while advanced atoms remain unpromoted until dedicated verifier evidence exists.

## Phase 18 Self-Retro

- Coverage improved: the project now has an explicit bridge from
  organism coverage to runtime productization planning.
- Remaining template needs: these are contracts and harness scenarios, not an
  autonomous classifier, planner, executor, or organism-level verifier
  implementation.
- Verifier strength: the harness validates artifact integrity and supported
  atom boundaries. It does not execute a multi-operation organism plan yet.
- Advisory claim risk: advanced atom decisions intentionally keep native pivot,
  matrix growth, timeline projection, calculation schedule verification, and
  printable render QA out of supported runtime until separate implementations
  and verifiers exist.
