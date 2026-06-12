# Sheet Ops

Scope: human-facing project intro. Derived view - for package authority see
`AGENTS.md`, for runtime contracts see `contracts/` and `runtime/`.

Sheet Ops is an experimental skill harness for LLM-assisted workbook work. It
assumes strong models can already reason about many Excel tasks; its job is to
make that reasoning cheaper, more repeatable, more inspectable, and safer to
hand off to deterministic execution.

It treats a natural-language workbook request as the start of an orchestration
process, not as the whole execution authority. The harness guides a model
through goal setting, request fact capture, capability selection, runtime
handoff, verification, and evidence review. The process separates the facts
needed for a workbook change, writes them into structured files, compiles them
into runtime operations, verifies the result, and keeps evidence and knowledge
that can be reused by later similar tasks.

The goal is not to replace model reasoning or claim that Sheet Ops can
understand every workbook by itself. The goal is to keep repeated workbook
patterns as local process files so a person or agent can inspect, test, reuse,
and compose them over time instead of spending model context on rediscovering
the same procedure.

Core positioning:

- LLM = planner/compiler
- Go runtime = executor/verifier
- JSON schema/contracts = execution authority
- artifacts/evidence = audit trail
- template research = advisory retrieval and planning hints
- atom/molecule/organism records = reusable decision scaffolding, not runtime
  authority
- structured request files preserve the resolved task facts
- artifacts, evidence, and knowledge keep the process inspectable and reusable

## What The Skill Harness Adds

Modern LLMs can already solve many workbook problems directly. Sheet Ops is
useful when the same kinds of workbook work repeat and the agent should not
spend fresh context rebuilding the process from scratch.

The harness gives the model reusable guidance for:

- deciding whether a request is ready for execution or needs more facts
- choosing supported workbook atoms instead of inventing ad hoc operations
- using template-class hints to narrow larger workbook requests into bounded
  atom sequences
- stopping at advisory evidence when the request needs unsupported behavior
- handing only schema-authorized work to deterministic runtime execution
- checking runtime verifier output before making user-facing success claims
- leaving request, plan, verification, evidence, and knowledge artifacts that
  make the next similar task cheaper to reason about

In short: Sheet Ops is not an Excel omniscience layer. It is a reusable
orchestration and verification harness for LLM workbook agents.

## Why not just Codex/Claude?

Modern LLMs can already help with workbook tasks. Sheet Ops starts from that
assumption.

This project focuses on the part that becomes useful after repetition:
preserving the process behind a workbook edit as local files that can be
reviewed, tested, adapted, and assembled into later work. A run should not only
produce an output workbook; it should also leave behind the request shape,
operation contract, verification result, evidence, and knowledge that make the
next similar task cheaper to reason about.

In this boundary model, the LLM only plans and compiles. Go runtime executes and verifies
workbook work; schema contracts decide what is valid, and evidence records what happened for review and repair.

## Example: Process Into Files

A user may ask:

> In `orders.xlsx`, create a revenue summary by region.

Sheet Ops does not treat that sentence as a single opaque execution command. It
separates the facts needed for the workbook work:

- workbook: `orders.xlsx`
- source sheet: the sheet to summarize
- grouping key: the column that represents region
- metric: the column that represents revenue
- output target: the summary sheet or output workbook to create
- verification rule: how to check summary values, row counts, and source
  preservation

Those facts are kept as process files instead of being left only in chat
context:

- structured request file: the normalized workbook task facts
- `TaskSpec`: the pre-execution task specification validated against workbook
  facts and policy
- `OperationIR`: the deterministic workbook operations the runtime can execute
- verification result: the runtime's judgment about whether the output matches
  the contract
- evidence: a scenario-level record of what input produced what output
- knowledge: reusable patterns that can inform later similar work

Canonical execution shape:

`TaskSpec -> OperationIR -> Execute -> Verify -> Evidence`

## Current Public Path

The current public path supports a broader set of workbook composition atoms.
These are the deterministic actions the skill harness may select after request
facts are captured and schema validation allows execution.

- `append_structured_rows`
- `extend_table_formulas`
- `copy_period_sheet`
- `add_data_validation`
- `protect_formula_cells`
- `normalize_headers`
- `roll_forward_period`
- `reconcile_tables`
- `generate_printable_form`
- `group_summarize`
- `highlight_threshold`
- `join_lookup`

