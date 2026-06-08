# Refinement Round 006

## What changed
- Tightened `expense_policy_checker` to an explicit in-sheet enforcement boundary: the hypothesis now only counts when the workbook itself blocks or flags non-compliance through formulas, validations, and locked rule cells. Policy prose, routing notes, and reimbursement narrative stay outside the pattern.
- Refined `attendance_register` around matrix-verifier behavior: the hypothesis now calls out a locked roster band, date or session columns, protected totals, and roll-forward copies as the core shape.
- Split `health_metrics_tracker` toward threshold-aware measurement control rather than lightweight append logs, keeping repeated measurements but excluding simple personal logging sheets.
- Preserved the queue-first boundary for `service_ticket_queue` so intake-first request trackers do not drift back into the same hypothesis.
- Narrowed `construction_cost_tracker` to active cost control with change orders and forecast-to-complete mechanics, while keeping estimate, bid-comparison, and generic budget shapes out.
- Reframed `purchase_order_control` around request, approval, PO issuance, and receipt or invoice handoff checkpoints so it stays distinct from procurement reconciliation.

## What remains ambiguous
- `expense_policy_checker` still needs workbook-native evidence to justify continued activity; if the next round stays at the policy-prose level, this pattern should be treated as parked rather than broadened.
- `attendance_register` still needs sharper verifier detail around protected totals and roll-forward mechanics before it can be treated as anything more than a matrix hypothesis.
- `health_metrics_tracker` may still split again if future evidence separates regulated measurement workflows from casual tracking logs.
- `purchase_order_control` remains intentionally separate from `procurement_reconciliation`; the two only overlap at the handoff boundary, not the matching boundary.

## Caveats
- No evidence was promoted.
- Every `evidence_status` in `pattern_hypotheses.json` remains `hypothesis_only`.
- No schemas, judgments, raw observations, patterns, opportunities, or tests were edited.
