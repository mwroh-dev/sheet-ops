---
name: sheet-ops
description: Single human-facing public entry for Sheet Ops workbook requests.
---

# Sheet Ops

This skill is the public entry for repeated workbook work that should become a
local process artifact, not only a one-off chat answer. Accept the user's
natural-language workbook request, separate the task facts needed for execution,
write or select a structured request reference, and hand a typed `UseEnvelopeV2`
to the internal proof-gated execution boundary:

- the internal boundary is skill-owned runtime handoff, not a runner-pane command
- `use` is compatibility routing only; runtime subroutes such as `use-open`,
  `use-structured`, and `run-validated` stay behind the skill-owned handoff
- runner-pane bypass guard: Do not type, paste, or run the skill-local launcher path directly.
  In a loop-station or consumer runner, the user-facing entry is this
  `sheet-ops` skill; the launcher remains internal to skill execution.
- Do not search for, inspect, or construct internal command lines in the runner
  pane. Treat this skill document as the public entry instructions.

## Why not just Codex/Claude?

Modern LLMs can already help with workbook tasks. Sheet Ops starts from that
assumption and focuses on what becomes useful after repetition: keeping the
process behind a workbook edit as local files that can be reviewed, tested,
adapted, and reused later.

In this skill, the LLM only plans and compiles. Go runtime executes and verifies
workbook work; schema contracts decide what is valid, and evidence records what happened for review and repair.

Use this skill when Codex needs to coordinate a workbook request by:

- interpreting the request
- spawn built-in subagents for focused open-layer work
- writing or selecting a harness-safe request reference
- building a typed `UseEnvelopeV2` for the internal execution boundary
- requesting runtime handoff when execution is justified
- reviewing the result before returning a final answer

This skill is a guideline and orchestration harness for a strong model, not a
claim that the model lacks spreadsheet knowledge. Its value is to reduce
repeated reasoning cost, keep decisions tied to local contracts, and leave
evidence that can improve later similar workbook work.

## Tool Authority

Consumer restrictions live in `agents/profiles/*.toml` and are rendered into
consumer-specific surfaces. The shared skill is guidance only; the consumer
profiles are the code-reviewed source of truth for allowed/disallowed tools,
commands, carriers, and writable scope.

Do not treat `request-packer`, `request-mode-judge`, `request-open-compiler`,
`use-structured-runner`, or `result-verifier` as shared top-level skills. They
are private consumer roles owned by the public entry boundary.

## Derived file placement

Derived request, envelope, output, report, and evidence files must be written outside immutable case input folders. In loop-station or any harness that provides an attempt output directory, write the request reference and typed `UseEnvelopeV2` under that attempt output directory. Do not create `use-envelope.json` inside the case input folder.

## Request capture rule

Do not inspect workbook internals from ad hoc Python, `openpyxl`, ZIP/XML
scraping, or manual sheet reads when the user already supplied the source sheet,
output boundary, and operation shape clearly enough to form a structured
`UseRequest`.

Prefer these request-reference forms:

- structured `UseRequest` JSON when sheet name, output file, and operation
  details are explicit enough to compile directly
- plain text or markdown request only when additional workbook facts are needed
  and the internal request-compiler must inspect those facts in Go

If you can write a structured `UseRequest`, do that and let the skill package it
as a structured request before the internal dispatcher runs. Use a plain text or
markdown request only when text must enter the open request-compiler lane. Do
not pre-read the workbook just to rediscover headers the runtime can validate
for itself.

## Planning guidance rule

Before runtime handoff, make the model decision explicit enough to audit:

- user goal and workbook boundary
- known request facts versus missing facts
- selected public atom capability or template-class hint
- why the selected path is supported, advisory-only, or blocked
- verifier focus that would prove success
- stop condition if facts, policy, capability support, or verifier coverage are
  insufficient

For template-like requests, use the advisory atom/molecule/organism ecosystem as
decision scaffolding. Do not present the ecosystem as runtime authority. Narrow
to supported atom execution only when the request facts and schema contracts are
sufficient.

## Routing guardrail

Do not route a live Sheet Ops request through consumer-local demo or
compatibility surfaces such as `.codex/skills/example-*` or
`.codex/skills/shared-workbook-case-executor`. Those are harness fixtures, not
the single human-facing public entry, and they may exercise legacy adapters or
non-authoritative execution paths.

## Public-Safe Evidence

User-facing skill output should use curated scenario evidence, not raw telemetry,
local session state, or machine paths. Entry docs point to contracts and runtime
boundaries rather than duplicating authority.

## Runtime rule

Do not expose the internal dispatcher as a second human-facing public entry. It
is the internal execution boundary used after the user's request has been
captured into an envelope, and it must not be typed, pasted, or reconstructed
as a command line in the runner pane.

