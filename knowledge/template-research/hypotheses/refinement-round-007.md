# Refinement Round 007

## What changed
- Tightened `expense_policy_checker` into a strict in-sheet enforcement boundary: the hypothesis now only counts workbook-native rule gates driven by formulas, validations, and locked rule cells. If the sheet only carries policy prose or reimbursement routing notes, the pattern is treated as retired rather than expanded.
- Expanded `attendance_register` to make the matrix mechanics explicit: the hypothesis now calls out locked roster bands, protected totals, per-date formulas, roll-forward copies, and row or period growth that must preserve protected summary cells. Append-only check-in logs remain outside the pattern.
- Clarified `health_metrics_tracker` as threshold-aware measurement control rather than a generic tracker. Regimen workflows and durable records are included only when they support measurement control; standalone diary, archive, or casual logging sheets are split-worthy.
- Left `service_ticket_queue`, `construction_cost_tracker`, `maintenance_issue_log`, and `purchase_order_control` stable and bounded to their existing verifier paths.

## What remains ambiguous
- `expense_policy_checker` still depends on workbook-native proof. Without direct in-sheet enforcement, the safer move is retirement instead of widening the pattern.
- `attendance_register` still needs evidence that row growth and period roll-forward can coexist with protected totals and locked roster bands without collapsing into a log-like shape.
- `health_metrics_tracker` may still split if future evidence separates threshold-controlled measurement work from durable recordkeeping or regimen scheduling.

## Caveats
- No evidence was promoted.
- Every `evidence_status` in `pattern_hypotheses.json` remains `hypothesis_only`.
- No judgments, raw observations, schemas, patterns, opportunities, or tests were edited.
