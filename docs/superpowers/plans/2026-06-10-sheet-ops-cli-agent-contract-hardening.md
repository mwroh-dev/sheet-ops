# Sheet Ops CLI Agent Contract Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Lock the Sheet Ops CLI agent contract follow-up gaps with phase-scoped commits, explicit lane ownership, and verifier/executor separation.

**Architecture:** The follow-up work is one dependent `contract-hardening` lane because installed smoke, error taxonomy shape, capabilities root shape, and final evidence docs build on the same public JSON contract. Implementation runs sequentially by phase; each phase ends with a commit, a verifier that is not the executor, and a backlog self-review before moving forward. Parallel subagents are reserved for independent verification/review work only, with lower-tier models for mechanical checks and higher-tier models for broad contract judgment.

**Tech Stack:** Go 1.25+, Cobra, JSON CLI contracts, `go test`, project-local installed skill smoke tests, public docs under `docs/superpowers/plans`.

---

## Lanes

### Lane A - Contract Hardening

Sequential phase lane. These phases share the same JSON contract and must not be parallelized as independent code edits:

1. Phase 8: installed contract smoke tightening.
2. Phase 9: error taxonomy producer semantics.
3. Phase 10: capabilities root surface shape.
4. Phase 11: final evidence review matrix.

### Lane B - Independent Verification

Verifier lane. Each phase gets a separate verifier from the executor after implementation and focused tests. Verifiers must inspect the diff, run or request evidence for the phase-specific command, and report pass/fail plus any blocker. Close the verifier subagent immediately after its result is consumed.

Model tier guidance:

- Mechanical smoke/doc mirror checks: `gpt-5.4-mini` or equivalent small/fast model.
- Error taxonomy and root contract semantics: standard model or higher, because these require contract judgment.
- Final cross-phase review: highest available reasoning tier appropriate for broad contract review.

## Phase 8 - Installed Contract Smoke Tightening

Lane: Contract Hardening.

### TODO

- [ ] Add failing installed smoke assertion for `schema command prepare-use --json`.
  - Evaluation: installed `.codex/skills/sheet-ops/bin/sheet-ops-codex` can emit the `prepare-use` schema from the materialized install tree.
  - Result: installed smoke checks both discovery (`capabilities`) and the contract-critical typed handoff surface (`prepare-use`), not only `preflight`.
  - Likely files: `cmd/sheet-ops-codex/install_skill_test.go`.
  - Risk: low.
  - Rollback: keep existing installed smoke and document why `prepare-use` cannot be invoked from installed binary.
- [ ] Verify red.
  - Evaluation: focused test fails before implementation because installed smoke does not yet assert `prepare-use`.
  - Result: test demonstrates the missing coverage gap.
- [ ] Implement the minimal installed smoke coverage.
  - Evaluation: assertion validates `prepare-use` name, `agent_contract` classification, `mutating=true`, `read_only=false`, and `--envelope-file` option presence.
  - Result: installed smoke proves the agent handoff schema exists after install.
- [ ] Run focused and package tests.
  - Evaluation: `go test ./cmd/sheet-ops-codex -run 'TestInstallSkillBundledCLIExposesAgentContract' -count=1` and `go test ./cmd/sheet-ops-codex -count=1` pass.
  - Result: Phase 8 implementation is locally green.
- [ ] Run separate verifier.
  - Evaluation: verifier is not the executor, checks the diff and focused command evidence, and reports pass/fail.
  - Result: verification is separated before commit.
- [ ] Commit Phase 8.
  - Evaluation: only Phase 8 files are staged.
  - Result: `phase8/install-smoke: cover prepare-use installed contract`.

### Phase-End Backlog Review

Ask:

- Does installed smoke still call only installed binaries, not source repo paths?
- Is a full workbook execution smoke now necessary, or still a separate high-risk fixture task?
- Does this phase expose any install/runtime issue that blocks the agent contract?

Only proceed to workbook execution smoke if the installed meta-contract cannot prove the CLI boundary without it.

## Phase 9 - Error Taxonomy Producer Semantics

Lane: Contract Hardening.

### TODO

- [ ] Add failing tests for emitted/reserved error metadata.
  - Evaluation: every `error_contract.codes[]` item exposes `status` and `producer`.
  - Result: agents can distinguish currently emitted errors from reserved runtime categories.
  - Likely files: `cmd/sheet-ops-codex/cli_schema_test.go`, `cmd/sheet-ops-codex/cli_errors.go`.
  - Risk: medium.
  - Rollback: keep taxonomy flat but document why reserved status is not machine-readable.
