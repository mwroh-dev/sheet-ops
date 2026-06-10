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
- Dry-run semantics: [Kubernetes dry-run KEP](https://github.com/kubernetes/enhancements/blob/master/keps/sig-api-machinery/576-dry-run/README.md) frames dry-run as running normal admission/validation without persistence, which is why Sheet Ops must not call preview a dry-run until runtime execution planning is also non-persisting.
- Plan automation: [Terraform plan command reference](https://developer.hashicorp.com/terraform/cli/commands/plan) keeps machine-readable plan output and detailed exit-code semantics explicit for automation callers.

## Lanes

### Lane A - Execution Contract

Sequential implementation lane:

1. Phase 12: installed end-to-end workbook smoke discovery.
2. Phase 13: common result envelope contract.
3. Phase 14: runtime error typed producers.
4. Phase 15: installed end-to-end workbook smoke promotion.
5. Phase 17: non-mutating plan/impact preview.
6. Phase 19: preview trust metadata and fingerprints.
7. Phase 20: invalid data typed producer promotion.
8. Phase 21: preview-run input identity sealing.
9. Phase 22: output workbook identity sealing.
10. Phase 23: verification artifact output identity sealing.
11. Phase 24: successful verification hash requirement.
12. Phase 25: verification output-file hash requirement.
13. Phase 26: executed failure public fingerprint contract.
14. Phase 27: executed failure artifact role contract.
15. Phase 28: failed run-intent public envelope emission.
16. Phase 29: run-validated internal handoff envelope.
17. Phase 30: installed run-validated handoff smoke.
18. Phase 31: internal handoff result schema.
19. Phase 32: final review refresh after handoff hardening.
20. Phase 33: hidden run-request alias proof.
21. Phase 34: installed run-validated failure-path smoke.

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

- [x] Identify the smallest stable workbook/request fixture.
  - Evaluation: fixture can run through installed skill surface without network, manual intervention, or long runtime.
  - Result: a candidate fixture path, request path, expected output workbook path, and expected evidence artifacts are documented.
  - Likely files: `internal/testfixtures/**`, `examples/**`, runtime package tests.
  - Risk: medium.
  - Rollback: document fixture blocker and stop before adding flaky smoke.
- [x] Add a skipped or failing E2E smoke skeleton.
  - Evaluation: test installs the package into a temp project and attempts to execute the installed surface only.
  - Result: red evidence shows the missing authoritative success contract or missing stable fixture.
  - Likely files: `cmd/sheet-ops-codex/install_e2e_test.go` or `cmd/sheet-ops-codex/install_skill_test.go`.
  - Risk: high.
  - Rollback: keep this phase as discovery-only if runtime invocation is too expensive.
- [x] Define authoritative success fields.
  - Evaluation: success cannot be based on stdout alone; it must require exit code, result JSON, output workbook existence, and evidence artifact existence.
  - Result: phase note lists the exact fields/files Phase 15 must assert.
- [x] Run separate verifier.
  - Evaluation: verifier checks fixture stability and whether the proposed success fields are authoritative.
  - Result: pass/fail before commit.
- [x] Commit Phase 12.
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

- [x] Add failing tests for a shared execution envelope shape.
  - Evaluation: mutating/diagnostic execution commands return JSON with stable outer fields: `schema_version`, `ok`, `status`, `command`, `artifacts`, `next_actions`, `recoverable`, and optional `error`.
  - Result: tests fail on commands whose success or blocked output is currently inconsistent.
  - Likely files: `cmd/sheet-ops-codex/*_test.go`, runtime result helpers.
  - Risk: high.
  - Rollback: start with one command and document commands not yet migrated.
- [x] Implement minimal envelope adapter.
  - Evaluation: adapter wraps existing outputs without hiding existing detailed payloads.
  - Result: commands keep useful domain data while adding a predictable agent-facing outer contract.
- [x] Define artifact references.
  - Evaluation: `artifacts` entries include kind, path, required flag, and success role.
  - Result: agent can decide which files prove success.
- [x] Preserve exit-code semantics.
  - Evaluation: success returns 0; invalid usage/data/runtime failures return the existing non-zero codes.
  - Result: machine-readable stdout and exit code agree.
- [x] Run separate verifier.
  - Evaluation: verifier checks that the envelope helps an agent recover and does not overclaim runtime success.
  - Result: pass/fail before commit.
- [x] Commit Phase 13.
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

- [x] Add failing producer tests for reserved runtime categories.
  - Evaluation: at least `state_root_mismatch`, `validation_blocked`, `request_checkpoint`, `execution_failed`, and `verification_failed` have explicit producer tests or remain documented as reserved with a blocker.
  - Result: every emitted runtime error has a live path and JSON envelope.
  - Likely files: `cmd/sheet-ops-codex/cli_errors_test.go`, runtime command tests.
  - Risk: high.
  - Rollback: promote only categories with deterministic producers.
- [x] Wire typed errors at runtime boundaries.
  - Evaluation: runtime failures map to stable `error.code`, `recoverable`, `suggested_commands`, and artifact references.
  - Result: agents can distinguish retryable repair from hard failure.
- [x] Update `error_contract.codes[]`.
  - Evaluation: statuses move from `reserved` to `emitted` only when producer tests pass.
  - Result: taxonomy stays honest.
- [x] Run separate verifier.
  - Evaluation: verifier confirms no reserved category was promoted without evidence.
  - Result: pass/fail before commit.
- [x] Commit Phase 14.
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

- [x] Promote the Phase 12 smoke to a passing test.
  - Evaluation: installed surface executes a minimal workbook request and exits 0.
  - Result: test proves installed workbook execution, not only metadata discovery.
- [x] Assert authoritative artifacts.
  - Evaluation: result envelope `ok:true`, output workbook exists, evidence artifacts exist, and verifier/result artifact reports success.
  - Result: success claim is based on files and fields.
- [x] Assert no manual intervention.
  - Evaluation: smoke output explicitly records no manual checkpoint or manual override.
  - Result: test cannot be mistaken for a manually assisted success.
- [x] Run separate verifier.
  - Evaluation: verifier checks that stdout alone is not used as proof.
  - Result: pass/fail before commit.
- [x] Commit Phase 15.
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

- [x] Add JSON Schema files for public CLI outputs.
  - Evaluation: schemas exist for capabilities, command schema, error contract, result envelope, preflight, and installed E2E result.
  - Result: external agents can validate outputs without reading Go structs.
  - Likely files: `contracts/cli/*.schema.json`, `cmd/sheet-ops-codex/*_test.go`.
  - Risk: medium.
  - Rollback: start with capabilities and result envelope only.
- [x] Add golden JSON fixtures.
  - Evaluation: representative command outputs are captured with stable fields and validated against schemas.
  - Result: compatibility drift is visible in review.
- [x] Add compatibility policy tests.
  - Evaluation: `schema_version` changes are required for breaking field removals/renames; additive optional fields are allowed.
  - Result: tests encode backward compatibility expectations.
- [x] Add docs for schema versioning.
  - Evaluation: docs define when to bump `schema_version` and how agents should handle unknown fields.
  - Result: consumers know what is stable.
- [x] Run separate verifier.
  - Evaluation: verifier checks schemas use JSON Schema 2020-12 and validate generated outputs.
  - Result: pass/fail before commit.
- [x] Commit Phase 16.
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

- [x] Add failing tests for a read-only preview command.
  - Evaluation: command returns planned reads/writes/state root/artifacts without creating output workbook or mutation artifacts.
  - Result: agents can inspect impact before mutation.
  - Likely command names: `plan-request`, `inspect-request`, or `preview-request`.
  - Risk: high.
  - Rollback: document why existing runtime cannot preview without mutation.
- [x] Implement the narrowest truthful preview.
  - Evaluation: preview may compile/validate request metadata but must not write workbook outputs.
  - Result: it is not called dry-run and does not claim execution success.
- [x] Expose preview in capabilities/schema.
  - Evaluation: command is `read_only:true`, `mutating:false`, `dry_run_capable:false`.
  - Result: agents can discover it safely.
- [x] Run separate verifier.
  - Evaluation: verifier checks no write side effects and no fake dry-run claim.
  - Result: pass/fail before commit.
- [x] Commit Phase 17.
  - Evaluation: focused preview tests pass.
  - Result: `phase17/preview: add truthful impact inspection`.

### Phase-End Backlog Review

Ask:

- Does preview require enough compiler work that it should be split into a separate architectural track?
- Does it expose sensitive workbook paths or data in a way agents should redact?
- Is a real dry-run now feasible, or still a later runtime planning mode?

Keep real dry-run deferred unless the runtime can prove non-mutation.

## Phase 19 - Preview Trust Metadata And Fingerprints

Lane: Execution Contract.

Web-search value: medium. Kubernetes dry-run and Terraform plan guidance both
confirm that a truthful preview must expose what was validated, what would
mutate, and where the plan is weaker than execution. MCP structured-output
guidance confirms these fields belong in the JSON/schema contract, not
human-only text.

### TODO

- [x] Add failing tests for preview trust metadata.
  - Evaluation: `preview-request` output exposes `planner`,
    `plan_confidence`, `operation`, `would_mutate`, `mutation_summary`, and
    input fingerprints.
  - Result: agents can distinguish compiler-validated preview from runtime
    dry-run and can compare request/workbook inputs across preview and run.
  - Likely files: `cmd/sheet-ops-codex/preview_request_test.go`,
    `contracts/cli/preview_request_result.schema.json`.
  - Risk: medium.
  - Rollback: keep Phase 17 output and document the missing trust metadata.
- [x] Implement minimal read-only metadata.
  - Evaluation: implementation runs non-persisting request-compiler validation
    over normalized intent and input workbook facts; it does not create
    `.sheet-ops-state`, output workbooks, or runtime artifacts.
  - Result: preview reports `planner:"requestcompiler_validate_intent"` and
    `plan_confidence:"compiler_validated_boundary"` rather than claiming dry-run equivalence.
- [x] Update schema and public docs.
  - Evaluation: live preview output validates against the published schema, and
    docs tell agents to treat the fingerprints as input identity evidence only.
  - Result: contract consumers see the same fields in JSON, schema, and docs.
- [x] Run separate verifier.
  - Evaluation: verifier checks no dry-run overclaim, no write side effects, and
    schema/docs alignment.
  - Result: pass/fail before commit.
- [x] Commit Phase 19.
  - Evaluation: focused preview/schema/release tests pass.
  - Result: `phase19/preview: expose trust metadata`.

### Phase-End Backlog Review

Ask:

- Is `compiler_validated_boundary` strong enough for agent execution decisions,
  or should the next phase add a compiler-backed planner artifact?
- Should fingerprints later be repeated in `run-intent` results so agents can
  prove preview/run input identity?
- Does request-compiler validation expose enough plan detail, or should preview
  include structural signals and validation notes next?

Do not set `dry_run:true` or `dry_run_capable:true` until the runtime can prove
normal validation/planning ran without persistence.

## Phase 20 - Invalid Data Typed Producer Promotion

Lane: Execution Contract.

Web-search value: low. Phase 14 already captured MCP validation-error guidance:
agent-correctable input/schema failures should be visible as tool execution
errors rather than generic internal failures. This phase applies that guidance
to a deterministic local producer.

### TODO

- [x] Add failing producer tests for malformed or schema-invalid normalized
  intent input.
  - Evaluation: a public preview entry that receives invalid normalized intent
    JSON is classified as `invalid_json_or_schema`, recoverable, exit 65, and
    gives schema/preview repair commands.
  - Result: invalid input can be corrected by an agent without treating it as
    internal CLI failure.
  - Likely files: `cmd/sheet-ops-codex/cli_errors_test.go`,
    `cmd/sheet-ops-codex/preview_request_test.go`.
  - Risk: low.
  - Rollback: keep the code reserved and document why the producer is not
    deterministic.
- [x] Implement typed invalid-data producer.
  - Evaluation: schema/JSON load errors from normalized intent boundaries map
    to `newCLIError(cliErrorInvalidData, ...)`.
  - Result: generated JSON error envelopes and `error_contract.codes[]` agree.
- [x] Promote `invalid_json_or_schema` to emitted.
  - Evaluation: `capabilities --json` marks the code emitted only after producer
    tests pass.
  - Result: taxonomy no longer reserves a producer that now has live evidence.
- [x] Run separate verifier.
  - Evaluation: verifier confirms the promoted code has a deterministic
    producer and does not swallow unrelated internal errors.
  - Result: pass/fail before commit.
- [x] Commit Phase 20.
  - Evaluation: focused error/preview/capabilities tests pass.
  - Result: `phase20/errors: emit invalid data producer`.

### Phase-End Backlog Review

Ask:

- Should request checkpoint and validation blocked be separate JSON error
  producers or remain success-like preview decisions?
- Do typed input errors need artifact references, or are suggested commands
  enough for this phase?
- Does any producer leak absolute paths in error messages that should be
  normalized later?

Keep execution/verification runtime errors reserved until deterministic failing
fixtures exist for those paths.

## Phase 21 - Preview-Run Input Identity Sealing

Lane: Execution Contract.

Web-search value: low. Phase 19 already used plan/dry-run automation guidance.
This phase applies the local backlog item: preview fingerprints are only useful
for agent workflow decisions if mutating execution repeats the same input
identity fields.

### TODO

- [x] Add failing installed/source workflow test for preview-run fingerprint
  identity.
  - Evaluation: `preview-request` and subsequent `run-intent` expose matching
    normalized intent and input workbook SHA-256 fingerprints.
  - Result: an agent can prove the run used the same request/workbook that it
    previewed before trusting execution artifacts.
  - Likely files: `cmd/sheet-ops-codex/public_entry_result_test.go`,
    `cmd/sheet-ops-codex/install_e2e_test.go`.
  - Risk: medium.
  - Rollback: keep preview fingerprints advisory-only.
- [x] Add execution fingerprints to public run results.
  - Evaluation: executed `run-intent` envelopes include
    `fingerprints.normalized_intent_sha256` and
    `fingerprints.input_workbook_sha256`.
  - Result: public execution result can be compared with preview output.
- [x] Update public result schema/golden.
  - Evaluation: schema requires fingerprints on executed success and golden
    fixture validates.
  - Result: compatibility drift is visible in schema tests.
- [x] Run separate verifier.
  - Evaluation: verifier checks the fingerprints are derived from actual input
    files, not invented constants, and no success claim relies only on stdout.
  - Result: pass/fail before commit.
- [x] Commit Phase 21.
  - Evaluation: focused public result, installed E2E, schema, and release
    contract tests pass.
  - Result: `phase21/results: seal preview run identity`.

### Phase-End Backlog Review

Ask:

- Should output workbook fingerprints be added after execution to prove the
  produced workbook identity?
- Should preflight also emit input fingerprints, or is preview the right
  identity checkpoint?
- Do fingerprints need redaction or hashing policy docs before non-local use?
- Should the normalized intent fingerprint be computed from the same byte slice
  that is parsed, rather than re-reading the file, to eliminate narrow
  concurrent file mutation ambiguity?

Do not claim full workflow matrix coverage from this phase; it seals one
identity invariant for the existing append-rows installed workflow.

### Phase 21 Result Note

Implementation outcome:

- Executed public entry results now include
  `fingerprints.normalized_intent_sha256` and
  `fingerprints.input_workbook_sha256`.
- `preview-request` and `run-intent` share the same file-hash helper so the
  two command outputs can be compared directly by an agent.
- The installed E2E workflow now runs installed `preview-request` before
  installed `run-intent` and asserts both fingerprint fields match.
- `contracts/results/public_entry_result.schema.json` requires fingerprints
  for `status:"executed"` results while keeping terminal compiler results
  free of runtime identity fields.
- Public docs and installed skill references tell agents to compare
  preview/run fingerprints before trusting runtime artifacts, without treating
  fingerprints as proof of runtime success.

Verification:

- Red evidence:
  `go test ./cmd/sheet-ops-codex -run 'TestInstallSkillBundledCLIRunIntentExecutesWorkbookEndToEnd|TestExecutedPublicEntryResultSchemaRejectsMissingFingerprints' -count=1 -v`
  failed because `run-intent` emitted empty fingerprints and the public result
  schema still accepted executed envelopes without fingerprints.
- `go test ./cmd/sheet-ops-codex -run 'TestExecutedPublicEntryResult.*Fingerprint|TestInstallSkillBundledCLIRunIntentExecutesWorkbookEndToEnd|TestPublicEntryResultGoldenValidatesAgainstPublicSchema|Test.*PublicEntryResult.*Schema|TestCLIJSONOutputsValidateAgainstPublishedSchemas|TestPreviewRequest' -count=1 -v`: pass.
- `go test ./internal/releasecontracts -run 'TestReleaseSchemasCompile|TestReleaseSchemaReferencesAreLocallyResolvable|TestCLIAgentContractDocsStayAligned' -count=1 -v`: pass.
- `go test ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts -count=1`: pass.
- `git diff --check`: pass.
- Separate verifier: no blockers. Non-blocking risk is a narrow TOCTOU gap
  where `run-intent` loads normalized intent before hashing the file, so a
  concurrent file mutation could make the emitted fingerprint describe bytes
  loaded after parsing.

Backlog self-review:

- Add output workbook fingerprinting after execution if agents need to prove
  produced workbook identity, not only input identity.
- Consider computing normalized intent fingerprint from the same byte slice
  parsed by the loader to remove the narrow concurrent mutation ambiguity.
- This phase proves the installed append-rows preview/run identity invariant;
  it does not claim full workflow matrix coverage.

## Phase 22 - Output Workbook Identity Sealing

Lane: Execution Contract.

Web-search value: low. This phase extends the existing local SHA-256 identity
contract from inputs to the produced workbook. No new external protocol or
time-sensitive best practice is needed before design.

### TODO

- [x] Add failing public result and installed workflow tests for output
  workbook fingerprints.
  - Evaluation: executed `run-intent` envelopes expose
    `fingerprints.output_workbook_sha256`, and the installed E2E verifies it
    matches the bytes of the output workbook on disk.
  - Result: an agent can bind `ok:true` and
    `runtime.verification.output_file` to a concrete produced workbook
    identity.
  - Likely files: `cmd/sheet-ops-codex/public_entry_result_test.go`,
    `cmd/sheet-ops-codex/install_e2e_test.go`.
  - Risk: medium.
  - Rollback: keep Phase 21 input identity only and document output identity as
    unresolved.
- [x] Split execution fingerprints from preview fingerprints.
  - Evaluation: preview output remains unchanged with only input fingerprints,
    while executed public results add output identity.
  - Result: preview does not claim a produced workbook hash before execution.
- [x] Update public result schema/golden/docs.
  - Evaluation: `status:"executed"` result schema requires
    `output_workbook_sha256`; preview schema remains unchanged.
  - Result: schema drift catches missing output identity.
- [x] Run separate verifier.
  - Evaluation: verifier checks output hash is computed after execution from
    the actual output file and that docs avoid claiming output quality from a
    hash alone.
  - Result: pass/fail before commit.
- [x] Commit Phase 22.
  - Evaluation: focused public result, installed E2E, preview schema, release
    contract, and package tests pass.
  - Result: `phase22/results: seal output identity`.

### Phase-End Backlog Review

Ask:

- Should verification artifacts also record the output workbook hash, so the
  public envelope and verification report can cross-check each other?
- Should failed-but-runtime-started executions emit a typed public failure
  envelope with partial fingerprints?
- Should output identity be promoted into a separate `artifacts[].sha256`
  field later, or is the root `fingerprints` object sufficient for agents?
- Should output hashing be moved into verification/runtime evidence to remove
  the narrow post-verification mutation window?
- If terminal compiler results later need input-only fingerprints, split the
  public result schema into terminal and execution fingerprint definitions
  rather than weakening executed-result requirements.

Do not treat output hash equality as proof of workbook semantic correctness.
Semantic success remains `runtime.verification.pass` plus required artifacts.

### Phase 22 Result Note

Implementation outcome:

- Executed public entry results now expose
  `fingerprints.output_workbook_sha256` in addition to normalized intent and
  input workbook fingerprints.
- `preview-request` remains input-only; it does not emit an output workbook
  fingerprint before execution.
- `run-intent` computes output identity after runtime execution from
  `runtime.verification.output_file`.
- Installed E2E now hashes the output workbook on disk and compares it to the
  public execution result fingerprint.
- `contracts/results/public_entry_result.schema.json` requires
  `output_workbook_sha256` for executed results through an
  `execution_fingerprints` definition, while terminal compiler results still
  omit fingerprints.
- Public docs and skill references describe output fingerprints as byte
  identity, not proof of semantic workbook correctness.

Verification:

- Red evidence:
  `go test ./cmd/sheet-ops-codex -run 'TestInstallSkillBundledCLIRunIntentExecutesWorkbookEndToEnd|TestExecutedPublicEntryResultExposesInputFingerprints|TestExecutedPublicEntryResultSchemaRejectsMissingOutputFingerprint' -count=1 -v`
  failed because `run-intent` emitted no output workbook fingerprint and the
  public result schema accepted executed results without one.
- `go test ./cmd/sheet-ops-codex -run 'TestInstallSkillBundledCLIRunIntentExecutesWorkbookEndToEnd|TestExecutedPublicEntryResultExposesInputFingerprints|TestExecutedPublicEntryResultSchemaRejectsMissingOutputFingerprint|TestExecutedPublicEntryResultValidatesAgainstPublicSchema|TestPublicEntryResultGoldenValidatesAgainstPublicSchema' -count=1 -v`: pass.
- `go test ./cmd/sheet-ops-codex -run 'TestExecutedPublicEntryResult.*Fingerprint|TestInstallSkillBundledCLIRunIntentExecutesWorkbookEndToEnd|TestPublicEntryResultGoldenValidatesAgainstPublicSchema|Test.*PublicEntryResult.*Schema|TestCLIJSONOutputsValidateAgainstPublishedSchemas|TestPreviewRequest' -count=1 -v`: pass.
- `go test ./internal/releasecontracts -run 'TestReleaseSchemasCompile|TestReleaseSchemaReferencesAreLocallyResolvable|TestCLIAgentContractDocsStayAligned' -count=1 -v`: pass.
- `go test ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts -count=1`: pass.
- `git diff --check`: pass.
- Separate verifier: no blockers. Non-blocking risks are the narrow
  post-verification output hash TOCTOU window and future schema splitting if
  terminal compiler results later need input-only fingerprints.

Backlog self-review:

- Do not move output hash into `artifacts[].sha256` until agents need per-
  artifact generic hashing across more artifact kinds.
- Consider hashing inside runtime verification so `verification.json` and the
  public result can cross-check the same output bytes.
- Keep semantic success tied to verification pass and required artifacts; a
  matching hash alone only identifies bytes.

## Phase 23 - Verification Artifact Output Identity Sealing

Lane: Execution Contract.

Web-search value: low. The design choice is internal contract alignment:
runtime verification and public result should report the same output workbook
SHA-256. No new external dependency or time-sensitive standard is needed before
implementation.

### TODO

- [x] Add failing installed E2E test for verification/public output hash
  equality.
  - Evaluation: installed `run-intent` result
    `fingerprints.output_workbook_sha256`, `runtime.verification`, and the
    on-disk verification artifact all report the same output workbook hash.
  - Result: an agent can cross-check the public envelope against authoritative
    runtime evidence.
  - Likely files: `cmd/sheet-ops-codex/install_e2e_test.go`.
  - Risk: medium.
  - Rollback: keep Phase 22 public-envelope output identity only.
- [x] Add runtime verification field and schema coverage.
  - Evaluation: `runtimeworkbookcase.VerificationResult` and
    `contracts/verification/verification_result.schema.json` include
    `output_workbook_sha256` for emitted verification summaries.
  - Result: runtime evidence carries the same byte identity as the public
    envelope.
- [x] Reuse verification hash for public execution fingerprints.
  - Evaluation: public result derives `output_workbook_sha256` from
    `result.Verification.OutputWorkbookSHA256` and only falls back to hashing
    the file when needed.
  - Result: the public envelope and verification artifact do not drift through
    independent hashing paths.
- [x] Run separate verifier.
  - Evaluation: verifier checks source of truth, schema compatibility, installed
    E2E assertions, and no semantic-success overclaim.
  - Result: pass/fail before commit.
- [x] Commit Phase 23.
  - Evaluation: focused installed/runtime/schema tests, release contract tests,
    package tests, and `git diff --check` pass.
  - Result: `phase23/verification: mirror output identity`.

### Phase-End Backlog Review

Ask:

- Should operation-specific verification implementations in `runtime/verify`
  expose source/output hashes directly, or is workbookcase summary enrichment
  the right boundary?
- Should failed verification summaries also carry output hash when the output
  workbook exists?
- Should input fingerprints also move into verification artifacts for a full
  input/output evidence tuple?
- Should `output_workbook_sha256` become required in the verification schema
  after all historical fixtures and non-success paths are audited?
- Should missing output hash become a hard verification artifact emission error
  instead of an omitted field?

Do not claim this eliminates all TOCTOU risk until hashing is performed at the
same boundary as semantic verification for every operation.

### Phase 23 Result Note

Implementation outcome:

- `runtimeworkbookcase.VerificationResult` now carries
  `output_workbook_sha256` when the output workbook can be read.
- The persisted `verification-summary.json`, `runtime.verification`, and public
  `fingerprints.output_workbook_sha256` now agree for the installed
  append-rows E2E path.
- Public execution fingerprints use
  `result.Verification.OutputWorkbookSHA256` as the source of truth when it is
  present, falling back to direct output-file hashing only for older or partial
  runtime results.
- `contracts/verification/verification_result.schema.json` accepts
  `output_workbook_sha256` with the SHA-256 pattern.
- Docs and skill references tell agents to cross-check the public fingerprint
  against `runtime.verification.output_workbook_sha256` without treating hash
  equality as semantic workbook correctness.

Verification:

- Red evidence:
  `go test ./cmd/sheet-ops-codex -run TestInstallSkillBundledCLIRunIntentExecutesWorkbookEndToEnd -count=1 -v`
  failed because `runtime.verification.output_workbook_sha256` was empty while
  the public result already emitted an output workbook fingerprint.
- `go test ./runtime/workbookcase ./cmd/sheet-ops-codex -run 'TestVerificationSummaryCarriesOutputWorkbookFingerprint|TestRunAppendStructuredRowsEndToEnd|TestInstallSkillBundledCLIRunIntentExecutesWorkbookEndToEnd' -count=1 -v`: pass.
- `go test ./runtime/workbookcase ./cmd/sheet-ops-codex -run 'TestVerificationSummaryCarriesOutputWorkbookFingerprint|TestRunAppendStructuredRowsEndToEnd|TestInstallSkillBundledCLIRunIntentExecutesWorkbookEndToEnd|TestCLIJSONOutputsValidateAgainstPublishedSchemas|TestPublicEntryResultGoldenValidatesAgainstPublicSchema|TestExecutedPublicEntryResult.*Fingerprint' -count=1 -v`: pass.
- `go test ./internal/releasecontracts -run 'TestReleaseSchemasCompile|TestReleaseSchemaReferencesAreLocallyResolvable|TestCLIAgentContractDocsStayAligned' -count=1 -v`: pass.
- `go test ./runtime/workbookcase ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts -count=1`: pass.
- `git diff --check`: pass.
- Separate verifier: no blockers. Non-blocking risks are that unreadable output
  files currently omit the hash rather than failing hard, and the verification
  schema accepts but does not require `output_workbook_sha256`.

Backlog self-review:

- Do not make the verification schema field required until historical fixtures
  and failure/partial-result paths are audited.
- Consider changing `fileSHA256IfReadable` into an error-returning helper once
  every operation can fail explicitly on missing output hash.
- Input fingerprints in verification artifacts remain a later evidence-tuple
  enhancement; this phase only cross-checks output identity.

## Phase 24 - Successful Verification Hash Requirement

Lane: Execution Contract.

Web-search value: low. This phase tightens a local JSON Schema 2020-12
contract already used in the repository: successful verification artifacts must
include output workbook identity. No new external methodology is needed.

### TODO

- [x] Add failing schema test for successful verification without output hash.
  - Evaluation: `contracts/verification/verification_result.schema.json`
    rejects `pass:true` verification summaries that omit
    `output_workbook_sha256`.
  - Result: schema validation alone proves successful verification artifacts
    carry output identity.
  - Likely files: `runtime/workbookcase/run_test.go` or a focused schema test.
  - Risk: medium.
  - Rollback: keep Phase 23 E2E proof and document the schema gap.
- [x] Add conditional schema requirement.
  - Evaluation: `pass:true` requires `output_workbook_sha256`; failure or
    partial summaries remain compatible while failure-path evidence is audited.
  - Result: successful runtime verification cannot be emitted without a hash.
- [x] Update docs/phase evidence.
  - Evaluation: public docs distinguish success artifact requirement from
    optional failure-path evidence.
  - Result: agents know successful verification summaries should always expose
    output identity.
- [x] Run separate verifier.
  - Evaluation: verifier checks schema condition, test coverage, no failure-path
    overclaim, and no preview contract drift.
  - Result: pass/fail before commit.
- [x] Commit Phase 24.
  - Evaluation: focused runtime/schema tests, release contract tests, package
    tests, and `git diff --check` pass.
  - Result: `phase24/verification: require successful output identity`.

### Phase-End Backlog Review

Ask:

- Should failed verification summaries also require output hash whenever
  `output_file` is non-empty?
- Should `fileSHA256IfReadable` return an error so missing hash failures are
  explicit before schema validation?
- Should runtime verification hash the output before and after semantic checks
  to close the remaining mutation window?

Do not claim failure-path identity is fully sealed in this phase; the
requirement is limited to successful verification summaries.

### Phase 24 Result Note

- Implementation: `contracts/verification/verification_result.schema.json`
  conditionally requires `output_workbook_sha256` when `pass:true`. Failed
  verification summaries remain schema-compatible without the hash until the
  failure-path policy is separately audited.
- Coverage: `runtime/workbookcase/run_test.go` rejects successful verification
  summaries without output identity, allows failed summaries without output
  identity, and keeps the runtime summary hash proof.
- Documentation: public CLI output contracts and mirrored skill references now
  tell agents to cross-check successful verification artifacts against
  `runtime.verification.output_workbook_sha256`.
- Red evidence: before the schema change,
  `go test ./runtime/workbookcase -run TestVerificationSchemaRejectsSuccessfulResultWithoutOutputFingerprint -count=1 -v`
  failed because the schema accepted a successful verification summary without
  `output_workbook_sha256`.
- Verification evidence: focused runtime/schema tests pass. Separate verifier
  reported no blockers and identified only non-blocking follow-up risk around
  explicit `pass:false` coverage and typed file-read failure handling; the
  compatibility test covers the former.

## Phase 25 - Verification Output-File Hash Requirement

Lane: Execution Contract.

Web-search value: low. This phase is a local JSON Schema contract tightening
based on the evidence tuple policy already established in Phases 22-24. The
design question is not external API behavior; it is whether a verification
artifact that names an output file may omit that file's byte identity.

### TODO

- [x] Add failing schema test for failed verification with output file but no
  output hash.
  - Evaluation: a `pass:false` verification summary with non-empty
    `output_file` and no `output_workbook_sha256` is rejected.
  - Result: failure-path artifacts cannot claim an output workbook path without
    binding the workbook bytes.
  - Likely files: `runtime/workbookcase/run_test.go`.
  - Risk: medium.
  - Rollback: keep Phase 24's success-only requirement and document the
    failure-path gap.
- [x] Preserve compatibility for failure summaries that have no output file.
  - Evaluation: `pass:false`, empty `output_file`, and no
    `output_workbook_sha256` remains schema-compatible.
  - Result: early failures before output materialization can still be reported.
- [x] Add conditional schema requirement for non-empty `output_file`.
  - Evaluation: `output_file` with `minLength: 1` requires
    `output_workbook_sha256`; the existing `pass:true` requirement remains.
  - Result: all materialized-output verification artifacts carry byte identity.
- [x] Update docs/phase evidence.
  - Evaluation: public docs and mirrored skill references distinguish
    materialized-output identity from semantic verification success.
  - Result: agents know to require hash whenever a verification artifact names
    an output file.
- [x] Run separate verifier.
  - Evaluation: verifier checks schema condition, compatibility for early
    failures, no preview drift, and no overclaim on unreadable file handling.
  - Result: pass/fail before commit.
- [x] Commit Phase 25.
  - Evaluation: focused runtime/schema tests, release contract tests, package
    tests, and `git diff --check` pass.
  - Result: `phase25/verification: require output-file identity`.

### Phase-End Backlog Review

Ask:

- Should `fileSHA256IfReadable` become `fileSHA256` with an explicit error
  return so unreadable output files fail before schema validation?
- Should the runtime verification layer hash before and after semantic checks
  to reduce the remaining mutation window?
- Should input workbook fingerprints move into verification artifacts next so
  each verification summary carries the complete input/output evidence tuple?

Do not claim this phase proves semantic correctness from hashes. It only
requires byte identity for verification artifacts that name a materialized
output workbook.

### Phase 25 Result Note

- Implementation: `contracts/verification/verification_result.schema.json`
  now requires `output_workbook_sha256` whenever `output_file` is a non-empty
  string. The Phase 24 `pass:true` requirement remains in place.
- Compatibility: failed verification summaries with an empty `output_file` can
  still omit the hash, preserving early failure reporting before output
  materialization.
- Coverage: `runtime/workbookcase/run_test.go` now rejects failed verification
  summaries that name an output file without output identity, allows early
  failures with an empty output file, rejects successful summaries without
  output identity, and keeps the runtime summary hash proof.
- Public result alignment: `contracts/results/public_entry_result.schema.json`
  mirrors the same nested `runtime.verification` hash requirement, and the
  executed public-result fixture/golden now include
  `runtime.verification.output_workbook_sha256`.
- Documentation: public CLI output contracts and mirrored skill references now
  tell agents that any verification artifact naming a non-empty `output_file`
  must carry `runtime.verification.output_workbook_sha256`.
- Red evidence:
  `go test ./runtime/workbookcase -run TestVerificationSchemaRejectsFailedResultWithOutputFileWithoutOutputFingerprint -count=1 -v`
  failed before the schema change because the schema accepted `pass:false`,
  non-empty `output_file`, and missing `output_workbook_sha256`.
- Focused verification:
  `go test ./runtime/workbookcase -run 'TestVerificationSchemaRejectsSuccessfulResultWithoutOutputFingerprint|TestVerificationSchemaRejectsFailedResultWithOutputFileWithoutOutputFingerprint|TestVerificationSchemaAllowsFailedResultWithoutOutputFileOrFingerprint|TestVerificationSummaryCarriesOutputWorkbookFingerprint|TestRunAppendStructuredRowsEndToEnd' -count=1 -v`
  passes.
- Verifier blocker resolved: the first separate verifier found that the public
  executed-result schema still allowed nested `runtime.verification.output_file`
  without `output_workbook_sha256`. The public schema, fixture, and golden were
  aligned, then
  `go test ./cmd/sheet-ops-codex -run 'TestPublicEntryResultGoldenValidatesAgainstPublicSchema|TestExecutedPublicEntryResult.*Fingerprint|TestCLIJSONOutputsValidateAgainstPublishedSchemas|TestInstallSkillBundledCLIRunIntentExecutesWorkbookEndToEnd' -count=1 -v`
  passed.
- Second verifier blocker resolved: the public executed-result schema had kept
  `runtime.verification.output_file` stricter than the canonical verification
  schema and also treated `status:"executed"` as `ok:true`. The public schema
  now allows empty nested `output_file`, keeps the non-empty output-file hash
  requirement, and allows executed envelopes whose verification failed
  (`ok:false`) while still requiring `recoverable:false` and fingerprints.
  `go test ./cmd/sheet-ops-codex -run 'TestExecutedPublicEntryResultSchemaRejectsVerificationOutputFileWithoutHash|TestExecutedPublicEntryResultSchemaAllowsVerificationFailureWithOutputFileAndHash|TestExecutedPublicEntryResultSchemaAllowsEarlyVerificationFailureWithoutOutputFileOrHash|TestPublicEntryResultGoldenValidatesAgainstPublicSchema|TestExecutedPublicEntryResultValidatesAgainstPublicSchema|TestExecutedPublicEntryResultSchemaRejectsMissingOutputFingerprint' -count=1 -v`
  passes.
- Third verifier blocker resolved: the public executed-result schema still
  required an `output_workbook` artifact for every `status:"executed"` result,
  even though real early verification failures omit that artifact when
  `runtime.verification.output_file` is empty. The schema now requires the
  primary output workbook artifact only for executed success (`ok:true`), while
  still requiring verification and evidence artifacts for executed envelopes.
  The early-failure public schema test now builds the envelope through
  `newExecutedPublicEntryResult` and asserts the output workbook artifact is
  absent. `go test ./cmd/sheet-ops-codex -run 'TestExecutedPublicEntryResultSchemaRejectsVerificationOutputFileWithoutHash|TestExecutedPublicEntryResultSchemaAllowsVerificationFailureWithOutputFileAndHash|TestExecutedPublicEntryResultSchemaAllowsEarlyVerificationFailureWithoutOutputFileOrHash|TestExecutedPublicEntryResultSchemaRejectsMissingSuccessArtifacts|TestExecutedPublicEntryResultValidatesAgainstPublicSchema' -count=1 -v`
  passes.

## Phase 26 - Executed Failure Public Fingerprint Contract

Lane: Execution Contract.

Web-search value: low. This phase aligns local public result schema and emitter
behavior. The external methodology is unchanged: execution identity should
preserve input fingerprints for every executed result and output fingerprints
only when a materialized output workbook exists.

### TODO

- [x] Add failing public schema/emitter test for early executed verification
  failure without output fingerprint.
  - Evaluation: an executed public result with `ok:false`, empty
    `runtime.verification.output_file`, no output workbook artifact, and no
    `fingerprints.output_workbook_sha256` validates.
  - Result: agents can receive structured executed failure results even before
    output materialization.
  - Likely files: `cmd/sheet-ops-codex/public_entry_result_test.go`,
    `contracts/results/public_entry_result.schema.json`.
  - Risk: medium.
  - Rollback: keep Phase 25 schema allowance but document that the public
    emitter cannot yet represent the early failure path.
- [x] Keep executed success output fingerprint required.
  - Evaluation: `status:"executed"`, `ok:true` without
    `fingerprints.output_workbook_sha256` remains rejected.
  - Result: success artifacts still bind the materialized workbook bytes.
- [x] Split output fingerprint production from input fingerprint production.
  - Evaluation: public result construction can carry normalized intent and
    input workbook hashes when no output hash is available, without hashing an
    empty output path.
  - Result: emitter behavior matches the public schema for both success and
    early failure.
- [x] Update docs/phase evidence.
  - Evaluation: docs distinguish input execution identity from optional output
    identity on early failures.
  - Result: agents do not expect output workbook identity before an output
    workbook exists.
- [x] Run separate verifier.
  - Evaluation: verifier checks success contract, early failure contract,
    public schema/golden alignment, no preview drift, and no semantic
    correctness overclaim.
  - Result: pass/fail before commit.
- [x] Commit Phase 26.
  - Evaluation: focused runtime/public tests, release contract tests, package
    tests, and `git diff --check` pass.
  - Result: `phase26/results: allow early failure fingerprints`.

### Phase-End Backlog Review

Ask:

- Should `runIntentEntry` emit a public entry result when orchestration returns
  an error after runtime start, instead of returning only the internal result?
- Should `executedFingerprints` be renamed to clarify that output identity is
  optional for executed failures?
- Should public docs add a concrete failure envelope example?
- Should executed failures with a materialized output workbook use a
  non-success artifact role instead of `primary_success`?

Do not relax output identity for successful execution. This phase only removes
the false requirement that an output workbook hash exists before an output
workbook exists.

### Phase 26 Result Note

- Implementation: public execution fingerprints now require normalized intent
  and input workbook hashes for all executed results, while
  outer `fingerprints.output_workbook_sha256` is required only for executed
  success (`ok:true`). Runtime verification still requires its nested
  `output_workbook_sha256` whenever it names a non-empty output file.
- Emitter behavior: `executedFingerprints` no longer attempts to hash an empty
  output path. It returns input identity with an omitted output hash for early
  executed failures before output materialization.
- Coverage: `cmd/sheet-ops-codex/public_entry_result_test.go` proves early
  executed failure validates without outer `fingerprints.output_workbook_sha256`
  by inspecting the marshaled public envelope, and that `executedFingerprints`
  preserves input identity without output identity. Existing success tests still
  reject missing output fingerprints.
- Documentation: public CLI docs and mirrored skill references now say output
  fingerprint binding applies when an output workbook exists; input
  fingerprint comparison still applies to every executed public result.
- Red evidence:
  `go test ./cmd/sheet-ops-codex -run TestExecutedPublicEntryResultSchemaAllowsEarlyFailureWithoutOutputFingerprint -count=1 -v`
  failed before the schema change because
  `fingerprints.output_workbook_sha256` was always required.
- Verification evidence:
  focused public schema/emitter tests, release contract tests, related package
  tests, and `git diff --check` pass. Separate verifier reported no blockers
  and confirmed preview contracts did not drift.
- Backlog self-review: keep the artifact-role issue for the next phase. An
  executed failure that has a materialized output workbook can still label that
  artifact as `primary_success`; that is a result semantics problem, not part
  of this phase's fingerprint contract.

## Phase 27 - Executed Failure Artifact Role Contract

Lane: Execution Contract.

Web-search value: low. This phase tightens the local public result contract so
artifact roles do not overstate failed execution. No external methodology is
needed; the existing envelope already separates `ok`, artifact `required`, and
`success_role`.

### TODO

- [x] Add failing test for materialized-output executed failure artifact role.
  - Evaluation: an executed public result with `ok:false` and non-empty
    `runtime.verification.output_file` does not label `output_workbook` as
    `primary_success` or `required:true`.
  - Result: agents cannot misread failed workbook output as the primary success
    artifact.
  - Likely files: `cmd/sheet-ops-codex/public_entry_result_test.go`.
  - Risk: medium.
  - Rollback: keep Phase 26 fingerprint semantics and document the role gap.
- [x] Split success and failure output workbook artifact roles.
  - Evaluation: executed success keeps `output_workbook` as `required:true`,
    `primary_success`; executed failure with output uses `required:false`,
    `failure_evidence`.
  - Result: materialized failed output remains visible for diagnosis without
    being treated as success.
- [x] Add schema guard for failed output artifact role.
  - Evaluation: public schema rejects `status:"executed"`, `ok:false` results
    whose `output_workbook` artifact is marked `primary_success` or
    `required:true`.
  - Result: the contract enforces the generator's failure semantics.
- [x] Update docs/phase evidence.
  - Evaluation: public docs and skill mirrors distinguish primary success
    output from diagnostic failed output.
  - Result: agents inspect failed output as evidence, not as success.
- [x] Run separate verifier.
  - Evaluation: verifier checks success artifact contract, failed materialized
    output contract, early failure without output artifact, schema/golden
    alignment, and no preview drift.
  - Result: pass/fail before commit.
- [x] Commit Phase 27.
  - Evaluation: focused public schema/emitter tests, release contract tests,
    package tests, and `git diff --check` pass.
  - Result: `phase27/results: label failed output evidence`.

### Phase-End Backlog Review

Ask:

- Should the field name `success_role` become a neutral `role` in a future
  schema version?
- Should public docs include a table of artifact roles by `ok` state?
- Should failure evidence paths become required when `ok:false`?

Do not change successful execution artifact semantics in this phase.

### Phase 27 Result Note

- Implementation: public result artifact generation now labels
  `output_workbook` as `required:true`, `primary_success` only when
  `runtime.verification.pass` is true. If verification fails but an output
  workbook exists, the artifact is emitted as `required:false`,
  `failure_evidence`.
- Contract guard: `contracts/results/public_entry_result.schema.json` rejects
  executed `ok:false` envelopes that mark `output_workbook` as
  `primary_success` or `required:true`.
- Coverage: `cmd/sheet-ops-codex/public_entry_result_test.go` proves failed
  materialized output is labeled as failure evidence, schema rejects the old
  primary-success shape, early failures without output still omit the artifact,
  and executed success keeps the primary success artifact.
- Documentation: public docs and mirrored skill references now tell agents to
  treat output workbooks on `ok:false` results as failure evidence.
- Red evidence:
  `go test ./cmd/sheet-ops-codex -run TestExecutedPublicEntryResultLabelsFailedOutputAsFailureEvidence -count=1 -v`
  failed before the generator change because the failed output artifact was
  still `required:true`.
- Verification evidence: focused public schema/emitter tests, release contract
  tests, related package tests, and `git diff --check` pass. Separate verifier
  reported no blockers and confirmed preview contract files did not drift.
- Backlog self-review: keep failed `run-intent` end-to-end coverage as a later
  enhancement. This phase locks the public generator/schema semantics with
  focused coverage.

## Phase 28 - Failed Run-Intent Public Envelope Emission

Lane: Execution Contract.

Web-search value: low. Phase 28 applies the established local contract: a CLI
agent must receive structured stdout whenever runtime execution started, even
if verification fails and the command exits non-zero.

### TODO

- [x] Add failing test for runtime-started `run-intent` orchestration failure.
  - Evaluation: `runIntentEntry` returns a non-nil `PublicEntryResult` and a
    non-nil orchestration error when the runtime has a run id or telemetry path.
  - Result: agents can parse a failed execution envelope instead of receiving
    only stderr.
  - Likely files: `cmd/sheet-ops-codex/run_intent_entry_test.go`.
  - Risk: medium.
  - Rollback: keep Phase 27 generator/schema coverage and document the CLI path
    gap.
- [x] Add an orchestrator seam for the command adapter.
  - Evaluation: tests can stub `OrchestrateValidated` without invoking the full
    workbook runtime.
  - Result: command adapter behavior is covered deterministically.
- [x] Emit public result on runtime-started orchestration errors.
  - Evaluation: failed verification output is labeled as `failure_evidence`,
    fingerprints are preserved, and the original orchestration error still
    controls the non-zero exit.
  - Result: stdout remains machine-readable while exit status remains truthful.
- [x] Run separate verifier.
  - Evaluation: verifier checks the failure command path, schema compatibility,
    success path preservation, and no preview drift.
  - Result: pass/fail before commit.
- [x] Commit Phase 28.
  - Evaluation: focused command adapter tests, related package tests, and
    `git diff --check` pass.
  - Result: `phase28/run-intent: emit failed execution envelope`.

### Phase-End Backlog Review

Ask:

- Should there be a command-level stdout/stderr smoke test for `run-intent`
  after this adapter-level coverage?
- Should `run-validated` also gain a public entry envelope instead of raw
  runtime JSON?
- Should fingerprint construction failures produce a smaller structured
  terminal envelope instead of falling back to stderr only?

Do not change successful execution semantics in this phase.

### Phase 28 Result Note

- Implementation: `run-intent` now emits an executed public entry result when
  orchestration returns an error after the runtime has started. The original
  orchestration error is still returned so the CLI remains non-zero.
- Contract guard: pre-runtime orchestration failures still return no executed
  public envelope, avoiding false runtime evidence when no run id or telemetry
  path exists.
- Coverage: focused tests cover helper-level failed envelope emission,
  command-level stdout JSON plus non-zero error behavior, and the pre-runtime
  failure branch.
- Red evidence:
  `go test ./cmd/sheet-ops-codex -run TestRunIntentEntryEmitsPublicEnvelopeWhenRuntimeStartedThenFails -count=1 -v`
  initially failed because there was no orchestrator seam and the helper
  returned a nil public entry result for runtime-started failures.
- Verification evidence: focused run-intent failure tests, focused public
  schema/emitter tests, related package tests, and `git diff --check` pass.
  Separate verifier reported no blockers; its two residual risks were converted
  into command-level and pre-runtime branch tests before commit.
- Backlog self-review: keep `run-validated` public envelope migration and
  structured fallback for fingerprint-construction failures as future work.

## Phase 29 - Run-Validated Internal Handoff Envelope

Lane: Execution Contract.

Web-search value: low. This is a local contract-boundary phase. Existing docs
and capabilities classify `run-validated` as `internal_handoff`, so the goal is
not to promote it to a public entry. The goal is to make its stdout agent-stable
instead of raw Go runtime JSON.

### TODO

- [x] Add failing command-level test for `run-validated` stdout shape.
  - Evaluation: successful `run-validated` writes an outer envelope with
    `schema_version`, `ok`, `command`, `recoverable`, `artifacts`,
    `next_actions`, `status`, `classification`, and `runtime`.
  - Result: internal handoff callers can parse the same top-level control
    fields as public execution entries without treating this as a public
    workbook request entry.
  - Likely files: `cmd/sheet-ops-codex/run_validated_entry_test.go`.
  - Risk: medium.
  - Rollback: keep raw runtime JSON and document the handoff gap.
- [x] Preserve non-public boundary.
  - Evaluation: `run-validated` remains classified as `internal_handoff` in
    capabilities/schema surfaces and docs.
  - Result: humans still use the `sheet-ops` skill / `run-intent` route, while
    internal agents get machine-readable handoff output.
- [x] Emit structured artifacts for the runtime result.
  - Evaluation: output workbook, verification, evidence dir, execution,
    outcome, and report use the same artifact role semantics as public executed
    results.
  - Result: callers can locate authoritative evidence without parsing raw
    runtime internals first.
- [x] Run separate verifier.
  - Evaluation: verifier checks handoff stdout shape, classification
    preservation, success/failure artifact roles, and public entry schema drift.
  - Result: pass/fail before commit.
- [x] Commit Phase 29.
  - Evaluation: focused handoff tests, related package tests, and
    `git diff --check` pass.
  - Result: `phase29/run-validated: wrap internal handoff result`.

### Phase-End Backlog Review

Ask:

- Should a separate JSON Schema be published for internal handoff results?
- Should `run-request`, the hidden compatibility alias, advertise the same
  envelope command value or preserve `run-request` as the command value?
- Should fingerprint-construction failures for public entries produce a smaller
  structured terminal envelope?

Do not add `run-validated` to the public entry result schema in this phase.

### Phase 29 Result Note

- Implementation: `run-validated` and hidden `run-request` now write an
  internal handoff envelope with `schema_version`, `ok`, `command`, `status`,
  `classification:"internal_handoff"`, `recoverable`, `artifacts`,
  `next_actions`, and `runtime`.
- Contract guard: pre-runtime orchestration failures return the original error
  without emitting an executed handoff envelope, matching the `run-intent`
  runtime-started guard.
- Coverage: command-level tests cover successful `run-validated`, runtime-started
  failure with non-zero error plus JSON stdout, and pre-runtime failure with no
  executed stdout.
- Boundary: `run-validated` remains classified as `internal_handoff`; public
  entry schema was not broadened to include it.
- Red evidence:
  `go test ./cmd/sheet-ops-codex -run TestRunValidatedCommandWritesInternalHandoffEnvelope -count=1 -v`
  failed before implementation because stdout was raw `RunResult` JSON with Go
  field names such as `IDs` and no outer `schema_version`.
- Verification evidence: focused handoff tests, capability/schema tests, related
  package tests, and `git diff --check` pass. Separate verifier initially found
  a pre-runtime failure blocker; Phase 29 added the guard and regression test
  before commit.
- Backlog self-review: keep installed-bundle `run-validated` execution smoke
  and a formal internal handoff JSON Schema as future work. The installed bundle
  currently rebuilds successfully through existing install tests.

## Phase 30 - Installed Run-Validated Handoff Smoke

Lane: Execution Contract.

Web-search value: low. This phase verifies the installed package/runtime
surface, not a new external CLI methodology. The authoritative evidence is an
installed bundled `sheet-ops-codex` binary executing `run-validated` and
emitting the internal handoff envelope.

### TODO

- [x] Add failing installed-bundle `run-validated` smoke.
  - Evaluation: after `install-skill`, the installed
    `.codex/skills/sheet-ops/bin/sheet-ops-codex run-validated --request`
    executes a validated append-rows request.
  - Result: test fails until installed bundled runtime exposes the Phase 29
    envelope contract end to end.
  - Likely files: `cmd/sheet-ops-codex/install_e2e_test.go`.
  - Risk: medium.
  - Rollback: keep source-level handoff tests and document installed smoke gap.
- [x] Assert installed handoff envelope semantics.
  - Evaluation: stdout includes `schema_version`, `ok:true`,
    `command:"run-validated"`, `classification:"internal_handoff"`,
    `status:"executed"`, runtime verification, and required artifacts.
  - Result: installed internal callers can parse the same contract as source
    tests.
- [x] Assert workbook and authoritative evidence.
  - Evaluation: output workbook contains appended row; verification artifact,
    execution artifact, outcome artifact, and evidence dir exist.
  - Result: installed smoke is not stdout-only.
- [x] Run separate verifier.
  - Evaluation: verifier checks installed package boundary, source tests,
    artifact roles, and no public schema drift.
  - Result: pass/fail before commit.
- [x] Commit Phase 30.
  - Evaluation: focused installed smoke, related package tests, and
    `git diff --check` pass.
  - Result: `phase30/install: smoke run-validated handoff`.

### Phase-End Backlog Review

Ask:

- Should installed `run-request` hidden compatibility alias get its own smoke?
- Should a formal internal handoff JSON Schema become the next phase?
- Should installed failure-path `run-validated` smoke be added, or is
  source-level failure coverage enough until schema exists?

Do not broaden public entry schema in this phase.

### Phase 30 Result Note

- Implementation: installed E2E coverage now runs the bundled
  `.codex/skills/sheet-ops/bin/sheet-ops-codex run-validated --request`
  command against a validated append-rows request.
- Contract guard: the installed binary must emit the Phase 29 internal handoff
  envelope with `schema_version`, `ok:true`, `command:"run-validated"`,
  `classification:"internal_handoff"`, `status:"executed"`, non-recoverable
  status, artifacts, and runtime verification.
- Evidence guard: the smoke verifies output workbook contents, output workbook
  SHA-256 in runtime verification, verification artifact identity, execution
  artifact, outcome artifact, and required artifact roles.
- Red evidence: this phase is a promotion of Phase 29 source-level behavior into
  installed binary evidence; the new smoke would fail on pre-Phase 29 installed
  binaries that emitted raw `RunResult` JSON without the outer handoff envelope.
- Verification evidence: focused installed `run-validated` smoke, installed
  `run-intent` E2E, source handoff tests, related package tests, and
  `git diff --check` pass. Separate verifier found no code/schema blocker; it
  flagged only the temporal issue that the commit checkbox was marked before
  the commit existed, which is resolved by this phase commit.
- Backlog self-review: keep hidden installed `run-request` smoke and formal
  internal handoff JSON Schema as future work. Installed `run-validated` success
  is now covered by authoritative workbook/evidence checks.

## Phase 31 - Internal Handoff Result Schema

Lane: Schema Compatibility.

Web-search value: low. This phase applies the existing JSON Schema 2020-12
contract pattern already used by Sheet Ops CLI outputs. No new external
methodology is needed because the envelope shape was stabilized in Phases 29
and 30.

### TODO

- [x] Add failing schema/golden tests for internal handoff results.
  - Evaluation: a golden `run-validated` internal handoff result validates
    against a dedicated schema, and the schema rejects missing top-level control
    fields.
  - Result: internal handoff output is formally specified without broadening the
    public entry result schema.
  - Likely files:
    `cmd/sheet-ops-codex/internal_handoff_result_test.go`,
    `cmd/sheet-ops-codex/testdata/internal_handoff_run_validated.golden.json`,
    `contracts/cli/internal_handoff_result.schema.json`.
  - Risk: medium.
  - Rollback: keep Phase 29/30 behavioral tests and document schema gap.
- [x] Publish a dedicated internal handoff schema.
  - Evaluation: schema allows `command:"run-validated"` and hidden
    `command:"run-request"` only, requires `classification:"internal_handoff"`
    for `run-validated`, requires `runtime.verification`, and enforces artifact
    roles for executed success/failure.
  - Result: agents can validate internal handoff output separately from public
    entry results.
- [x] Keep public entry schema narrow.
  - Evaluation: `contracts/results/public_entry_result.schema.json` still does
    not include `run-validated` or `run-request`.
  - Result: internal handoff remains separate from public workbook entry.
- [x] Run separate verifier.
  - Evaluation: verifier checks schema validity, golden compatibility,
    public-schema separation, release contract compatibility, and docs truth.
  - Result: pass/fail before commit.
- [x] Commit Phase 31.
  - Evaluation: focused schema/golden tests, release contract tests, related
    package tests, and `git diff --check` pass.
  - Result: `phase31/schema: publish internal handoff result`.

### Phase-End Backlog Review

Ask:

- Should schema command/capabilities expose this new internal handoff schema
  path explicitly?
- Should installed failure-path `run-validated` smoke be added now that schema
  exists?
- Should the stale Final Review Phase be reopened to include Phases 28-31?

Do not add internal handoff commands to public entry schemas in this phase.

### Phase 31 Result Note

- Implementation: published
  `contracts/cli/internal_handoff_result.schema.json` for `run-validated` and
  hidden compatibility `run-request` internal handoff envelopes.
- Contract guard: the schema requires `schema_version`, `ok`, `command`,
  `status:"executed"`, `classification`, `recoverable:false`, `artifacts`,
  `next_actions`, and `runtime.verification`.
- Artifact guard: the schema requires `verification` as `success_evidence`,
  `evidence_dir` as `audit_trail`, requires successful `output_workbook` as
  `primary_success`, and rejects failed output workbooks marked as primary
  success.
- Boundary guard: focused tests prove
  `contracts/results/public_entry_result.schema.json` still rejects
  `run-validated` internal handoff envelopes.
- Compatibility: added a golden
  `cmd/sheet-ops-codex/testdata/internal_handoff_run_validated.golden.json`
  and included the new schema in published CLI schema version pinning.
- Red evidence:
  `go test ./cmd/sheet-ops-codex -run 'TestInternalHandoff.*Schema|TestPublicEntryResultSchemaRejectsRunValidatedHandoff' -count=1 -v`
  initially failed because the internal handoff schema and golden did not exist.
- Verification evidence: focused schema/golden tests, published schema pinning,
  installed `run-validated` smoke with schema validation, release contract
  schema compilation, related package tests, and `git diff --check` pass.
  Separate verifier reported no blockers.
- Backlog self-review: installed failure-path `run-validated` smoke, stronger
  mirror-drift enforcement, and stale final-review refresh remain follow-up
  candidates; public entry schema remains intentionally narrow.

## Phase 32 - Final Review Refresh After Handoff Hardening

Lane: Independent Verification.

Web-search value: low. This phase updates local evidence and no-overclaim
constraints after Phases 28-31. The web-researched methodology sources remain
the same as the earlier final review.

### TODO

- [x] Refresh final review status and implemented contract summary.
  - Evaluation: final review artifact includes failed `run-intent` envelope,
    internal handoff envelope, installed `run-validated` smoke, and internal
    handoff schema.
  - Result: reviewers do not rely on stale Phase 18-only evidence.
  - Likely files:
    `docs/superpowers/plans/2026-06-10-sheet-ops-cli-agent-contract-final-review.md`.
  - Risk: low.
  - Rollback: leave final review stale and record blocker.
- [x] Update evidence matrix and verification commands.
  - Evaluation: matrix points to authoritative current tests and schemas for
    Phases 28-31, including installed and schema evidence.
  - Result: final review proves the latest branch state.
- [x] Update backlog/no-overclaim constraints.
  - Evaluation: remaining gaps are explicit: installed failure-path smoke,
    hidden `run-request` installed smoke, mirror-drift hardening, and any
    runtime categories still reserved.
  - Result: final review does not overclaim complete runtime coverage.
- [x] Run separate verifier.
  - Evaluation: verifier checks final review truthfulness, current test
    coverage, no stale Phase 18 language, and no public-entry/internal-handoff
    confusion.
  - Result: separate verifier passed with no blockers. It confirmed
    `run-request` is not overclaimed, verifier closure is not used as
    repo-verifiable matrix proof, Phase 18 is historical, and backlog remains
    explicit.
- [x] Commit Phase 32.
  - Evaluation: final review docs, focused tests, related package tests, and
    `git diff --check` pass.
  - Result: `phase32/review: refresh handoff contract evidence`.

### Phase-End Backlog Review

Ask:

- Is the active goal now complete enough to mark complete, or do remaining
  backlog items still require more phases?
- Should installed failure-path handoff smoke be immediate now that schemas
  exist?
- Should mirror-drift hardening become the next release-contract phase?

Do not mark the overall goal complete until a fresh completion audit proves all
goal requirements and current plan evidence are satisfied.

### Phase 32 Result

- Focused handoff/public/schema tests passed.
- Focused preview, typed-error, and public-envelope tests passed.
- Related package regression passed:
  `go test ./runtime/workbookcase ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts -count=1`.
- `schema command run-validated --json` reports
  `classification:"internal_handoff"`.
- `git diff --check` passed.
- Separate verifier passed with no blockers.
- Active overall goal remains open pending a fresh requirement-by-requirement
  completion audit.

## Phase 33 - Hidden Run-Request Alias Proof

Lane: Schema Compatibility / Installed Boundary.

Web-search value: low. This phase closes a local compatibility-proof gap for
an already implemented hidden alias. No external methodology question is open.

### TODO

- [x] Add direct source-level `run-request` success coverage.
  - Evaluation: invoking `run-request --file <request>` emits an internal
    handoff envelope with `command:"run-request"`,
    `classification:"internal_handoff"`, runtime evidence, artifacts, and
    schema-valid JSON.
  - Result: the hidden alias is no longer inferred only from shared handler
    wiring. `TestRunRequestCommandWritesInternalHandoffEnvelope` covers
    `run-request --file`, `command:"run-request"`, `classification`,
    artifacts, runtime evidence, and internal handoff schema validation.
  - Likely files:
    `cmd/sheet-ops-codex/run_validated_entry_test.go`.
  - Risk: low.
  - Rollback: keep alias behavior unproven and leave it in backlog.
- [x] Add installed-bundle `run-request` smoke.
  - Evaluation: the installed `.codex/skills/sheet-ops/bin/sheet-ops-codex`
    binary accepts `run-request --file <request>`, emits a schema-valid
    internal handoff envelope, writes the workbook, and records evidence hashes.
  - Result: legacy compatibility callers have installed-boundary proof.
  - Likely files:
    `cmd/sheet-ops-codex/install_e2e_test.go`.
  - Risk: medium because installed E2E tests are slower and exercise installer
    packaging.
  - Rollback: retain source-level alias proof only and keep installed proof in
    backlog.
- [x] Update final review/backlog language.
  - Evaluation: `run-request` no longer appears as an unproven alias once the
    source and installed smokes pass; remaining backlog still names true gaps.
  - Result: final review evidence now lists direct and installed
    `run-request` proof and removes the stale alias-proof backlog.
- [x] Run separate verifier.
  - Evaluation: verifier confirms `run-request` proof is real, still hidden,
    still internal-only, and not added to public entry schemas.
  - Result: separate verifier passed with no blockers. It confirmed source and
    installed alias proof, hidden/internal-only status, public-entry separation,
    final-review backlog cleanup, and sufficient executor verification.
- [x] Commit Phase 33.
  - Evaluation: focused source/installed tests, related package tests, and
    `git diff --check` pass.
  - Result: `phase33/compat: prove run-request handoff alias`.

### Phase-End Backlog Review

Ask:

- Should installed failure-path `run-validated` smoke be the next immediate
  phase?
- Should mirror-drift hardening move ahead of failure-path coverage?
- Does proving `run-request` change any public docs, or should it stay hidden
  and compatibility-only?

Result:

- Keep `run-request` hidden and compatibility-only; no public docs promotion is
  needed.
- Installed failure-path `run-validated` smoke remains the next direct
  execution-contract gap.
- Mirror-drift hardening remains a separate release-contract hardening phase.
- Active overall goal remains open pending a fresh requirement-by-requirement
  completion audit.

## Phase 34 - Installed Run-Validated Failure-Path Smoke

Lane: Installed Boundary / Failure Evidence.

Web-search value: low. This phase verifies local installed-binary behavior for
an already designed failure envelope. No external methodology question is open.

### TODO

- [x] Add installed-bundle runtime-started failure smoke for `run-validated`.
  - Evaluation: installed `.codex/skills/sheet-ops/bin/sheet-ops-codex
    run-validated --request <request>` returns non-zero for a deterministic
    runtime-started failure, still writes a schema-valid internal handoff
    envelope to stdout, reports `ok:false`, preserves
    `command:"run-validated"` and `classification:"internal_handoff"`, and
    points to authoritative evidence artifacts.
  - Result: installed internal handoff callers can parse failure evidence
    without relying on source-level stubs.
    `TestInstallSkillBundledCLIRunValidatedFailureEmitsHandoffEnvelope` covers
    non-zero exit, stdout JSON, `ok:false`, internal handoff schema validation,
    `failure_evidence` output workbook role, verification/review/repair
    artifacts, and output hash agreement.
  - Likely files:
    `cmd/sheet-ops-codex/install_e2e_test.go`.
  - Risk: medium because the installed smoke must distinguish pre-runtime
    validation failure from runtime-started failure.
  - Rollback: keep source-level failure coverage only and leave installed
    failure proof in backlog.
- [x] Update final review/backlog language.
  - Evaluation: installed failure-path `run-validated` is moved out of backlog
    only if the installed smoke proves schema-valid failure evidence.
  - Result: final review evidence now lists installed `run-validated`
    failure-path coverage and removes the stale backlog item.
- [x] Run separate verifier.
  - Evaluation: verifier confirms the failure is runtime-started, stdout JSON
    is schema-valid, the command exits non-zero, and no success artifact is
    overclaimed.
  - Result: separate verifier passed with no blockers. It confirmed installed
    binary invocation, non-zero exit, runtime-started schema-valid stdout
    envelope, `failure_evidence` output role, evidence hash/review/repair
    checks, and public/internal contract separation.
- [x] Commit Phase 34.
  - Evaluation: focused installed failure test, related package tests, and
    `git diff --check` pass.
  - Result: `phase34/install: smoke run-validated failure handoff`.

### Phase-End Backlog Review

Ask:

- Should mirror-drift hardening become the next release-contract phase?
- Should reserved runtime error categories now be promoted with deterministic
  installed/source producers, or stay reserved?
- Is a hidden `run-request` installed failure smoke necessary, or is
  `run-validated` failure proof plus shared handler sufficient?

Result:

- Hidden `run-request` installed failure smoke stays optional; `run-validated`
  failure proof covers the internal handler boundary, and the alias remains
  hidden compatibility.
- Mirror-drift hardening is the next direct release-contract candidate.
- Reserved runtime error categories stay reserved until each has deterministic
  producer tests.
- Active overall goal remains open pending a fresh requirement-by-requirement
  completion audit.

## Final Review Phase - Agent Execution Contract Completion

Lane: Independent Verification.

Status: historical Phase 18 closure, superseded by Phase 32 for the current
post-handoff-hardening evidence matrix.

### TODO

- [x] Run full verification.
  - Evaluation: Phase 18 focused phase tests, generated JSON validation,
    installed E2E smoke, and schema compatibility tests pass.
  - Result: Phase 18 completion evidence was fresh at the time of that commit;
    Phase 32 refreshes the current evidence after Phases 28-31.
- [x] Dispatch final verifier.
  - Evaluation: Phase 18 verifier reviewed then-current phases,
    authoritative artifacts, web-research-backed design choices, and
    no-overclaim constraints.
  - Result: pass/fail before the Phase 18 final review commit.
- [x] Update final review artifact.
  - Evaluation: Phase 18 evidence matrix covered installed E2E, typed errors,
    result envelopes, schemas/goldens, preview, dry-run deferral, and remaining
    backlog.
  - Result: Phase 32 updates the same final review artifact for the current
    branch state.
- [x] Commit final review.
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
- Separate verifier: retrospective Phase 12 verifier passed with no blockers. It
  confirmed the deterministic `append_structured_rows` candidate, the
  non-stdout-only authoritative success fields, and that Phase 15 implemented
  those fields in the installed E2E smoke. It also reran the workbookcase,
  requestcompiler, and installed E2E focused commands successfully.

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
  `request_checkpoint`, `validation_blocked`, `execution_failed`, and
  `verification_failed`. Phase 20 later promoted `invalid_json_or_schema` after
  adding a deterministic normalized intent loader producer.

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

### Phase 16 Result Note

Web research:

- JSON Schema 2020-12 references confirmed the use of explicit root `$schema`,
  `$id`, and modular `$defs`.
- MCP tool output-schema guidance reinforced the contract shape: structured
  outputs should have a schema and clients should be able to validate them.
- Schema-evolution references informed the local compatibility policy:
  breaking output changes require a new `schema_version`; additive optional
  fields can stay on `sheet-ops-cli/v1`.

Implementation outcome:

- Added published CLI output schemas for:
  - `capabilities --json`,
  - `schema --json` and `schema command <name> --json`,
  - JSON error envelopes,
  - `preflight --json`.
- Updated `contracts/results/public_entry_result.schema.json` for the Phase 13
  outer agent envelope fields.
- Added a public entry result golden fixture for a minimal executed
  `run-intent` success envelope.
- JSON error envelopes now include `schema_version`.
- Added schema tests against live CLI outputs, generated public-entry
  envelopes, terminal compiler-stop envelopes, the golden fixture, negative
  malformed public-entry cases, and schema-version pins.
- Added `docs/public/cli-output-contracts.md` with versioning and unknown-field
  consumer guidance.

Verification:

- Red evidence:
  `go test ./cmd/sheet-ops-codex -run 'TestExecutedPublicEntryResultValidatesAgainstPublicSchema' -count=1 -v`
  failed because `public_entry_result.schema.json` rejected the Phase 13 outer
  fields `schema_version`, `ok`, `command`, `recoverable`, `artifacts`, and
  `next_actions`.
- `go test ./cmd/sheet-ops-codex -run 'TestSchemaCommandUnknownJSONEmitsTypedError|TestCLIJSONOutputsValidateAgainstPublishedSchemas|TestCLIErrorEnvelopeValidatesAgainstPublishedSchema|TestPublishedCLISchemasPinSchemaVersion|Test.*PublicEntryResult.*Schema|TestTerminalCompilerPublicEntryResultValidatesAgainstPublicSchema' -count=1 -v`: pass.
- `go test ./internal/releasecontracts -run 'TestReleaseSchemasCompile|TestReleaseSchemaReferencesAreLocallyResolvable' -count=1`: pass.
- `go test ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts -count=1`: pass.
- `go test ./... -count=1`: pass.
- Separate verifier initially found three blockers: executed success artifacts
  were too permissive, checkpoint/blocked status consistency was not enforced,
  and error envelopes lacked `schema_version`; all three were closed before
  commit.

Backlog self-review:

- Hand-written schemas are acceptable for this phase because they encode the
  agent-facing contract rather than every internal Go struct detail.
- Schema generation remains a possible future maintenance improvement if CLI
  output structs grow, but introducing it now would add toolchain surface before
  the contract finishes stabilizing.
- `sheet-ops-cli/v1` is sufficient for now; no semantic sub-version is needed
  until a real breaking output migration exists.

Commit: `phase16/schema: publish cli output contracts`.

### Phase 17 Result Note

Implementation outcome:

- Added `preview-request` as a read-only, non-dry-run impact inspection command.
- The command validates the normalized intent JSON and input workbook boundary,
  computes the expected workbook-case state root, and returns planned reads,
  planned writes, planned artifacts, and limitations.
- The command does not persist request-compiler artifacts, does not execute
  runtime orchestration, does not create `.sheet-ops-state`, and does not write
  the output workbook.
- Exposed `preview-request` through help, capabilities, command schema, installed
  CLI smoke, public skill references, and `contracts/cli/preview_request_result.schema.json`.

Verification:

- Red evidence:
  `go test ./cmd/sheet-ops-codex -run 'TestPreviewRequest' -count=1 -v`
  initially failed because `preview-request` was an unknown command and was not
  present in the agent-contract command group.
- `go test ./cmd/sheet-ops-codex -run 'TestPreviewRequest|TestInstallSkillBundledCLIExposesAgentContract|TestCLIJSONOutputsValidateAgainstPublishedSchemas|TestPublishedCLISchemasPinSchemaVersion|TestCLIHelpSeparatesCommandSurfaces' -count=1 -v`: pass.
- `go test ./internal/releasecontracts -run 'TestCLIAgentContractDocsStayAligned|TestReleaseSchemasCompile|TestReleaseSchemaReferencesAreLocallyResolvable' -count=1 -v`: pass after public skill/reference mirrors were aligned.
- Separate verifier: pass, no blockers. It confirmed `preview-request` is
  `read_only:true`, `mutating:false`, `dry_run_capable:false`, does not create
  output workbook or state artifacts, does not claim dry-run or execution
  success, is present in the install manifest and installed CLI smoke, and has a
  schema-validated live output. The verifier's residual risk that generic CLI
  schema sweep did not include preview was closed by adding the preview case to
  `TestCLIJSONOutputsValidateAgainstPublishedSchemas`.

Backlog self-review:

- A real dry-run is still deferred. `preview-request` only proves read-only
  impact inspection, not execution success or runtime mutation simulation.
- Preview currently validates normalized intent schema and workbook file
  boundary, but intentionally does not persist compiler decisions. A deeper
  compiler-backed preview should be a separate runtime planning track if needed.
- Preview reports full local paths for agent utility. Redaction can be revisited
  if this output becomes user-shareable or leaves the local agent boundary.

Commit: `phase17/preview: add truthful impact inspection`.

### Phase 19 Result Note

Implementation outcome:

- Extended `preview-request` output with agent trust metadata:
  `planner`, `plan_confidence`, `operation`, `would_mutate`,
  `mutation_summary`, and `fingerprints`.
- The command still reports `dry_run:false`, `read_only:true`, and
  `plan_confidence:"compiler_validated_boundary"` so agents do not mistake it for runtime
  dry-run evidence.
- `operation` comes from request-compiler validation's selected operation, so
  multi-candidate normalized intents follow compiler priority rather than input
  array order.
- `would_mutate` is true only for compiled preview decisions. Unresolved,
  checkpoint, or blocked decisions report `operation:"unresolved"` and
  no-mutation summary values with empty `planned_writes` and
  `planned_artifacts` instead of predicting workbook/artifact writes.
- `fingerprints.normalized_intent_sha256` and
  `fingerprints.input_workbook_sha256` identify the exact request/workbook
  inputs read by preview without writing state or output artifacts.
- Updated `contracts/cli/preview_request_result.schema.json` so the new fields
  are required and schema-validated.

Verification:

- Red evidence:
  `go test ./cmd/sheet-ops-codex -run TestPreviewRequestReportsImpactWithoutMutating -count=1 -v`
  failed because `preview_request_result.schema.json` rejected
  `planner`, `plan_confidence`, `operation`, `would_mutate`,
  `mutation_summary`, and `fingerprints`.
- Verifier blocker evidence:
  an independent verifier found that multi-candidate intents could report the
  first composition candidate rather than the compiler-selected operation. A
  second verifier found that `unresolved` decisions were reachable but not
  schema-valid and still reported `would_mutate:true`. A later verifier found
  that unresolved previews still populated required planned write/artifact
  arrays; those arrays are now empty for unresolved preview decisions. A final
  blocker pass found that the schema did not enforce no-mutation invariants and
  command schema read artifacts omitted `SHEET_OPS_ARTIFACT_ROOT` and
  `SHEET_OPS_KNOWLEDGE_ROOT`; both contract gaps are now covered by tests and
  schema/command metadata.
- `go test ./cmd/sheet-ops-codex -run TestPreviewRequestReportsImpactWithoutMutating -count=1 -v`: pass.
- `go test ./cmd/sheet-ops-codex -run 'TestPreviewRequest|TestInstallSkillBundledCLIExposesAgentContract|TestCLIJSONOutputsValidateAgainstPublishedSchemas|TestPublishedCLISchemasPinSchemaVersion' -count=1 -v`: pass.
- `go test ./internal/releasecontracts -run 'TestCLIAgentContractDocsStayAligned|TestReleaseSchemasCompile|TestReleaseSchemaReferencesAreLocallyResolvable' -count=1`: pass.
- `go test ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts -count=1`: pass.
- `git diff --check`: pass.
- Separate verifier: final pass reported no blockers. Residual risks are limited
  to a multi-candidate regression that proves compiler selection without a
  simultaneously satisfiable second operation, and installed-path metadata
  coverage that is narrower than source-path read-artifact assertions.

Backlog self-review:

- This phase intentionally does not promote real dry-run. The preview still
  validates the normalized intent and workbook boundary only.
- The next architectural step should repeat the fingerprints in mutating
  `run-intent` results, or add a compiler-backed planner result, so agents can
  prove preview/run input identity.
- Operation selection now follows request-compiler validation. Unknown or
  unresolved selections remain possible and must not be presented as executed
  runtime operations.

### Phase 20 Result Note

Implementation outcome:

- Added a typed `invalid_json_or_schema` producer for normalized intent loading
  failures on `preview-request --json`.
- Invalid normalized intent JSON now returns a JSON error envelope on stdout
  with `ok:false`, `error.code:"invalid_json_or_schema"`,
  `recoverable:true`, exit code 65, and repair-oriented suggested commands.
- Added `newInvalidDataError` so deterministic input/schema failures can be
  separated from fallback `internal_error`.
- Promoted `invalid_json_or_schema` from `reserved` to `emitted` in
  `capabilities --json` because it now has a focused producer test.

Verification:

- Red evidence:
  `go test ./cmd/sheet-ops-codex -run TestPreviewRequestInvalidIntentJSONEmitsInvalidDataError -count=1 -v`
  initially failed because stdout was empty and no JSON error envelope was
  emitted for malformed normalized intent JSON.
- Separate verifier blocker evidence:
  the first verifier found that wrapping `LoadNormalizedIntent` wholesale
  incorrectly classified missing intent files as `invalid_json_or_schema`.
  The loader path now separates file I/O from JSON/schema validation, and
  `TestPreviewRequestMissingIntentFileDoesNotClassifyAsInvalidData` covers the
  boundary.
- `go test ./cmd/sheet-ops-codex -run 'TestPreviewRequestInvalidIntentJSONEmitsInvalidDataError|TestCapabilitiesJSONReportsSafeEntryBoundaries|TestCLIErrorEnvelopeValidatesAgainstPublishedSchema|TestCLIJSONOutputsValidateAgainstPublishedSchemas' -count=1 -v`: pass.
- `go test ./internal/releasecontracts -run 'TestCLIAgentContractDocsStayAligned|TestReleaseSchemasCompile|TestReleaseSchemaReferencesAreLocallyResolvable' -count=1`: pass.
- Separate verifier: first pass found that wrapping the whole normalized intent
  loader swallowed missing intent files as `invalid_json_or_schema`; after
  separating file I/O from JSON/schema validation, re-verifier reported no
  blockers.

Backlog self-review:

- This phase only promotes deterministic normalized intent load/schema failures.
  It does not claim `request_checkpoint`, `validation_blocked`,
  `execution_failed`, or `verification_failed` producers.
- `run-intent` invalid intent loading can be considered in a later phase, but
  this slice intentionally starts with the agent-contract preview boundary that
  already requires JSON mode.
- Error messages still include local paths for agent repair context. Redaction
  can be revisited if error envelopes become user-shareable outside the local
  agent boundary.
