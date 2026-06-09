# Verification

Verification is deterministic runtime work. Agent review cannot turn a failed verifier into success.

Evidence is scenario-level proof, not raw logs. Telemetry remains append-only raw execution history, and knowledge remains durable cross-run learning.

The skill should inspect verification artifacts only after the skill-owned
runtime handoff completes.

Layered verification contract:

- Level 1: file opens
- Level 2: expected sheets/ranges exist
- Level 3: values/formulas match
- Level 4: formulas calculate
- Level 5: source preserved
- Level 6: render artifact emission
- Level 7: semantic task-specific check

The render artifact status is evidence, not a guess. Public installs default to
`SHEET_OPS_RENDER_MODE=never`, so Level 6 is skipped unless a maintainer
explicitly opts into render capture. When enabled and LibreOffice/soffice and
`pdftoppm` are available, the runtime records whether the preview file exists
and has nonzero bytes. The LLM must not infer visual success from workbook
structure alone. Level 6 is render artifact emission, not visual quality
verification.

Public evidence contract:

- `contracts/public_preview/claims.json` records the supported public claims
  and preview limitations.
- Raw or generated demo evidence under `artifacts/` is source-side verification
  material and is not part of the public package surface.
