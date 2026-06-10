# Sheet Ops CLI Agent Contract Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Productize the existing Sheet Ops CLI surfaces as stable agent-readable contracts without changing the core Sheet Ops skill/runtime architecture.

**Architecture:** Keep `sheet-ops` as the single human-facing workbook request entry. Treat CLI commands as install, skill-owned handoff, diagnostic, and agent-contract surfaces that expose stable metadata, errors, state behavior, and verification evidence for Codex, Claude, CI, and future agents.

**Tech Stack:** Go 1.25+, Cobra, JSON schema contracts, project-local Codex skill install under `.codex/skills/sheet-ops`, existing Sheet Ops runtime packages.

---

## Branch And Commit Rules

- Work branch: `codex/cli-agent-contract`.
- Do not work directly on `main`.
- Every meaningful phase or fix must end with its own commit.
- Commit message format: `phase<N>/<lane>: <concise task summary>`.
- Keep unrelated dirty files untouched.
- If a phase exposes a deeper architectural issue, record it as backlog first, then run the phase-end self-review before deciding whether it is an immediate blocker.

## Lane Map

The work has one sequential critical lane and three dependent lanes.

- Lane A, bootstrap and inventory: Phase 0 only.
- Lane B, CLI contract core: Phases 1-4, sequential because help, schema, errors, and mutation metadata must share one command contract.
- Lane C, install/runtime proof: Phase 5, after the command contract exists.
- Lane D, docs and final consistency: Phases 6-7, after the contract and install proof stabilize.

Parallel work is allowed only when files do not overlap. Examples:

- A lower-reasoning worker may enumerate docs/help strings after command metadata is stable.
- A higher-reasoning worker must own command classification, error taxonomy, mutation semantics, and state-root safety.
- A verifier worker must be separate from the executor worker for phase-end verification. The verifier reads artifacts, command output, and tests, and reports pass/fail without making production edits.

Subagent memory rule:

- Dispatch fresh subagents per bounded task.
- Assign model tier by reasoning need: simple enumeration/docs smoke at lower tier; contract taxonomy, safety, mutation, and schema compatibility at higher tier.
- After a subagent returns, extract only decisions, file paths, commands, and evidence needed for the parent phase note.
- Do not keep subagent transcript details in active context after the task is reviewed.
- Do not allow subagents to edit the same files in parallel.

## Current Baseline

Commands checked on `codex/cli-agent-contract` before implementation:

- `go test ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts`: pass.
- `go run ./cmd/sheet-ops-codex --help`: pass.
- `go run ./cmd/sheet-ops-agent --help`: pass.

Observed command surfaces:

- `sheet-ops-codex`: `install-skill`, `prepare-use`, `run-validated`, `run-intent`, Cobra default `completion`, `help`.
- `sheet-ops-agent`: `use`, `use-open`, `use-structured`, `run-validated`, `trace`, Cobra default `completion`, `help`.

Known gaps:

- No explicit agent-readable `capabilities --json`.
- No explicit `schema --json` or `schema command <name> --json`.
- No Sheet Ops-specific command classification for `public_install`, `agent_contract`, `internal_handoff`, and `maintainer_diagnostic`.
- No stable CLI error taxonomy or JSON error shape.
- No explicit mutation/state metadata for `.sheet-ops-state`, workbook output writes, or bundled runtime state.
- Installed project-local CLI proof exists at install level, but not yet as a full agent-contract smoke tied to schema/help/error metadata.

## Phase 0 - Baseline Audit And Branch Setup

Lane: bootstrap.

### TODO

- [x] Create `codex/cli-agent-contract`.
  - Evaluation: `git branch --show-current` returns `codex/cli-agent-contract`.
  - Result: no CLI contract work happens on `main`.
  - Likely files: none.
  - Risk: low.
  - Rollback: switch to `main` and delete the branch before commits if the branch is wrong.
- [x] Run current targeted baseline tests.
  - Evaluation: current command packages and release contracts pass or failures are recorded before implementation.
  - Result: later regressions can be separated from baseline.
  - Likely files: this plan only.
  - Risk: low.
  - Rollback: no production code changes in this phase.
- [x] Inventory current CLI surfaces.
  - Evaluation: current `sheet-ops-codex` and `sheet-ops-agent` commands are listed with current gaps.
  - Result: Phase 1 classification starts from current evidence.
  - Likely files: this plan only.
  - Risk: low.
  - Rollback: update the inventory if later audit finds missing commands.

