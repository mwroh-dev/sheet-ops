# Agent Native CLI Six Pack Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `sheet-ops-codex` more agent-native across discovery, operation schemas, stdin payloads, evidence inspection, and installed skill guidance while preserving `sheet-ops` as the single human-facing workbook entry.

**Architecture:** Add small read-only agent-contract surfaces around the existing CLI contract rather than creating a second public workbook runner. Keep mutation paths unchanged except for accepting stdin as an equivalent normalized-intent source. Each phase has explicit evaluation evidence, a commit boundary, and a backlog regression review before moving on.

**Tech Stack:** Go 1.23, Cobra, JSON Schema draft 2020-12 contracts, existing Sheet Ops requestcompiler/runtime evidence types, Markdown skill/docs.

---

## Phase / Lane Map

### Lane A: Read-only discovery surfaces

Independent from mutation execution. Can run before Lane B/C.

- Phase A1 / Task 1: `agent-guide --json`
- Phase A2 / Task 2: operation schema/example introspection

### Lane B: Intent input ergonomics

Depends on existing preview/run-intent command behavior. Must preserve path-based compatibility.

- Phase B1 / Task 3: `--intent-file -` stdin support for `preview-request`
- Phase B2 / Task 4: `--intent-file -` stdin support for `run-intent`

### Lane C: Evidence trust summary

Independent read-only CLI surface, but should reference artifacts produced by existing runtime commands.

- Phase C1 / Task 5: `evidence-summary --json`

### Lane D: Public guidance alignment

Depends on command names finalized by A/B/C.

- Phase D1 / Task 6: update README and `skills/sheet-ops/SKILL.md`

## Subagent / Model Policy

- Simple mechanical implementation or docs review: `gpt-5.3-codex-spark`, low/medium reasoning.
- Integration-heavy implementation or contract review: `gpt-5.4`, high reasoning.
- Architecture/spec verification and final review: `gpt-5.5`, high reasoning.
- Implementation and verification roles must be separate. A verifier agent may inspect code, run tests, and report findings, but must not implement fixes in the same pass.
- Close each subagent immediately after its result is consumed.

## Backlog Regression Review Rule

At the end of every phase:

1. Record any fundamental issue, broader refactor, or uncovered design gap in the phase backlog.
2. Ask whether the backlog item invalidates the current phase result or creates immediate regression risk.
3. If yes, fix it before committing the phase.
4. If no, leave it in the plan backlog and continue.

## Task 1: Add Agent Guide Command

**Files:**
- Create: `cmd/sheet-ops-codex/agent_guide_command.go`
- Modify: `cmd/sheet-ops-codex/main.go`
- Modify: `cmd/sheet-ops-codex/cli_contract.go`
- Modify: `cmd/sheet-ops-codex/cli_schema_test.go`
- Modify: `cmd/sheet-ops-codex/cli_contract_test.go`

**Evaluation:**
- `go test ./cmd/sheet-ops-codex -run 'TestAgentGuide|TestCapabilities|TestSchema|TestCLICommandContracts'`
- Result must prove `agent-guide --json` is read-only, listed in capabilities/schema, emits ordered phases with commands and expected evidence, and refuses non-JSON mode with a JSON error envelope.

**Backlog checks:**
- If command naming conflicts with existing `schema` semantics, rename before commit.
- If guide duplicates long prose from README, reduce it to terse machine-usable steps before commit.

- [ ] Step 1: Write failing tests for `agent-guide --json`.
- [ ] Step 2: Run the focused tests and confirm they fail because the command is missing.
- [ ] Step 3: Implement `newAgentGuideCommand` with stable JSON payload.
- [ ] Step 4: Register command and CLI contract.
- [ ] Step 5: Run evaluation command and verify green.
- [ ] Step 6: Commit with `feat: add agent guide cli surface`.

## Task 2: Add Operation Schema / Example Introspection

**Files:**
- Create: `cmd/sheet-ops-codex/operation_schema_command.go`
- Modify: `cmd/sheet-ops-codex/main.go`
- Modify: `cmd/sheet-ops-codex/cli_contract.go`
- Modify: `cmd/sheet-ops-codex/cli_schema_test.go`
- Modify: `cmd/sheet-ops-codex/cli_contract_test.go`

**Evaluation:**
- `go test ./cmd/sheet-ops-codex -run 'TestOperationSchema|TestCapabilities|TestSchema|TestCLICommandContracts'`
- Result must prove agents can list supported workbook operations, inspect one operation schema, and obtain at least one concrete normalized-intent example without invoking runtime execution.

**Backlog checks:**
- If operation metadata duplicates runtime capability records in a fragile way, add a backlog item to derive it from `runtime/capabilities`.
- If examples do not validate against `agents/request-compiler/contract/normalized_intent.schema.json`, fix before commit.

- [ ] Step 1: Write failing tests for `operation list --json`, `operation schema <name> --json`, and `operation example <name> --json`.
- [ ] Step 2: Run focused tests and confirm missing command failure.
- [ ] Step 3: Implement static operation metadata for public atom operations.
- [ ] Step 4: Register command and contract.
- [ ] Step 5: Validate examples in tests against existing schema helper where practical.
- [ ] Step 6: Commit with `feat: expose operation schema introspection`.