- [ ] Define status and producer fields.
  - Evaluation: emitted codes use `status:"emitted"` and a concrete producer such as `cobra_usage`, `schema_lookup`, `install_dependency`, or `fallback_classifier`.
  - Result: reserved runtime codes use `status:"reserved"` and a producer namespace describing the future owner.
- [ ] Update tests for actual error samples.
  - Evaluation: invalid usage, missing required option, unknown schema command, and fallback/internal error remain covered.
  - Result: producer semantics do not weaken existing error behavior.
- [ ] Run focused error/schema tests.
  - Evaluation: `go test ./cmd/sheet-ops-codex -run 'Test.*Error.*|TestCapabilitiesJSONReportsSafeEntryBoundaries' -count=1` passes.
  - Result: Phase 9 implementation is locally green.
- [ ] Run separate verifier.
  - Evaluation: verifier checks the JSON shape and whether reserved categories are no longer ambiguous.
  - Result: verification is separated before commit.
- [ ] Commit Phase 9.
  - Evaluation: only Phase 9 files are staged.
  - Result: `phase9/errors: mark emitted and reserved taxonomy`.

### Phase-End Backlog Review

Ask:

- Should any reserved runtime code be wired to a typed producer now?
- Does adding producer metadata reveal a real missing error classification that should block this phase?
- Does the taxonomy claim more runtime certainty than tests prove?

Promote a reserved code to emitted only when a focused test proves the actual producer path.

## Phase 10 - Capabilities Root Surface Shape

Lane: Contract Hardening.

### TODO

- [ ] Add failing tests for root command shape.
  - Evaluation: `capabilities --json` exposes `root_command`, and `read_only_commands` contains only executable subcommands.
  - Result: root metadata is not mixed into subcommand discovery lists.
  - Likely files: `cmd/sheet-ops-codex/cli_schema_test.go`, `cmd/sheet-ops-codex/cli_contract.go`.
  - Risk: medium.
  - Rollback: retain current shape and add explicit documentation that `read_only_commands` includes the root command.
- [ ] Implement `root_command`.
  - Evaluation: `root_command.name == cli_name`, `root_command.classification == builtin_support`, `root_command.read_only == true`, `root_command.mutating == false`, and `root_command.hidden == false`.
  - Result: agents can inspect root metadata without treating the root as a subcommand.
- [ ] Remove root from `read_only_commands`.
  - Evaluation: `read_only_commands` includes `capabilities`, `completion`, `help`, `preflight`, and `schema`, but not `sheet-ops-codex`.
  - Result: subcommand lists are symmetric with `commands`.
- [ ] Run focused schema/capabilities tests.
  - Evaluation: `go test ./cmd/sheet-ops-codex -run 'TestCapabilitiesJSONReportsSafeEntryBoundaries|TestSchemaJSONReportsCommandContracts' -count=1` passes.
  - Result: Phase 10 implementation is locally green.
- [ ] Run separate verifier.
  - Evaluation: verifier checks the capability JSON and confirms the root/subcommand split is intentional.
  - Result: verification is separated before commit.
- [ ] Commit Phase 10.
  - Evaluation: only Phase 10 files are staged.
  - Result: `phase10/capabilities: separate root command metadata`.

### Phase-End Backlog Review

Ask:

- Does this break any documented consumer expectation?
- Should `schema --json` also expose root metadata differently, or is root already acceptable in schema command metadata?
- Does `machine_entry_commands` need any change after root separation?

Only change schema root behavior if a concrete consumer ambiguity appears.

## Phase 11 - Final Evidence Review Matrix

Lane: Contract Hardening.

### TODO

- [ ] Update final review artifact with a browser-flow-style evidence matrix.
  - Evaluation: document lists pass/fail evidence for installed smoke, source CLI schema, error emitted/reserved split, root surface shape, dry-run deferral, and workbook execution smoke deferral.
  - Result: final status no longer relies only on command lists.
  - Likely files: `docs/superpowers/plans/2026-06-10-sheet-ops-cli-agent-contract-final-review.md`, this hardening plan.
  - Risk: low.
  - Rollback: keep final review concise if matrix duplicates committed plan content without adding proof.
- [ ] Add unsupported assumptions.
  - Evaluation: final review names assumptions that are not proven, including full workbook execution smoke and real dry-run.
  - Result: no unsupported claim is hidden inside a pass statement.
- [ ] Add allowed/rejected strategy notes.
  - Evaluation: final review distinguishes acceptable follow-ups from rejected shortcuts, especially fake dry-run and treating internal launchers as public human entries.
  - Result: future agents have review guardrails.