### Phase-End Backlog Review

Immediate blockers: none.

Deferred:

- Full `go test ./...` may be expensive and should be run in a later verifier phase after focused contract tests exist.
- Whether `sheet-ops-agent` should expose schema/capabilities directly or only through `sheet-ops-codex` must be decided in Phase 1 from caller boundaries.

## Phase 1 - Command Classification And Help Contract

Lane: CLI contract core.

### TODO

- [ ] Add failing tests for command classification and help coverage.
  - Evaluation: tests fail before metadata exists.
  - Result: `sheet-ops-codex` command help and command classification become contract-covered.
  - Likely files: `cmd/sheet-ops-codex/cli_contract_test.go`.
  - Risk: low.
  - Rollback: remove the new tests if the classification model changes before implementation.
- [ ] Add a focused command metadata module for `sheet-ops-codex`.
  - Evaluation: every `sheet-ops-codex` command has name, classification, intended caller, usage, options, outputs, side effects, artifacts, state behavior, safety notes, and related commands.
  - Result: help, schema, capabilities, docs, and tests can share one source of truth.
  - Likely files: `cmd/sheet-ops-codex/cli_contract.go`, `cmd/sheet-ops-codex/main.go`.
  - Risk: medium.
  - Rollback: keep command execution untouched and remove only metadata/help rendering.
- [ ] Improve command help without changing command execution.
  - Evaluation: top-level help distinguishes install, agent-contract, internal handoff, and maintainer diagnostic surfaces.
  - Result: help does not teach internal launchers as a second human-facing workbook entry.
  - Likely files: `cmd/sheet-ops-codex/main.go`, `cmd/sheet-ops-codex/cli_contract.go`.
  - Risk: medium.
  - Rollback: revert help wiring while preserving metadata for schema work if useful.
- [ ] Commit Phase 1.
  - Evaluation: focused help/classification tests pass.
  - Result: `phase1/help: classify cli command surfaces`.

### Phase-End Backlog Review

Ask:

- Did this phase reveal a hidden public workbook request path that conflicts with `sheet-ops` skill ownership?
- Is the issue necessary for the CLI agent-contract goal?
- Can it be blocked or demoted without changing runtime behavior?

Immediate blockers should be fixed in Phase 1 only if they directly expose a stale human-facing command path. Otherwise record them for Phase 6 docs consistency.

## Phase 2 - Agent-Readable Schema And Capabilities

Lane: CLI contract core.

### TODO

- [ ] Add failing tests for `capabilities --json`, `schema --json`, and `schema command <name> --json`.
  - Evaluation: tests prove agents can inspect commands without parsing prose.
  - Result: failures exist before implementation because the commands do not exist.
  - Likely files: `cmd/sheet-ops-codex/cli_schema_test.go`.
  - Risk: low.
  - Rollback: remove tests if the contract command names change before implementation.
- [ ] Implement read-only schema/capabilities commands.
  - Evaluation: the commands emit stable JSON and do not invoke workbook runtime paths.
  - Result: agents can discover command contracts safely.
  - Likely files: `cmd/sheet-ops-codex/cli_contract.go`, `cmd/sheet-ops-codex/main.go`.
  - Risk: medium.
  - Rollback: remove dispatch branches; existing workflow commands remain intact.
- [ ] Include Sheet Ops-specific metadata.
  - Evaluation: JSON includes command classification, intended caller, input schema path, output shape, read/write artifacts, state root behavior, mutation behavior, safety implications, and related commands.
  - Result: contract is useful to another agent or CI script.
  - Likely files: `cmd/sheet-ops-codex/cli_contract.go`.
  - Risk: medium.
  - Rollback: keep only fields proven by tests and defer optional metadata.
- [ ] Commit Phase 2.
  - Evaluation: schema/capabilities tests and Phase 1 tests pass.
  - Result: `phase2/schema: expose agent-readable cli contract`.

### Phase 2 Red Test Evidence (Pre-Implementation)

- Added focused Phase 2 tests in `cmd/sheet-ops-codex/cli_schema_test.go` before
  production changes.
- Red command:
  `go test ./cmd/sheet-ops-codex -run 'Test.*Schema.*|Test.*Capabilities.*|Test.*Contract.*' -count=1`
