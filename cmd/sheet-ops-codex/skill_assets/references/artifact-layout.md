# Artifact Layout

Purpose boundaries:

- `artifacts/telemetry`: append-only raw execution records
- `artifacts/evidence`: scenario-level proof bundles
- `artifacts/reports`: human-facing derived summaries
- `knowledge`: durable cross-run learning

The main evidence root is `artifacts/evidence`. Evidence may point to telemetry, proofs, verification, and repair advice artifacts, but it does not replace raw logs.

