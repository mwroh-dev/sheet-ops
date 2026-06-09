# Knowledge Corpus Layout

This directory separates reusable learning material from raw telemetry and scenario evidence.

- `golden_cases/`: known-good scenarios and stable reference behavior.
- `failure_cases/`: known failures and repair examples.
- `recipes/`: reusable operation patterns.
- `capability_registry/`: machine-readable or copied capability summaries for retrieval.
- `verification/episodic/failures/`: append-only verification failure episodes.
- `verification/semantic/`: promoted cross-run verification knowledge.

## Public Knowledge Promotion Rule

- raw episodic records are not public docs
- reusable lessons must be rewritten as semantic knowledge with no local paths,
  account names, run IDs, or private workbook content
- knowledge never overrides `TaskSpec`, `OperationIR`, runtime execution, or
  verification authority