- Observed failure before implementation:
  - `unknown command "capabilities" for "sheet-ops-codex"`
  - `unknown command "schema" for "sheet-ops-codex"`

### Phase-End Backlog Review

Ask:

- Are any contract fields guesses rather than evidence from runtime/docs?
- Does any field imply hosted, global, or direct human workbook execution behavior?
- Are unsupported commands better classified as diagnostic than public?

Fields that cannot be made truthful in this phase should be removed or marked explicitly as deferred; do not ship speculative metadata.

## Phase 2 Result Note

Implementation outcome:

- Added read-only `sheet-ops-codex capabilities --json` discovery output.
- Added read-only `sheet-ops-codex schema --json` and
  `sheet-ops-codex schema command <name> --json` output.
- Output uses consistent `schema_version` keys.
- Command schema now reports command name, classification, intended caller,
  description, usage, options, output mode, side effects, read artifacts,
  written artifacts, state behavior, safety notes, related commands, and
  mutation/read-only visibility metadata.

Phase 2 schema-scope decision:

- Include the hidden `run-request` compatibility command in machine-readable
  schema, marked hidden, because it is a real Sheet Ops CLI surface with
  behavior that existing wrappers may still call.
- Exclude Cobra-generated completion leaf subcommands (`bash`, `zsh`, `fish`,
  `powershell`) from machine-readable schema because they are generated support
  details rather than stable Sheet Ops command-contract surfaces.
- Exclude the hidden `install-skill --sheet-ops-codex-bin` flag from
  machine-readable schema because it is internal install plumbing, not a stable
  public or agent-facing contract boundary.

Phase 2 backlog:

- Phase 3 should decide whether unknown schema-command lookup errors need a
  typed JSON error payload in `--json` mode or can remain plain Cobra/Go errors
  until the broader error taxonomy lands.
- Independent verifier result: `PASS_WITH_CONCERNS`.
- Verification commands passed:
  `go test ./cmd/sheet-ops-codex -run 'Test.*Schema.*|Test.*Capabilities.*|Test.*Contract.*' -count=1`,
  `go test ./cmd/sheet-ops-codex -count=1`,
  `go run ./cmd/sheet-ops-codex capabilities --json`, and
  `go run ./cmd/sheet-ops-codex schema command prepare-use --json`.
- Concern to resolve or document in Phase 3/4: `read_only_commands` currently
  includes the root command while `capabilities.commands` omits the root. This
  is not a blocker because root help/dispatch is read-only, but the contract
  should clarify root representation.
- Concern to resolve or document in Phase 3/4: command schema describes
  options and artifacts but does not yet expose first-class input schema
  references for agents that want to validate inputs directly from the CLI
  contract.

## Phase 1 Result Note

Implementation commit: `00076c4 phase1/help: classify cli command surfaces`.

Executor evidence:

- Red test was reported by the executor as:
  `go test ./cmd/sheet-ops-codex -run 'Test.*CLI.*|Test.*Help.*|Test.*Contract.*' -count=1`
  failing with missing contract symbols before implementation.
- Green verification reported and re-run by the parent:
  `go test ./cmd/sheet-ops-codex -run 'Test.*CLI.*|Test.*Help.*|Test.*Contract.*' -count=1`.
- Full command package verification reported and re-run by the parent:
  `go test ./cmd/sheet-ops-codex -count=1`.

Independent verifier: `Gauss`, closed after review to satisfy subagent memory management.

Verifier result: `PASS_WITH_CONCERNS`.

Verifier findings:

- No blocking Phase 1 failure.
- Scope is limited to `cmd/sheet-ops-codex/cli_contract.go`,
  `cmd/sheet-ops-codex/cli_contract_test.go`, and
  `cmd/sheet-ops-codex/main.go`.
- Top-level help distinguishes install, agent-contract, internal handoff,
  maintainer diagnostic, and support surfaces.
- Top-level help does not teach `sheet-ops-agent use` as a user entry and keeps
  `sheet-ops` as the workbook request entry.
- Execution behavior appears limited to help/metadata wiring.

Phase 1 backlog classification:

- TDD sequencing is not independently provable from the single implementation
  commit. Treat this as a process-evidence concern, not a code blocker, because
  the executor supplied red-test output and the parent/verifier re-ran green
  checks. For future phases, keep red-test evidence in phase notes before
  production-code commits whenever practical.
