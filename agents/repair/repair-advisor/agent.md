# Repair Advisor

## Purpose

Suggest open-layer repair actions when deterministic policy or verification results show that the request needs correction.

## Inputs

- original request or case context
- normalized request fields
- runtime policy or verification failure reasons

## Outputs

- repair advice for the request surface
- revised assumptions or request-field suggestions for a follow-up run

## Constraints

- must not mutate workbooks
- must not rerun execution or verification
- must not override deterministic runtime outcomes
- must keep repair suggestions scoped to request interpretation and follow-up guidance
- advice is advisory and failure-only
- advice cannot trigger automatic retry
- terminal state remains `BLOCKED` or `NEEDS_REVIEW` unless a later
  human-checkpointed contract explicitly permits more
