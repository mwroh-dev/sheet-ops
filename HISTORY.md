# Sheet Ops History

<!-- sheet-ops-history:start -->
## 2026-06-11 - CLI agent contract

- Added a machine-readable `sheet-ops-codex` agent-contract surface for
  discovery, schema inspection, preflight readiness checks, preview impact
  inspection, install diagnostics, and internal validated-request handoff.
- Kept `sheet-ops` as the single human-facing workbook request entry while
  documenting `sheet-ops-codex` as an install, diagnostic, and agent-oriented
  contract surface.
- Added JSON schemas for CLI capabilities, command schema, error envelopes,
  preflight results, preview-request results, internal handoff results, and
  public entry results.
- Added stable CLI error taxonomy and exit code handling for usage errors,
  missing required options, unknown commands, invalid JSON/schema data,
  state-root mismatches, validation stops, execution failures, verification
  failures, and internal errors.
- Hardened JSON-mode failure behavior so emitted JSON error envelopes remain on
  stdout without duplicate stderr noise.
- Added read-only `preview-request` impact inspection with normalized-intent and
  input workbook fingerprints. It is explicitly not runtime dry-run evidence.
- Added installed-package e2e coverage for project-local `.codex/skills/sheet-ops`
  installs, bundled CLI binaries, run-intent execution, output workbook
  fingerprints, public result envelopes, and internal handoff envelopes.
- Exercised the installed CLI from a throwaway project with 22 red-input and
  happy-path cases covering discovery, schemas, preflight failures,
  preview-request failures, state-root mismatch, prepare-use validation,
  missing handoff requests, and successful run-intent execution.

Limits: this release does not make `sheet-ops-codex` a second human-facing
workbook entry, does not claim preview-request is a dry run, and does not make
Sheet Ops hosted or multi-tenant ready.

## 2026-06-09 - Template research runtime bridge

- Expanded the public workbook composition surface from the original
  summary/highlight/lookup path to 12 public capability families.
- Reframed the package as a local skill harness for LLM-assisted workbook work:
  goal/fact capture, capability selection, verifier focus, stop conditions,
  runtime handoff, and evidence are now the documented decision path.
- Added the template research advisory front door: curated template patterns,
  hypotheses, classification, atom/molecule/organism catalogs, atom-builder
  audit records, and runtime bridge documentation.
- Added explicit organism execution request contracts, draft planner coverage,
  operation planner records, template-class harness records, and standalone
  final-workbook semantic verifier slices for 21 roadmap organisms.
- Refactored operation/request schemas away from large field-exclusion
  `not.anyOf.required` blocks into operation-specific `oneOf` variants.
- Cleaned request-compiler normalized intent schema boundaries so printable
  form fields only appear on the printable form intent surface, not on summary,
  aggregate, validation, mapping, materialization, or ambiguity shapes.
- Removed duplicated capability `operation_family` source records; runtime help
  output now derives the operation family from capability name.
- Removed raw collection shards and scratch research logs from the release
  mirror; release knowledge now keeps curated records and schema-validated
  advisory artifacts.
- Added an assist orchestration backlog for remaining skill-harness gaps:
  decision artifact enforcement, live guidance compliance, failure-to-knowledge
  promotion, template-class hint adoption checks, evidence-centered final
  answers, real workbook regression fixtures, and stop-condition taxonomy.

Limits: this release does not claim full template generation, native pivot or
matrix expansion, domain-specific policy correctness, financial advice, safety
certification, hosted readiness, or rich visual/print QA.
<!-- sheet-ops-history:end -->