- Cobra generated completion subcommands and the hidden
  `--sheet-ops-codex-bin` install flag are not individually modeled. This is not
  a Phase 1 blocker because Phase 1 treats `completion` as one support surface,
  but Phase 2 must decide whether machine-readable schema output should expose
  generated completion subcommands and hidden install flags.

## Phase 3 - Structured Errors And Exit Codes

Lane: CLI contract core.

### TODO

- [ ] Add failing tests for invalid usage, missing required options, unknown schema command, and JSON error mode.
  - Evaluation: failures show current errors are not stable enough for agents.
  - Result: typed error contract is test-driven.
  - Likely files: `cmd/sheet-ops-codex/cli_errors_test.go`.
  - Risk: low.
  - Rollback: remove tests if taxonomy is revised.
- [ ] Define Sheet Ops CLI exit-code taxonomy.
  - Evaluation: taxonomy distinguishes invalid usage, missing option, invalid JSON/schema, install dependency failure, state-root mismatch, request checkpoint/block, validation block, execution failure, verification failure, and internal runtime error.
  - Result: agents can recover or stop based on error category.
  - Likely files: `cmd/sheet-ops-codex/cli_errors.go`, `cmd/sheet-ops-codex/cli_contract.go`.
  - Risk: high.
  - Rollback: keep Cobra default errors and defer taxonomy if it requires broad runtime rewrites.
- [ ] Add JSON error shape for `--json` where supported.
  - Evaluation: JSON failures write stable stdout payload and avoid mixing stderr into parseable output.
  - Result: `{ "ok": false, "error": { "code": "...", "message": "...", "recoverable": true, "suggested_commands": [] } }`.
  - Likely files: `cmd/sheet-ops-codex/main.go`, `cmd/sheet-ops-codex/cli_errors.go`.
  - Risk: high.
  - Rollback: restrict JSON errors to meta commands first and defer runtime command wrapping.
- [ ] Commit Phase 3.
  - Evaluation: Phase 1-3 focused tests pass.
  - Result: `phase3/errors: stabilize cli failure contract`.

### Phase 3 Red Test Evidence (Pre-Implementation)

- Added focused Phase 3 tests in `cmd/sheet-ops-codex/cli_errors_test.go`
  before production implementation.
- Initial red command:
  `go test ./cmd/sheet-ops-codex -run 'Test.*Error.*|Test.*Unknown.*|Test.*Missing.*' -count=1`.
- Red result: build failed because `classifyCLIError` was undefined, and the
  current CLI has no structured error taxonomy or JSON error envelope for
  `schema command <name> --json` failures.
- Added a second focused expectation that `capabilities --json` exposes an
  `error_contract`. Red result:
  `TestCapabilitiesJSONReportsSafeEntryBoundaries` failed with
  `json_failure_shape = ""`.

## Phase 3 Result Note

Implementation outcome:

- Added wrapper-level Sheet Ops CLI error classification with exit-code
  metadata.
- Added stable JSON failure envelope for supported `--json` meta-command
  failures: `{ "ok": false, "error": { ... } }`.
- Added `error_contract` to `capabilities --json` so agents can discover
  recoverability and exit-code semantics without parsing help prose.
- Kept successful runtime/stdout contracts unchanged.

Verification commands passed:

- `go test ./cmd/sheet-ops-codex -run 'Test.*Error.*|Test.*Unknown.*|Test.*Missing.*|Test.*Schema.*|Test.*Capabilities.*|Test.*Contract.*' -count=1`.
- `go test ./cmd/sheet-ops-codex -count=1`.
- `go run ./cmd/sheet-ops-codex capabilities --json`.
- `go run ./cmd/sheet-ops-codex schema command nope --json`.

### Phase-End Backlog Review

Ask:

- Did command execution errors need deeper runtime error typing?
- Is deeper typing required now, or can wrapper-level classification cover the public contract truthfully?
- Did JSON error mode risk changing existing stdout contracts for successful runtime commands?

Do not rewrite runtime orchestration in this phase unless a tested CLI contract cannot be made truthful at the boundary.

## Phase 4 - State, Mutation, Doctor, And Dry-Run Clarity

Lane: CLI contract core.

### TODO

