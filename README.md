# Sheet Ops

Scope: human-facing project intro. Derived view - for package authority see
`AGENTS.md`, for runtime contracts see `contracts/` and `runtime/`.

Sheet Ops is an experimental local workflow surface for repeated workbook work.
It is still a safety belt + dashboard around deterministic workbook execution,
but the front door is process reuse rather than a claim that models cannot do
the work.

It treats a natural-language workbook request as the start of a process, not as
the whole execution authority. The process separates the facts needed for a
workbook change, writes them into structured files, compiles them into runtime
operations, verifies the result, and keeps evidence and knowledge that can be
reused by later similar tasks.

The goal is not to replace model reasoning. The goal is to keep repeated
workbook patterns as local process files so a person or agent can inspect,
test, reuse, and compose them over time instead of spending model context on
rediscovering the same procedure.

Core positioning:

- LLM = planner/compiler
- Go runtime = executor/verifier
- JSON schema/contracts = execution authority
- artifacts/evidence = audit trail
- structured request files preserve the resolved task facts
- artifacts, evidence, and knowledge keep the process inspectable and reusable

## Why not just Codex/Claude?

Modern LLMs can already help with workbook tasks. Sheet Ops starts from that
assumption.

This project focuses on the part that becomes useful after repetition: keeping
the process behind a workbook edit as local files that can be reviewed, tested,
adapted, and assembled into later work. A run should not only produce an output
workbook; it should also leave behind the request shape, operation contract,
verification result, evidence, and knowledge that make the next similar task
cheaper to reason about.

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

The current public path supports a broader set of workbook composition
families, while the template research layer supplies advisory patterns for
larger workbook workflows.

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

## Template Research Front Door

The front end is no longer only a thin natural-language-to-operation path.
Sheet Ops now carries an advisory template research layer that helps classify
spreadsheet work before it becomes runtime execution:

- raw template observations and model-generated hypotheses
- pattern classification and capability gap maps
- atom, molecule, and organism composition catalogs
- atom-builder audit records for code-generation assistance
- draft planner coverage for 21 roadmap organisms
- standalone final-workbook semantic verifier slices for those 21 organisms

This layer is advisory. It helps shape request compilation and runtime
coverage, but it does not make organism IDs public capabilities and it does not
claim full template generation.

## Human-Facing Skill Entry

The `sheet-ops` skill is the single human-facing public entry for local
Codex/Claude workbook requests.

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

- experimental local workflow surface for Codex/Claude workbook agents
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
- explicit organism execution request path with draft planner coverage and
  final-workbook semantic verifier slices for 21 roadmap organisms
- deterministic fixture-backed harness smoke with checked-in frozen demo
  evidence
- render artifact emission when renderer tools exist, or explicit unavailable
  evidence when they do not

Preview limitations:

- live Codex delegation smoke is a non-blocking opt-in diagnostic, not a
  required release gate
- render artifact emission records file-open and preview artifact properties;
  it is not visual quality verification
- template research artifacts are advisory; they are not execution authority
  and do not imply full template generation
- organism-level verifier slices prove bounded workbook evidence, not
  domain-specific correctness, financial advice, safety certification, or rich
  visual/print QA
- Sheet Ops is not hosted or multi-tenant ready
- `write_values` is not a public agent capability; it remains an internal
  runtime primitive until public orchestration and e2e tests exist

Follow-up:

- harden the optional live smoke with stronger timeout/process supervision
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
- [Release](docs/public/release.md): release bundle generation, mirror rules,
  privacy scan, and release verification
- [Preview Status](docs/public/preview.md): supported capabilities, limitations,
  and follow-up work

## Canonical Package Contract

The source repository root is the canonical package.

The release repository root is a sanitized mirror of the canonical package.

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

## Latest Release

<!-- sheet-ops-latest:start -->
- Source branch: `main`
- Source SHA: `f98d0b562f8c`
- Updated: 2026-06-04T18:14:43.089Z
- Bundle payload file count: 227
- Work item: none
<!-- sheet-ops-latest:end -->
