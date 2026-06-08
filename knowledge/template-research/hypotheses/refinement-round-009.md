# Refinement Round 009

## What changed
- Narrowed `health_metrics_tracker` to a verifier-first measurement-control boundary. The hypothesis now requires repeated measurements, explicit targets or ranges, protected summaries or locked flags, and durable trend or period review.
- Kept diary, regimen, archive, and planning-style health sheets outside the core pattern unless they can directly prove the same measurement-control loop.
- Left `attendance_register` and `expense_policy_checker` boundaries untouched so the round-008 promotion and retirement decisions remain stable.

## What remains ambiguous
- Whether some wellness or regimen sheets are only thin wrappers around measurement control or are separate workflow patterns that should be split later.
- Whether the next evidence set can demonstrate the same verifier across repeated measurements, threshold logic, and period review without drifting into diary-only or schedule-only workbooks.

## Caveats
- No evidence was promoted.
- Every `evidence_status` in `pattern_hypotheses.json` remains `hypothesis_only`.
- No judgments, raw observations, schemas, patterns, opportunities, or tests were edited.
