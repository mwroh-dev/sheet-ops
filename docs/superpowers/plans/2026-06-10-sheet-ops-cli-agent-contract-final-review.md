# Sheet Ops Agent Execution Contract Final Review

Branch: `codex/cli-agent-contract`

Review status: refreshed after Phase 35 mirror hardening. The current evidence covers
the original execution-contract work plus the later handoff hardening phases
28-35.

## Implemented Contract

- `sheet-ops` remains the single human-facing workbook request entry.
- `sheet-ops-codex` exposes classified install, discovery, typed handoff,
  maintainer diagnostic, internal execution, preview, and support surfaces.
- Installed `.codex/skills/sheet-ops/bin/sheet-ops-codex` is tested for
  metadata discovery, public `run-intent` workbook execution, and internal
  `run-validated` plus hidden `run-request` handoff execution. Installed
  `run-validated` also has runtime-started failure-path coverage.
- Public `run-intent` results expose a stable agent envelope with
  `schema_version`, `ok`, `command`, `recoverable`, `artifacts`,
  `next_actions`, compiler status, runtime evidence, and execution
  fingerprints.
- Failed public `run-intent` executions that start the runtime emit a public
  envelope while still returning a non-zero orchestration error. Materialized
  failed output workbooks are labeled `failure_evidence`, not primary success.
- Internal `run-validated` handoffs emit a separate internal handoff envelope
  with `classification:"internal_handoff"` and runtime evidence under
  `runtime`. Hidden `run-request` compatibility preserves `command:"run-request"`
  and is covered at both source and installed-bundle boundaries.
- Public CLI outputs and internal handoff outputs have JSON Schema 2020-12
  contracts and compatibility tests.
- `preview-request` is a truthful non-mutating impact inspection surface. It is
  not a dry-run and does not claim execution success.

## Evidence Matrix

| Area | Status | Authoritative Evidence | Success Claim |
| --- | --- | --- | --- |
| Phase/lane plan | pass | `docs/superpowers/plans/2026-06-10-sheet-ops-agent-execution-contract.md` defines execution, schema compatibility, and independent verification lanes through Phase 35. | Work proceeded phase-by-phase with dependency-aware lanes and phase notes. |
| Installed public workbook E2E | pass | `TestInstallSkillBundledCLIRunIntentExecutesWorkbookEndToEnd` runs public `install-skill.sh`, executes installed `bin/sheet-ops-codex run-intent`, asserts `ok:true`, artifacts, verification pass, evidence paths, hashes, and output workbook row `LineItems!A3:C3`. | Installed CLI can execute a real public workbook append with authoritative artifact evidence. |
| Installed internal handoff E2E | pass | `TestInstallSkillBundledCLIRunValidatedEmitsHandoffEnvelope` and `TestInstallSkillBundledCLIRunRequestEmitsHandoffEnvelope` install the package, run installed `bin/sheet-ops-codex run-validated --request` and hidden `run-request --file`, validate the handoff schema, check workbook contents, output SHA-256, verification artifact, execution artifact, outcome artifact, and artifact roles. | Installed internal handoff callers receive schema-valid envelopes backed by workbook and evidence artifacts. |
| Public result envelope | pass | `TestExecutedPublicEntryResultExposesAgentEnvelope`, `TestExecutedPublicEntryResultValidatesAgainstPublicSchema`, `TestTerminalCompilerPublicEntryResultValidatesAgainstPublicSchema`, and public result golden validation. | Agents get stable success and recoverable compiler-stop envelopes. |
| Failed public execution envelope | pass | `TestRunIntentCommandWritesPublicEnvelopeThenReturnsFailure`, `TestRunIntentEntryEmitsPublicEnvelopeWhenRuntimeStartedThenFails`, and `TestRunIntentEntryDoesNotEmitExecutionEnvelopeBeforeRuntimeStarts`. | Runtime-started failures still provide machine-readable public evidence while pre-runtime failures do not invent executed evidence. |
| Failed output artifact semantics | pass | `TestExecutedPublicEntryResultLabelsFailedOutputAsFailureEvidence`, `TestExecutedPublicEntryResultSchemaRejectsFailedOutputMarkedAsPrimarySuccess`, and schema checks for early failures with or without output. | Failed materialized output workbooks are diagnostic evidence, not success artifacts. |
| Internal handoff envelope | pass | `TestRunValidatedCommandWritesInternalHandoffEnvelope`, `TestRunRequestCommandWritesInternalHandoffEnvelope`, `TestRunValidatedCommandWritesFailureHandoffEnvelopeThenReturnsFailure`, `TestRunValidatedCommandDoesNotEmitExecutedEnvelopeBeforeRuntimeStarts`, and `TestInstallSkillBundledCLIRunValidatedFailureEmitsHandoffEnvelope`. | `run-validated` and hidden `run-request` success paths have stable control fields; source and installed failure coverage preserves non-zero errors while still emitting schema-valid runtime-started evidence. |
| Internal handoff schema | pass | `contracts/cli/internal_handoff_result.schema.json`, `TestInternalHandoffRunResultValidatesAgainstSchema`, `TestInternalHandoffResultGoldenValidatesAgainstSchema`, `TestInternalHandoffResultSchemaRejectsPublicEntryCommand`, and `TestPublicEntryResultSchemaRejectsRunValidatedHandoff`. | Internal handoff output is formally specified and separated from public entry output. |
| Typed errors | pass with reserved backlog | `TestStateRootMismatchClassifiesAsConfigurationError`, `TestPreviewRequestInvalidIntentJSONEmitsInvalidDataError`; `capabilities --json` marks `state_root_mismatch` and `invalid_json_or_schema` as `emitted`, while unproven runtime categories remain `reserved`. | Emitted error taxonomy is evidence-backed; unproven categories are not overclaimed. |
| Published schemas | pass | `TestCLIJSONOutputsValidateAgainstPublishedSchemas`, `TestPublishedCLISchemasPinSchemaVersion`, internal handoff golden validation, public entry golden validation, and release schema compile/reference tests. | CLI JSON outputs and published schemas are version-pinned and compile under release-contract checks. |
| Preview | pass | `TestPreviewRequestReportsImpactWithoutMutating`, `TestPreviewRequestIsExposedAsReadOnlyNonDryRunCommand`, installed preview smoke, and `contracts/cli/preview_request_result.schema.json`. | Agents can inspect planned impact without workbook output or state mutation. |
| Docs alignment | pass | Public and bundled skill docs mention the same internal handoff schema path, distinguish it from `contracts/results/public_entry_result.schema.json`, preserve preview non-dry-run guidance, output hash guidance, and failure-evidence semantics; `TestCLIAgentContractDocsStayAligned` guards these public/bundled mirrors. | Installed docs and public docs teach the same safe contract, and release-contract tests now catch the prior mirror-drift hardening gap. |
| Full regression | pass | `go test ./runtime/workbookcase ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts -count=1`; `git diff --check`. | No known regression in the packages most directly affected by the agent execution contract work. |

