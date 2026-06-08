# Template Research Corpus

This corpus reverse-engineers spreadsheet template intent. It does not publish,
bundle, or reuse third-party template files.

## Layers

- `raw/shards/*.jsonl`: Agent A evidence shards from source-specific or
  category-specific collection rounds.
- `extracted/templates.jsonl`: source observations with URL, category, terms
  note, observed why, observed how, and input/output shape.
- `hypotheses/`: Agent B trigger catalogs, pattern hypotheses, and level
  requirements generated from model knowledge rather than source evidence.
- `judgments/`: Agent C synthesis outputs that compare evidence against
  hypotheses and assign promotion status.
- `classification/`: post-research grouping of roadmap candidates by Sheet Ops
  implementation family, fixture priority, capability gap, and verifier
  strategy.
- `composition/`: advisory atom, molecule, and organism composition records.
  These include molecule criteria, organism combinations, ecosystem principles,
  atom-builder audit notes, runtime bridge rules for code-generation
  assistance, non-runtime organism preview harness notes, and the executable
  organism coverage ladder that separates fixture-backed previews from
  supported-atom plans and planned-atom blockers.
- `patterns/*.json`: normalized template patterns grouped by user workflow and
  spreadsheet mechanics.
- `opportunities/*.json`: capability backlog candidates that are intentionally
  separate from the supported capability registry.
- `taxonomy.md`: shared labels for domains, visual patterns, data patterns,
  validation patterns, and workflow patterns.
- `../atom-builders/`: runtime-adjacent advisory records for workbook app
  primitives, atom-builder mirrors, and plan shapes. These records intentionally
  mirror existing supported runtime paths and planned opportunity atoms without
  claiming new support.

## Collection Rules

- Treat free template sites as observation sources only.
- Do not copy third-party workbook files, screenshots, layouts, formulas, or
  prose into release artifacts.
- Prefer source summaries written in original language over quoted text.
- Record a terms note for every observation.
- Use low-cost parallel workers for raw extraction and first-pass summaries.
- Use reviewer passes for taxonomy merge, duplicate detection, and promotion
  decisions.
- Do not add unimplemented opportunities to
  `contracts/capabilities/capability.schema.json`.
- Do not treat atom, molecule, or organism composition records as runtime
  support claims.

## Current Seed Scope

The checked-in records are an expanded seed corpus, not an exhaustive crawl.
They cover the major catalog categories visible in initial research so the
schema, taxonomy, and backlog workflow can be reviewed before deeper parallel
collection adds more shards.

## Triple-Agent Research Loop

Each research round separates source evidence, model-generated hypotheses, and
promotion judgment.

- Agent A, Evidence Researcher: collects real source observations into
  `raw/shards/`. It records provenance and terms notes, but does not promote
  patterns.
- Agent B, Hypothesis Modeler: creates trigger catalogs, pattern hypotheses, and
  L1-L5 requirements in `hypotheses/`. It does not claim evidence.
- Agent C, Judge Synthesizer: compares Agent A evidence and Agent B hypotheses
  in `judgments/`, assigns status, and identifies the next collection targets.

Promotion status is owned by Agent C:

- `hypothesis_only`: plausible from model knowledge, no matching source
  evidence yet.
- `observed`: at least one source observation supports the pattern.
- `recurring`: independent observations support the pattern across sources or
  categories.
- `roadmap_candidate`: recurring pattern plus concrete Sheet Ops capability and
  verification strategy.
- `rejected`: weak, duplicated, unsafe, or unsupported by evidence.

## Promotion Rules

- A pattern is `single_source_seed` until it has evidence from at least two
  independent sources.
- A capability opportunity needs a concrete workflow and verification strategy.
- A supported Sheet Ops capability requires separate implementation,
  runtime/schema authority, fixture coverage, and verifier coverage.
- Supported capability status is decided only by the capability registry,
  runtime contracts, deterministic execution, and operation-specific verifier
  coverage. Template research artifacts are advisory until promoted through
  those release surfaces.
- The executable organism coverage ladder records claim strength for roadmap
  organisms. It does not make organism IDs public capabilities and does not
  replace operation-level verifier coverage.
- Atom-builder records are an intermediate planning layer. A supported mirror
  must reference an existing capability record and runtime paths; a planned
  builder must reference an opportunity record and no runtime implementation.