`write_values` remains a runtime primitive, not a public agent capability.

The public entry is the `sheet-ops` skill for local Codex/Claude workbook
requests. This repository is currently a local source-install harness; the
installed runtime is materialized under `.codex/skills/sheet-ops/agent-system`,
and local runtime state defaults to `.sheet-ops-state/`.

## Template Research Guidance Layer

The front end is no longer only a thin natural-language-to-operation path.
Sheet Ops now carries an advisory template research layer that helps classify
spreadsheet work before it becomes runtime execution. This layer exists to guide
the LLM toward known workbook patterns and away from unsupported claims:

- curated template patterns and model-generated hypotheses
- pattern classification and capability gap maps
- atom, molecule, and organism composition catalogs
- atom-builder audit records for code-generation assistance
- draft planner coverage for 21 roadmap organisms
- standalone final-workbook semantic verifier slices for those 21 organisms

This layer is advisory. It helps shape request compilation, capability choice,
runtime coverage, and verification focus, but it does not make organism IDs
public capabilities and it does not claim full template generation.
For template-like requests, compiler artifacts retain `template_class_plan`
evidence before the request is narrowed to supported atom execution.

## Human-Facing Skill Entry

The `sheet-ops` skill is the single human-facing public entry for local
Codex/Claude workbook requests.

## CLI Agent Contract

The `sheet-ops` skill is the single human-facing workbook request entry.
`sheet-ops-codex` is an install, diagnostic, and agent-contract CLI surface,
not a second human-facing workbook entry.

Agent and CI discovery should use:

- `sheet-ops-codex agent-guide --json`
- `sheet-ops-codex capabilities --json`
- `sheet-ops-codex operation list --json`
- `sheet-ops-codex operation schema <operation> --json`
- `sheet-ops-codex operation example <operation> --json`
- `sheet-ops-codex schema command preflight --json`
- `sheet-ops-codex preflight --json`
- `sheet-ops-codex preview-request --json --intent-file <path-or-> --input-file <path> --output-file <path>`
- `sheet-ops-codex evidence-summary --json --evidence-dir <path>`

The CLI contract is machine-readable and intended for agent orchestration:
agent-guide reports the recommended phase order, capabilities and schema
commands expose stable command metadata, operation commands expose public
workbook operation contracts and normalized-intent examples, preflight reports
read-only readiness checks, preview-request reports planned impact without
creating workbooks or state artifacts, and evidence-summary checks runtime
verification artifacts before an agent claims success. Emitted JSON failure
envelopes keep stderr quiet so agents can treat stdout as the authoritative
contract channel. Large or generated normalized intents may be passed through
stdin with `--intent-file -` for `preview-request` and `run-intent`.

`sheet-ops-codex run-validated` is an internal handoff surface. It is owned by
the installed skill handoff and must not be presented as the normal public
request route. Mutating workbook commands currently report
`dry_run_capable: false`; `preview-request` is read-only impact inspection with
`planner:"requestcompiler_validate_intent"` and
`plan_confidence:"compiler_validated_boundary"`, not dry-run evidence. Do not
claim dry-run behavior until a truthful runtime planning mode exists.

After execution, do not report workbook success from stdout alone. Use
`sheet-ops-codex evidence-summary --json --evidence-dir <dir>` and require a
passing `verification.json` plus `output_workbook_sha256`; failed or incomplete
evidence must route to repair advice or human review.

The installed CLI has been exercised from a throwaway project-local
`.codex/skills/sheet-ops` install with red-input coverage for discovery,
schema lookup, preflight failure states, preview-request JSON failures,
state-root mismatch, prepare-use validation, missing internal handoff requests,
and a successful run-intent append path.

## Internal Execution Boundary

The bundled launcher behind this handoff is part of the skill-owned internal proof-gated boundary behind the public skill entry. It exists so the public entry and the runtime can meet at a typed contract rather than an unstructured conversation.

## Codex Skill Install

Install into a target workspace with the repository-local installer:

```bash
/path/to/sheet-ops/install-skill.sh --project /path/to/target/workspace
```