Verifier separation was performed in-thread and the phase notes record separate
verifier outcomes. Session closure is operational conversation evidence rather
than a repository artifact, so it is not used as a repo-verifiable success row
in this matrix.

## Verification Commands

- `go test ./cmd/sheet-ops-codex -run 'TestInstallSkillBundledCLIRunIntentExecutesWorkbookEndToEnd|TestInstallSkillBundledCLIRunValidatedEmitsHandoffEnvelope|TestInstallSkillBundledCLIRunRequestEmitsHandoffEnvelope|TestInstallSkillBundledCLIRunValidatedFailureEmitsHandoffEnvelope|TestRunIntentCommandWritesPublicEnvelopeThenReturnsFailure|TestRunValidatedCommandWritesInternalHandoffEnvelope|TestRunRequestCommandWritesInternalHandoffEnvelope|TestRunValidatedCommandWritesFailureHandoffEnvelopeThenReturnsFailure|TestRunValidatedCommandDoesNotEmitExecutedEnvelopeBeforeRuntimeStarts|TestInternalHandoff|TestPublicEntryResultSchemaRejectsRunValidatedHandoff|TestCLIJSONOutputsValidateAgainstPublishedSchemas|TestPublishedCLISchemasPinSchemaVersion' -count=1 -v`: pass.
- `go test ./cmd/sheet-ops-codex -run 'TestPreviewRequestReportsImpactWithoutMutating|TestPreviewRequestIsExposedAsReadOnlyNonDryRunCommand|TestStateRootMismatchClassifiesAsConfigurationError|TestPreviewRequestInvalidIntentJSONEmitsInvalidDataError|TestExecutedPublicEntryResultExposesAgentEnvelope|TestExecutedPublicEntryResultValidatesAgainstPublicSchema|TestTerminalCompilerPublicEntryResultValidatesAgainstPublicSchema|TestExecutedPublicEntryResultLabelsFailedOutputAsFailureEvidence|TestExecutedPublicEntryResultSchemaRejectsFailedOutputMarkedAsPrimarySuccess' -count=1 -v`: pass.
- `go test ./runtime/workbookcase ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts -count=1`: pass.
- `go test ./internal/releasecontracts -run 'TestReleaseSchemasCompile|TestReleaseSchemaReferencesAreLocallyResolvable|TestCLIAgentContractDocsStayAligned|TestBundledSchemasStayInSyncWithSourceContracts' -count=1 -v`: pass.
- `go run ./cmd/sheet-ops-codex schema command run-validated --json`: pass and reports `classification:"internal_handoff"` plus internal handoff output mode.
- `git diff --check`: pass.

## Web-Research-Backed Design Choices

- CLI stdout/exit-code shape follows the CLI Guidelines source class.
- Tool execution versus protocol/business errors follows MCP tool error guidance
  and SEP-1303's agent-visible validation-error direction.
- JSON contracts use JSON Schema 2020-12 with explicit root `$schema`, `$id`,
  and `$defs` based on JSON Schema and MCP schema guidance.

## No-Overclaim Constraints

- `preview-request` is not a dry-run. It validates input boundaries and reports
  planned impact only.
- Preflight is readiness diagnostics only. It does not prove future workbook
  mutation success.
- `run-validated` is an internal handoff surface, not a second public workbook
  request entry. Validate it with
  `contracts/cli/internal_handoff_result.schema.json`, not
  `contracts/results/public_entry_result.schema.json`.
- Public entry schema intentionally rejects internal handoff envelopes.
- Hidden `run-request` remains compatibility-only and hidden. Its proof does
  not make it a public workbook request entry.
- Runtime categories remain `reserved` until deterministic producers and tests
  exist. `invalid_json_or_schema` is no longer reserved after Phase 20 because
  malformed normalized intent JSON now has a deterministic typed producer.
- Installed workbook success is claimed only from installed E2E tests that check
  JSON result fields, evidence paths, hashes, and workbook contents.
- No success claim is based on stdout alone when an authoritative artifact or
  schema-validated JSON result exists.

## Remaining Backlog

- Refresh or supersede older final-review checklist language if additional
  phases are added after Phase 32.
- A deeper compiler-backed preview could be added later as a separate runtime
  planning track.
- A real dry-run remains deferred until runtime planning can prove non-mutation
  while simulating execution semantics.
- Runtime `request_checkpoint`, `validation_blocked`, `execution_failed`, and
  `verification_failed` should move from `reserved` to `emitted` only with
  deterministic producer tests.