- [ ] Add tests for mutation/state metadata in schema output.
  - Evaluation: mutating commands disclose workbook and `.sheet-ops-state` writes; read-only commands are not falsely mutating.
  - Result: agents can decide when to ask for confirmation or run preflight.
  - Likely files: `cmd/sheet-ops-codex/cli_schema_test.go`.
  - Risk: medium.
  - Rollback: keep metadata read-only and do not add dry-run behavior until commands are classified.
- [ ] Add or document `doctor`/`preflight` command scope.
  - Evaluation: preflight checks Go availability where relevant, bundled binary presence, schema asset availability, state-root behavior, input workbook readability, output workbook writability, and project-local install shape.
  - Result: agents can diagnose readiness before mutating workbooks.
  - Likely files: `cmd/sheet-ops-codex/preflight_command.go`, `cmd/sheet-ops-codex/preflight_command_test.go`, `cmd/sheet-ops-codex/main.go`.
  - Risk: medium-high.
  - Rollback: ship preflight as read-only diagnostic only; defer invasive runtime readiness checks.
- [ ] Decide dry-run boundaries.
  - Evaluation: each mutating command is either dry-run capable or explicitly documented as not dry-run capable with reason.
  - Result: no command silently claims preview behavior it cannot satisfy.
  - Likely files: `cmd/sheet-ops-codex/cli_contract.go`, optional command files only if a bounded dry-run is added.
  - Risk: high.
  - Rollback: prefer truthful metadata over shallow fake dry-run.
- [ ] Commit Phase 4.
  - Evaluation: mutation/state/preflight tests pass and earlier contract tests still pass.
  - Result: `phase4/safety: expose state and mutation preflight`.

### Phase-End Backlog Review

Ask:

- Would implementing dry-run require duplicating workbook execution planning?
- Can preflight prove enough without changing runtime behavior?
- Is any write path missing from metadata?

If dry-run is not truthful for workbook mutation, record it as deferred and keep the contract honest.

## Phase 5 - Installed Project-Local Contract Smoke

Lane: install/runtime proof.

### TODO

- [ ] Add an installed-surface smoke test.
  - Evaluation: temp workspace install uses repository-local `install-skill.sh --project <tmp>`.
  - Result: test uses installed `.codex/skills/sheet-ops` assets rather than source-only paths.
  - Likely files: `cmd/sheet-ops-codex/install_skill_test.go` or `internal/releasecontracts/release_contracts_test.go`.
  - Risk: medium.
  - Rollback: keep existing install tests and add a narrower proof if full workbook smoke is too slow.
- [ ] Verify installed CLI contract commands.
  - Evaluation: installed `bin/sheet-ops-codex` or wrapper can run `--help`, `capabilities --json`, and `schema command prepare-use --json`.
  - Result: contract works after materialized install.
  - Likely files: install test files.
  - Risk: medium.
  - Rollback: separate binary-build proof from workbook runtime proof.
- [ ] Verify a minimal workbook request path if feasible.
  - Evaluation: use a fixture workbook/request to produce JSON result, output workbook, and verification/evidence artifacts from installed surface.
  - Result: success is based on authoritative result fields and artifact existence, not stdout alone.
  - Likely files: install/runtime smoke test file.
  - Risk: high.
  - Rollback: document fixture/runtime blocker and keep meta-command install proof as immediate deliverable.
- [ ] Commit Phase 5.
  - Evaluation: installed contract smoke passes.
  - Result: `phase5/install: verify installed cli contract surface`.

### Phase-End Backlog Review

Ask:

- Does the installed smoke accidentally call source repo paths?
- Does it rely on machine-specific Go paths?
- Does it prove actual installed behavior or just copy shape?

Machine-specific paths and source-only invocations are blockers for this phase.

## Phase 6 - Documentation And Mirror Consistency

Lane: docs/final consistency.

### TODO

- [ ] Update public docs for the CLI agent contract.
  - Evaluation: README/docs explain CLI as an agent-contract/install/diagnostic surface, not a second human workbook request entry.
  - Result: users still invoke the `sheet-ops` skill for normal workbook requests.
  - Likely files: `README.md`, `docs/public/install.md`, `skills/sheet-ops/references/quickstart.md`, `skills/sheet-ops/references/capabilities.md`.
  - Risk: medium.
  - Rollback: revert docs only if product positioning changes.