- [ ] Run full verification.
  - Evaluation: `go test ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts -count=1`, `go test ./... -count=1`, `go run ./cmd/sheet-ops-codex capabilities --json`, `go run ./cmd/sheet-ops-codex schema command prepare-use --json`, and `go run ./cmd/sheet-ops-codex preflight --json` pass.
  - Result: final evidence is fresh.
- [ ] Run final separate verifier.
  - Evaluation: verifier reviews the full hardening diff and final evidence matrix.
  - Result: final verification is separated from execution before commit.
- [ ] Commit Phase 11.
  - Evaluation: final review and plan notes are staged.
  - Result: `phase11/review: lock hardening evidence matrix`.

### Phase-End Backlog Review

Ask:

- Are any backlog items immediate blockers rather than safe follow-ups?
- Is every success claim tied to a command, JSON field, test, or committed artifact?
- Are any browser-flow verification lessons missing from the final matrix?

Do not mark the goal complete until the completion audit proves every explicit objective requirement.

## Execution Result Notes

### Phase 8 Result Note

Implementation outcome:

- Extended installed bundled CLI smoke so it runs installed
  `.codex/skills/sheet-ops/bin/sheet-ops-codex schema command prepare-use --json`.
- The installed smoke now verifies `prepare-use` name, `agent_contract`
  classification, `mutating:true`, `read_only:false`, and `--envelope-file`.
- Kept the smoke scoped to installed meta-contract surfaces; full workbook
  execution remains a separate high-risk fixture task.

Verification:

- Red evidence: `rg -n "prepare-use" cmd/sheet-ops-codex/install_skill_test.go`
  exited 1 before the test was added.
- `go test ./cmd/sheet-ops-codex -run 'TestInstallSkillBundledCLIExposesAgentContract' -count=1`: pass.
- `go test ./cmd/sheet-ops-codex -count=1`: pass.
- Separate verifier: pass, no blockers.

Commit: `phase8/install-smoke: cover prepare-use installed contract`.

### Phase 9 Result Note

Implementation outcome:

- Added `status` and `producer` to each `error_contract.codes[]` descriptor.
- Marked currently observed producers as `emitted`: `invalid_usage`,
  `missing_required_option`, `unknown_command`, and `internal_error`.
- Marked runtime/future categories as `reserved` until typed producers and
  focused tests exist.

Verification:

- Red evidence:
  `go test ./cmd/sheet-ops-codex -run 'TestCapabilitiesJSONReportsSafeEntryBoundaries' -count=1`
  failed because `invalid_usage` had empty `status`.
- `go test ./cmd/sheet-ops-codex -run 'Test.*Error.*|TestCapabilitiesJSONReportsSafeEntryBoundaries' -count=1`: pass.
- `go test ./cmd/sheet-ops-codex -count=1`: pass.
- Separate verifier: pass, no blockers.

Commit: `phase9/errors: mark emitted and reserved taxonomy`.

### Phase 10 Result Note

Implementation outcome:

- Added `root_command` to `capabilities --json`.
- Removed root `sheet-ops-codex` from `read_only_commands` so that list contains
  executable read-only subcommands only.
- Kept `commands` as subcommands only.

Verification:

- Red evidence:
  `go test ./cmd/sheet-ops-codex -run 'TestCapabilitiesJSONReportsSafeEntryBoundaries' -count=1`
  failed because `root_command.name` was empty.
- `go test ./cmd/sheet-ops-codex -run 'TestCapabilitiesJSONReportsSafeEntryBoundaries|TestSchemaJSONReportsCommandContracts' -count=1`: pass.
- `go test ./cmd/sheet-ops-codex -count=1`: pass.
- `go run ./cmd/sheet-ops-codex capabilities --json`: pass; output includes
  `root_command` and excludes `sheet-ops-codex` from `read_only_commands`.
- Separate verifier: pass, no blockers.

Commit: `phase10/capabilities: separate root command metadata`.

### Phase 11 Result Note

Implementation outcome:

- Updated the final review artifact with a browser-flow-style evidence matrix.
- Added unsupported assumptions, allowed strategy, and rejected strategy notes.
- Recorded hardening result notes in this plan to preserve phase/lane evidence.

Verification:

- `go test ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts -count=1`: pass.
- `go run ./cmd/sheet-ops-codex capabilities --json`: pass.
- `go run ./cmd/sheet-ops-codex schema command prepare-use --json`: pass.
- `go run ./cmd/sheet-ops-codex preflight --json`: pass.
- `go test ./... -count=1`: pass.
- Separate final verifier: pass, no blockers.