## Task 3: Support Stdin Intent for Preview Request

**Files:**
- Modify: `cmd/sheet-ops-codex/preview_request_command.go`
- Modify: `cmd/sheet-ops-codex/preview_request_test.go`
- Modify: `contracts/cli/preview_request_result.schema.json` if input provenance requires a schema addition.

**Evaluation:**
- `go test ./cmd/sheet-ops-codex -run TestPreviewRequest`
- Result must prove `preview-request --intent-file - --json` reads normalized intent JSON from stdin, fingerprints stdin content deterministically, keeps read-only behavior, and still supports file paths exactly as before.

**Backlog checks:**
- If stdin fingerprint naming is ambiguous, add a field such as `normalized_intent_source` instead of overloading path semantics.
- If tests require temp workbooks, reuse existing fixture helpers rather than adding large binary fixtures.

- [ ] Step 1: Write failing stdin preview test.
- [ ] Step 2: Run focused preview tests and confirm failure.
- [ ] Step 3: Implement stdin read path and fingerprint handling.
- [ ] Step 4: Preserve existing file path behavior.
- [ ] Step 5: Run evaluation command.
- [ ] Step 6: Commit with `feat: accept preview intent from stdin`.

## Task 4: Support Stdin Intent for Run Intent

**Files:**
- Modify: `cmd/sheet-ops-codex/main.go`
- Modify: `cmd/sheet-ops-codex/run_intent_entry_test.go`

**Evaluation:**
- `go test ./cmd/sheet-ops-codex -run TestRunIntent`
- Result must prove `run-intent --intent-file -` accepts stdin without weakening output fingerprint binding or public entry result validation.

**Backlog checks:**
- If stdin makes default scenario IDs unstable, fix by deriving a stable fallback from input/output path plus intent hash.
- If shared stdin loading logic grows beyond two call sites, extract a helper before commit.

- [ ] Step 1: Write failing stdin run-intent test.
- [ ] Step 2: Run focused run-intent tests and confirm failure.
- [ ] Step 3: Implement shared normalized-intent source helper if Task 3 introduced one.
- [ ] Step 4: Run focused run-intent tests.
- [ ] Step 5: Run preview tests to guard shared helper behavior.
- [ ] Step 6: Commit with `feat: accept run intent from stdin`.

## Task 5: Add Evidence Summary Command

**Files:**
- Create: `cmd/sheet-ops-codex/evidence_summary_command.go`
- Modify: `cmd/sheet-ops-codex/main.go`
- Modify: `cmd/sheet-ops-codex/cli_contract.go`
- Modify: `cmd/sheet-ops-codex/cli_schema_test.go`
- Modify: `cmd/sheet-ops-codex/cli_contract_test.go`
- Create or modify focused tests in `cmd/sheet-ops-codex/evidence_summary_command_test.go`

**Evaluation:**
- `go test ./cmd/sheet-ops-codex -run 'TestEvidenceSummary|TestCapabilities|TestSchema|TestCLICommandContracts'`
- Result must prove `evidence-summary --json --evidence-dir <dir>` reports pass/fail status for `verification.json`, `execution.json`, `failure.json`, output workbook hash presence, and recommended next actions without mutating files.

**Backlog checks:**
- If evidence layouts vary across runtime paths, support the current authoritative layout and record unsupported legacy layouts as backlog.
- If command starts making success claims from weak evidence, tighten status to `unknown` or `incomplete` before commit.

- [ ] Step 1: Write failing tests with temp evidence directories.
- [ ] Step 2: Run focused tests and confirm missing command failure.
- [ ] Step 3: Implement read-only evidence summary.
- [ ] Step 4: Register command and contract.
- [ ] Step 5: Run evaluation command.
- [ ] Step 6: Commit with `feat: summarize runtime evidence for agents`.

## Task 6: Update Skill and README Agent-Native Guidance

**Files:**
- Modify: `README.md`
- Modify: `skills/sheet-ops/SKILL.md`

**Evaluation:**
- `go test ./cmd/sheet-ops-codex ./runtime/...`
- `rg -n "agent-guide|operation schema|evidence-summary|--intent-file -" README.md skills/sheet-ops/SKILL.md`
- Result must prove public docs mention the new safe discovery, stdin, and evidence-summary surfaces while preserving the single human-facing `sheet-ops` entry rule.

**Backlog checks:**
- If docs imply dry-run execution, correct the wording before commit.
- If docs invite humans to call internal handoff commands directly, correct before commit.

- [ ] Step 1: Write doc updates for CLI agent contract.
- [ ] Step 2: Add skill routing guidance for new commands.
- [ ] Step 3: Run evaluation commands.
- [ ] Step 4: Commit with `docs: document agent-native cli workflow`.

## Final Verification

- Run: `go test ./cmd/sheet-ops-codex ./runtime/...`
- Run: `git log --oneline --max-count=8` and confirm at least six task commits after the branch point.
- Run final separate verifier agent over branch diff and close it after consuming results.
- Do not mark complete unless all phase evaluations pass and backlog regression review has no immediate unresolved item.
