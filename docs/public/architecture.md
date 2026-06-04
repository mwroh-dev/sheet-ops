# Sheet Ops Architecture

Sheet Ops keeps natural-language interpretation in the open layer and workbook
mutation in the closed layer.

- open layer: capture the request, separate the task facts needed to form a
  typed request boundary, and stop when the request is not ready to cross that
  boundary safely
- closed layer: inspect workbook facts, build `TaskSpec`, compile
  `OperationIR`, execute deterministically, verify the outcome, and emit
  evidence

The boundary exists so natural-language context is turned into files and
contracts before workbook mutation happens.

Canonical flow:

- open: `request -> request reference -> UseEnvelopeV2 -> validated execution boundary`
- closed: `TaskSpec -> OperationIR -> Execute -> Verify -> Evidence`

## How The Harness Works

Humans invoke Sheet Ops through the `sheet-ops` Codex skill using natural
language. That public entry captures the workbook request, chooses or writes a
request reference, creates a typed `UseEnvelopeV2`, and hands execution to the
internal proof-gated runtime boundary.

When the request already contains the task facts needed for execution, the
public entry can write a structured request file and skip open-layer
compilation. When more workbook facts are needed, the request stays on the
prompt-text path so the request-compiler lane can resolve only the facts needed
to produce a typed validated boundary. If the request is unsupported or not
ready for safe execution, the run stops before mutation and returns a blocked
outcome with repair guidance.

Once the typed boundary exists, the closed runtime owns workbook inspection,
policy checks, `TaskSpec`, `OperationIR`, deterministic execution,
verification, and evidence emission. Result interpretation can help explain
that output, but runtime validation remains the authority on pass/fail.

## Human-Facing Skill Entry

The `sheet-ops` skill is the single human-facing public entry for local
Codex/Claude workbook requests. The public entry captures the workbook request,
resolves only the facts needed to form a typed boundary, and remains
responsible for the user-facing answer.

The bundled launcher behind this handoff is part of the skill-owned internal
proof-gated boundary behind the public skill entry. It exists so the public
entry and the runtime can meet at a typed contract rather than an unstructured
conversation.

## Current Implementation Shape

Sheet Ops started as an Excel command execution skill and evolved into a
workbook agent harness. The current public docs keep the operational boundary
model here; deeper design-history notes live in the source repository's
architecture docs.

The repository currently exposes a local source-install public path for
Codex/Claude workbook agents. In that path:

- the single human-facing public entry is the `sheet-ops` skill
- the currently implemented public path compositions are `group_summarize`,
  `highlight_threshold`, and `join_lookup`
- the runtime path is schema-authorized and evidence-producing
- deterministic fixture-backed smoke is the release gate for the harness shape

Some live subagent-gated and bootstrap/compatibility paths coexist today
because the harness is still being extended. That coexistence is a transitional
implementation state, not a change in direction. The direction of travel
remains:

`proof-gated open layer -> typed execution boundary -> deterministic runtime -> verification/evidence`

## Repository Model

- `contracts/` define execution authority for the pipeline
- `runtime/` performs deterministic workbook inspection, policy, execution,
  verification, and evidence emission
- `agents/` resolve open-layer ambiguity only
- `skills/` are shared reusable and discoverable tool surfaces only
- `artifacts/` store versioned episodes, telemetry, evidence, and reports

The old skill-centric mental model is deprecated; public skill names are entry
surfaces into the harness, not execution authority.

Boundary invariants:

- hard restrictions live on consumer agents/sessions, not shared skills
- shared skills are reusable tool surfaces, not a security boundary
- single-consumer procedures stay under their owning agent subtree
- the public entry captures the workbook request
- private capability ownership belongs under `agents/`

## Canonical Package Contract

The source repository root is the canonical package.

The release repository root is a sanitized mirror of the canonical package.

The installed workspace tree is a materialized copy of that package under
`.codex/skills/sheet-ops`.

`agents/` and `skills/` are required package/runtime surfaces, not optional
docs.