- [ ] Update bundled skill assets where needed.
  - Evaluation: `cmd/sheet-ops-codex/skill_assets/*` mirrors canonical docs and does not teach stale internal commands.
  - Result: installed consumers get the same contract language.
  - Likely files: `cmd/sheet-ops-codex/skill_assets/SKILL.md`, `cmd/sheet-ops-codex/skill_assets/references/*`.
  - Risk: medium.
  - Rollback: restore asset mirror from canonical docs and rerun release contract tests.
- [ ] Add consistency tests.
  - Evaluation: help/schema/capabilities/docs mention the same command groups and do not expose internal handoff as public human entry.
  - Result: future command changes cannot silently drift.
  - Likely files: `internal/releasecontracts/release_contracts_test.go`, `cmd/sheet-ops-codex/*_test.go`.
  - Risk: low-medium.
  - Rollback: narrow assertions if they overfit prose.
- [ ] Commit Phase 6.
  - Evaluation: docs/mirror tests and earlier CLI tests pass.
  - Result: `phase6/docs: align cli contract documentation`.

### Phase-End Backlog Review

Ask:

- Did docs add claims not proven by tests?
- Are bundled assets exactly as authoritative as canonical docs where required?
- Does any stale `run-text`, `run-prompt`, or direct internal launcher workflow remain?

Stale public guidance is an immediate blocker.

## Phase 7 - Final Verification And Review

Lane: final review.

### TODO

- [ ] Dispatch a verifier separate from the executor.
  - Evaluation: verifier reviews branch diff, phase notes, tests, and installed smoke evidence.
  - Result: completion claim is based on independent verification, not executor intent.
  - Likely files: none unless verifier report is committed.
  - Risk: low.
  - Rollback: parent performs manual verification if no subagent tool is available, but must record that verifier separation was unavailable.
- [ ] Run focused CLI contract suite.
  - Evaluation: Phase 1-6 tests pass together.
  - Result: help, schema, capabilities, errors, state/preflight, install smoke, and docs consistency are green.
  - Likely files: none.
  - Risk: low.
  - Rollback: fix the first failing contract owner and commit separately.
- [ ] Run broader release/runtime checks.
  - Evaluation: at minimum `go test ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts`; broader `go test ./...` if runtime cost is acceptable.
  - Result: no hidden regression in existing runtime packages.
  - Likely files: none.
  - Risk: medium.
  - Rollback: classify broad failures as blocker, pre-existing baseline, or deferred repo-health issue with evidence.
- [ ] Write final review note.
  - Evaluation: note includes pass/fail matrix, backlog classification, and exact commands run.
  - Result: branch is reviewable.
  - Likely files: `docs/superpowers/plans/2026-06-10-sheet-ops-cli-agent-contract-final-review.md` or phase note under the same plan directory.
  - Risk: low.
  - Rollback: update note if evidence changes.
- [ ] Commit Phase 7.
  - Evaluation: final review evidence is committed.
  - Result: `phase7/review: lock cli agent contract evidence`.

### Final Done Criteria

- Dedicated branch exists and contains phase-mapped commits.
- Every phase has TODO evaluation/result criteria and phase-end backlog/self-review.
- `sheet-ops-codex` exposes truthful human help for contract commands.
- `sheet-ops-codex capabilities --json` works.
- `sheet-ops-codex schema command <name> --json` works for contract-relevant commands.
- JSON error contract and exit codes are documented and tested.
- Mutation and `.sheet-ops-state` behavior are visible to agents.
- Installed project-local surface can run the contract discovery commands.
- Docs and bundled assets do not present internal launchers as a second human-facing workbook request entry.
- Executor and verifier roles are separated for final verification, or lack of subagent tooling is explicitly recorded.
- Backlog is classified as immediate blocker, safe follow-up, or separate architectural task.

## Non-Goals

- Do not redesign Sheet Ops runtime execution.
- Do not make `sheet-ops-agent` a new human-facing CLI product.
- Do not bypass the `sheet-ops` skill as the normal workbook request entry.
- Do not add hosted or global service behavior.
- Do not invent dry-run behavior that cannot truthfully prove non-mutation.
- Do not rewrite workbook compiler/orchestrator logic unless a CLI contract test proves the boundary cannot be made truthful otherwise.
