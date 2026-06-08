# Refinement Round 004

## What changed
- Split the attendance boundary conceptually by keeping `attendance_register` as a matrix attendance hypothesis only: the trigger set now points at a roster cross-tab with a frozen reference band, while append-only check-in logic is excluded from this entry.
- Recast `expense_policy_checker` as a spreadsheet-native enforcement pattern rather than a generic policy artifact, with trigger language centered on threshold checks, category validation, cross-field rules, and formula cascades.
- Separated `safety_compliance_register` from `compliance_action_register` by making the former about risk and control state, and the latter about corrective work that continues until closure.
- Narrowed `construction_cost_tracker` to active cost tracking and phase rollups, while explicitly excluding estimate sheets and bid comparison sheets from the same hypothesis.
- Cleaned `event_planning_runbook` trigger hygiene so the trigger list only contains trigger-like cues; printable output remains a capability, not a trigger.
- Tightened `project_timeline_tracker` versus `project_status_dashboard` by anchoring the former on dated task sequencing and keeping the latter on broad dashboard rollups.

## What remains ambiguous
- `attendance_register` still sits near append-only check-in sheets, but the current hypothesis boundary treats those as a separate pattern rather than a shared umbrella.
- `expense_policy_checker` still depends on the later evidence staying worksheet-native; if the next round drifts back toward policy docs, the pattern should stay parked.
- `construction_cost_tracker` may still split further if future sheets cleanly separate estimate, scheduled spend, active tracking, and bid comparison into distinct workbook shapes.
- `event_planning_runbook` remains a rejected pattern, and its trigger cleanup should not be read as a promotion signal.

## Caveats
- No evidence was promoted; every `evidence_status` remains `hypothesis_only`.
- No schema changes were required, and no new pattern ids were added.
