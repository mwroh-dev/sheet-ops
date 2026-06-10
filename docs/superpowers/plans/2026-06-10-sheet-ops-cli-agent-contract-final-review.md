# Sheet Ops Agent Execution Contract Final Review

Branch: `codex/cli-agent-contract`

Final status: pass after Phase 18 documentation closure.

## Implemented Contract

- `sheet-ops` remains the single human-facing workbook request entry.
- `sheet-ops-codex` exposes classified install, discovery, typed handoff,
  maintainer diagnostic, internal execution, preview, and support surfaces.
- Installed `.codex/skills/sheet-ops/bin/sheet-ops-codex` is tested for both
  metadata discovery and real workbook execution.
- Public `run-intent` results expose a stable agent envelope with
  `schema_version`, `ok`, `command`, `recoverable`, `artifacts`,
  `next_actions`, compiler status, and runtime evidence.
- Public CLI outputs have JSON Schema 2020-12 contracts and compatibility
  tests.
- `preview-request` is a truthful non-mutating impact inspection surface. It is
  not a dry-run and does not claim execution success.

## Evidence Matrix

| Area | Status | Authoritative Evidence | Success Claim |
| --- | --- | --- | --- |
| Phase/lane plan | pass | `docs/superpowers/plans/2026-06-10-sheet-ops-agent-execution-contract.md` defines Lane A execution, Lane B schema compatibility, and Lane C independent verification. | Work proceeded phase-by-phase with dependency-aware lanes. |
| Installed workbook E2E | pass | `TestInstallSkillBundledCLIRunIntentExecutesWorkbookEndToEnd` runs public `install-skill.sh`, executes installed `bin/sheet-ops-codex run-intent`, asserts `ok:true`, artifacts, verification pass, evidence paths, and output workbook row `LineItems!A3:C3`. | Installed CLI can execute a real workbook append with authoritative artifact evidence. |
| Result envelope | pass | `TestExecutedPublicEntryResultExposesAgentEnvelope`, `TestExecutedPublicEntryResultValidatesAgainstPublicSchema`, and `TestTerminalCompilerPublicEntryResultValidatesAgainstPublicSchema`. | Agents get stable success and recoverable compiler-stop envelopes. |
| Typed errors | pass with reserved backlog | `TestStateRootMismatchClassifiesAsConfigurationError`; `capabilities --json` marks `state_root_mismatch` as `emitted` and unproven runtime categories as `reserved`. | Emitted error taxonomy is evidence-backed; unproven categories are not overclaimed. |
| JSON Schemas and golden fixture | pass | `TestCLIJSONOutputsValidateAgainstPublishedSchemas`, `TestPublishedCLISchemasPinSchemaVersion`, `TestPublicEntryResultGoldenValidatesAgainstPublicSchema`, release schema compile/reference tests. | Public CLI JSON outputs are schema-valid and version-pinned. |
| Preview | pass | `TestPreviewRequestReportsImpactWithoutMutating`, `TestPreviewRequestIsExposedAsReadOnlyNonDryRunCommand`, installed CLI smoke, and `contracts/cli/preview_request_result.schema.json`. | Agents can inspect planned impact without workbook output or state mutation. |
| Docs alignment | pass | `TestCLIAgentContractDocsStayAligned` keeps public and bundled skill references aligned. | Installed docs and public docs teach the same safe contract. |
| Verifier separation and closure | pass | Separate verifier agents reviewed Phases 13-17 and final review; final verifier found documentation closure gaps, was closed after reporting, and Phase 12 was retrospectively verified before Phase 18. | Execution and verification roles are separated, and verifier sessions used in this thread were closed after use. |
| Full regression | pass | `go test ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts -count=1`; `go test ./... -count=1`; `git diff --check`. | No known repo regression from the agent execution contract work. |

## Verification Commands

- `go test ./cmd/sheet-ops-codex -run 'TestInstallSkillBundledCLIRunIntentExecutesWorkbookEndToEnd|TestPreviewRequest|TestCLIJSONOutputsValidateAgainstPublishedSchemas|TestPublishedCLISchemasPinSchemaVersion|Test.*PublicEntryResult.*Schema|TestTerminalCompilerPublicEntryResultValidatesAgainstPublicSchema' -count=1 -v`: pass.
- `go test ./cmd/sheet-ops-codex ./cmd/sheet-ops-agent ./internal/releasecontracts -count=1`: pass.
- `go test ./internal/releasecontracts -run 'TestReleaseSchemasCompile|TestReleaseSchemaReferencesAreLocallyResolvable|TestCLIAgentContractDocsStayAligned' -count=1 -v`: pass.
- `go test ./... -count=1`: pass.
- `go run ./cmd/sheet-ops-codex capabilities --json`: pass.
- `go run ./cmd/sheet-ops-codex schema command preview-request --json`: pass.
- `go run ./cmd/sheet-ops-codex preflight --json`: pass.
- `go run ./cmd/sheet-ops-codex --help`: pass.
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
- Runtime categories remain `reserved` until deterministic producers and tests
  exist.
- Installed workbook success is claimed only from the installed E2E test that
  checks JSON result fields, evidence paths, and workbook contents.
- No success claim is based on stdout alone when an authoritative artifact or
  schema-validated JSON result exists.

## Final Verifier Closure

The final verifier found no code-level contract regressions, but did find three
evidence/closure blockers:

- Final review checklist and artifact still described a pending state.
- Phase 12 and Phase 17 did not both record separate verifier outcomes in their
  phase notes.
- Subagent closure was not represented in the durable evidence matrix.

Actions before Phase 18 commit:

- Final review checklist was marked complete only after fresh verification and
  final verifier review.
- Phase 12 received a retrospective independent verifier pass confirming the
  deterministic installed E2E candidate and authoritative success fields.
- Phase 17 result note now records its independent verifier pass and the closed
  residual schema-sweep risk.
- The evidence matrix now includes verifier separation and closure. All
  verifier subagents used in this final review turn were closed after their
  reports were consumed.

## Remaining Backlog

- A deeper compiler-backed preview could be added later as a separate runtime
  planning track.
- A real dry-run remains deferred until runtime planning can prove non-mutation
  while simulating execution semantics.
- Runtime `validation_blocked`, `execution_failed`, and `verification_failed`
  should move from `reserved` to `emitted` only with deterministic producer
  tests.
- Schema generation can be reconsidered if hand-written output schemas become
  hard to maintain.
