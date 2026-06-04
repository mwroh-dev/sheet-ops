# Release

Publish the public bundle into the sibling release repository:

```bash
node scripts/publish/publish-release.mjs
```

The publish command updates the release working tree, runs the privacy scan,
and records `.release-state.json`, `.release-source.json`, and `HISTORY.md`.
Before the release is explicitly marked published, repeated publishes amend the
single bootstrap release commit. After it is marked published, each changed
publish creates an append-only release commit and requires a release note:

```bash
node scripts/publish/publish-release.mjs --mark-published
node scripts/publish/publish-release.mjs --release-note release-note.json
```

To publish into a different release repository, set `SHEET_OPS_RELEASE_DIR`:

```bash
SHEET_OPS_RELEASE_DIR=../sheet-ops-released node scripts/publish/publish-release.mjs
```

## Release Mirror

The release repository root is a sanitized mirror of the canonical package. It
is generated from the source repository allowlist, then verified by the release
publish flow and privacy scan.

The public release checkout should not be used as a scratch workspace. Do not
leave raw artifacts, telemetry, private knowledge records, temporary test
output, or untracked files there.

## Smoke

The deterministic release smoke path exercises the public harness shape with
fixture-backed specialist decisions, without depending on a live LLM call:

```bash
SHEET_OPS_SMOKE_ROOT=$(mktemp -d) GO_BIN=$(command -v go) bash scripts/run_codex_sheet_ops_agent_system_smoke.sh
```

The smoke proves the skill-local launcher, `UseEnvelope`, `TaskSpec`,
`OperationIR`, execution, verification, evidence, and delegated lifecycle proof
contracts fit together as one harness. Live Codex delegation is intentionally a
separate non-blocking smoke because it depends on external model latency.

Optional live diagnostic:

```bash
SHEET_OPS_ENABLE_LIVE_CODEX_SMOKE=1 bash scripts/run_codex_live_delegation_smoke.sh
```

Without `SHEET_OPS_ENABLE_LIVE_CODEX_SMOKE=1`, the live diagnostic exits 0 with
`status: skipped`. It is not the deterministic release smoke and does not prove
that live Codex delegation is always available.