This is a project-local install. Go 1.25 or newer is required; check it with
`command -v go && go version`. If Go is installed but not visible in the current
shell, pass it explicitly with `--go-bin /absolute/path/to/go`. Do not guess the Go path.

If Go 1.25 or newer is already available, `install-skill` does not prompt. If
Go is missing or outdated, it asks before continuing. The install flow does not
use a zip file, tarball, or prebuilt platform binary.

The installed surface lives under `./.codex/skills/sheet-ops`. Local runtime
state defaults to `.sheet-ops-state/` in local testing mode. This is not deployment-safe yet for hosted multi-tenant execution.

See [Install](docs/public/install.md) for the full install contract, PATH
guidance, and local state behavior.

## Status

Supported today:

- experimental local skill harness for Codex/Claude workbook agents
- single human-facing `sheet-ops` skill entry
- currently implemented public path compositions:
  `append_structured_rows`, `extend_table_formulas`, `copy_period_sheet`,
  `add_data_validation`, `protect_formula_cells`, `normalize_headers`,
  `roll_forward_period`, `reconcile_tables`, `generate_printable_form`,
  `group_summarize`, `highlight_threshold`, and `join_lookup`
- schema-authorized `TaskSpec -> OperationIR -> Execute -> Verify -> Evidence`
  runtime path
- advisory template research corpus with atom/molecule/organism composition
  records and 21 roadmap organism coverage artifacts
- template-class planning evidence that helps the model choose bounded
  supported atom sequences before runtime handoff
- explicit organism execution request path with draft planner coverage and
  final-workbook semantic verifier slices for 21 roadmap organisms
- deterministic fixture-backed harness smoke with checked-in frozen demo
  evidence
- installed CLI agent-contract smoke covering 22 discovery, preview, preflight,
  handoff, and red-input cases from a project-local `.codex/skills/sheet-ops`
  install
- render artifact emission when renderer tools exist, or explicit unavailable
  evidence when they do not

Preview limitations:

- live Codex delegation smoke is a non-blocking opt-in diagnostic, not a
  required public harness gate
- render artifact emission records file-open and preview artifact properties;
  it is not visual quality verification
- template research artifacts are advisory; they are not execution authority
  and do not imply full template generation
- the harness improves repeated agent workflow quality, but it is not a claim
  that arbitrary Excel templates can be fully inferred or generated
- organism-level verifier slices prove bounded workbook evidence, not
  domain-specific correctness, financial advice, safety certification, or rich
  visual/print QA
- Sheet Ops is not hosted or multi-tenant ready
- `write_values` is not a public agent capability; it remains an internal
  runtime primitive until public orchestration and e2e tests exist

Follow-up:

- harden the optional live smoke with stronger timeout/process supervision
- strengthen skill-level guidance that forces model decisions to cite request
  facts, selected atoms, verifier focus, and stop conditions
- add cross-platform visual quality assertions only after renderer support is
  proven
- add richer template generation, native pivot/matrix expansion, and
  domain-specific correctness layers after the current advisory/runtime bridge
- harden state/privacy boundaries before any hosted claim
- promote `write_values` only after public orchestration and e2e coverage exist

## Documentation

- [Architecture](docs/public/architecture.md): boundary model, execution shape,
  repository model, and package ownership
- [Install](docs/public/install.md): Go requirements, installer usage, and local
  state
- [Preview Status](docs/public/preview.md): supported capabilities, limitations,
  and follow-up work
- [Assist Orchestration Backlog](docs/public/orchestration-backlog.md): known
  gaps in live skill guidance, evidence feedback, and model-orchestration
  compliance

## Canonical Package Contract

This repository root is the canonical public package and install source.

The installed workspace tree is a materialized copy of that package under `.codex/skills/sheet-ops`.

`agents/` and `skills/` are required package/runtime surfaces, not optional docs.

## Boundary Notes

The old skill-centric mental model is deprecated; public skill names are entry surfaces into the harness, not execution authority.

Boundary invariants:

- hard restrictions live on consumer agents/sessions, not shared skills
- shared skills are reusable tool surfaces, not a security boundary
- single-consumer procedures stay under their owning agent subtree
- the public entry captures the workbook request
- private capability ownership belongs under `agents/`
