# Refinement Round 008

## What changed
- Tightened `attendance_register` around the verifier boundary: the hypothesis now explicitly requires protected total cells, locked roster structure, per-date formula cascades, and roll-forward copies that survive row and period growth. The boundary stays matrix-first and keeps append-only logs out of scope.
- Narrowed `health_metrics_tracker` to threshold-controlled measurement work. The core hypothesis now centers explicit goal, limit, or alert checks paired with repeated measurements and protected summaries, while diary-like, regimen, archive, and schedule-only variants remain split-worthy unless they actually drive the measurement loop.
- Kept `expense_policy_checker` retired unless a workbook shows a true in-sheet rule gate. The hypothesis text now explicitly excludes reimbursement routing, approval prose, and printable claim forms from the pattern boundary.
- Preserved the existing roadmap boundaries for the rest of the hypothesis set, including `expense_reimbursement`, `service_ticket_queue`, `construction_cost_tracker`, `maintenance_issue_log`, and `purchase_order_control`.

## What remains ambiguous
- `attendance_register` still needs direct evidence that row growth and period roll-forward can coexist with protected totals and locked roster sections without the workbook collapsing into a simple log.
- `health_metrics_tracker` may still split if later evidence separates measurement control from broader diary, regimen, or archive behavior.
- `expense_policy_checker` remains a reopen-only hypothesis until a workbook-native gate appears inside the sheet itself rather than in surrounding reimbursement flow.

## Caveats
- No evidence was promoted.
- Every `evidence_status` in `pattern_hypotheses.json` remains `hypothesis_only`.
- No judgments, raw observations, schemas, patterns, opportunities, or tests were edited.
