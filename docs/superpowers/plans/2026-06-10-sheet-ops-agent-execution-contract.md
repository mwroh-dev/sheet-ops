# Sheet Ops Agent Execution Contract Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move Sheet Ops from an agent-discoverable CLI contract to an agent-executable CLI contract with authoritative end-to-end evidence, typed runtime errors, stable result envelopes, formal JSON Schemas, and a truthful non-mutating preview surface.

**Architecture:** The work is split into two lanes. Lane A is the execution contract lane and must run sequentially because installed E2E smoke, result envelopes, runtime error producers, and preview semantics depend on the same runtime evidence model. Lane B is the schema compatibility lane; it can be researched in parallel but should be implemented after the envelope shape stabilizes, otherwise schema snapshots will churn. Every phase has red evidence, focused implementation, separate verifier review, phase-end backlog self-review, and a phase commit.

**Tech Stack:** Go 1.25+, Cobra, JSON Schema 2020-12, existing Sheet Ops runtime evidence artifacts, installed project-local `.codex/skills/sheet-ops` package, `go test`.

---

## Web Research Inputs

These sources have design value before implementation:

- CLI contract basics: [Command Line Interface Guidelines](https://clig.dev/) reinforces correct exit codes and machine-readable stdout.
- Agent tool error split: [MCP Tools error handling](https://modelcontextprotocol.io/specification/2025-11-25/server/tools) separates protocol errors from tool execution errors.
- Agent-visible validation errors: [MCP SEP-1303](https://modelcontextprotocol.io/seps/1303-input-validation-errors-as-tool-execution-errors) argues that input/business errors should reach the model so it can self-correct.
- JSON Schema dialect: [MCP JSON Schema usage](https://modelcontextprotocol.io/specification/2025-11-25/basic) recommends JSON Schema 2020-12.
- JSON Schema declaration: [JSON Schema dialect reference](https://json-schema.org/understanding-json-schema/reference/schema) recommends root `$schema` to tell tooling the intended dialect.
- Tool output schemas: [MCP schema reference](https://modelcontextprotocol.io/specification/2025-06-18/schema) includes `outputSchema` for structured tool results.

## Lanes

### Lane A - Execution Contract

Sequential implementation lane:

1. Phase 12: installed end-to-end workbook smoke discovery.
2. Phase 13: common result envelope contract.
3. Phase 14: runtime error typed producers.
4. Phase 15: installed end-to-end workbook smoke promotion.
5. Phase 17: non-mutating plan/impact preview.

Phase 12 comes before Phase 13 because the existing runtime output must be observed before a stable envelope is imposed. Phase 15 comes after Phases 13-14 so the E2E smoke can assert the final contract rather than a temporary shape.

### Lane B - Schema Compatibility

Research can run early, implementation waits for Phase 13:

1. Phase 16: formal JSON Schemas and golden compatibility.

Schema work depends on stabilized envelopes and error fields. It should not be implemented before Phase 13 unless only research notes are being added.

### Lane C - Independent Verification

Verifier lane. Every phase requires a verifier that is not the executor. Close each verifier subagent immediately after consuming the result.

Model tier guidance:

- Mechanical installed smoke and golden-file checks: `gpt-5.4-mini`.
- Result envelope/error producer design: `gpt-5.4` or higher.
- Final cross-phase execution contract review: `gpt-5.5` or highest available reasoning tier.

## Phase 12 - Installed E2E Workbook Smoke Discovery

Lane: Execution Contract.

Web-search value: low. Use local runtime fixtures and existing browser-flow authoritative-artifact methodology.

### TODO

- [ ] Identify the smallest stable workbook/request fixture.
  - Evaluation: fixture can run through installed skill surface without network, manual intervention, or long runtime.
  - Result: a candidate fixture path, request path, expected output workbook path, and expected evidence artifacts are documented.
  - Likely files: `internal/testfixtures/**`, `examples/**`, runtime package tests.
  - Risk: medium.
  - Rollback: document fixture blocker and stop before adding flaky smoke.
- [ ] Add a skipped or failing E2E smoke skeleton.
  - Evaluation: test installs the package into a temp project and attempts to execute the installed surface only.
  - Result: red evidence shows the missing authoritative success contract or missing stable fixture.
  - Likely files: `cmd/sheet-ops-codex/install_e2e_test.go` or `cmd/sheet-ops-codex/install_skill_test.go`.
  - Risk: high.
  - Rollback: keep this phase as discovery-only if runtime invocation is too expensive.
- [ ] Define authoritative success fields.
  - Evaluation: success cannot be based on stdout alone; it must require exit code, result JSON, output workbook existence, and evidence artifact existence.
  - Result: phase note lists the exact fields/files Phase 15 must assert.
- [ ] Run separate verifier.
  - Evaluation: verifier checks fixture stability and whether the proposed success fields are authoritative.
  - Result: pass/fail before commit.
- [ ] Commit Phase 12.
  - Evaluation: only discovery notes and/or smoke skeleton are staged.
  - Result: `phase12/e2e: discover installed workbook smoke contract`.

### Phase-End Backlog Review

Ask:

- Is the fixture fast and deterministic enough to live in normal package tests?
- Does the runtime require external tools or manual review that make this a CI-only or long-running test?
- Is a missing fixture a blocker for agent execution contract, or can envelope/error work proceed first?

Proceed to Phase 13 even if the E2E smoke is only a skeleton, but do not claim installed workbook execution until Phase 15.

## Phase 13 - Common Result Envelope Contract

Lane: Execution Contract.

Web-search value: medium. Use CLI stdout/exit-code guidance and MCP tool-result/error split, but adapt to Sheet Ops artifacts.

### TODO

- [ ] Add failing tests for a shared execution envelope shape.
  - Evaluation: mutating/diagnostic execution commands return JSON with stable outer fields: `schema_version`, `ok`, `status`, `command`, `artifacts`, `next_actions`, `recoverable`, and optional `error`.
  - Result: tests fail on commands whose success or blocked output is currently inconsistent.
  - Likely files: `cmd/sheet-ops-codex/*_test.go`, runtime result helpers.
  - Risk: high.
  - Rollback: start with one command and document commands not yet migrated.
- [ ] Implement minimal envelope adapter.
  - Evaluation: adapter wraps existing outputs without hiding existing detailed payloads.
  - Result: commands keep useful domain data while adding a predictable agent-facing outer contract.
- [ ] Define artifact references.
  - Evaluation: `artifacts` entries include kind, path, required flag, and success role.
  - Result: agent can decide which files prove success.
- [ ] Preserve exit-code semantics.
  - Evaluation: success returns 0; invalid usage/data/runtime failures return the existing non-zero codes.
  - Result: machine-readable stdout and exit code agree.
- [ ] Run separate verifier.
  - Evaluation: verifier checks that the envelope helps an agent recover and does not overclaim runtime success.
  - Result: pass/fail before commit.
- [ ] Commit Phase 13.
  - Evaluation: focused envelope tests pass.
  - Result: `phase13/envelope: standardize agent execution results`.

### Phase-End Backlog Review

Ask:

- Did envelope wrapping remove any domain detail needed by humans or agents?
- Are there commands that should intentionally remain discovery-only and not use execution envelopes?
- Does any error path still produce human-only text where an agent needs JSON?

Treat human-only execution failure output as a blocker for this phase.

## Phase 14 - Runtime Error Typed Producers

Lane: Execution Contract.

Web-search value: medium. Use MCP protocol vs execution error split and agent-visible validation error guidance.

### TODO

- [ ] Add failing producer tests for reserved runtime categories.
  - Evaluation: at least `state_root_mismatch`, `validation_blocked`, `request_checkpoint`, `execution_failed`, and `verification_failed` have explicit producer tests or remain documented as reserved with a blocker.
  - Result: every emitted runtime error has a live path and JSON envelope.
  - Likely files: `cmd/sheet-ops-codex/cli_errors_test.go`, runtime command tests.
  - Risk: high.
  - Rollback: promote only categories with deterministic producers.
- [ ] Wire typed errors at runtime boundaries.
  - Evaluation: runtime failures map to stable `error.code`, `recoverable`, `suggested_commands`, and artifact references.
  - Result: agents can distinguish retryable repair from hard failure.
- [ ] Update `error_contract.codes[]`.
  - Evaluation: statuses move from `reserved` to `emitted` only when producer tests pass.
  - Result: taxonomy stays honest.
- [ ] Run separate verifier.
  - Evaluation: verifier confirms no reserved category was promoted without evidence.
  - Result: pass/fail before commit.
- [ ] Commit Phase 14.
  - Evaluation: focused error tests pass.
  - Result: `phase14/errors: wire runtime typed producers`.

### Phase-End Backlog Review

Ask:

- Are any runtime failures too broad and need subcodes?
- Can a model self-correct from the message and `next_actions`?
- Does any producer leak implementation paths that should be artifact references instead?

Do not broaden error claims just to make the taxonomy look complete.

## Phase 15 - Installed E2E Workbook Smoke Promotion

Lane: Execution Contract.

Web-search value: low. Use Phase 12 local discovery and browser-flow authoritative-artifact methodology.

### TODO

- [ ] Promote the Phase 12 smoke to a passing test.
  - Evaluation: installed surface executes a minimal workbook request and exits 0.
  - Result: test proves installed workbook execution, not only metadata discovery.
- [ ] Assert authoritative artifacts.
  - Evaluation: result envelope `ok:true`, output workbook exists, evidence artifacts exist, and verifier/result artifact reports success.
  - Result: success claim is based on files and fields.
- [ ] Assert no manual intervention.
  - Evaluation: smoke output explicitly records no manual checkpoint or manual override.
  - Result: test cannot be mistaken for a manually assisted success.
- [ ] Run separate verifier.
  - Evaluation: verifier checks that stdout alone is not used as proof.
  - Result: pass/fail before commit.
- [ ] Commit Phase 15.
  - Evaluation: installed E2E smoke passes in focused mode.
  - Result: `phase15/e2e: prove installed workbook execution`.

### Phase-End Backlog Review

Ask:

- Is the test too slow for default `go test ./...`?
- Should it be guarded behind an integration build tag?
- Does the fixture prove enough workbook behavior or only a trivial no-op?

If the test is too slow, classify it explicitly as integration coverage and keep a fast focused command.

## Phase 16 - Formal JSON Schemas And Golden Compatibility

Lane: Schema Compatibility.

Web-search value: high. Use JSON Schema 2020-12, root `$schema`, modular `$defs`, and output-schema patterns.

### TODO

- [ ] Add JSON Schema files for public CLI outputs.
  - Evaluation: schemas exist for capabilities, command schema, error contract, result envelope, preflight, and installed E2E result.
  - Result: external agents can validate outputs without reading Go structs.
  - Likely files: `contracts/cli/*.schema.json`, `cmd/sheet-ops-codex/*_test.go`.
  - Risk: medium.
  - Rollback: start with capabilities and result envelope only.
- [ ] Add golden JSON fixtures.
  - Evaluation: representative command outputs are captured with stable fields and validated against schemas.
  - Result: compatibility drift is visible in review.
- [ ] Add compatibility policy tests.
  - Evaluation: `schema_version` changes are required for breaking field removals/renames; additive optional fields are allowed.
  - Result: tests encode backward compatibility expectations.
- [ ] Add docs for schema versioning.
  - Evaluation: docs define when to bump `schema_version` and how agents should handle unknown fields.
  - Result: consumers know what is stable.
- [ ] Run separate verifier.
  - Evaluation: verifier checks schemas use JSON Schema 2020-12 and validate generated outputs.
  - Result: pass/fail before commit.
- [ ] Commit Phase 16.
  - Evaluation: schema/golden tests pass.
  - Result: `phase16/schema: publish cli output contracts`.

### Phase-End Backlog Review

Ask:

- Are schemas duplicating Go structs in a way that will drift?
- Should schema generation be introduced, or are hand-written schemas clearer for now?
- Does compatibility policy need semantic versioning beyond `sheet-ops-cli/v1`?

Do not publish schemas for unstable envelopes until Phase 13 is complete.

## Phase 17 - Non-Mutating Plan And Impact Preview

Lane: Execution Contract.

Web-search value: low. This is mostly Sheet Ops domain design; external CLI guidance only confirms that it must not be mislabeled as dry-run.

### TODO

- [ ] Add failing tests for a read-only preview command.
  - Evaluation: command returns planned reads/writes/state root/artifacts without creating output workbook or mutation artifacts.
  - Result: agents can inspect impact before mutation.
  - Likely command names: `plan-request`, `inspect-request`, or `preview-request`.
  - Risk: high.
  - Rollback: document why existing runtime cannot preview without mutation.
- [ ] Implement the narrowest truthful preview.
  - Evaluation: preview may compile/validate request metadata but must not write workbook outputs.
  - Result: it is not called dry-run and does not claim execution success.
- [ ] Expose preview in capabilities/schema.
  - Evaluation: command is `read_only:true`, `mutating:false`, `dry_run_capable:false`.
  - Result: agents can discover it safely.
- [ ] Run separate verifier.
  - Evaluation: verifier checks no write side effects and no fake dry-run claim.
  - Result: pass/fail before commit.
- [ ] Commit Phase 17.
  - Evaluation: focused preview tests pass.
  - Result: `phase17/preview: add truthful impact inspection`.

### Phase-End Backlog Review

Ask:

- Does preview require enough compiler work that it should be split into a separate architectural track?
- Does it expose sensitive workbook paths or data in a way agents should redact?
- Is a real dry-run now feasible, or still a later runtime planning mode?

Keep real dry-run deferred unless the runtime can prove non-mutation.

## Final Review Phase - Agent Execution Contract Completion

Lane: Independent Verification.

### TODO

- [ ] Run full verification.
  - Evaluation: focused phase tests, `go test ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts -count=1`, `go test ./... -count=1`, generated JSON validation, installed E2E smoke, and schema compatibility tests pass.
  - Result: completion evidence is fresh.
- [ ] Dispatch final verifier.
  - Evaluation: verifier reviews all phases, authoritative artifacts, web-research-backed design choices, and no-overclaim constraints.
  - Result: pass/fail before final commit.
- [ ] Update final review artifact.
  - Evaluation: evidence matrix covers installed E2E, typed errors, result envelopes, schemas/goldens, preview, dry-run deferral, and remaining backlog.
  - Result: branch is reviewable.
- [ ] Commit final review.
  - Evaluation: worktree is clean after commit.
  - Result: `phase18/review: lock agent execution contract evidence`.

### Completion Audit Requirements

- Every phase has a commit.
- Every phase has evaluation/result criteria recorded before implementation.
- Every phase has phase-end backlog self-review.
- Web-search-backed phases cite the source class used for design.
- Verifier and executor are separated for each phase.
- Subagent sessions are closed after use.
- No success claim relies on stdout alone when authoritative artifacts exist.

## Execution Result Notes

### Phase 12 Result Note

Discovery outcome:

- The strongest installed E2E candidate is an `append_structured_rows` workbook
  request using an in-test generated workbook rather than checked-in `.xlsx`
  files. This avoids relying on missing example workbooks and keeps the fixture
  deterministic.
- Existing runtime evidence for this candidate:
  - `runtime/workbookcase.TestRunAppendStructuredRowsEndToEnd` creates a
    `LineItems` workbook, executes the append rows task, verifies
    `Verification.Pass`, checks operation `append_structured_rows`, checks three
    written cells, opens the output workbook, and asserts `LineItems!A3`.
  - `runtime/openlayer/requestcompiler.TestAppendRowsIntentCompilesThroughValidation`
    proves the normalized intent can compile/validate to
    `structured_row_append`.
- Recommended Phase 15 installed path:
  1. create a temp project,
  2. install the project-local skill,
  3. create the `LineItems` input workbook in the temp project,
  4. write a normalized intent JSON for append rows,
  5. execute installed `bin/sheet-ops-codex run-intent --intent-file ...`,
  6. parse JSON result and assert authoritative fields,
  7. inspect output workbook and required evidence artifacts.

Authoritative success fields for Phase 15:

- command exit code is 0;
- result JSON parses and has `entry:"run-intent"`;
- result status is executed or the Phase 13 envelope equivalent has `ok:true`;
- output workbook exists and opens;
- output workbook contains appended row `LineItems!A3 == "B002"`;
- runtime verification pass is true;
- runtime paths include evidence/report artifacts and required files exist;
- no manual checkpoint or manual override is recorded.

Verification:

- `go test ./runtime/workbookcase -run 'TestRunAppendStructuredRowsEndToEnd|TestRunUsesClosedArtifactsForSummaryAndHighlight' -count=1`: pass.
- `go test ./runtime/openlayer/requestcompiler -run 'TestAppendRowsIntentCompilesThroughValidation|TestNormalizeHeadersIntentCompilesThroughValidation' -count=1`: pass.
- `go test ./runtime/openlayer/useorchestrator -run 'TestNormalizeHeadersValidatedRequestBuildsHeaderNormalizationTaskSpec' -count=1`: pass.

Backlog self-review:

- The fixture is fast enough at runtime-package level, but installed E2E may be
  slower because it builds/installs the skill first.
- There is not yet a committed installed E2E smoke skeleton; Phase 15 should add
  the passing smoke after Phase 13 stabilizes the envelope.
- Missing checked-in workbook files are not a blocker because the test can
  generate the workbook deterministically.

Commit: `phase12/e2e: discover installed workbook smoke contract`.

### Phase 13 Result Note

Implementation outcome:

- Added a stable outer envelope to `PublicEntryResult` while preserving existing
  detailed payload fields.
- New outer fields:
  - `schema_version`
  - `ok`
  - `command`
  - `recoverable`
  - `artifacts`
  - `next_actions`
- Executed public results now expose required artifact references for:
  - `output_workbook` as `primary_success`,
  - `verification` as `success_evidence`,
  - `evidence_dir` as `audit_trail`.
- Terminal compiler stops now return a recoverable envelope with
  `compiler_decision` as `failure_context`.

Verification:

- Red evidence:
  `go test ./cmd/sheet-ops-codex -run 'Test.*PublicEntryResult.*Envelope' -count=1`
  failed because `PublicEntryResult` had no `schema_version`, `ok`, `command`,
  `recoverable`, `artifacts`, or `next_actions` fields.
- `go test ./cmd/sheet-ops-codex -run 'Test.*PublicEntryResult.*Envelope' -count=1 -v`: pass and runs both executed-success and terminal-compiler envelope tests.
- `go test ./cmd/sheet-ops-codex -count=1`: pass.
- Separate verifier: first pass found the focused regex did not cover terminal
  compiler stop; after renaming the test, re-verifier passed with no blockers.

Backlog self-review:

- The public `run-intent` entry now has a stable agent envelope without hiding
  runtime detail.
- `run-validated` remains an internal handoff surface and should be evaluated in
  Phase 14 when runtime typed producers are wired.
- Formal JSON Schema for the new envelope remains deferred to Phase 16 so schema
  files do not churn while runtime error producers are still changing.

Commit: `phase13/envelope: standardize agent execution results`.

### Phase 14 Result Note

Implementation outcome:

- Added a typed `state_root_mismatch` producer for public-entry state-root guard
  failures.
- `configureStateRootsFromStateRoot` and `validateDerivedStateRootEnv` now return
  a recoverable `state_root_mismatch` CLI error instead of a generic internal
  error when state roots conflict.
- Updated `error_contract.codes[]` so `state_root_mismatch` is `emitted` with
  producer `state_root_guard`.
- Kept runtime categories without deterministic producer tests as `reserved`:
  `invalid_json_or_schema`, `request_checkpoint`, `validation_blocked`,
  `execution_failed`, and `verification_failed`.

Verification:

- Red evidence:
  `go test ./cmd/sheet-ops-codex -run 'TestStateRootMismatchClassifiesAsConfigurationError' -count=1`
  failed because the mismatch classified as `internal_error`.
- `go test ./cmd/sheet-ops-codex -run 'TestStateRootMismatchClassifiesAsConfigurationError|Test.*Error.*|TestCapabilitiesJSONReportsSafeEntryBoundaries' -count=1`: pass.
- `go test ./cmd/sheet-ops-codex -count=1`: pass.
- `go run ./cmd/sheet-ops-codex capabilities --json`: pass; output marks
  `state_root_mismatch` as `emitted` and leaves unproven runtime categories
  `reserved`.
- Separate verifier: pass, no blockers.

Backlog self-review:

- `state_root_mismatch` is safe to promote because it has a deterministic guard
  and focused classification test.
- `request_checkpoint` is represented in the public result envelope but should
  not move to `emitted` in the CLI error taxonomy until the error/envelope split
  is explicitly designed.
- Runtime execution and verification failures still need deterministic producer
  tests before promotion.

Commit: `phase14/errors: emit state root mismatch producer`.

### Phase 15 Result Note

Implementation outcome:

- Added an installed-package E2E smoke for the bundled `sheet-ops-codex`
  launcher.
- The test now creates a temp target project, runs the public
  `install-skill.sh --project ...` entrypoint, builds the installed CLI binary,
  generates a workbook, executes installed
  `bin/sheet-ops-codex run-intent`, and parses the public JSON result.
- The smoke proves the agent-facing success contract for a real workbook append:
  `schema_version`, `ok:true`, `command:"run-intent"`, `recoverable:false`,
  required success artifacts, runtime verification pass, evidence paths, and
  output workbook mutation across `LineItems!A3:C3`.
- The normalized intent fixture includes the data-validation default surface
  required by the current internal normalization/schema pass, even though the
  selected operation is `structured_row_append`.

Verification:

- Red evidence:
  `go test ./cmd/sheet-ops-codex -run 'TestInstallSkillBundledCLIRunIntentExecutesWorkbookEndToEnd' -count=1 -v`
  initially failed with a normalized-intent schema error at
  `/add_data_validation/validation_rule/rule_type`.
- `go test ./cmd/sheet-ops-codex -run 'TestInstallSkillBundledCLIRunIntentExecutesWorkbookEndToEnd' -count=1 -v`: pass.
- `go test ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts -count=1`: pass.
- Separate verifier initially found two gaps: the smoke used the in-process root
  command instead of public `install-skill.sh`, and it only checked `A3`; both
  were closed before commit.

Backlog self-review:

- This phase locks the installed `run-intent` workbook path for one concrete
  append operation, not every public operation family.
- The data-validation default fixture shape is a useful signal that a future
  schema/normalization cleanup should avoid forcing unrelated operation defaults
  into agent-authored normalized intents.
- Phase 16 should still promote a formal result schema/golden fixture so agents
  can validate the envelope without executing a workbook.

Commit: `phase15/e2e: prove installed workbook execution`.
