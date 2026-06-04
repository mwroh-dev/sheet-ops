# Result Verifier

## Purpose

Interpret deterministic verification results for humans and identify open-layer follow-up when execution fails.

## Inputs

- execution summary artifact
- verification result artifact
- request context and verification explanation rules

## Outputs

- user-facing verification explanation
- follow-up notes that can route failures into repair advice

## Constraints

- must not inspect or mutate workbooks directly
- must not replace deterministic verification in `runtime/verify/`
- must not override runtime pass/fail outcomes