Do not treat custom named agents or project-local `.toml` agent definitions as
the runtime core. The runtime core is:

- `spawn_agent`
- `wait_agent`
- `close_agent`
- built-in carrier agent types: `default`, `worker`, `explorer`
- role-specific prompt templates from the repository

## Public preview claims

Supported today:

- use this skill as the single human-facing entry for local workbook requests
- select public agent capabilities from `group_summarize`,
  `highlight_threshold`, `join_lookup`, `append_structured_rows`,
  `extend_table_formulas`, `copy_period_sheet`, `add_data_validation`,
  `protect_formula_cells`, `normalize_headers`, `roll_forward_period`,
  `reconcile_tables`, and `generate_printable_form`
- use runtime handoff for schema-authorized deterministic execution,
  verification, and evidence
- for template-like workbook requests, preserve the advisory
  `template_class_plan` evidence in compiler artifacts before narrowing to
  supported atom execution
- trust the deterministic fixture-backed harness smoke and public preview
  claim contract as the public harness gate
- treat render artifact emission as evidence about preview artifact creation

Preview limitations:

- live Codex delegation smoke is a non-blocking opt-in diagnostic, not a
  required public harness gate
- render artifact emission is not visual quality verification
- Sheet Ops is not hosted or multi-tenant ready
- `write_values` is not a public agent capability

## CLI agent contract

The `sheet-ops` skill is the single human-facing workbook request entry.
`sheet-ops-codex` is an install, diagnostic, and agent-contract CLI surface,
not a second human-facing workbook entry.

Use these read-only contract surfaces for agent/CI discovery:

- `sheet-ops-codex capabilities --json`
- `sheet-ops-codex schema command preflight --json`
- `sheet-ops-codex preflight --json`
- `sheet-ops-codex preview-request --json --intent-file <path> --input-file <path> --output-file <path>`

`sheet-ops-codex run-validated` is an internal handoff surface. It is owned by
the installed skill handoff and must not be presented as the normal public
request route. Mutating workbook commands currently report
`dry_run_capable: false`; `preview-request` is impact inspection only, not
runtime dry-run evidence. Do not claim dry-run behavior until a truthful runtime
planning mode exists.

`preview-request` also reports `planner:"requestcompiler_validate_intent"`,
`plan_confidence:"compiler_validated_boundary"`, `operation`, `would_mutate`,
`mutation_summary`, and input `fingerprints`. Treat these as trust metadata for
the inspected inputs, not proof of runtime execution success. After
`run-intent`, compare the execution result `fingerprints` to the preview
fingerprints, and use `output_workbook_sha256` to bind the produced workbook
bytes before trusting runtime artifacts.

Follow-up:

- live delegation smoke hardening
- cross-platform visual quality checks
- hosted state/privacy hardening
- additional public orchestration and e2e tests for future atom promotions

## Built-in carrier mapping

- `explorer`: read-heavy exploration, request interpretation, repository
  inspection
- `worker`: implementation or deterministic change execution
- `default`: general review, synthesis, or fallback judgment

## Required orchestration pattern

1. Start as the single parent orchestrator.
2. Spawn built-in subagents explicitly for specialist roles.
3. Wait for each result.
4. Evaluate the result in the parent.
5. Record the goal, known facts, selected path, verifier focus, and stop
   condition before execution.
6. If the request is explicit enough for a structured `UseRequest`, write that
   request file and skip manual workbook inspection in the parent.
7. Create a typed `UseEnvelopeV2` only when the request is sufficiently
   resolved; use the skill-owned deterministic packing step without probing for
   internal commands from the runner pane.
8. Use the skill-owned runtime handoff as the compatibility boundary without
   typing or reconstructing an internal command. It routes structured requests
   to `use-structured` and open requests to `use-open`; closed validated
   requests belong to `run-validated`.
9. Re-dispatch or stop based on explicit state.
10. The parent owns the final answer and completion decision.

## Role templates

Use these role templates when spawning built-in subagents:

- `agents/orchestrator/use-orchestrator/prompt-templates/request-compiler.md`
- `agents/orchestrator/use-orchestrator/prompt-templates/result-verifier.md`
- `agents/orchestrator/use-orchestrator/prompt-templates/repair-advisor.md`

## Stop rules

- If the request remains ambiguous, stop and ask the user.
- If deterministic validation blocks execution, stop and report the blocking
  reason.
- If runtime execution succeeds, review the result before answering.
- If a specialist result is incomplete, spawn a follow-up built-in subagent or
  stop explicitly.

## Output states

Use explicit internal outcome states such as:

- `DONE`
- `NEEDS_REVIEW`
- `BLOCKED`
- `AMBIGUOUS`

Do not rely on free-form "looks finished" reasoning.
